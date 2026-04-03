package app

// RenderMode 表示渲染模式。
// 这个类型的作用是明确表达：
// 当前程序在生成 RouterOS .rsc 脚本时，到底采用哪一种策略。
//
// 为什么不用裸 string？
// 主要有三个原因：
// 1. 可读性更好，代码里一看就知道这是“渲染模式”
// 2. 后续校验时更方便限制合法值
// 3. 后续如果要扩展第三种模式，也更容易维护
type RenderMode string

const (
	// RenderModeReplaceAll 表示“全量替换”模式。
	//
	// 该模式的典型行为是：
	// 1. 先删除某个 list 中由程序管理的旧条目
	// 2. 再把当前期望状态中的条目全部重新 add 回去
	//
	// 这种模式的优点：
	// - 实现简单
	// - 行为稳定
	// - 不依赖当前在线状态
	// - 适合离线生成 .rsc 后手工导入 RouterOS
	RenderModeReplaceAll RenderMode = "replace_all"

	// RenderModeDiff 表示“差异更新”模式。
	//
	// 该模式的典型行为是：
	// 1. 对比“当前快照”和“目标快照”
	// 2. 仅输出新增项和删除项
	//
	// 这种模式的优点：
	// - 生成脚本更短
	// - 变更更清晰
	// - 更适合审计
	//
	// 但它依赖一个前提：
	// 必须能拿到当前 RouterOS 上已有条目的快照。
	RenderModeDiff RenderMode = "diff"
)

// IPFamily 表示地址族。
// 之所以单独抽出来，而不是让程序仅凭 list 名字猜测，
// 是因为未来你会同时支持：
// 1. IPv4 address-list
// 2. IPv6 address-list
//
// RouterOS 中两者对应的命令路径不同：
// - IPv4: /ip firewall address-list
// - IPv6: /ipv6 firewall address-list
//
// 所以地址族必须成为正式模型的一部分。
type IPFamily string

const (
	// FamilyIPv4 表示 IPv4 地址族。
	FamilyIPv4 IPFamily = "ipv4"

	// FamilyIPv6 表示 IPv6 地址族。
	FamilyIPv6 IPFamily = "ipv6"
)

// ListDefinition 表示一个 address-list 的定义信息。
// 注意：这里表示的是“管理元数据”，不是该列表里的实际地址条目。
//
// 这个结构体很重要，因为它承担了“配置中心”的角色。
// 后续无论是 CLI、HTTP API、还是 Web 管理端，
// 都会围绕它对 address-list 做增删改查。
type ListDefinition struct {
	// Name 是 address-list 的名称。
	//
	// 例如：
	// - toWanTelecom
	// - toWanMobile
	// - toWanIpv4
	// - toWanIpv6
	//
	// 这是该列表的唯一标识符。
	Name string `json:"name"`

	// Family 表示该列表对应的地址族。
	//
	// 取值通常为：
	// - ipv4
	// - ipv6
	//
	// 程序后续会据此决定渲染到：
	// - /ip firewall address-list
	// - /ipv6 firewall address-list
	Family IPFamily `json:"family"`

	// Enabled 表示该 list 是否启用。
	//
	// 启用时：
	// - 会参与目标状态构建
	// - 会参与渲染
	//
	// 禁用时：
	// - 不再保留目标条目
	// - 但在 replace_all 模式下，后续仍可能输出 remove，
	//   用于清理历史残留
	Enabled bool `json:"enabled"`

	// Description 表示该 address-list 的说明文字。
	//
	// 这个字段是“管理信息”，主要给人看：
	// - CLI 输出摘要
	// - HTTP API 返回给前端
	// - Web 管理端展示
	//
	// 后续你可以随时修改它，而不会影响条目本身。
	Description string `json:"description"`
}

// SourceConfig 表示一个输入源的配置。
// 该输入源可以是：
// 1. 本地文件
// 2. 远程 URL
//
// 你要求支持“多个 URL、多个 JSON”，
// 所以未来配置中会有多个 SourceConfig 组成的数组。
type SourceConfig struct {
	// Name 是该 source 的逻辑名称。
	//
	// 这个字段主要用于：
	// - 日志输出
	// - 错误定位
	// - 后续 Web 管理端显示
	Name string `json:"name"`

	// Type 表示来源类型。
	//
	// 目前计划支持：
	// - file：从本地 JSON 文件读取
	// - url：从远程 HTTP 地址读取
	Type string `json:"type"`

	// Path 是本地文件路径。
	// 当 Type=file 时使用。
	Path string `json:"path,omitempty"`

	// URL 是远程地址。
	// 当 Type=url 时使用。
	URL string `json:"url,omitempty"`

	// Headers 用于保存访问 URL 时需要带上的请求头。
	// 例如未来如果某些接口需要 Authorization，
	// 或者需要自定义 User-Agent，就可以放在这里。
	Headers map[string]string `json:"headers,omitempty"`

	// TimeoutSeconds 表示访问 URL 时的超时时间，单位为秒。
	//
	// 这个字段只对 url 类型有意义。
	// 之所以现在就设计进去，是为了避免未来远程源阻塞整个程序。
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`

	// Enabled 表示该 source 是否启用。
	//
	// 关闭某个 source 后：
	// - 程序不再读取它
	// - 但该 source 的配置仍然保留，便于以后恢复
	Enabled bool `json:"enabled"`

	// Priority 表示该 source 的优先级。
	//
	// 数值越大，优先级越高。
	// 后续在合并多个来源时，会用这个值决定冲突时谁覆盖谁。
	Priority int `json:"priority"`
}

// ManualRule 表示手工维护规则。
// 这是本项目非常重要的扩展能力之一。
//
// 你原本要求的是：
// - 支持追加手工维护名单
// - 支持配置优先级
//
// 我这里把它升级成更完整的形式：
// 1. 支持 add
// 2. 支持 remove
// 3. 支持 priority
//
// 这样就不只是“追加”，而是真正的“人工覆盖层”。
type ManualRule struct {
	// ID 是规则的唯一标识。
	//
	// 后续通过 HTTP API 修改或删除规则时，
	// 一般会依赖这个字段。
	ID string `json:"id"`

	// ListName 表示该规则作用于哪个 address-list。
	ListName string `json:"list_name"`

	// Action 表示规则动作。
	//
	// 计划合法值：
	// - add
	// - remove
	//
	// 未来合并阶段会据此决定该条目最终是保留还是移除。
	Action string `json:"action"`

	// Priority 表示该规则的优先级。
	//
	// 一般来说，手工规则优先级通常会高于普通 source，
	// 这样才能达到“人工强制覆盖”的效果。
	Priority int `json:"priority"`

	// Enabled 表示该规则是否启用。
	Enabled bool `json:"enabled"`

	// Description 表示这条规则的说明。
	//
	// 这个字段很重要，后续排查为什么某个 IP 被加进来或删掉时，
	// 你可以直接从说明中看出原因。
	Description string `json:"description"`

	// Entries 是这条规则管理的地址条目列表。
	//
	// 未来会在校验阶段做：
	// - 合法性检查
	// - 地址族一致性检查
	// - 去重
	// - 规范化
	Entries []string `json:"entries"`
}

// OutputConfig 表示输出相关配置。
type OutputConfig struct {
	// Path 表示生成的 RouterOS .rsc 文件默认输出位置。
	Path string `json:"path"`

	// Mode 表示当前默认使用的渲染模式。
	Mode RenderMode `json:"mode"`

	// ManagedComment 表示程序写入 RouterOS 时统一使用的 comment。
	//
	// 后续 replace_all 和 diff 模式都会依赖它来识别：
	// 哪些条目是程序管理的，哪些是人工手工维护的。
	//
	// 这样程序就不会误删手工条目。
	ManagedComment string `json:"managed_comment"`
}

// ServerConfig 表示 HTTP 服务与未来 Web 管理端相关配置。
type ServerConfig struct {
	// Listen 表示 HTTP 服务监听地址。
	// 例如：
	// - :8090
	// - 127.0.0.1:8090
	Listen string `json:"listen"`

	// EnableWeb 表示是否启用内置 Web 静态资源目录。
	//
	// 当前阶段先保留该字段，后续进入 Web 管理端阶段时会真正使用。
	EnableWeb bool `json:"enable_web"`

	// WebDir 表示前端构建产物所在目录。
	//
	// 未来前端打包后，可以把 dist 目录挂在这里。
	WebDir string `json:"web_dir"`
}

// AppConfig 是整个应用程序的总配置。
// 这是当前项目最核心的配置结构体。
//
// 后续：
// - CLI 会读取它
// - HTTP API 会返回它
// - 配置持久化层会存储它
// - Web 管理端会编辑它
type AppConfig struct {
	// AutoCreateLists 表示是否允许程序在发现未知 list 时自动创建定义。
	//
	// 如果为 true：
	// - 当 source 或 manual rule 中出现一个未在 lists 中定义的新名称时，
	//   后续流程可以自动为它补出一个最小定义
	//
	// 如果为 false：
	// - 所有 list 都必须先在 lists 中显式声明
	AutoCreateLists bool `json:"auto_create_lists"`

	// LogFile 表示日志文件位置。
	LogFile string `json:"log_file"`

	// Lists 保存所有 address-list 的元数据定义。
	Lists []ListDefinition `json:"lists"`

	// DesiredSources 表示“目标状态”的输入源。
	//
	// 这些 source 决定“我希望 RouterOS 最终有哪些条目”。
	DesiredSources []SourceConfig `json:"desired_sources"`

	// CurrentStateSources 表示“当前状态快照”的输入源。
	//
	// 这些 source 主要给 diff 模式使用，
	// 用来表示 RouterOS 当前已经有哪些条目。
	CurrentStateSources []SourceConfig `json:"current_state_sources"`

	// ManualRules 表示人工维护规则。
	ManualRules []ManualRule `json:"manual_rules"`

	// Output 表示输出配置。
	Output OutputConfig `json:"output"`

	// Server 表示服务配置。
	Server ServerConfig `json:"server"`
}

// SourceList 表示某个 source 中的一组 list 数据。
// 这里表示的是“来源数据中的实际条目”。
//
// 它和 ListDefinition 的区别是：
// - ListDefinition：管理元数据
// - SourceList：实际数据载体
type SourceList struct {
	// Name 表示 address-list 名称。
	Name string `json:"name"`

	// Entries 表示实际地址条目。
	Entries []string `json:"entries"`

	// Description 表示来源里附带的说明。
	//
	// 注意：
	// 这个说明只是“来源附加信息”。
	// 最终真正的管理描述，仍以后续 AppConfig.Lists 中的 Description 为准。
	Description string `json:"description,omitempty"`

	// Family 表示来源里附带的地址族。
	//
	// 如果来源不写，后续程序可以根据配置或地址内容推断。
	Family IPFamily `json:"family,omitempty"`
}

// SourcePayload 表示标准输入 JSON 的结构。
// 推荐格式示例：
//
//	{
//	  "lists": [
//	    {
//	      "name": "toWanTelecom",
//	      "entries": ["1.1.1.1", "10.0.0.0/24"]
//	    }
//	  ]
//	}
type SourcePayload struct {
	Lists []SourceList `json:"lists"`
}

// Snapshot 表示某一时刻程序视角下的 address-list 快照。
// 它是后续“合并引擎”和“diff 渲染器”的核心输入。
type Snapshot struct {
	// Definitions 保存所有 list 的定义信息。
	Definitions map[string]ListDefinition

	// Entries 保存每个 list 当前有哪些地址条目。
	//
	// key 是 list 名称
	// value 是该 list 下的地址列表
	Entries map[string][]string
}

// ExecuteResult 表示一次执行结果摘要。
// 这个结构体后续会被：
// - CLI 打印
// - HTTP API 返回
// - Web 管理端展示
type ExecuteResult struct {
	// Script 表示最终生成的 RouterOS .rsc 文本内容。
	Script string

	// OutputPath 表示本次写出的目标文件路径。
	OutputPath string

	// Mode 表示本次执行使用的渲染模式。
	Mode RenderMode

	// ListCount 表示本次涉及的 list 数量。
	ListCount int

	// EntryCount 表示本次涉及的总条目数量。
	EntryCount int
}

// ApplyDefaults 用于给配置补充默认值。
// 这样做的目的，是让配置文件不必写得过于冗长，
// 同时也保证程序运行时能拿到稳定的默认行为。
//
// 注意：
// 这个函数当前只负责“补默认值”，不负责“校验合法性”。
// 合法性校验会在下一步单独实现。
func (cfg *AppConfig) ApplyDefaults() {
	// 如果没有设置默认 managed comment，则使用 managed-by-go。
	if cfg.Output.ManagedComment == "" {
		cfg.Output.ManagedComment = "managed-by-go"
	}

	// 如果没有设置默认模式，则使用 replace_all。
	if cfg.Output.Mode == "" {
		cfg.Output.Mode = RenderModeReplaceAll
	}

	// 如果没有设置默认监听地址，则使用 :8090。
	if cfg.Server.Listen == "" {
		cfg.Server.Listen = ":8090"
	}

	// 如果没有设置前端目录，则给一个默认值。
	if cfg.Server.WebDir == "" {
		cfg.Server.WebDir = "./web/dist"
	}

	// 如果没有设置日志文件位置，则给一个默认值。
	if cfg.LogFile == "" {
		cfg.LogFile = "./logs/app.log"
	}

	// 对 lists 中未填写 family 的项，先默认补成 ipv4。
	// 这是一个偏保守的默认行为，因为大多数 address-list 初期都是 IPv4。
	for i := range cfg.Lists {
		if cfg.Lists[i].Family == "" {
			cfg.Lists[i].Family = FamilyIPv4
		}
	}

	// 对 desired sources 中没有填写 timeout 的项，补上默认值。
	for i := range cfg.DesiredSources {
		if cfg.DesiredSources[i].TimeoutSeconds <= 0 {
			cfg.DesiredSources[i].TimeoutSeconds = 15
		}
	}

	// 对 current state sources 中没有填写 timeout 的项，补上默认值。
	for i := range cfg.CurrentStateSources {
		if cfg.CurrentStateSources[i].TimeoutSeconds <= 0 {
			cfg.CurrentStateSources[i].TimeoutSeconds = 15
		}
	}
}
