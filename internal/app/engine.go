package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Execute 是当前项目“单次执行模式”的核心入口。
// 它会把前面已经实现的各层能力串起来，形成一次完整流程：
//
// 1. 补默认值
// 2. 校验配置
// 3. 构建目标快照
// 4. 构建当前快照
// 5. 渲染脚本
// 6. 写出到文件
// 7. 返回执行结果摘要
//
// 为什么要单独做一个 Execute 函数，而不是把这些流程写在 main 里？
// 因为后续：
// - CLI
// - HTTP API
// - Web 管理端
// 都会需要“触发一次渲染并输出结果”的能力。
// 把流程收敛在 Execute 中，可以避免业务逻辑散落在入口层。
func Execute(cfg AppConfig, logger *log.Logger) (ExecuteResult, error) {
	// 先补默认值，保证后续逻辑拿到的是完整配置。
	cfg.ApplyDefaults()

	logger.Printf("开始执行单次渲染流程，mode=%s output=%s", cfg.Output.Mode, cfg.Output.Path)

	// 第一步：校验配置。
	if err := ValidateConfig(cfg); err != nil {
		return ExecuteResult{}, fmt.Errorf("配置校验失败: %w", err)
	}
	logger.Printf("配置校验通过")

	// 第二步：构建目标快照。
	desired, err := BuildDesiredSnapshot(cfg)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("构建目标快照失败: %w", err)
	}
	logger.Printf("目标快照构建完成")

	// 第三步：构建当前快照。
	// 注意：
	// 如果 current_state_sources 为空，这里也允许继续执行，
	// 只是 current 会是一个空快照。
	current := Snapshot{
		Definitions: map[string]ListDefinition{},
		Entries:     map[string][]string{},
	}
	if len(cfg.CurrentStateSources) > 0 {
		current, err = BuildCurrentSnapshot(cfg)
		if err != nil {
			return ExecuteResult{}, fmt.Errorf("构建当前快照失败: %w", err)
		}
		logger.Printf("当前快照构建完成")
	} else {
		logger.Printf("当前配置未提供 current_state_sources，使用空快照")
	}

	// 第四步：渲染脚本。
	script, err := RenderScript(
		desired,
		current,
		cfg.Output.Mode,
		cfg.Output.ManagedComment,
	)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("渲染脚本失败: %w", err)
	}
	logger.Printf("脚本渲染完成")

	// 第五步：写出脚本到文件。
	outputPath := cfg.Output.Path
	if outputPath != "" {
		if err := writeOutputFile(outputPath, script); err != nil {
			return ExecuteResult{}, fmt.Errorf("写出脚本失败: %w", err)
		}
		logger.Printf("脚本已写出到文件：%s", outputPath)
	} else {
		logger.Printf("未配置输出路径，本次不写文件")
	}

	// 统计结果摘要。
	listCount, entryCount := countSnapshot(desired)

	result := ExecuteResult{
		Script:     script,
		OutputPath: outputPath,
		Mode:       cfg.Output.Mode,
		ListCount:  listCount,
		EntryCount: entryCount,
	}

	logger.Printf(
		"执行完成：mode=%s list_count=%d entry_count=%d output=%s",
		result.Mode,
		result.ListCount,
		result.EntryCount,
		result.OutputPath,
	)

	return result, nil
}

// writeOutputFile 将渲染出的脚本写入目标文件。
// 如果上级目录不存在，会自动创建。
func writeOutputFile(path string, content string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

// countSnapshot 用于统计一个快照中：
// 1. 启用的 list 数量
// 2. 条目总数
//
// 这里 listCount 的含义是：
// “当前定义中处于 enabled=true 的 list 数量”。
//
// 这里 entryCount 的含义是：
// “所有 list 下实际条目数量之和”。
func countSnapshot(s Snapshot) (int, int) {
	listCount := 0
	entryCount := 0

	for _, def := range s.Definitions {
		if def.Enabled {
			listCount++
		}
	}

	for _, entries := range s.Entries {
		entryCount += len(entries)
	}

	return listCount, entryCount
}
