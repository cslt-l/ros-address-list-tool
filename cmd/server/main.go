package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ros-address-list-tool/internal/app"
)

func main() {
	// configPath 表示配置文件路径。
	// 当前阶段仍然只保留这一个命令行参数。
	// 到下一步 CLI 阶段，我们会继续增加：
	// - output 覆盖
	// - mode 覆盖
	// - serve 模式
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 执行引擎阶段")
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

	// 先读取配置，主要是为了拿到日志路径。
	cfg, err := app.LoadConfigFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
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

	// 统一调用执行引擎。
	result, err := app.Execute(cfg, logger)
	if err != nil {
		logger.Printf("执行失败：%v", err)
		fmt.Fprintf(os.Stderr, "执行失败：%v\n", err)
		os.Exit(1)
	}

	// 打印最终结果摘要。
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

	fmt.Println("第 7 步完成：执行引擎与日志已落地。")
	fmt.Println("下一步我们将实现更完整的命令行入口：支持参数覆盖 output、mode，以及 serve 模式。")
}
