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
	// 保留为命令行参数，是为了方便不同环境使用不同配置文件。
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 多来源加载阶段")
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
	if err := app.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "配置校验失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("配置校验通过。")

	// 继续做一个“地址规范化与去重”的最小演示，
	// 这样我们可以确认前一步的能力仍然正常。
	demoEntries := []string{
		"10.0.0.1/24",
		"10.0.0.0/24",
		"1.1.1.1",
		"1.1.1.1",
		"223.5.5.5",
	}

	normalizedEntries, err := app.NormalizeAndDeduplicateEntries(demoEntries, app.FamilyIPv4)
	if err != nil {
		fmt.Fprintf(os.Stderr, "地址规范化示例失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("地址规范化与去重示例（IPv4）：")
	fmt.Printf("- 原始输入: %v\n", demoEntries)
	fmt.Printf("- 规范化后: %v\n", normalizedEntries)

	// 加载目标来源。
	desiredLoaded, err := app.LoadSources(cfg.DesiredSources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载 desired_sources 失败：%v\n", err)
		os.Exit(1)
	}

	// 加载当前状态来源。
	currentLoaded, err := app.LoadSources(cfg.CurrentStateSources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载 current_state_sources 失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("已成功加载 desired_sources：%d 个\n", len(desiredLoaded))
	for i, loaded := range desiredLoaded {
		fmt.Printf("  [%d] source=%s type=%s priority=%d loaded_lists=%d\n",
			i+1, loaded.Source.Name, loaded.Source.Type, loaded.Source.Priority, len(loaded.Lists))

		for j, list := range loaded.Lists {
			fmt.Printf("      - [%d] list=%s entries=%d family=%s description=%s\n",
				j+1, list.Name, len(list.Entries), list.Family, list.Description)
		}
	}

	fmt.Println()
	fmt.Printf("已成功加载 current_state_sources：%d 个\n", len(currentLoaded))
	for i, loaded := range currentLoaded {
		fmt.Printf("  [%d] source=%s type=%s priority=%d loaded_lists=%d\n",
			i+1, loaded.Source.Name, loaded.Source.Type, loaded.Source.Priority, len(loaded.Lists))

		for j, list := range loaded.Lists {
			fmt.Printf("      - [%d] list=%s entries=%d family=%s description=%s\n",
				j+1, list.Name, len(list.Entries), list.Family, list.Description)
		}
	}

	fmt.Println()
	fmt.Println("第 4 步完成：多来源数据加载器已落地。")
	fmt.Println("下一步我们将实现合并引擎：把多个来源和手工规则收敛成最终快照。")
}
