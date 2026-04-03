package app

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"
)

// ValidationError 用于承载“多个校验错误”。
// 之所以不在发现第一个错误时立刻返回，
// 是因为配置类错误通常成组出现：
// - 一个字段写错
// - 另一个字段也漏了
// - 某个 list 名称还重复了
//
// 如果一次只返回一个错误，用户会进入“修一个错再跑一次”的低效循环。
// 所以这里设计成“聚合错误”，让用户尽量一次修完。
type ValidationError struct {
	// Messages 保存所有错误文本。
	Messages []string
}

// Error 实现 error 接口。
// 这样 ValidationError 就可以像普通 error 一样被返回。
func (e *ValidationError) Error() string {
	return strings.Join(e.Messages, "\n")
}

// add 用于追加一条错误信息。
func (e *ValidationError) add(msg string) {
	e.Messages = append(e.Messages, msg)
}

// hasError 判断当前是否已经收集到错误。
func (e *ValidationError) hasError() bool {
	return len(e.Messages) > 0
}

// listNamePattern 用于限制 address-list 名称格式。
// 这里使用一个偏保守、工程上更容易维护的规则：
// 1. 必须以字母或数字开头
// 2. 后续只能包含字母、数字、下划线、短横线
// 3. 总长度限制在 64 个字符以内
//
// 说明：
// RouterOS 实际可接受的名称可能更宽松，
// 但工程上建议统一收敛，避免后续 source、API、前端、脚本渲染时产生歧义。
var listNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

// ValidateConfig 对整个 AppConfig 做结构层面的合法性校验。
// 当前它负责的内容包括：
// 1. 渲染模式是否合法
// 2. list 定义是否合法
// 3. source 配置是否合法
// 4. manual rule 定义是否合法
//
// 注意：
// 这里暂时只校验“配置自身”是否合法，
// 不会去读取 source 的实际内容。
// source 内容中的地址校验会在后续 source 加载和合并阶段继续复用本文件中的地址校验函数。
func ValidateConfig(cfg AppConfig) error {
	// 先补默认值，保证某些默认字段已经就位。
	cfg.ApplyDefaults()

	var verr ValidationError

	// 用于检测重复 list 名称。
	listSeen := make(map[string]struct{})

	// 用于检测重复规则 ID。
	ruleSeen := make(map[string]struct{})

	// 先校验渲染模式。
	switch cfg.Output.Mode {
	case RenderModeReplaceAll, RenderModeDiff:
		// 合法，什么都不做
	default:
		verr.add(fmt.Sprintf("output.mode 非法：%q，只支持 replace_all 或 diff", cfg.Output.Mode))
	}

	// 校验 managed comment。
	if strings.TrimSpace(cfg.Output.ManagedComment) == "" {
		verr.add("output.managed_comment 不能为空")
	}

	// 校验日志文件路径。
	if strings.TrimSpace(cfg.LogFile) == "" {
		verr.add("log_file 不能为空")
	}

	// 校验输出路径。
	if strings.TrimSpace(cfg.Output.Path) == "" {
		verr.add("output.path 不能为空")
	}

	// 校验服务监听地址。
	if strings.TrimSpace(cfg.Server.Listen) == "" {
		verr.add("server.listen 不能为空")
	}

	// 校验所有 list 定义。
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

	// 统一校验 source 数组。
	validateSources := func(kind string, sources []SourceConfig) {
		for i, src := range sources {
			if strings.TrimSpace(src.Name) == "" {
				verr.add(fmt.Sprintf("%s[%d].name 不能为空", kind, i))
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

			// 这里不强制 priority 必须 > 0。
			// 因为我们允许用户显式写 0，后续仍可以作为一个有效优先级参与比较。
			// 但超时时间如果 <=0，就说明配置不合理。
			if src.Type == "url" && src.TimeoutSeconds <= 0 {
				verr.add(fmt.Sprintf("%s[%d].timeout_seconds 必须大于 0", kind, i))
			}
		}
	}

	validateSources("desired_sources", cfg.DesiredSources)
	validateSources("current_state_sources", cfg.CurrentStateSources)

	// 校验 manual rules。
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

			// 如果不允许自动创建 list，那么手工规则引用的 list 必须已经存在于定义中。
			if !cfg.AutoCreateLists {
				if _, ok := listSeen[listName]; !ok {
					verr.add(fmt.Sprintf("manual_rules[%d].list_name=%q 未在 lists 中定义，且当前 auto_create_lists=false", i, listName))
				}
			}
		}

		switch rule.Action {
		case "add", "remove":
			// 合法
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
// 处理顺序如下：
// 1. 去除首尾空格
// 2. 校验地址或 CIDR 是否合法
// 3. 根据地址族做一致性校验
// 4. 规范化输出
// 5. 去重
// 6. 排序，保证输出稳定
//
// 参数：
// - entries: 原始地址条目
// - family: 期望地址族；如果传入 ipv4，则 IPv6 条目会报错；反之亦然
//
// 返回：
// - 规范化后的地址列表
// - 如果有任何非法项，返回聚合错误
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

		// 去重：如果已经出现过，直接跳过。
		if _, ok := seen[normalized]; ok {
			continue
		}

		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	if verr.hasError() {
		return nil, &verr
	}

	// 做稳定排序，保证后续输出一致。
	// 这对后面的：
	// - git diff
	// - 日志审计
	// - diff 渲染
	// 都很重要。
	sort.Strings(out)

	return out, nil
}

// NormalizeAddress 对单个地址或 CIDR 做校验与规范化。
// 它支持：
// 1. 单个 IPv4
// 2. 单个 IPv6
// 3. IPv4 CIDR
// 4. IPv6 CIDR
//
// 如果 expectFamily 非空，还会额外校验地址族是否匹配。
//
// 返回值：
// 1. 规范化后的字符串
// 2. 实际识别出的地址族
// 3. 错误
func NormalizeAddress(raw string, expectFamily IPFamily) (string, IPFamily, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", "", fmt.Errorf("空地址")
	}

	// 情况一：带斜杠，按 CIDR 处理。
	if strings.Contains(s, "/") {
		ip, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return "", "", fmt.Errorf("不是合法 CIDR：%w", err)
		}

		var family IPFamily
		if ip.To4() != nil {
			family = FamilyIPv4
		} else {
			family = FamilyIPv6
		}

		// 如果调用方指定了期望地址族，则校验是否匹配。
		if expectFamily != "" && family != expectFamily {
			return "", "", fmt.Errorf("地址族不匹配，期望=%s，实际=%s", expectFamily, family)
		}

		// 为了确保 CIDR 规范化结果稳定，这里显式取网络地址。
		networkIP := ip.Mask(ipNet.Mask)

		ones, bits := ipNet.Mask.Size()

		// IPv4 必须对应 32 位掩码空间。
		if family == FamilyIPv4 {
			if bits != 32 {
				return "", "", fmt.Errorf("不是合法 IPv4 CIDR")
			}

			ip4 := networkIP.To4()
			if ip4 == nil {
				return "", "", fmt.Errorf("CIDR 解析结果不是合法 IPv4")
			}

			return fmt.Sprintf("%s/%d", ip4.String(), ones), family, nil
		}

		// IPv6 必须对应 128 位掩码空间。
		if bits != 128 {
			return "", "", fmt.Errorf("不是合法 IPv6 CIDR")
		}

		return fmt.Sprintf("%s/%d", networkIP.String(), ones), family, nil
	}

	// 情况二：不带斜杠，按单个 IP 处理。
	ip := net.ParseIP(s)
	if ip == nil {
		return "", "", fmt.Errorf("不是合法 IP")
	}

	// 优先判断 IPv4。
	if ip4 := ip.To4(); ip4 != nil {
		if expectFamily != "" && expectFamily != FamilyIPv4 {
			return "", "", fmt.Errorf("地址族不匹配，期望=%s，实际=%s", expectFamily, FamilyIPv4)
		}
		return ip4.String(), FamilyIPv4, nil
	}

	// 否则按 IPv6 处理。
	if expectFamily != "" && expectFamily != FamilyIPv6 {
		return "", "", fmt.Errorf("地址族不匹配，期望=%s，实际=%s", expectFamily, FamilyIPv6)
	}

	return ip.String(), FamilyIPv6, nil
}
