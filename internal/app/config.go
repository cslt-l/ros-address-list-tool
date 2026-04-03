package app

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfigFile 从指定路径读取配置文件，并解析为正式的 AppConfig。
//
// 这个函数当前只做三件事：
// 1. 读取文件
// 2. 反序列化 JSON
// 3. 补默认值
//
// 为什么这一步不做完整校验？
// 因为我们当前还处在“核心模型设计”阶段，
// 下一步会专门实现 ValidateConfig 等校验逻辑。
// 这样职责会更清晰：
// - LoadConfigFile：负责读取
// - ValidateConfig：负责校验
func LoadConfigFile(path string) (AppConfig, error) {
	var cfg AppConfig

	// 读取原始文件内容。
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将 JSON 反序列化到 AppConfig。
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 补充默认值，保证程序后续打印摘要时行为稳定。
	cfg.ApplyDefaults()

	return cfg, nil
}
