package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigStore 用于统一管理应用配置的内存态与磁盘持久化。
type ConfigStore struct {
	mu   sync.RWMutex
	path string
	cfg  AppConfig
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

	if err := saveConfigFileSafely(s.path, next); err != nil {
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

func upsertSourceSlice(items []SourceConfig, src SourceConfig) []SourceConfig {
	for i := range items {
		if items[i].Name == src.Name {
			items[i] = src
			return items
		}
	}

	return append(items, src)
}

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

// saveConfigFileSafely 将配置先写入随机临时文件，再尽量安全地替换目标文件。
// 注意：
// - 这里不再宣称“原子保存”，因为 Windows 上 os.Rename 覆盖已有文件并不可靠。
// - 在替换前会先备份旧文件为 .bak。
// - 临时文件名使用随机后缀，避免多个进程或异常中断时固定 .tmp 文件名互相冲突。
func saveConfigFileSafely(path string, cfg AppConfig) (err error) {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	} else {
		dir = "."
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	if _, statErr := os.Stat(path); statErr == nil {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}

	if err := os.Rename(tmpPath, path); err == nil {
		renamed = true
		return nil
	}

	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		return removeErr
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	renamed = true
	return nil
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return dst.Sync()
}
