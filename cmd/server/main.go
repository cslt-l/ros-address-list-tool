package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"ros-address-list-tool/internal/app"
)

func main() {
	// configPath 表示配置文件路径。
	configPath := flag.String("config", "./configs/config.json", "配置文件路径")
	flag.Parse()

	fmt.Println("================================================")
	fmt.Println("RouterOS Address List Tool - 渲染器阶段")
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

	cfg, err := app.LoadConfigFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载正式配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := app.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "配置校验失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("配置校验通过。")

	// 保留最小规范化演示，确保基础能力仍然正常。
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

	desiredSnapshot, err := app.BuildDesiredSnapshot(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "构建目标快照失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	currentSnapshot, err := app.BuildCurrentSnapshot(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "构建当前快照失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("目标快照（desired snapshot）：")
	printSnapshot(desiredSnapshot)

	fmt.Println()
	fmt.Println("当前快照（current snapshot）：")
	printSnapshot(currentSnapshot)

	// 用配置中的默认模式进行渲染。
	script, err := app.RenderScript(
		desiredSnapshot,
		currentSnapshot,
		cfg.Output.Mode,
		cfg.Output.ManagedComment,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "渲染默认模式脚本失败：%v\n", err)
		os.Exit(1)
	}

	// 同时再额外生成一个 diff 预览，便于你比较两种模式。
	diffScript, err := app.RenderScript(
		desiredSnapshot,
		currentSnapshot,
		app.RenderModeDiff,
		cfg.Output.ManagedComment,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "渲染 diff 脚本失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("默认模式脚本输出（mode=%s）：\n", cfg.Output.Mode)
	fmt.Println("------------------------------------------------")
	fmt.Println(script)

	fmt.Println("diff 模式脚本预览：")
	fmt.Println("------------------------------------------------")
	fmt.Println(diffScript)

	fmt.Println("第 6 步完成：渲染器已落地。")
	fmt.Println("下一步我们将实现执行引擎与日志：把脚本写到文件，并形成真正的单次执行流程。")
}

// printSnapshot 用于稳定打印 Snapshot。
// 当前仍然保留“定义全集 + 条目列表”的展示方式，
// 这样可以更清楚地看出：
// - 哪些 list 已定义
// - 哪些 list 当前没有条目
func printSnapshot(s app.Snapshot) {
	var listNames []string
	for name := range s.Definitions {
		listNames = append(listNames, name)
	}
	sort.Strings(listNames)

	for _, name := range listNames {
		def := s.Definitions[name]
		entries := s.Entries[name]

		fmt.Printf("- list=%s family=%s enabled=%v description=%s entries=%d\n",
			def.Name, def.Family, def.Enabled, def.Description, len(entries))

		for i, entry := range entries {
			fmt.Printf("    [%d] %s\n", i+1, entry)
		}
	}
}
