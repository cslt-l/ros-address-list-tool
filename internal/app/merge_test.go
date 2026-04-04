package app

import (
	"reflect"
	"testing"
)

// TestApplyDecision 验证优先级覆盖逻辑。
// 这个测试很重要，因为“谁覆盖谁”是整个项目最核心的业务规则之一。
func TestApplyDecision(t *testing.T) {
	decisions := make(map[string]map[string]ruleDecision)

	// 先写一个低优先级 add
	applyDecision(decisions, "toWanGlobal", "8.8.8.8", ruleDecision{
		Action:   "add",
		Priority: 100,
		Source:   "source:a",
	})

	// 再写一个更高优先级 remove，应覆盖前者
	applyDecision(decisions, "toWanGlobal", "8.8.8.8", ruleDecision{
		Action:   "remove",
		Priority: 200,
		Source:   "manual:b",
	})

	got := decisions["toWanGlobal"]["8.8.8.8"]
	if got.Action != "remove" || got.Priority != 200 {
		t.Fatalf("优先级覆盖失败，got=%+v", got)
	}
}

// TestBuildDesiredSnapshot 做一个最小化的快照构建测试。
// 这里不依赖文件，而是直接构造配置。
// 目的是验证：
// 1. 手工规则能覆盖来源数据
// 2. 最终快照中条目是收敛后的结果
func TestBuildDesiredSnapshot(t *testing.T) {
	cfg := AppConfig{
		AutoCreateLists: true,
		Lists: []ListDefinition{
			{
				Name:        "toWanGlobal",
				Family:      FamilyIPv4,
				Enabled:     true,
				Description: "国际出口地址列表",
			},
		},
		// 这里不走文件 source，后续有文件加载测试，这里只专注测试 merge 层核心规则。
		DesiredSources: []SourceConfig{},
		ManualRules: []ManualRule{
			{
				ID:       "r1",
				ListName: "toWanGlobal",
				Action:   "add",
				Priority: 100,
				Enabled:  true,
				Entries:  []string{"8.8.8.8"},
			},
			{
				ID:       "r2",
				ListName: "toWanGlobal",
				Action:   "remove",
				Priority: 200,
				Enabled:  true,
				Entries:  []string{"8.8.8.8"},
			},
			{
				ID:       "r3",
				ListName: "toWanGlobal",
				Action:   "add",
				Priority: 300,
				Enabled:  true,
				Entries:  []string{"1.1.1.1"},
			},
		},
		Output: OutputConfig{
			Path:           "./output/test.rsc",
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{
			Listen: ":8090",
		},
		LogFile: "./logs/test.log",
	}

	got, err := BuildDesiredSnapshot(cfg)
	if err != nil {
		t.Fatalf("不期望报错，但实际报错：%v", err)
	}

	want := []string{"1.1.1.1"}

	if !reflect.DeepEqual(got.Entries["toWanGlobal"], want) {
		t.Fatalf("最终快照条目不正确，got=%v want=%v", got.Entries["toWanGlobal"], want)
	}
}
