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
	// 当前仍保留命令行参数形式，
	// 是为了让程序在开发环境、测试环境、生产环境中都能方便切换配置。
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 校验与规范化阶段")
	fmt.Println("================================================")

	// 输出当前工作目录，便于排查相对路径问题。
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取当前工作目录失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("当前工作目录: %s\n", wd)

	// 打印配置文件绝对路径，确认程序实际读取的是哪份配置。
	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析配置文件绝对路径失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("准备读取配置文件: %s\n", absConfigPath)

	// 加载正式配置。
	cfg, err := app.LoadConfigFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载正式配置失败: %v\n", err)
		os.Exit(1)
	}

	// 对配置做正式校验。
	// 这一层的职责是尽早发现“结构性错误”，
	// 让后续 source、merge、render 都运行在可信输入之上。
	if err := app.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "配置校验失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("配置校验通过。当前配置摘要如下：")
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

	// 为了验证“地址规范化与去重”这一步已经真正落地，
	// 这里做一个非常小的演示。
	//
	// 说明：
	// 这段演示不是最终业务逻辑的一部分，
	// 它只是当前开发阶段用于验证 validate.go 能力是否正常工作的最小示例。
	demoEntries := []string{
		"10.0.0.1/24", // 应规范化为 10.0.0.0/24
		"10.0.0.0/24", // 与上面规范化后重复，应被去重
		"1.1.1.1",
		"1.1.1.1", // 重复，应被去重
		"223.5.5.5",
	}

	normalizedEntries, err := app.NormalizeAndDeduplicateEntries(demoEntries, app.FamilyIPv4)
	if err != nil {
		fmt.Fprintf(os.Stderr, "地址规范化示例失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("地址规范化与去重示例（IPv4）：")
	fmt.Printf("- 原始输入: %v\n", demoEntries)
	fmt.Printf("- 规范化后: %v\n", normalizedEntries)

	fmt.Println()
	fmt.Println("第 3 步完成：配置校验、地址合法性检查、去重与规范化已落地。")
	fmt.Println("下一步我们将实现多来源数据加载器（支持多个本地 JSON 与多个 URL）。")
}
