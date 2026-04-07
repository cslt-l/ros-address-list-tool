package app

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
)

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

var listNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)
var sourceNamePattern = listNamePattern

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
	if strings.ContainsAny(cfg.Output.ManagedComment, "\"\r\n") {
		verr.add("output.managed_comment 不能包含双引号、回车或换行")
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
	if requiresAuthTokenForListen(cfg.Server.Listen) && strings.TrimSpace(cfg.Server.AuthToken) == "" {
		verr.add("server.listen 暴露到非本机时，server.auth_token 不能为空（也可通过环境变量 ROS_LIST_API_TOKEN 提供）")
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

			srcType := strings.TrimSpace(src.Type)
			switch srcType {
			case "file", "url":
			default:
				verr.add(fmt.Sprintf("%s[%d].type 非法：%q，只支持 file 或 url", kind, i, src.Type))
				continue
			}

			format := normalizeSourceFormat(src.Format)
			switch format {
			case "", "json", "plain_cidr":
			default:
				verr.add(fmt.Sprintf("%s[%d].format 非法：%q，只支持 json 或 plain_cidr", kind, i, src.Format))
			}

			if !src.Enabled {
				continue
			}

			if srcType == "file" && strings.TrimSpace(src.Path) == "" {
				verr.add(fmt.Sprintf("%s[%d] type=file 时必须提供 path", kind, i))
			}
			if srcType == "url" {
				if strings.TrimSpace(src.URL) == "" {
					verr.add(fmt.Sprintf("%s[%d] type=url 时必须提供 url", kind, i))
				} else if err := validateSourceURLString(src.URL); err != nil {
					verr.add(fmt.Sprintf("%s[%d].url 非法：%v", kind, i, err))
				}
				if src.TimeoutSeconds <= 0 {
					verr.add(fmt.Sprintf("%s[%d].timeout_seconds 必须大于 0", kind, i))
				}
			}

			if format == "plain_cidr" {
				if strings.TrimSpace(src.TargetListName) == "" {
					verr.add(fmt.Sprintf("%s[%d] format=plain_cidr 时必须提供 target_list_name", kind, i))
				}
				if src.TargetListFamily != "" && src.TargetListFamily != FamilyIPv4 && src.TargetListFamily != FamilyIPv6 {
					verr.add(fmt.Sprintf("%s[%d].target_list_family 非法：%q", kind, i, src.TargetListFamily))
				}
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
