package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// SourceProbeResult 表示一次 source 连接测试与预检的结果。
// 这个结果不会写回配置文件，仅用于前端页面临时展示。
type SourceProbeResult struct {
	OK bool `json:"ok"`

	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Location string `json:"location,omitempty"`

	HeadersCount int   `json:"headers_count"`
	DurationMS   int64 `json:"duration_ms"`

	// 对于 URL 来源，这里会展示真实 HTTP 状态。
	// 对于 file 来源，这里会展示 LOCAL_FILE。
	StatusCode int    `json:"status_code,omitempty"`
	StatusText string `json:"status_text,omitempty"`

	ContentType string `json:"content_type,omitempty"`
	BodyBytes   int    `json:"body_bytes"`

	JSONValid bool   `json:"json_valid"`
	JSONType  string `json:"json_type,omitempty"`

	DetectedListCount  int      `json:"detected_list_count"`
	DetectedEntryCount int      `json:"detected_entry_count"`
	ListNames          []string `json:"list_names,omitempty"`

	// 仅针对 URL 请求返回的响应头做预览。
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`

	RawPreview string   `json:"raw_preview,omitempty"`
	Error      string   `json:"error,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

// ProbeSource 对单个 source 做“连接测试 + 内容预检”。
// 注意：
// 1. 不会修改配置文件
// 2. 不会加入真正的合并流程
// 3. 这里只做“单次即时检查”
func ProbeSource(src SourceConfig) SourceProbeResult {
	start := time.Now()

	result := SourceProbeResult{
		Name:         src.Name,
		Type:         src.Type,
		HeadersCount: len(src.Headers),
		Warnings:     []string{},
	}

	defer func() {
		result.DurationMS = time.Since(start).Milliseconds()
	}()

	switch src.Type {
	case "file":
		result.Location = src.Path
		return probeFileSource(src, result)

	case "url":
		result.Location = src.URL
		return probeURLSource(src, result)

	default:
		result.OK = false
		result.Error = fmt.Sprintf("不支持的 source type: %s", src.Type)
		return result
	}
}

func probeFileSource(src SourceConfig, result SourceProbeResult) SourceProbeResult {
	if strings.TrimSpace(src.Path) == "" {
		result.OK = false
		result.Error = "file 类型 source 缺少 path"
		return result
	}

	body, err := os.ReadFile(src.Path)
	if err != nil {
		result.OK = false
		result.Error = fmt.Sprintf("读取文件失败: %v", err)
		return result
	}

	result.StatusText = "LOCAL_FILE"
	result.BodyBytes = len(body)
	result.RawPreview = buildRawPreview(body)

	fillJSONPreview(&result, body)
	result.OK = true
	return result
}

func probeURLSource(src SourceConfig, result SourceProbeResult) SourceProbeResult {
	if strings.TrimSpace(src.URL) == "" {
		result.OK = false
		result.Error = "url 类型 source 缺少 url"
		return result
	}

	timeout := time.Duration(src.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := buildProbeHTTPRequest(src)
	if err != nil {
		result.OK = false
		result.Error = fmt.Sprintf("构造请求失败: %v", err)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		result.OK = false
		result.Error = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.StatusText = resp.Status
	result.ContentType = strings.TrimSpace(resp.Header.Get("Content-Type"))
	result.ResponseHeaders = flattenResponseHeaders(resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.OK = false
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	result.BodyBytes = len(body)
	result.RawPreview = buildRawPreview(body)

	// 即使不是 2xx，也尽量继续做 body/JSON 预览，方便前端排错。
	fillJSONPreview(&result, body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.OK = false
		result.Error = fmt.Sprintf("返回非 2xx 状态码: %s", resp.Status)
		return result
	}

	result.OK = true
	return result
}

func buildProbeHTTPRequest(src SourceConfig) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, src.URL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range src.Headers {
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		if key == "" {
			continue
		}
		req.Header.Set(key, value)
	}

	return req, nil
}

// fillJSONPreview 尝试把 body 当作 JSON 解析，并做最基础的 source 结构预检。
// 这里不依赖你现有的正式 loader，仅做“轻量预判”。
func fillJSONPreview(result *SourceProbeResult, body []byte) {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		result.JSONValid = false
		result.Warnings = append(result.Warnings, "响应内容不是合法 JSON")
		return
	}

	result.JSONValid = true
	result.JSONType = detectJSONType(data)

	// 尝试识别常见结构：
	// {
	//   "lists": [
	//     {"name":"toWanTelecom","entries":["1.1.1.1","223.5.5.5"]},
	//     ...
	//   ]
	// }
	switch root := data.(type) {
	case map[string]any:
		listsValue, ok := root["lists"]
		if !ok {
			result.Warnings = append(result.Warnings, "JSON 中未检测到 lists 字段")
			return
		}

		lists, ok := listsValue.([]any)
		if !ok {
			result.Warnings = append(result.Warnings, "lists 字段存在，但不是数组")
			return
		}

		result.DetectedListCount = len(lists)

		listNames := make([]string, 0, 10)
		entryCount := 0

		for _, item := range lists {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}

			if name, ok := obj["name"].(string); ok && name != "" {
				if len(listNames) < 10 {
					listNames = append(listNames, name)
				}
			}

			if entries, ok := obj["entries"].([]any); ok {
				entryCount += len(entries)
			}
		}

		result.DetectedEntryCount = entryCount
		result.ListNames = listNames

	default:
		result.Warnings = append(result.Warnings, "JSON 是合法的，但不是预期的对象结构")
	}
}

func detectJSONType(v any) string {
	switch v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

func buildRawPreview(body []byte) string {
	// 预览最多取前 1200 个字符，避免页面太大。
	const maxChars = 1200

	text := string(body)
	if !utf8.ValidString(text) {
		return "[响应不是有效 UTF-8 文本，无法直接预览]"
	}

	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}

	return string(runes[:maxChars]) + "\n\n[内容过长，已截断预览]"
}

func flattenResponseHeaders(h http.Header) map[string]string {
	if len(h) == 0 {
		return nil
	}

	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string]string, len(keys))
	for _, k := range keys {
		values := h.Values(k)
		out[k] = strings.Join(values, "; ")
	}
	return out
}
