package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// LoadSources 用于按顺序加载多个来源。
// 它的职责非常单一：
// 1. 遍历配置中的 source
// 2. 跳过未启用的 source
// 3. 读取原始数据
// 4. 解析为统一的 []SourceList
// 5. 返回 []LoadedSource
//
// 注意：
// - 它不做业务合并
// - 它不做最终渲染
// - 它不做日志记录（日志会在后续执行引擎中统一接入）
//
// 这样设计的好处是：
// source 层只负责“取数据”和“整理格式”，
// 后续 merge / render / api 都能复用它。
func LoadSources(sources []SourceConfig) ([]LoadedSource, error) {
	var out []LoadedSource

	for i, src := range sources {
		// 如果 source 被禁用，则直接跳过。
		if !src.Enabled {
			continue
		}

		data, err := loadSourceData(src)
		if err != nil {
			return nil, fmt.Errorf("加载 source 失败（index=%d, name=%s）: %w", i, src.Name, err)
		}

		lists, err := ParseSourcePayload(data)
		if err != nil {
			return nil, fmt.Errorf("解析 source 数据失败（index=%d, name=%s）: %w", i, src.Name, err)
		}

		out = append(out, LoadedSource{
			Source: src,
			Lists:  lists,
		})
	}

	return out, nil
}

// loadSourceData 根据 source 类型读取原始字节数据。
// 当前支持：
// 1. file：从本地文件读取
// 2. url：通过 HTTP GET 获取
//
// 为什么拆成单独函数？
// 因为“读取原始字节”本身是一个独立动作，
// 后续如果要新增：
// - s3
// - github raw
// - 自定义接口
// - 本地目录批量扫描
// 也可以在这一层继续扩展，而不影响解析逻辑。
func loadSourceData(src SourceConfig) ([]byte, error) {
	switch src.Type {
	case "file":
		return os.ReadFile(src.Path)

	case "url":
		timeout := time.Duration(src.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 15 * time.Second
		}

		client := &http.Client{
			Timeout: timeout,
		}

		req, err := http.NewRequest(http.MethodGet, src.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
		}

		// 将配置中的自定义请求头写入请求。
		for k, v := range src.Headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求 URL 失败: %w", err)
		}
		defer resp.Body.Close()

		// 只接受 2xx。
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP 状态码异常: %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取 HTTP 响应体失败: %w", err)
		}

		return data, nil

	default:
		return nil, fmt.Errorf("不支持的 source.type: %s", src.Type)
	}
}

// ParseSourcePayload 将原始 JSON 数据解析为统一的 []SourceList。
// 当前支持三种格式：
//
// 格式一：推荐格式（对象 + lists 字段）
//
//	{
//	  "lists": [
//	    {
//	      "name": "toWanTelecom",
//	      "entries": ["1.1.1.1", "10.0.0.0/24"]
//	    }
//	  ]
//	}
//
// 格式二：简写 map 格式
//
//	{
//	  "toWanTelecom": ["1.1.1.1", "10.0.0.0/24"],
//	  "toWanGlobal": ["8.8.8.8"]
//	}
//
// 格式三：根直接就是数组
// [
//
//	{
//	  "name": "toWanTelecom",
//	  "entries": ["1.1.1.1"]
//	}
//
// ]
//
// 为什么要支持多种格式？
// 因为你明确要求支持多个 JSON 来源，
// 而现实里不同来源的数据格式往往不统一。
// 如果 source 层能在这里统一格式，
// 后续 merge 层就能大幅简化。
func ParseSourcePayload(data []byte) ([]SourceList, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("source 数据为空")
	}

	// 先尝试识别“对象 + lists 字段”的标准结构。
	{
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(data, &probe); err == nil {
			if _, ok := probe["lists"]; ok {
				var payload SourcePayload
				if err := json.Unmarshal(data, &payload); err != nil {
					return nil, fmt.Errorf("解析标准 lists 结构失败: %w", err)
				}
				normalizeSourceLists(payload.Lists)
				return payload.Lists, nil
			}
		}
	}

	// 再尝试识别“根是数组”的结构。
	{
		var arr []SourceList
		if err := json.Unmarshal(data, &arr); err == nil {
			normalizeSourceLists(arr)
			return arr, nil
		}
	}

	// 最后尝试识别“map[string][]string”的简写结构。
	{
		var m map[string][]string
		if err := json.Unmarshal(data, &m); err == nil {
			var lists []SourceList
			for name, entries := range m {
				lists = append(lists, SourceList{
					Name:    strings.TrimSpace(name),
					Entries: entries,
				})
			}
			normalizeSourceLists(lists)
			return lists, nil
		}
	}

	return nil, fmt.Errorf("无法识别的 source JSON 格式")
}

// normalizeSourceLists 对解析后的来源列表做轻量整理。
// 这里故意只做“轻量处理”，不做最终校验：
// 1. 去掉 name 和 description 的首尾空格
// 2. 按 list 名称排序，保证输出稳定
//
// 为什么不在这里做地址合法性校验？
// 因为地址校验属于 validate 层的职责，
// source 层应该只负责“读取”和“格式统一”。
func normalizeSourceLists(lists []SourceList) {
	for i := range lists {
		lists[i].Name = strings.TrimSpace(lists[i].Name)
		lists[i].Description = strings.TrimSpace(lists[i].Description)
	}

	sort.Slice(lists, func(i, j int) bool {
		return lists[i].Name < lists[j].Name
	})
}
