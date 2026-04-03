package app

import (
	"fmt"
	"sort"
	"strings"
)

// ruleDecision 表示某个具体条目在“决策层”上的最终动作。
// 例如：
// - 某个 IP 先来自 source A（priority=100, add）
// - 又被 source B 覆盖（priority=110, add）
// - 最后又被 manual rule 覆盖（priority=2000, remove）
//
// 那么这个结构体就用于保存：
// 该条目当前最终应该执行什么动作、来自哪一层、优先级是多少。
//
// 为什么不直接在 map 里只存字符串 action？
// 因为后续比较时我们需要同时知道：
// 1. action
// 2. priority
// 3. 来源说明（便于调试和日志）
//
// 所以这里单独抽一个结构体，后续维护会更清晰。
type ruleDecision struct {
	// Action 表示最终动作：
	// - add
	// - remove
	Action string

	// Priority 表示当前生效决策的优先级。
	// 数值越大，优先级越高。
	Priority int

	// Source 表示这条决策来自哪里。
	// 例如：
	// - source:desired-local-1
	// - source:desired-local-2
	// - manual:rule-global-remove
	//
	// 当前阶段它主要用于调试输出；
	// 后续做日志和 Web 管理端审计时也会非常有帮助。
	Source string
}

// BuildDesiredSnapshot 根据：
// 1. list 定义
// 2. desired_sources
// 3. manual_rules
//
// 构建出“目标快照”。
//
// 这是整个项目最关键的函数之一。
// 它真正回答的是：
// “综合所有来源和手工规则之后，我最终希望 RouterOS 上有哪些条目？”
//
// 主要流程：
// 1. 先构建 list definition map
// 2. 加载 desired sources
// 3. 将每个 source 中的地址条目做规范化和去重
// 4. 按 priority 做决策
// 5. 再应用 manual rules（通常优先级更高）
// 6. 最终收敛成 Snapshot
func BuildDesiredSnapshot(cfg AppConfig) (Snapshot, error) {
	cfg.ApplyDefaults()

	// 先把配置中的 list 定义转成 map，便于快速查找。
	definitions := buildDefinitionMap(cfg.Lists)

	// 加载目标来源。
	loaded, err := LoadSources(cfg.DesiredSources)
	if err != nil {
		return Snapshot{}, err
	}

	// decisions 用于记录：
	// 每个 list 下，每个 entry 当前最终由谁决定、动作为何。
	//
	// 第一层 key：listName
	// 第二层 key：normalized entry
	// value：当前生效决策
	decisions := make(map[string]map[string]ruleDecision)

	var verr ValidationError

	// 先处理 desired sources。
	for _, src := range loaded {
		for _, list := range src.Lists {
			// 先保证这个 list 有 definition。
			// 如果配置中没有，但允许 auto_create_lists，则自动创建。
			def, err := ensureDefinition(
				cfg.AutoCreateLists,
				definitions,
				list.Name,
				list.Family,
				list.Description,
				list.Entries,
			)
			if err != nil {
				verr.add(fmt.Sprintf("source=%s list=%s 无法建立 definition：%v", src.Source.Name, list.Name, err))
				continue
			}
			definitions[list.Name] = def

			// 对该来源中的地址条目做规范化与去重。
			normalizedEntries, err := NormalizeAndDeduplicateEntries(list.Entries, def.Family)
			if err != nil {
				verr.add(fmt.Sprintf("source=%s list=%s 地址校验失败：%v", src.Source.Name, list.Name, err))
				continue
			}

			// desired source 默认动作都是 add。
			for _, entry := range normalizedEntries {
				applyDecision(decisions, list.Name, entry, ruleDecision{
					Action:   "add",
					Priority: src.Source.Priority,
					Source:   "source:" + src.Source.Name,
				})
			}
		}
	}

	// 再处理手工规则。
	// 这一步非常关键，因为它是“人工覆盖层”。
	for _, rule := range cfg.ManualRules {
		if !rule.Enabled {
			continue
		}

		def, err := ensureDefinition(
			cfg.AutoCreateLists,
			definitions,
			rule.ListName,
			"",
			"",
			rule.Entries,
		)
		if err != nil {
			verr.add(fmt.Sprintf("manual_rule=%s 无法建立 definition：%v", rule.ID, err))
			continue
		}
		definitions[rule.ListName] = def

		normalizedEntries, err := NormalizeAndDeduplicateEntries(rule.Entries, def.Family)
		if err != nil {
			verr.add(fmt.Sprintf("manual_rule=%s 地址校验失败：%v", rule.ID, err))
			continue
		}

		for _, entry := range normalizedEntries {
			applyDecision(decisions, rule.ListName, entry, ruleDecision{
				Action:   rule.Action,
				Priority: rule.Priority,
				Source:   "manual:" + rule.ID,
			})
		}
	}

	if verr.hasError() {
		return Snapshot{}, &verr
	}

	// 将 decisions 收敛成最终 snapshot。
	entries := make(map[string][]string)

	for listName, entryMap := range decisions {
		def, ok := definitions[listName]
		if !ok {
			// 正常情况下不会发生，因为 ensureDefinition 已经兜底。
			continue
		}

		// 如果 list 被禁用，则目标快照中不保留任何条目。
		if !def.Enabled {
			entries[listName] = []string{}
			continue
		}

		for entry, decision := range entryMap {
			if decision.Action == "add" {
				entries[listName] = append(entries[listName], entry)
			}
		}

		sort.Strings(entries[listName])
	}

	return Snapshot{
		Definitions: definitions,
		Entries:     entries,
	}, nil
}

// BuildCurrentSnapshot 根据 current_state_sources 构建“当前快照”。
// 它和 BuildDesiredSnapshot 的区别在于：
// 1. 当前快照不应用 manual rules
// 2. 当前快照主要是为了 diff 模式和状态展示
//
// 这里我们仍然会：
// - 加载多个 current source
// - 做地址规范化与去重
// - 合并为统一 snapshot
//
// 对 current source 而言，通常只表示“目前已有”，
// 所以默认都按 add 语义处理即可。
func BuildCurrentSnapshot(cfg AppConfig) (Snapshot, error) {
	cfg.ApplyDefaults()

	definitions := buildDefinitionMap(cfg.Lists)

	loaded, err := LoadSources(cfg.CurrentStateSources)
	if err != nil {
		return Snapshot{}, err
	}

	// 当前快照不需要复杂的优先级覆盖语义，
	// 只需要把当前看到的所有条目归并成集合即可。
	entrySet := make(map[string]map[string]struct{})

	var verr ValidationError

	for _, src := range loaded {
		for _, list := range src.Lists {
			def, err := ensureDefinition(
				true,
				definitions,
				list.Name,
				list.Family,
				list.Description,
				list.Entries,
			)
			if err != nil {
				verr.add(fmt.Sprintf("current source=%s list=%s 无法建立 definition：%v", src.Source.Name, list.Name, err))
				continue
			}
			definitions[list.Name] = def

			normalizedEntries, err := NormalizeAndDeduplicateEntries(list.Entries, def.Family)
			if err != nil {
				verr.add(fmt.Sprintf("current source=%s list=%s 地址校验失败：%v", src.Source.Name, list.Name, err))
				continue
			}

			if _, ok := entrySet[list.Name]; !ok {
				entrySet[list.Name] = make(map[string]struct{})
			}

			for _, entry := range normalizedEntries {
				entrySet[list.Name][entry] = struct{}{}
			}
		}
	}

	if verr.hasError() {
		return Snapshot{}, &verr
	}

	entries := make(map[string][]string)
	for listName, set := range entrySet {
		for entry := range set {
			entries[listName] = append(entries[listName], entry)
		}
		sort.Strings(entries[listName])
	}

	return Snapshot{
		Definitions: definitions,
		Entries:     entries,
	}, nil
}

// buildDefinitionMap 将 slice 形式的 list 定义转换成 map。
// 这样后续查询 list definition 时复杂度更低。
func buildDefinitionMap(in []ListDefinition) map[string]ListDefinition {
	out := make(map[string]ListDefinition, len(in))
	for _, item := range in {
		out[item.Name] = item
	}
	return out
}

// ensureDefinition 确保某个 list 一定存在 definition。
// 处理逻辑如下：
// 1. 如果 definitions 中已经有，直接返回
// 2. 如果没有，但 autoCreate=true，则自动创建一个最小 definition
// 3. 如果没有且 autoCreate=false，则报错
//
// family 的确定顺序：
// 1. 优先使用已存在 definition 的 family
// 2. 其次使用来源中显式给出的 family
// 3. 再其次从 entries 中推断
// 4. 如果还推断不出来，则默认 ipv4
func ensureDefinition(
	autoCreate bool,
	definitions map[string]ListDefinition,
	name string,
	family IPFamily,
	desc string,
	entries []string,
) (ListDefinition, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ListDefinition{}, fmt.Errorf("list name 为空")
	}

	// 如果已经存在定义，优先使用已有定义。
	if def, ok := definitions[name]; ok {
		// 如果已有 definition 的描述为空，而来源带了描述，则顺手补上。
		if def.Description == "" && desc != "" {
			def.Description = desc
		}
		// 如果已有 definition 没写 family，而来源给了，则补上。
		if def.Family == "" && family != "" {
			def.Family = family
		}
		// 仍然兜底默认值。
		if def.Family == "" {
			def.Family = FamilyIPv4
		}
		return def, nil
	}

	if !autoCreate {
		return ListDefinition{}, fmt.Errorf("list=%q 未定义，且 auto_create_lists=false", name)
	}

	// 尝试自动推断 family。
	inferredFamily := family

	// 如果来源没给 family，则从 entries 推断。
	if inferredFamily == "" {
		for _, entry := range entries {
			_, detectedFamily, err := NormalizeAddress(entry, "")
			if err == nil {
				inferredFamily = detectedFamily
				break
			}
		}
	}

	// 最后兜底默认 IPv4。
	if inferredFamily == "" {
		inferredFamily = FamilyIPv4
	}

	def := ListDefinition{
		Name:        name,
		Family:      inferredFamily,
		Enabled:     true,
		Description: desc,
	}
	return def, nil
}

// applyDecision 用于把某个 entry 的决策写入 decisions。
// 规则非常重要：
// 1. 如果当前还没有决策，则直接写入
// 2. 如果已经有决策，则“优先级更高者覆盖优先级更低者”
// 3. 如果优先级相同，则“后写入者覆盖前写入者”
//
// 为什么同优先级后写入者覆盖前写入者？
// 因为这样行为是稳定且容易理解的：
// - source 按配置顺序处理
// - manual rules 也按配置顺序处理
// 当优先级相同时，越靠后的规则越像“更晚的管理员决策”。
func applyDecision(
	decisions map[string]map[string]ruleDecision,
	listName string,
	entry string,
	next ruleDecision,
) {
	if _, ok := decisions[listName]; !ok {
		decisions[listName] = make(map[string]ruleDecision)
	}

	current, ok := decisions[listName][entry]
	if !ok {
		decisions[listName][entry] = next
		return
	}

	if next.Priority > current.Priority {
		decisions[listName][entry] = next
		return
	}

	if next.Priority == current.Priority {
		decisions[listName][entry] = next
	}
}
