package app

import (
	"fmt"
	"sort"
	"strings"
)

func RenderScript(desired Snapshot, current Snapshot, mode RenderMode, managedComment string) (string, error) {
	switch mode {
	case RenderModeReplaceAll:
		return renderReplaceAll(desired, current, managedComment), nil
	case RenderModeDiff:
		return renderDiff(desired, current, managedComment), nil
	default:
		return "", fmt.Errorf("未知渲染模式：%s", mode)
	}
}

func renderReplaceAll(desired Snapshot, current Snapshot, managedComment string) string {
	var ipv4Lines []string
	var ipv6Lines []string

	for _, listName := range unionListNames(desired.Definitions, current.Definitions) {
		def, ok := desired.Definitions[listName]
		if !ok {
			def = current.Definitions[listName]
		}

		removeCmd := fmt.Sprintf(
			`remove [find where list="%s" comment=%s]`,
			listName,
			quoteRouterOSString(managedComment),
		)

		target := &ipv4Lines
		if def.Family == FamilyIPv6 {
			target = &ipv6Lines
		}
		*target = append(*target, removeCmd)

		desiredDef, existsInDesired := desired.Definitions[listName]
		if !existsInDesired || !desiredDef.Enabled {
			continue
		}

		for _, entry := range desired.Entries[listName] {
			*target = append(*target, fmt.Sprintf(
				`add list=%s address=%s comment=%s`,
				listName,
				entry,
				quoteRouterOSString(managedComment),
			))
		}
	}

	return joinSections(ipv4Lines, ipv6Lines)
}

func renderDiff(desired Snapshot, current Snapshot, managedComment string) string {
	var ipv4Lines []string
	var ipv6Lines []string

	for _, listName := range unionListNames(desired.Definitions, current.Definitions) {
		def, ok := desired.Definitions[listName]
		if !ok {
			def = current.Definitions[listName]
		}

		target := &ipv4Lines
		if def.Family == FamilyIPv6 {
			target = &ipv6Lines
		}

		currentSet := sliceToSet(current.Entries[listName])
		desiredSet := make(map[string]struct{})
		desiredDef, existsInDesired := desired.Definitions[listName]
		if existsInDesired && desiredDef.Enabled {
			desiredSet = sliceToSet(desired.Entries[listName])
		}

		var removed []string
		for entry := range currentSet {
			if _, ok := desiredSet[entry]; !ok {
				removed = append(removed, entry)
			}
		}
		sort.Strings(removed)
		for _, entry := range removed {
			*target = append(*target, fmt.Sprintf(
				`remove [find where list="%s" address="%s" comment=%s]`,
				listName,
				entry,
				quoteRouterOSString(managedComment),
			))
		}

		if existsInDesired && desiredDef.Enabled {
			var added []string
			for entry := range desiredSet {
				if _, ok := currentSet[entry]; !ok {
					added = append(added, entry)
				}
			}
			sort.Strings(added)
			for _, entry := range added {
				*target = append(*target, fmt.Sprintf(
					`add list=%s address=%s comment=%s`,
					listName,
					entry,
					quoteRouterOSString(managedComment),
				))
			}
		}
	}

	return joinSections(ipv4Lines, ipv6Lines)
}

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

func sliceToSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		out[item] = struct{}{}
	}
	return out
}

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

func quoteRouterOSString(raw string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\r", " ", "\n", " ")
	return `"` + replacer.Replace(raw) + `"`
}
