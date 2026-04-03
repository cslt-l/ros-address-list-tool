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
	fmt.Println("RouterOS Address List Tool - 合并引擎阶段")
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

	// 保留前一步的最小规范化演示，确保底层能力仍然工作正常。
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

	// 构建目标快照。
	desiredSnapshot, err := app.BuildDesiredSnapshot(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "构建目标快照失败：\n%s\n", err.Error())
		os.Exit(1)
	}

	// 构建当前快照。
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

	fmt.Println()
	fmt.Println("第 5 步完成：合并引擎已落地。")
	fmt.Println("下一步我们将实现渲染器：把快照变成 RouterOS 可导入的 .rsc 脚本。")
}

// printSnapshot 用于把 Snapshot 以稳定顺序打印出来。
// 这个函数当前只服务于命令行演示，
// 主要目的是让你能肉眼确认：
// 1. 哪些 list 进入了快照
// 2. 每个 list 最终有哪些条目
//
// 之所以单独抽成函数，是为了让 main 保持清晰，
// 同时后续如果要做更丰富的 CLI 输出，也方便继续扩展。
func printSnapshot(s app.Snapshot) {
	// 为了保证输出顺序稳定，先排序 list 名称。
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
