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
	//
	// 之所以保留为命令行参数，而不是写死，
	// 是因为后续我们会有这些实际场景：
	// 1. 本地开发环境用一份配置
	// 2. 测试环境用另一份配置
	// 3. 生产环境用第三份配置
	// 4. 自动化测试时用临时配置
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 正式模型加载阶段")
	fmt.Println("================================================")

	// 打印当前工作目录，便于排查相对路径问题。
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前工作目录失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("当前工作目录: %s\n", wd)

	// 打印配置文件绝对路径，便于确认程序到底读了哪份配置。
	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析配置文件绝对路径失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("准备读取配置文件: %s\n", absConfigPath)

	// 使用 internal/app 中的正式配置加载函数读取配置。
	cfg, err := app.LoadConfigFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载正式配置失败: %v\n", err)
		os.Exit(1)
	}

	// 打印正式配置摘要。
	// 当前阶段我们还不进入“执行逻辑”，
	// 只验证：正式模型已经能稳定承载整个项目所需信息。
	fmt.Println()
	fmt.Println("正式配置加载成功。当前配置摘要如下：")
	fmt.Printf("- 自动创建未知 list: %v\n", cfg.AutoCreateLists)
	fmt.Printf("- 日志文件路径: %s\n", cfg.LogFile)
	fmt.Printf("- 默认输出路径: %s\n", cfg.Output.Path)
	fmt.Printf("- 默认渲染模式: %s\n", cfg.Output.Mode)
	fmt.Printf("- managed comment: %s\n", cfg.Output.ManagedComment)
	fmt.Printf("- HTTP 监听地址: %s\n", cfg.Server.Listen)
	fmt.Printf("- 是否启用 Web 静态目录: %v\n", cfg.Server.EnableWeb)
	fmt.Printf("- Web 目录: %s\n", cfg.Server.WebDir)

	fmt.Println()
	fmt.Printf("已定义 address-list 数量: %d\n", len(cfg.Lists))
	for i, item := range cfg.Lists {
		fmt.Printf("  [%d] name=%s family=%s enabled=%v description=%s\n",
			i+1, item.Name, item.Family, item.Enabled, item.Description)
	}

	fmt.Println()
	fmt.Printf("目标来源 desired_sources 数量: %d\n", len(cfg.DesiredSources))
	for i, src := range cfg.DesiredSources {
		fmt.Printf("  [%d] name=%s type=%s enabled=%v priority=%d path=%s url=%s timeout=%d\n",
			i+1, src.Name, src.Type, src.Enabled, src.Priority, src.Path, src.URL, src.TimeoutSeconds)
	}

	fmt.Println()
	fmt.Printf("当前状态来源 current_state_sources 数量: %d\n", len(cfg.CurrentStateSources))
	for i, src := range cfg.CurrentStateSources {
		fmt.Printf("  [%d] name=%s type=%s enabled=%v priority=%d path=%s url=%s timeout=%d\n",
			i+1, src.Name, src.Type, src.Enabled, src.Priority, src.Path, src.URL, src.TimeoutSeconds)
	}

	fmt.Println()
	fmt.Printf("手工规则 manual_rules 数量: %d\n", len(cfg.ManualRules))
	for i, rule := range cfg.ManualRules {
		fmt.Printf("  [%d] id=%s list=%s action=%s enabled=%v priority=%d entries=%d description=%s\n",
			i+1, rule.ID, rule.ListName, rule.Action, rule.Enabled, rule.Priority, len(rule.Entries), rule.Description)
	}

	fmt.Println()
	fmt.Println("第 2 步完成：正式核心数据结构已落地。")
	fmt.Println("下一步我们将实现配置校验、地址合法性检查、去重与规范化。")
}
