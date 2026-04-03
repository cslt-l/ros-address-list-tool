package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigStore 用于统一管理应用配置的内存态与磁盘持久化。
// 它的设计目标是：
// 1. 任何地方都不要直接散着读写 config.json
// 2. 所有配置修改都通过同一层完成
// 3. 更新时保证线程安全
// 4. 写文件时尽量使用“原子替换”方式，降低写坏配置的风险
//
// 后续这些能力都会依赖它：
// - HTTP API
// - Web 管理端
// - list 增删改查
// - description 管理
// - 手工规则管理
type ConfigStore struct {
	// mu 用于保护 cfg 的并发访问。
	// 后续 HTTP 服务模式下，多个请求可能同时读写配置，所以必须加锁。
	mu sync.RWMutex

	// path 表示配置文件在磁盘上的实际位置。
	path string

	// cfg 表示当前内存中的配置副本。
	cfg AppConfig
}

// NewConfigStore 从配置文件创建一个新的 ConfigStore。
// 它负责：
// 1. 读取配置文件
// 2. 反序列化为 AppConfig
// 3. 补默认值
// 4. 做合法性校验
//
// 如果初始化成功，说明：
// - 内存中的配置是可用的
// - 磁盘上的配置当前也是合法的
func NewConfigStore(path string) (*ConfigStore, error) {
	cfg, err := LoadConfigFile(path)
	if err != nil {
		return nil, err
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return &ConfigStore{
		path: path,
		cfg:  cfg,
	}, nil
}

// GetConfig 返回当前配置的“深拷贝副本”。
// 这里不能直接把内部 cfg 暴露出去，
// 否则外部代码拿到后可以绕过锁直接修改，破坏存储层的封装。
func (s *ConfigStore) GetConfig() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return deepCopyConfig(s.cfg)
}

// Update 以事务式方式更新配置。
// 使用方式是：
// 1. 先复制一份当前配置
// 2. 在副本上做修改
// 3. 校验新配置是否合法
// 4. 持久化写回磁盘
// 5. 成功后再替换内存里的 cfg
//
// 这样做的好处是：
// - 修改过程不会直接污染当前内存配置
// - 如果中途出错，不会把内存态改坏
// - 如果写文件失败，也不会把内存态提前改掉
func (s *ConfigStore) Update(mutator func(*AppConfig) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 基于当前配置做深拷贝。
	next := deepCopyConfig(s.cfg)

	// 让调用方在副本上完成修改。
	if err := mutator(&next); err != nil {
		return err
	}

	// 重新补默认值，避免调用方只改了一部分字段后留下空值。
	next.ApplyDefaults()

	// 对修改后的配置做正式校验。
	if err := ValidateConfig(next); err != nil {
		return err
	}

	// 写回磁盘。
	if err := saveConfigFileAtomic(s.path, next); err != nil {
		return err
	}

	// 只有磁盘写成功后，才更新内存中的当前配置。
	s.cfg = next
	return nil
}

// UpsertList 新增或更新一个 address-list 定义。
// 行为规则：
// - 如果 name 已存在，则覆盖旧定义
// - 如果 name 不存在，则追加新定义
func (s *ConfigStore) UpsertList(def ListDefinition) error {
	return s.Update(func(cfg *AppConfig) error {
		for i := range cfg.Lists {
			if cfg.Lists[i].Name == def.Name {
				cfg.Lists[i] = def
				return nil
			}
		}

		cfg.Lists = append(cfg.Lists, def)
		return nil
	})
}

// DeleteList 删除一个 address-list 定义。
// 这里除了删除 list 本身，还会一并删除所有引用该 list 的手工规则，
// 避免留下悬空配置。
func (s *ConfigStore) DeleteList(name string) error {
	return s.Update(func(cfg *AppConfig) error {
		found := false

		var newLists []ListDefinition
		for _, item := range cfg.Lists {
			if item.Name == name {
				found = true
				continue
			}
			newLists = append(newLists, item)
		}

		if !found {
			return fmt.Errorf("list 不存在：%s", name)
		}

		cfg.Lists = newLists

		// 顺手清理关联的手工规则。
		var newRules []ManualRule
		for _, rule := range cfg.ManualRules {
			if rule.ListName == name {
				continue
			}
			newRules = append(newRules, rule)
		}
		cfg.ManualRules = newRules

		return nil
	})
}

// SetListDescription 单独修改某个 list 的 description。
// 这种单字段更新在后续 Web 管理端里会非常常见。
func (s *ConfigStore) SetListDescription(name, description string) error {
	return s.Update(func(cfg *AppConfig) error {
		for i := range cfg.Lists {
			if cfg.Lists[i].Name == name {
				cfg.Lists[i].Description = description
				return nil
			}
		}
		return fmt.Errorf("list 不存在：%s", name)
	})
}

// UpsertManualRule 新增或更新一条手工规则。
// 规则：
// - 如果 ID 已存在，则覆盖
// - 如果 ID 不存在，则追加
// - 如果调用方传入空 ID，则自动生成一个临时 ID
func (s *ConfigStore) UpsertManualRule(rule ManualRule) error {
	return s.Update(func(cfg *AppConfig) error {
		if rule.ID == "" {
			rule.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano())
		}

		for i := range cfg.ManualRules {
			if cfg.ManualRules[i].ID == rule.ID {
				cfg.ManualRules[i] = rule
				return nil
			}
		}

		cfg.ManualRules = append(cfg.ManualRules, rule)
		return nil
	})
}

// DeleteManualRule 删除一条手工规则。
func (s *ConfigStore) DeleteManualRule(id string) error {
	return s.Update(func(cfg *AppConfig) error {
		found := false

		var newRules []ManualRule
		for _, rule := range cfg.ManualRules {
			if rule.ID == id {
				found = true
				continue
			}
			newRules = append(newRules, rule)
		}

		if !found {
			return fmt.Errorf("manual rule 不存在：%s", id)
		}

		cfg.ManualRules = newRules
		return nil
	})
}

// deepCopyConfig 使用 JSON 序列化/反序列化做一个完整深拷贝。
// 对当前这种配置结构来说，这种方式简单且稳定。
// 优点：
// - 不需要手写大量 slice/map 的复制逻辑
// - 不容易漏字段
func deepCopyConfig(cfg AppConfig) AppConfig {
	data, _ := json.Marshal(cfg)

	var out AppConfig
	_ = json.Unmarshal(data, &out)

	out.ApplyDefaults()
	return out
}

// saveConfigFileAtomic 以“原子替换”的方式保存配置文件。
// 典型流程：
// 1. 序列化配置
// 2. 先写入同目录下的临时文件
// 3. 再用 rename 替换正式文件
//
// 这样即便中间失败，也尽量避免把原始配置写坏。
func saveConfigFileAtomic(path string, cfg AppConfig) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
