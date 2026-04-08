package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"ros-address-list-tool/internal/app"
)

func main() {
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	outputPath := flag.String("output", "", "临时覆盖输出文件路径")
	mode := flag.String("mode", "", "临时覆盖渲染模式：replace_all 或 diff")
	serve := flag.Bool("serve", false, "是否进入服务模式")
	listen := flag.String("listen", "", "临时覆盖服务监听地址")
	logFile := flag.String("log-file", "", "临时覆盖日志文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - HTTP API 服务阶段")
	fmt.Println("================================================")

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前工作目录失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("当前工作目录: %s\n", wd)

	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析配置文件绝对路径失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("准备读取配置文件: %s\n", absConfigPath)

	// 初始化配置存储层。
	store, err := app.NewConfigStore(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化配置存储失败: %v\n", err)
		os.Exit(1)
	}

	// 从 store 中取一份当前配置副本，用于本次运行。
	cfg := store.GetConfig()

	// 应用本次运行的临时命令行覆盖。
	if *outputPath != "" {
		cfg.Output.Path = *outputPath
	}
	if *mode != "" {
		cfg.Output.Mode = app.RenderMode(*mode)
	}
	if *listen != "" {
		cfg.Server.Listen = *listen
	}
	if *logFile != "" {
		cfg.LogFile = *logFile
	}

	// 覆盖完成后，必须重新补默认值并重新校验。
	cfg.ApplyDefaults()
	if err := app.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "命令行覆盖后的配置无效: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志器。
	logger, closeFn, err := app.NewLogger(cfg.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closeFn != nil {
			_ = closeFn()
		}
	}()

	logger.Printf("程序启动")
	logger.Printf("当前工作目录：%s", wd)
	logger.Printf("配置文件路径：%s", absConfigPath)
	logger.Printf("配置存储层初始化成功")
	logger.Printf("命令行参数覆盖结果：mode=%s output=%s log_file=%s serve=%v listen=%s",
		cfg.Output.Mode, cfg.Output.Path, cfg.LogFile, *serve, cfg.Server.Listen)

	// serve 模式：真正启动 HTTP API。
	if *serve {
		handler := app.NewHTTPHandler(store, logger)

		addr := cfg.Server.Listen
		if addr == "" {
			addr = ":8090"
		}

		logger.Printf("HTTP API 服务启动，监听地址：%s", addr)

		fmt.Println()
		fmt.Println("HTTP API 服务已启动。")
		fmt.Printf("监听地址: %s\n", addr)
		fmt.Println("可用接口：")
		fmt.Println("  GET  /healthz")
		fmt.Println("  POST /api/v1/auth/login")
		fmt.Println("  POST /api/v1/auth/logout")
		fmt.Println("  GET  /api/v1/auth/me")
		fmt.Println("  POST /api/v1/auth/change-password")
		fmt.Println("  GET  /api/v1/config")
		fmt.Println("  POST /api/v1/render")

		server := &http.Server{
			Addr:    addr,
			Handler: handler,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Printf("HTTP 服务异常退出：%v", err)
			fmt.Fprintf(os.Stderr, "HTTP 服务异常退出：%v\n", err)
			os.Exit(1)
		}
		return
	}

	// 非 serve 模式下，继续执行单次渲染流程。
	result, err := app.Execute(cfg, logger)
	if err != nil {
		logger.Printf("执行失败：%v", err)
		fmt.Fprintf(os.Stderr, "执行失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("执行结果摘要：")
	fmt.Printf("- 渲染模式: %s\n", result.Mode)
	fmt.Printf("- 目标 list 数量: %d\n", result.ListCount)
	fmt.Printf("- 目标条目总数: %d\n", result.EntryCount)
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)

	fmt.Println()
	fmt.Println("已生成的 RouterOS 脚本内容：")
	fmt.Println("------------------------------------------------")
	fmt.Println(result.Script)

	fmt.Println("第 10 步完成：HTTP API 服务已落地。")
	fmt.Println("下一步我们将实现 address-list 管理接口（新增、修改、删除、description 管理）。")
}
