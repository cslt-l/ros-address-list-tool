package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ros-address-list-tool/internal/app"
)

func main() {
	// 这一阶段开始，我们把命令行入口补完整。
	//
	// 这些参数分成两类：
	// 1. 真正已经生效的参数：
	//    -config
	//    -output
	//    -mode
	//    -log-file
	//
	// 2. 先预留结构、下一步接入功能的参数：
	//    -serve
	//    -listen
	//
	// 这样做的目的，是先把 CLI 层定型，
	// 后续做 HTTP API 时就不需要再反复改 main.go 的参数结构。
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	outputPath := flag.String("output", "", "临时覆盖输出文件路径")
	mode := flag.String("mode", "", "临时覆盖渲染模式：replace_all 或 diff")
	serve := flag.Bool("serve", false, "是否进入服务模式（当前阶段仅预留入口）")
	listen := flag.String("listen", "", "临时覆盖服务监听地址")
	logFile := flag.String("log-file", "", "临时覆盖日志文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 完整 CLI 入口阶段")
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

	// 第一步：先读取配置文件。
	// 之所以先读配置，而不是直接初始化日志，
	// 是因为日志文件路径通常来自配置。
	cfg, err := app.LoadConfigFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 第二步：用命令行参数临时覆盖配置。
	// 这里一定要清楚：
	// 这些覆盖只影响“本次运行”，不会回写到 config.json。
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

	// 第三步：初始化日志器。
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
	logger.Printf("命令行参数覆盖结果：mode=%s output=%s log_file=%s serve=%v listen=%s",
		cfg.Output.Mode, cfg.Output.Path, cfg.LogFile, *serve, cfg.Server.Listen)

	// 这一阶段先把 serve 模式入口立住，
	// 但不真正启动 HTTP 服务。
	// 下一阶段做完 store 和 server 后，
	// 这里会直接接入 HTTP API 监听。
	if *serve {
		logger.Printf("当前已进入 serve 模式入口，但 HTTP 服务尚未在本步骤实现。listen=%s", cfg.Server.Listen)

		fmt.Println()
		fmt.Println("serve 模式入口已识别。")
		fmt.Printf("当前监听地址配置为: %s\n", cfg.Server.Listen)
		fmt.Println("HTTP API 服务将在后续步骤正式接入。")

		fmt.Println()
		fmt.Println("第 8 步完成：完整 CLI 入口已落地。")
		fmt.Println("下一步我们将实现配置持久化存储层，为后续 HTTP API 和 Web 管理端做准备。")
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

	fmt.Println("第 8 步完成：完整 CLI 入口已落地。")
	fmt.Println("下一步我们将实现配置持久化存储层，为后续 HTTP API 和 Web 管理端做准备。")
}
