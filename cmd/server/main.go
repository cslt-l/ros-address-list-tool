package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ros-address-list-tool/internal/app"
)

func main() {
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	outputPath := flag.String("output", "", "临时覆盖输出文件路径")
	mode := flag.String("mode", "", "临时覆盖渲染模式：replace_all 或 diff")
	serve := flag.Bool("serve", false, "是否进入服务模式（当前阶段仅预留入口）")
	listen := flag.String("listen", "", "临时覆盖服务监听地址")
	logFile := flag.String("log-file", "", "临时覆盖日志文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 配置存储层阶段")
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

	// 先初始化配置存储层。
	store, err := app.NewConfigStore(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化配置存储失败: %v\n", err)
		os.Exit(1)
	}

	// 从 store 中取出当前配置副本。
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

	if *serve {
		logger.Printf("当前已进入 serve 模式入口，但 HTTP 服务尚未在本步骤实现。listen=%s", cfg.Server.Listen)

		fmt.Println()
		fmt.Println("serve 模式入口已识别。")
		fmt.Printf("当前监听地址配置为: %s\n", cfg.Server.Listen)
		fmt.Println("HTTP API 服务将在后续步骤正式接入。")

		fmt.Println()
		fmt.Println("第 9 步完成：配置存储层已落地。")
		fmt.Println("下一步我们将实现 HTTP API 服务，并把它接到 serve 模式中。")
		return
	}

	// 正常执行单次渲染流程。
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

	fmt.Println("第 9 步完成：配置存储层已落地。")
	fmt.Println("下一步我们将实现 HTTP API 服务，并把它接到 serve 模式中。")
}
