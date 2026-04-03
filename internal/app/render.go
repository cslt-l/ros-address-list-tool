package app

import (
	"fmt"
	"sort"
	"strings"
)

// RenderScript 根据：
// 1. 目标快照 desired
// 2. 当前快照 current
// 3. 渲染模式 mode
// 4. 程序管理 comment managedComment
//
// 生成最终的 RouterOS .rsc 文本。
//
// 这是“快照层”和“最终输出层”之间的桥梁。
// 它的职责是把抽象数据结构转换成 RouterOS 可以直接执行的脚本命令。
//
// 当前支持两种模式：
// 1. replace_all
// 2. diff
//
// 为什么要把渲染入口统一成一个函数？
// 因为调用方（CLI、HTTP API、Web 管理端）并不关心具体模式内部怎么实现，
// 它们只关心：
// - 给我输入
// - 返回一段脚本文本
func RenderScript(
	desired Snapshot,
	current Snapshot,
	mode RenderMode,
	managedComment string,
) (string, error) {
	switch mode {
	case RenderModeReplaceAll:
		return renderReplaceAll(desired, current, managedComment), nil
	case RenderModeDiff:
		return renderDiff(desired, current, managedComment), nil
	default:
		return "", fmt.Errorf("未知渲染模式：%s", mode)
	}
}

// renderReplaceAll 实现“全量替换”模式。
// 它的思路是：
// 1. 先删除每个 list 中 comment=managedComment 的旧条目
// 2. 再把目标快照中的条目全部 add 回去
//
// 这个模式的最大优点是：
// - 行为稳定
// - 不依赖当前 RouterOS 在线状态
// - 非常适合离线生成 .rsc 后导入
//
// 当前实现有一个重要增强：
// 它不仅遍历 desired 中存在的 list，
// 还会遍历 current 中存在但 desired 已经没有的 list。
// 这样当某个 list 被删除或禁用后，也能把旧残留清掉。
func renderReplaceAll(
	desired Snapshot,
	current Snapshot,
	managedComment string,
) string {
	var ipv4Lines []string
	var ipv6Lines []string

	// 取 desired / current definitions 的并集。
	union := unionListNames(desired.Definitions, current.Definitions)

	for _, listName := range union {
		// 先决定该 list 归属于 IPv4 还是 IPv6。
		// 优先使用 desired 中的 definition；
		// 如果 desired 没有，则退回 current。
		def, ok := desired.Definitions[listName]
		if !ok {
			def = current.Definitions[listName]
		}

		removeCmd := fmt.Sprintf(`remove [find where list="%s" comment="%s"]`, listName, managedComment)

		target := &ipv4Lines
		if def.Family == FamilyIPv6 {
			target = &ipv6Lines
		}

		// 无论目标条目是否为空，都先输出 remove。
		// 这是 replace_all 的核心行为。
		*target = append(*target, removeCmd)

		// 如果目标定义不存在，或者目标定义存在但被禁用，
		// 那么只需要清理旧条目，不需要重新 add。
		desiredDef, existsInDesired := desired.Definitions[listName]
		if !existsInDesired || !desiredDef.Enabled {
			continue
		}

		// 如果目标 entries 为空，也不报错，只表示该 list 当前不需要任何条目。
		for _, entry := range desired.Entries[listName] {
			addCmd := fmt.Sprintf(`add list=%s address=%s comment="%s"`, listName, entry, managedComment)
			*target = append(*target, addCmd)
		}
	}

	return joinSections(ipv4Lines, ipv6Lines)
}

// renderDiff 实现“差异更新”模式。
// 它的思路是：
// 1. 比较 desired 与 current
// 2. current 有但 desired 没有 -> 输出 remove
// 3. desired 有但 current 没有 -> 输出 add
//
// 这个模式的优点是：
// - 输出脚本更短
// - 变更更聚焦
// - 更适合审计和预览
//
// 与 replace_all 相比，它更依赖 current snapshot 的准确性。
func renderDiff(
	desired Snapshot,
	current Snapshot,
	managedComment string,
) string {
	var ipv4Lines []string
	var ipv6Lines []string

	union := unionListNames(desired.Definitions, current.Definitions)

	for _, listName := range union {
		def, ok := desired.Definitions[listName]
		if !ok {
			def = current.Definitions[listName]
		}

		target := &ipv4Lines
		if def.Family == FamilyIPv6 {
			target = &ipv6Lines
		}

		currentSet := sliceToSet(current.Entries[listName])

		// 如果目标里该 list 不存在，或者存在但被禁用，
		// 则 desiredSet 视为空集合，意味着当前已有条目都需要 remove。
		desiredSet := make(map[string]struct{})
		desiredDef, existsInDesired := desired.Definitions[listName]
		if existsInDesired && desiredDef.Enabled {
			desiredSet = sliceToSet(desired.Entries[listName])
		}

		// 先输出删除项。
		var removed []string
		for entry := range currentSet {
			if _, ok := desiredSet[entry]; !ok {
				removed = append(removed, entry)
			}
		}
		sort.Strings(removed)

		for _, entry := range removed {
			removeCmd := fmt.Sprintf(
				`remove [find where list="%s" address="%s" comment="%s"]`,
				listName, entry, managedComment,
			)
			*target = append(*target, removeCmd)
		}

		// 再输出新增项。
		if existsInDesired && desiredDef.Enabled {
			var added []string
			for entry := range desiredSet {
				if _, ok := currentSet[entry]; !ok {
					added = append(added, entry)
				}
			}
			sort.Strings(added)

			for _, entry := range added {
				addCmd := fmt.Sprintf(`add list=%s address=%s comment="%s"`, listName, entry, managedComment)
				*target = append(*target, addCmd)
			}
		}
	}

	return joinSections(ipv4Lines, ipv6Lines)
}

// unionListNames 返回两个 definition map 的 key 并集。
// 这么做是为了保证：
// - desired 中有的 list 会被处理
// - current 中有但 desired 已删除的 list 也会被处理
func unionListNames(a, b map[string]ListDefinition) []string {
	set := make(map[string]struct{})

	for name := range a {
		set[name] = struct{}{}
	}
	for name := range b {
		set[name] = struct{}{}
	}

	var out []string
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)

	return out
}

// sliceToSet 将切片转换为集合形式。
// 主要用于 diff 模式做集合比较。
func sliceToSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		out[item] = struct{}{}
	}
	return out
}

// joinSections 将 IPv4 与 IPv6 两部分脚本拼接成最终结果。
// 输出规则：
// 1. 如果有 IPv4 命令，则先输出 /ip firewall address-list 段
// 2. 如果有 IPv6 命令，则再输出 /ipv6 firewall address-list 段
// 3. 中间空一行，提高可读性
func joinSections(ipv4Lines, ipv6Lines []string) string {
	var b strings.Builder

	if len(ipv4Lines) > 0 {
		b.WriteString("/ip firewall address-list\n")
		for _, line := range ipv4Lines {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}

	if len(ipv6Lines) > 0 {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("/ipv6 firewall address-list\n")
		for _, line := range ipv6Lines {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}

	return b.String()
}
