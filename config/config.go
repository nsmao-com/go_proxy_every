package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// ProxyRule 代理规则
type ProxyRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`       // 本地路径前缀，如 /nsmao
	Target    string    `json:"target"`     // 目标地址，如 https://www.nsmao.com
	Enabled   bool      `json:"enabled"`    // 是否启用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Config 配置
type Config struct {
	Rules []ProxyRule `json:"rules"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	mu       sync.RWMutex
	config   Config
	filePath string
}

var (
	manager *ConfigManager
	once    sync.Once
)

// GetManager 获取配置管理器单例
func GetManager() *ConfigManager {
	once.Do(func() {
		manager = &ConfigManager{
			filePath: "data/rules.json",
			config:   Config{Rules: []ProxyRule{}},
		}
		manager.Load()
	})
	return manager
}

// Load 从文件加载配置
func (m *ConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建默认配置
			return m.saveWithoutLock()
		}
		return err
	}

	return json.Unmarshal(data, &m.config)
}

// Save 保存配置到文件
func (m *ConfigManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveWithoutLock()
}

func (m *ConfigManager) saveWithoutLock() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

// GetRules 获取所有规则
func (m *ConfigManager) GetRules() []ProxyRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rules := make([]ProxyRule, len(m.config.Rules))
	copy(rules, m.config.Rules)
	return rules
}

// GetEnabledRules 获取启用的规则
func (m *ConfigManager) GetEnabledRules() []ProxyRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var rules []ProxyRule
	for _, r := range m.config.Rules {
		if r.Enabled {
			rules = append(rules, r)
		}
	}
	return rules
}

// AddRule 添加规则
func (m *ConfigManager) AddRule(rule ProxyRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	m.config.Rules = append(m.config.Rules, rule)

	return m.saveWithoutLock()
}

// UpdateRule 更新规则
func (m *ConfigManager) UpdateRule(rule ProxyRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, r := range m.config.Rules {
		if r.ID == rule.ID {
			rule.CreatedAt = r.CreatedAt
			rule.UpdatedAt = time.Now()
			m.config.Rules[i] = rule
			return m.saveWithoutLock()
		}
	}

	return nil
}

// DeleteRule 删除规则
func (m *ConfigManager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, r := range m.config.Rules {
		if r.ID == id {
			m.config.Rules = append(m.config.Rules[:i], m.config.Rules[i+1:]...)
			return m.saveWithoutLock()
		}
	}

	return nil
}

// GetRuleByPath 根据路径获取规则
func (m *ConfigManager) GetRuleByPath(path string) *ProxyRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, r := range m.config.Rules {
		if r.Path == path && r.Enabled {
			return &r
		}
	}
	return nil
}
