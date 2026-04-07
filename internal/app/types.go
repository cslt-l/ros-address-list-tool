package app

import (
	"os"
	"strings"
)

// RenderMode 表示渲染模式。
type RenderMode string

const (
	// RenderModeReplaceAll 表示“全量替换”模式。
	RenderModeReplaceAll RenderMode = "replace_all"

	// RenderModeDiff 表示“差异更新”模式。
	RenderModeDiff RenderMode = "diff"
)

// IPFamily 表示地址族。
type IPFamily string

const (
	// FamilyIPv4 表示 IPv4 地址族。
	FamilyIPv4 IPFamily = "ipv4"

	// FamilyIPv6 表示 IPv6 地址族。
	FamilyIPv6 IPFamily = "ipv6"
)

// ListDefinition 表示一个 address-list 的定义信息。
type ListDefinition struct {
	// Name 是 address-list 名称。
	Name string `json:"name"`

	// Family 表示地址族。
	Family IPFamily `json:"family"`

	// Enabled 表示是否启用。
	Enabled bool `json:"enabled"`

	// Description 表示列表描述。
	Description string `json:"description"`
}

// SourceConfig 表示一个输入源配置。
type SourceConfig struct {
	// Name 是来源名称。
	Name string `json:"name"`

	// Type 表示来源类型，支持 file / url。
	Type string `json:"type"`

	// Path 用于 file 类型。
	Path string `json:"path,omitempty"`

	// URL 用于 url 类型。
	URL string `json:"url,omitempty"`

	// Headers 用于 url 类型附加请求头。
	Headers map[string]string `json:"headers,omitempty"`

	// Format 表示来源格式，支持 json / plain_cidr。
	Format string `json:"format,omitempty"`

	// TargetListName 仅在 plain_cidr 格式下有意义。
	TargetListName string `json:"target_list_name,omitempty"`

	// TargetListFamily 仅在 plain_cidr 格式下有意义。
	TargetListFamily IPFamily `json:"target_list_family,omitempty"`

	// LineCommentPrefixes 仅在 plain_cidr 格式下有意义。
	LineCommentPrefixes []string `json:"line_comment_prefixes,omitempty"`

	// Enabled 表示是否启用。
	Enabled bool `json:"enabled"`

	// Priority 表示来源优先级。
	Priority int `json:"priority"`

	// TimeoutSeconds 表示请求超时时间（秒）。
	TimeoutSeconds int `json:"timeout_seconds"`
}

// ManualRule 表示手工维护规则。
type ManualRule struct {
	// ID 是规则唯一标识。
	ID string `json:"id"`

	// ListName 表示作用的目标列表。
	ListName string `json:"list_name"`

	// Action 表示 add / remove。
	Action string `json:"action"`

	// Priority 表示优先级。
	Priority int `json:"priority"`

	// Enabled 表示是否启用。
	Enabled bool `json:"enabled"`

	// Description 表示规则说明。
	Description string `json:"description"`

	// Entries 表示规则条目。
	Entries []string `json:"entries"`
}

// OutputConfig 表示输出配置。
type OutputConfig struct {
	// Path 表示默认输出路径。
	Path string `json:"path"`

	// Mode 表示默认渲染模式。
	Mode RenderMode `json:"mode"`

	// ManagedComment 表示程序统一写入的 comment。
	ManagedComment string `json:"managed_comment"`
}

// ServerConfig 表示 HTTP 服务配置。
type ServerConfig struct {
	// Listen 表示监听地址。
	Listen string `json:"listen"`

	// EnableWeb 表示是否启用内置 Web 静态目录。
	EnableWeb bool `json:"enable_web"`

	// WebDir 表示前端静态目录。
	WebDir string `json:"web_dir"`

	// AuthToken 表示 HTTP API 的访问令牌。
	// 建议优先通过环境变量 ROS_LIST_API_TOKEN 注入，
	// 避免把敏感 token 直接写死在配置文件里。
	AuthToken string `json:"auth_token,omitempty"`
}

// AppConfig 是整个程序的总配置。
type AppConfig struct {
	// AutoCreateLists 表示是否允许自动创建未知 list。
	AutoCreateLists bool `json:"auto_create_lists"`

	// LogFile 表示日志文件路径。
	LogFile string `json:"log_file"`

	// Lists 保存所有 address-list 定义。
	Lists []ListDefinition `json:"lists"`

	// DesiredSources 表示目标状态输入源。
	DesiredSources []SourceConfig `json:"desired_sources"`

	// CurrentStateSources 表示当前状态快照输入源。
	CurrentStateSources []SourceConfig `json:"current_state_sources"`

	// ManualRules 表示手工规则。
	ManualRules []ManualRule `json:"manual_rules"`

	// Output 表示输出配置。
	Output OutputConfig `json:"output"`

	// Server 表示服务配置。
	Server ServerConfig `json:"server"`
}

// SourceList 表示某个来源中的一组 list 数据。
type SourceList struct {
	// Name 表示 list 名称。
	Name string `json:"name"`

	// Entries 表示条目。
	Entries []string `json:"entries"`

	// Description 表示来源附带的描述。
	Description string `json:"description,omitempty"`

	// Family 表示来源附带的地址族。
	Family IPFamily `json:"family,omitempty"`
}

// SourcePayload 表示标准 JSON 来源格式。
type SourcePayload struct {
	Lists []SourceList `json:"lists"`
}

// LoadedSource 表示已经完成加载的来源结果。
type LoadedSource struct {
	// Source 是原始来源配置。
	Source SourceConfig

	// Lists 是解析出的列表数据。
	Lists []SourceList
}

// Snapshot 表示某一时刻的地址快照。
type Snapshot struct {
	// Definitions 保存所有列表定义。
	Definitions map[string]ListDefinition

	// Entries 保存每个列表的条目。
	Entries map[string][]string
}

// ExecuteResult 表示一次执行的结果摘要。
type ExecuteResult struct {
	// Script 表示最终生成的 RouterOS 脚本文本。
	Script string

	// OutputPath 表示写出的目标文件路径。
	OutputPath string

	// Mode 表示本次执行使用的模式。
	Mode RenderMode

	// ListCount 表示涉及的 list 数量。
	ListCount int

	// EntryCount 表示涉及的总条目数量。
	EntryCount int
}

// ApplyDefaults 用于补充默认值。
func (cfg *AppConfig) ApplyDefaults() {
	// 默认 managed comment。
	if cfg.Output.ManagedComment == "" {
		cfg.Output.ManagedComment = "managed-by-go"
	}

	// 默认渲染模式。
	if cfg.Output.Mode == "" {
		cfg.Output.Mode = RenderModeReplaceAll
	}

	// 默认监听地址，改成仅本机。
	if cfg.Server.Listen == "" {
		cfg.Server.Listen = "127.0.0.1:8090"
	}

	// 默认前端目录。
	if cfg.Server.WebDir == "" {
		cfg.Server.WebDir = "./web/src"
	}

	// 默认日志文件。
	if cfg.LogFile == "" {
		cfg.LogFile = "./logs/app.log"
	}

	// 如果环境变量里提供了 token，则覆盖配置值。
	if envToken := strings.TrimSpace(os.Getenv("ROS_LIST_API_TOKEN")); envToken != "" {
		cfg.Server.AuthToken = envToken
	}

	// lists 中未填 family 的，默认补成 ipv4。
	for i := range cfg.Lists {
		if cfg.Lists[i].Family == "" {
			cfg.Lists[i].Family = FamilyIPv4
		}
	}

	// desired sources 默认 timeout。
	for i := range cfg.DesiredSources {
		if cfg.DesiredSources[i].TimeoutSeconds <= 0 {
			cfg.DesiredSources[i].TimeoutSeconds = 15
		}
	}

	// current state sources 默认 timeout。
	for i := range cfg.CurrentStateSources {
		if cfg.CurrentStateSources[i].TimeoutSeconds <= 0 {
			cfg.CurrentStateSources[i].TimeoutSeconds = 15
		}
	}
}
