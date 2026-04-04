package app

import (
	"strings"
	"testing"
)

// TestRenderReplaceAll 验证 replace_all 渲染结果中是否包含关键命令。
// 这里不要求逐字符全等，只检查最关键的行为点：
// 1. 是否先 remove
// 2. 是否 add 目标条目
// 3. 是否分出 IPv4 / IPv6 两段
func TestRenderReplaceAll(t *testing.T) {
	desired := Snapshot{
		Definitions: map[string]ListDefinition{
			"toWanTelecom": {
				Name:    "toWanTelecom",
				Family:  FamilyIPv4,
				Enabled: true,
			},
			"toWanIpv6": {
				Name:    "toWanIpv6",
				Family:  FamilyIPv6,
				Enabled: true,
			},
		},
		Entries: map[string][]string{
			"toWanTelecom": {"1.1.1.1"},
			"toWanIpv6":    {"2400:3200::1"},
		},
	}

	current := Snapshot{
		Definitions: map[string]ListDefinition{
			"toWanTelecom": {
				Name:    "toWanTelecom",
				Family:  FamilyIPv4,
				Enabled: true,
			},
			"toWanIpv6": {
				Name:    "toWanIpv6",
				Family:  FamilyIPv6,
				Enabled: true,
			},
		},
		Entries: map[string][]string{},
	}

	got, err := RenderScript(desired, current, RenderModeReplaceAll, "managed-by-go")
	if err != nil {
		t.Fatalf("不期望报错，但实际报错：%v", err)
	}

	checkContains(t, got, `/ip firewall address-list`)
	checkContains(t, got, `remove [find where list="toWanTelecom" comment="managed-by-go"]`)
	checkContains(t, got, `add list=toWanTelecom address=1.1.1.1 comment="managed-by-go"`)

	checkContains(t, got, `/ipv6 firewall address-list`)
	checkContains(t, got, `remove [find where list="toWanIpv6" comment="managed-by-go"]`)
	checkContains(t, got, `add list=toWanIpv6 address=2400:3200::1 comment="managed-by-go"`)
}

// TestRenderDiff 验证 diff 渲染是否只输出必要增删项。
func TestRenderDiff(t *testing.T) {
	desired := Snapshot{
		Definitions: map[string]ListDefinition{
			"toWanGlobal": {
				Name:    "toWanGlobal",
				Family:  FamilyIPv4,
				Enabled: true,
			},
		},
		Entries: map[string][]string{
			"toWanGlobal": {"1.1.1.1"},
		},
	}

	current := Snapshot{
		Definitions: map[string]ListDefinition{
			"toWanGlobal": {
				Name:    "toWanGlobal",
				Family:  FamilyIPv4,
				Enabled: true,
			},
		},
		Entries: map[string][]string{
			"toWanGlobal": {"8.8.8.8"},
		},
	}

	got, err := RenderScript(desired, current, RenderModeDiff, "managed-by-go")
	if err != nil {
		t.Fatalf("不期望报错，但实际报错：%v", err)
	}

	checkContains(t, got, `remove [find where list="toWanGlobal" address="8.8.8.8" comment="managed-by-go"]`)
	checkContains(t, got, `add list=toWanGlobal address=1.1.1.1 comment="managed-by-go"`)
}

// checkContains 是测试里的小工具函数，用于减少重复判断代码。
func checkContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("结果中缺少关键片段：%q\n实际内容：\n%s", sub, s)
	}
}
