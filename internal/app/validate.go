package app

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
)

// ValidationError 用于承载“多个校验错误”。
type ValidationError struct {
	Messages []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Messages, "\n")
}

func (e *ValidationError) add(msg string) {
	e.Messages = append(e.Messages, msg)
}

func (e *ValidationError) hasError() bool {
	return len(e.Messages) > 0
}

// listNamePattern 用于限制 address-list 名称格式。
var listNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

// sourceNamePattern 复用与 list 相同的保守规则。
// 这样可以避免 source 名称把 HTTP 路由路径、前端路由参数和存储主键语义搅乱。
var sourceNamePattern = listNamePattern

// ValidateConfig 对整个 AppConfig 做结构层面的合法性校验。
func ValidateConfig(cfg AppConfig) error {
	cfg.ApplyDefaults()

	var verr ValidationError

	listSeen := make(map[string]struct{})
	ruleSeen := make(map[string]struct{})

	switch cfg.Output.Mode {
	case RenderModeReplaceAll, RenderModeDiff:
	default:
		verr.add(fmt.Sprintf("output.mode 非法：%q，只支持 replace_all 或 diff", cfg.Output.Mode))
	}

	if strings.TrimSpace(cfg.Output.ManagedComment) == "" {
		verr.add("output.managed_comment 不能为空")
	}

	if strings.TrimSpace(cfg.LogFile) == "" {
		verr.add("log_file 不能为空")
	}

	if strings.TrimSpace(cfg.Output.Path) == "" {
		verr.add("output.path 不能为空")
	}

	if strings.TrimSpace(cfg.Server.Listen) == "" {
		verr.add("server.listen 不能为空")
	}

	for i, item := range cfg.Lists {
		name := strings.TrimSpace(item.Name)

		if name == "" {
			verr.add(fmt.Sprintf("lists[%d].name 不能为空", i))
			continue
		}

		if !listNamePattern.MatchString(name) {
			verr.add(fmt.Sprintf("lists[%d].name 非法：%q，只允许字母、数字、下划线、短横线，且需以字母或数字开头", i, name))
		}

		if _, ok := listSeen[name]; ok {
			verr.add(fmt.Sprintf("lists[%d].name 重复：%q", i, name))
		}
		listSeen[name] = struct{}{}

		if item.Family != FamilyIPv4 && item.Family != FamilyIPv6 {
			verr.add(fmt.Sprintf("lists[%d].family 非法：%q，只支持 ipv4 或 ipv6", i, item.Family))
		}
	}

	validateSources := func(kind string, sources []SourceConfig) {
		sourceSeen := make(map[string]struct{})

		for i, src := range sources {
			name := strings.TrimSpace(src.Name)
			if name == "" {
				verr.add(fmt.Sprintf("%s[%d].name 不能为空", kind, i))
			} else {
				if !sourceNamePattern.MatchString(name) {
					verr.add(fmt.Sprintf("%s[%d].name 非法：%q，只允许字母、数字、下划线、短横线，且需以字母或数字开头", kind, i, name))
				}

				if _, ok := sourceSeen[name]; ok {
					verr.add(fmt.Sprintf("%s[%d].name 重复：%q", kind, i, name))
				}
				sourceSeen[name] = struct{}{}
			}

			switch src.Type {
			case "file":
				if strings.TrimSpace(src.Path) == "" {
					verr.add(fmt.Sprintf("%s[%d] type=file 时必须提供 path", kind, i))
				}
			case "url":
				if strings.TrimSpace(src.URL) == "" {
					verr.add(fmt.Sprintf("%s[%d] type=url 时必须提供 url", kind, i))
				}
			default:
				verr.add(fmt.Sprintf("%s[%d].type 非法：%q，只支持 file 或 url", kind, i, src.Type))
			}

			if src.Type == "url" && src.TimeoutSeconds <= 0 {
				verr.add(fmt.Sprintf("%s[%d].timeout_seconds 必须大于 0", kind, i))
			}
		}
	}

	validateSources("desired_sources", cfg.DesiredSources)
	validateSources("current_state_sources", cfg.CurrentStateSources)

	for i, rule := range cfg.ManualRules {
		if strings.TrimSpace(rule.ID) == "" {
			verr.add(fmt.Sprintf("manual_rules[%d].id 不能为空", i))
		} else {
			if _, ok := ruleSeen[rule.ID]; ok {
				verr.add(fmt.Sprintf("manual_rules[%d].id 重复：%q", i, rule.ID))
			}
			ruleSeen[rule.ID] = struct{}{}
		}

		listName := strings.TrimSpace(rule.ListName)
		if listName == "" {
			verr.add(fmt.Sprintf("manual_rules[%d].list_name 不能为空", i))
		} else {
			if !listNamePattern.MatchString(listName) {
				verr.add(fmt.Sprintf("manual_rules[%d].list_name 非法：%q", i, listName))
			}

			if !cfg.AutoCreateLists {
				if _, ok := listSeen[listName]; !ok {
					verr.add(fmt.Sprintf("manual_rules[%d].list_name=%q 未在 lists 中定义，且当前 auto_create_lists=false", i, listName))
				}
			}
		}

		switch rule.Action {
		case "add", "remove":
		default:
			verr.add(fmt.Sprintf("manual_rules[%d].action 非法：%q，只支持 add 或 remove", i, rule.Action))
		}
	}

	if verr.hasError() {
		return &verr
	}

	return nil
}

// NormalizeAndDeduplicateEntries 对一组地址条目做统一处理。
func NormalizeAndDeduplicateEntries(entries []string, family IPFamily) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	var verr ValidationError

	for idx, raw := range entries {
		normalized, _, err := NormalizeAddress(raw, family)
		if err != nil {
			verr.add(fmt.Sprintf("entries[%d]=%q 非法：%v", idx, raw, err))
			continue
		}

		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	if verr.hasError() {
		return nil, &verr
	}

	sort.Strings(out)
	return out, nil
}

// NormalizeAddress 对单个地址或 CIDR 做校验与规范化。
func NormalizeAddress(raw string, expectFamily IPFamily) (string, IPFamily, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", "", fmt.Errorf("地址为空")
	}

	if strings.Contains(s, "/") {
		ip, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return "", "", fmt.Errorf("不是合法 CIDR")
		}

		family := detectIPFamily(ip)
		if family == "" {
			return "", "", fmt.Errorf("无法识别地址族")
		}

		if expectFamily != "" && family != expectFamily {
			return "", family, fmt.Errorf("地址族不匹配，期望 %s，实际 %s", expectFamily, family)
		}

		return ipNet.String(), family, nil
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return "", "", fmt.Errorf("不是合法 IP")
	}

	family := detectIPFamily(ip)
	if family == "" {
		return "", "", fmt.Errorf("无法识别地址族")
	}

	if expectFamily != "" && family != expectFamily {
		return "", family, fmt.Errorf("地址族不匹配，期望 %s，实际 %s", expectFamily, family)
	}

	if family == FamilyIPv4 {
		return ip.To4().String(), family, nil
	}
	return ip.String(), family, nil
}

func detectIPFamily(ip net.IP) IPFamily {
	if ip == nil {
		return ""
	}

	if ip.To4() != nil {
		return FamilyIPv4
	}

	if ip.To16() != nil {
		return FamilyIPv6
	}

	return ""
}
