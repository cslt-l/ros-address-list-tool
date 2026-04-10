package app

import (
	"fmt"
	"net"
	neturl "net/url"
	"path/filepath"
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
var sourceNamePattern = listNamePattern

func normalizeSourcePathForDuplicate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = filepath.Clean(raw)
	raw = strings.ReplaceAll(raw, "\\", "/")
	return raw
}

func normalizeSourceURLForDuplicate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	u, err := neturl.Parse(raw)
	if err != nil {
		return raw
	}

	u.Fragment = ""
	u.Host = strings.ToLower(u.Host)
	return u.String()
}

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

	if requiresAPIToken(cfg.Server.Listen) && strings.TrimSpace(cfg.Server.AuthToken) == "" && !cfg.Server.LoginEnabled {
		verr.add(fmt.Sprintf(
			"server.listen=%q 不是本机回环地址，必须配置 server.auth_token，或启用网页登录 login_enabled",
			cfg.Server.Listen,
		))
	}

	if cfg.Server.LoginEnabled {
		if strings.TrimSpace(cfg.Server.LoginUsername) == "" {
			verr.add("server.login_username 不能为空")
		}
		if cfg.Server.SessionTTLMinutes <= 0 {
			verr.add("server.session_ttl_minutes 必须大于 0")
		}
		if strings.TrimSpace(cfg.Server.SessionCookieName) == "" {
			verr.add("server.session_cookie_name 不能为空")
		}
		if strings.ContainsAny(cfg.Server.SessionCookieName, " =;\t\r\n") {
			verr.add("server.session_cookie_name 含有非法字符")
		}
		if strings.TrimSpace(cfg.Server.LoginPasswordHash) != "" && !strings.HasPrefix(strings.TrimSpace(cfg.Server.LoginPasswordHash), passwordHashPrefix+"$") {
			verr.add("server.login_password_hash 格式非法")
		}
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
		filePathSeen := make(map[string]int)
		urlSeen := make(map[string]int)

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

			// 禁用的 source 允许暂时不填 path/url，不继续做位置强校验
			if !src.Enabled {
				continue
			}

			if srcType == "file" {
				rawPath := strings.TrimSpace(src.Path)
				if rawPath == "" {
					verr.add(fmt.Sprintf("%s[%d] type=file 时必须提供 path", kind, i))
				} else {
					normPath := normalizeSourcePathForDuplicate(rawPath)
					if prev, ok := filePathSeen[normPath]; ok {
						verr.add(fmt.Sprintf("%s[%d].path 重复：%q（与 %s[%d] 重复）", kind, i, rawPath, kind, prev))
					} else {
						filePathSeen[normPath] = i
					}
				}
			}

			if srcType == "url" {
				rawURL := strings.TrimSpace(src.URL)
				if rawURL == "" {
					verr.add(fmt.Sprintf("%s[%d] type=url 时必须提供 url", kind, i))
				} else {
					parsed, err := neturl.Parse(rawURL)
					if err != nil || parsed.Scheme == "" || parsed.Host == "" {
						verr.add(fmt.Sprintf("%s[%d].url 非法：%q", kind, i, rawURL))
					} else {
						scheme := strings.ToLower(parsed.Scheme)
						if scheme != "http" && scheme != "https" {
							verr.add(fmt.Sprintf("%s[%d].url 非法：仅支持 http/https，当前为 %q", kind, i, parsed.Scheme))
						}
					}

					normURL := normalizeSourceURLForDuplicate(rawURL)
					if prev, ok := urlSeen[normURL]; ok {
						verr.add(fmt.Sprintf("%s[%d].url 重复：%q（与 %s[%d] 重复）", kind, i, rawURL, kind, prev))
					} else {
						urlSeen[normURL] = i
					}
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

	// diff 模式必须显式提供 current_state_sources。
	if cfg.Output.Mode == RenderModeDiff && len(cfg.CurrentStateSources) == 0 {
		verr.add("output.mode=diff 时必须提供 current_state_sources")
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
