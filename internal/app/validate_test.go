package app

import (
	"reflect"
	"testing"
)

// TestNormalizeAddress 用于验证单个地址或 CIDR 的规范化逻辑。
// 这一层很关键，因为后续：
// - source 加载
// - 手工规则
// - 合并引擎
// 都依赖它判断地址是否合法、是否需要规范化。
func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectFamily IPFamily
		want         string
		wantFamily   IPFamily
		wantErr      bool
	}{
		{
			name:         "IPv4 单地址",
			input:        "1.1.1.1",
			expectFamily: FamilyIPv4,
			want:         "1.1.1.1",
			wantFamily:   FamilyIPv4,
			wantErr:      false,
		},
		{
			name:         "IPv4 CIDR 需要规范化",
			input:        "10.0.0.1/24",
			expectFamily: FamilyIPv4,
			want:         "10.0.0.0/24",
			wantFamily:   FamilyIPv4,
			wantErr:      false,
		},
		{
			name:         "IPv6 单地址",
			input:        "2400:3200::1",
			expectFamily: FamilyIPv6,
			want:         "2400:3200::1",
			wantFamily:   FamilyIPv6,
			wantErr:      false,
		},
		{
			name:         "地址族不匹配",
			input:        "1.1.1.1",
			expectFamily: FamilyIPv6,
			wantErr:      true,
		},
		{
			name:         "非法地址",
			input:        "not-an-ip",
			expectFamily: FamilyIPv4,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, gotFamily, err := NormalizeAddress(tt.input, tt.expectFamily)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望报错，但实际没有报错")
				}
				return
			}

			if err != nil {
				t.Fatalf("不期望报错，但实际报错：%v", err)
			}

			if got != tt.want {
				t.Fatalf("规范化结果不正确，got=%q want=%q", got, tt.want)
			}

			if gotFamily != tt.wantFamily {
				t.Fatalf("地址族不正确，got=%q want=%q", gotFamily, tt.wantFamily)
			}
		})
	}
}

// TestNormalizeAndDeduplicateEntries 用于验证：
// 1. 规范化
// 2. 去重
// 3. 排序
//
// 这是后续渲染输出稳定性的关键前提。
func TestNormalizeAndDeduplicateEntries(t *testing.T) {
	input := []string{
		"10.0.0.1/24",
		"10.0.0.0/24",
		"223.5.5.5",
		"1.1.1.1",
		"1.1.1.1",
	}

	got, err := NormalizeAndDeduplicateEntries(input, FamilyIPv4)
	if err != nil {
		t.Fatalf("不期望报错，但实际报错：%v", err)
	}

	want := []string{
		"1.1.1.1",
		"10.0.0.0/24",
		"223.5.5.5",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("结果不符合预期，got=%v want=%v", got, want)
	}
}
