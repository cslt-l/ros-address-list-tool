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
type ConfigStore struct {
	// mu 用于保护 cfg 的并发访问。
	mu sync.RWMutex

	// path 表示配置文件在磁盘上的实际位置。
	path string

	// cfg 表示当前内存中的配置副本。
	cfg AppConfig
}

// NewConfigStore 从配置文件创建一个新的 ConfigStore。
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

// GetConfig 返回当前配置的深拷贝副本。
func (s *ConfigStore) GetConfig() AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return deepCopyConfig(s.cfg)
}

// Update 以事务式方式更新配置。
func (s *ConfigStore) Update(mutator func(*AppConfig) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := deepCopyConfig(s.cfg)

	if err := mutator(&next); err != nil {
		return err
	}

	next.ApplyDefaults()

	if err := ValidateConfig(next); err != nil {
		return err
	}

	if err := saveConfigFileAtomic(s.path, next); err != nil {
		return err
	}

	s.cfg = next
	return nil
}

// UpsertList 新增或更新一个 address-list 定义。
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

// DeleteList 删除一个 address-list 定义，并清理其关联 manual rules。
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

// UpsertDesiredSource 新增或更新一个 desired source。
func (s *ConfigStore) UpsertDesiredSource(src SourceConfig) error {
	return s.Update(func(cfg *AppConfig) error {
		cfg.DesiredSources = upsertSourceSlice(cfg.DesiredSources, src)
		return nil
	})
}

// DeleteDesiredSource 删除一个 desired source。
func (s *ConfigStore) DeleteDesiredSource(name string) error {
	return s.Update(func(cfg *AppConfig) error {
		next, found := deleteSourceSlice(cfg.DesiredSources, name)
		if !found {
			return fmt.Errorf("desired source 不存在：%s", name)
		}
		cfg.DesiredSources = next
		return nil
	})
}

// UpsertCurrentSource 新增或更新一个 current_state source。
func (s *ConfigStore) UpsertCurrentSource(src SourceConfig) error {
	return s.Update(func(cfg *AppConfig) error {
		cfg.CurrentStateSources = upsertSourceSlice(cfg.CurrentStateSources, src)
		return nil
	})
}

// DeleteCurrentSource 删除一个 current_state source。
func (s *ConfigStore) DeleteCurrentSource(name string) error {
	return s.Update(func(cfg *AppConfig) error {
		next, found := deleteSourceSlice(cfg.CurrentStateSources, name)
		if !found {
			return fmt.Errorf("current source 不存在：%s", name)
		}
		cfg.CurrentStateSources = next
		return nil
	})
}

// upsertSourceSlice 在来源切片中执行“按名称新增或覆盖”。
// 规则：
// 1. 如果名称已存在，则原位覆盖
// 2. 如果名称不存在，则追加到尾部
func upsertSourceSlice(items []SourceConfig, src SourceConfig) []SourceConfig {
	for i := range items {
		if items[i].Name == src.Name {
			items[i] = src
			return items
		}
	}

	return append(items, src)
}

// deleteSourceSlice 从来源切片中按名称删除一个 source。
// 返回值：
// 1. 删除后的切片
// 2. 是否找到
func deleteSourceSlice(items []SourceConfig, name string) ([]SourceConfig, bool) {
	found := false
	var next []SourceConfig

	for _, item := range items {
		if item.Name == name {
			found = true
			continue
		}
		next = append(next, item)
	}

	return next, found
}

// deepCopyConfig 使用 JSON 序列化/反序列化做完整深拷贝。
func deepCopyConfig(cfg AppConfig) AppConfig {
	data, _ := json.Marshal(cfg)

	var out AppConfig
	_ = json.Unmarshal(data, &out)

	out.ApplyDefaults()
	return out
}

// saveConfigFileAtomic 以原子替换的方式保存配置文件。
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
