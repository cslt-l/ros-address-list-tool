package app

import (
	"fmt"
	"sort"
	"strings"
)

type ruleDecision struct {
	Action   string
	Priority int
	Source   string
}

// BuildDesiredSnapshot 根据 list 定义、desired_sources、manual_rules 构建目标快照。
func BuildDesiredSnapshot(cfg AppConfig) (Snapshot, error) {
	cfg.ApplyDefaults()

	definitions := buildDefinitionMap(cfg.Lists)

	loaded, err := LoadSources(cfg.DesiredSources)
	if err != nil {
		return Snapshot{}, err
	}

	decisions := make(map[string]map[string]ruleDecision)
	var verr ValidationError

	for _, src := range loaded {
		for _, list := range src.Lists {
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

			normalizedEntries, err := NormalizeAndDeduplicateEntries(list.Entries, def.Family)
			if err != nil {
				verr.add(fmt.Sprintf("source=%s list=%s 地址校验失败：%v", src.Source.Name, list.Name, err))
				continue
			}

			for _, entry := range normalizedEntries {
				applyDecision(decisions, list.Name, entry, ruleDecision{
					Action:   "add",
					Priority: src.Source.Priority,
					Source:   "source:" + src.Source.Name,
				})
			}
		}
	}

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

	entries := make(map[string][]string)
	for listName, entryMap := range decisions {
		def, ok := definitions[listName]
		if !ok {
			continue
		}

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

// BuildCurrentSnapshot 根据 current_state_sources 构建当前快照。
func BuildCurrentSnapshot(cfg AppConfig) (Snapshot, error) {
	cfg.ApplyDefaults()

	definitions := buildDefinitionMap(cfg.Lists)

	loaded, err := LoadSources(cfg.CurrentStateSources)
	if err != nil {
		return Snapshot{}, err
	}

	entrySet := make(map[string]map[string]struct{})
	var verr ValidationError

	for _, src := range loaded {
		for _, list := range src.Lists {
			listName := strings.TrimSpace(list.Name)

			// 修复点：
			// 当 auto_create_lists=false 时，current_state_sources 里的未知 list
			// 不应该让整个 current snapshot 构建失败，而应该直接忽略。
			if !cfg.AutoCreateLists {
				if _, ok := definitions[listName]; !ok {
					continue
				}
			}

			def, err := ensureDefinition(
				cfg.AutoCreateLists,
				definitions,
				listName,
				list.Family,
				list.Description,
				list.Entries,
			)
			if err != nil {
				verr.add(fmt.Sprintf("current source=%s list=%s 无法建立 definition：%v", src.Source.Name, list.Name, err))
				continue
			}

			definitions[listName] = def

			normalizedEntries, err := NormalizeAndDeduplicateEntries(list.Entries, def.Family)
			if err != nil {
				verr.add(fmt.Sprintf("current source=%s list=%s 地址校验失败：%v", src.Source.Name, list.Name, err))
				continue
			}

			if _, ok := entrySet[listName]; !ok {
				entrySet[listName] = make(map[string]struct{})
			}
			for _, entry := range normalizedEntries {
				entrySet[listName][entry] = struct{}{}
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

func buildDefinitionMap(in []ListDefinition) map[string]ListDefinition {
	out := make(map[string]ListDefinition, len(in))
	for _, item := range in {
		out[item.Name] = item
	}
	return out
}

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

	if def, ok := definitions[name]; ok {
		if def.Description == "" && desc != "" {
			def.Description = desc
		}
		if def.Family == "" && family != "" {
			def.Family = family
		}
		if def.Family == "" {
			def.Family = FamilyIPv4
		}
		return def, nil
	}

	if !autoCreate {
		return ListDefinition{}, fmt.Errorf("list=%q 未定义，且 auto_create_lists=false", name)
	}

	inferredFamily := family
	if inferredFamily == "" {
		for _, entry := range entries {
			_, detectedFamily, err := NormalizeAddress(entry, "")
			if err == nil {
				inferredFamily = detectedFamily
				break
			}
		}
	}
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
