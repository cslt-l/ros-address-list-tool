package app

import (
	"reflect"
	"testing"
)

// TestParseSourcePayload 验证 source 解析器是否真的支持我们设计的三种 JSON 结构。
// 这一步很重要，因为你后面会长期依赖“多个来源、多种格式”的输入能力。
func TestParseSourcePayload(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNames []string
	}{
		{
			name: "标准 lists 结构",
			input: `{
				"lists": [
					{"name": "toWanTelecom", "entries": ["1.1.1.1"]},
					{"name": "toWanGlobal", "entries": ["8.8.8.8"]}
				]
			}`,
			wantNames: []string{"toWanGlobal", "toWanTelecom"},
		},
		{
			name: "简写 map 结构",
			input: `{
				"toWanTelecom": ["1.1.1.1"],
				"toWanGlobal": ["8.8.8.8"]
			}`,
			wantNames: []string{"toWanGlobal", "toWanTelecom"},
		},
		{
			name: "根数组结构",
			input: `[
				{"name": "toWanTelecom", "entries": ["1.1.1.1"]},
				{"name": "toWanGlobal", "entries": ["8.8.8.8"]}
			]`,
			wantNames: []string{"toWanGlobal", "toWanTelecom"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSourcePayload([]byte(tt.input))
			if err != nil {
				t.Fatalf("不期望报错，但实际报错：%v", err)
			}

			var gotNames []string
			for _, item := range got {
				gotNames = append(gotNames, item.Name)
			}

			if !reflect.DeepEqual(gotNames, tt.wantNames) {
				t.Fatalf("解析出的 list 名称不正确，got=%v want=%v", gotNames, tt.wantNames)
			}
		})
	}
}
