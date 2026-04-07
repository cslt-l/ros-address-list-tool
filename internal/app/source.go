package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// SourceProbeResult 表示 Source 测试结果。
// 这个结构体就是 Sources 页面“Source 测试结果”板块的数据来源。
type SourceProbeResult struct {
	OK                   bool              `json:"ok"`
	Name                 string            `json:"name,omitempty"`
	Location             string            `json:"location,omitempty"`
	StatusText           string            `json:"status_text,omitempty"`
	ContentType          string            `json:"content_type,omitempty"`
	BodyBytes            int               `json:"body_bytes"`
	JSONValid            bool              `json:"json_valid"`
	JSONType             string            `json:"json_type,omitempty"`
	DetectedFormat       string            `json:"detected_format,omitempty"`
	FormatDetail         string            `json:"format_detail,omitempty"`
	DetectedListCount    int               `json:"detected_list_count"`
	DetectedEntryCount   int               `json:"detected_entry_count"`
	TextLineCount        int               `json:"text_line_count"`
	TextValidLineCount   int               `json:"text_valid_line_count"`
	TextInvalidLineCount int               `json:"text_invalid_line_count"`
	HeadersCount         int               `json:"headers_count"`
	DurationMS           int64             `json:"duration_ms"`
	Error                string            `json:"error,omitempty"`
	Warnings             []string          `json:"warnings,omitempty"`
	ListNames            []string          `json:"list_names,omitempty"`
	ResponseHeaders      map[string]string `json:"response_headers,omitempty"`
	RawPreview           string            `json:"raw_preview,omitempty"`
}

type probeFetchMeta struct {
	StatusText      string
	ContentType     string
	ResponseHeaders map[string]string
}

type PlainCIDRProbeStats struct {
	LineCount        int
	ValidLineCount   int
	InvalidLineCount int
}

// LoadSources 用于按顺序加载多个来源。
func LoadSources(sources []SourceConfig) ([]LoadedSource, error) {
	var out []LoadedSource

	for i, src := range sources {
		if !src.Enabled {
			continue
		}

		data, err := loadSourceData(src)
		if err != nil {
			return nil, fmt.Errorf("加载 source 失败（index=%d, name=%s）: %w", i, src.Name, err)
		}

		lists, err := ParseSourcePayloadForSource(src, data)
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

// ProbeSource 用于 Sources 页面中的“测试 Source”。
// 这一步不落盘，只做即时探测。
func ProbeSource(src SourceConfig) SourceProbeResult {
	start := time.Now()

	result := SourceProbeResult{
		Name:     strings.TrimSpace(src.Name),
		Location: sourceLocation(src),
	}

	data, meta, err := loadSourceDataForProbe(src)
	result.DurationMS = time.Since(start).Milliseconds()

	if meta != nil {
		result.StatusText = meta.StatusText
		result.ContentType = meta.ContentType
		result.ResponseHeaders = meta.ResponseHeaders
		result.HeadersCount = len(meta.ResponseHeaders)
	}

	result.BodyBytes = len(data)
	result.RawPreview = buildRawPreview(data)

	jsonValid, jsonType := inspectJSON(data)
	result.JSONValid = jsonValid
	result.JSONType = jsonType

	if err != nil {
		result.OK = false
		result.Error = err.Error()
		return result
	}

	var (
		lists          []SourceList
		warnings       []string
		detectedFormat string
		formatDetail   string
	)

	// 测试阶段不盲信 format 配置，而是优先根据实际响应内容识别。
	configuredFormat := normalizeSourceFormat(src.Format)

	if parsed, jsonErr := ParseSourcePayload(data); jsonErr == nil {
		lists = parsed
		detectedFormat = "json"
		formatDetail = "响应内容已识别为 JSON"

		if configuredFormat != "" && configuredFormat != "json" {
			warnings = append(warnings, fmt.Sprintf("当前配置 format=%s，但响应内容更像 json。", configuredFormat))
		}
	} else {
		parsed, probeWarnings, stats, plainErr := ParsePlainCIDRPayloadForProbe(src, data)
		if plainErr == nil {
			lists = parsed
			warnings = append(warnings, probeWarnings...)
			detectedFormat = "plain_cidr"
			formatDetail = "响应内容已识别为纯文本 CIDR/IP 列表"

			result.TextLineCount = stats.LineCount
			result.TextValidLineCount = stats.ValidLineCount
			result.TextInvalidLineCount = stats.InvalidLineCount

			if configuredFormat != "" && configuredFormat != "plain_cidr" {
				warnings = append(warnings, fmt.Sprintf("当前配置 format=%s，但响应内容更像 plain_cidr。", configuredFormat))
			}
		} else {
			detectedFormat = "unknown"
			formatDetail = "无法识别为 json 或 plain_cidr"
			err = fmt.Errorf("JSON 解析失败：%v；plain_cidr 解析失败：%v", jsonErr, plainErr)
		}
	}

	result.DetectedFormat = detectedFormat
	result.FormatDetail = formatDetail
	result.Warnings = warnings

	if err != nil {
		result.OK = false
		result.Error = err.Error()
		return result
	}

	result.OK = true
	result.DetectedListCount = len(lists)

	totalEntries := 0
	var listNames []string
	for _, item := range lists {
		totalEntries += len(item.Entries)
		if strings.TrimSpace(item.Name) != "" {
			listNames = append(listNames, item.Name)
		}
	}
	sort.Strings(listNames)

	result.DetectedEntryCount = totalEntries
	result.ListNames = listNames

	return result
}

// loadSourceData 根据 source 类型读取原始字节数据。
func loadSourceData(src SourceConfig) ([]byte, error) {
	data, _, err := loadSourceDataWithMeta(src)
	return data, err
}

// loadSourceDataForProbe 读取 source 原始字节，同时返回更适合测试展示的元信息。
func loadSourceDataForProbe(src SourceConfig) ([]byte, *probeFetchMeta, error) {
	return loadSourceDataWithMeta(src)
}

func loadSourceDataWithMeta(src SourceConfig) ([]byte, *probeFetchMeta, error) {
	switch src.Type {
	case "file":
		body, err := os.ReadFile(src.Path)
		if err != nil {
			return nil, nil, err
		}

		contentType := ""
		if len(body) > 0 {
			sample := body
			if len(sample) > 512 {
				sample = sample[:512]
			}
			contentType = http.DetectContentType(sample)
		}

		return body, &probeFetchMeta{
			StatusText:      "FILE OK",
			ContentType:     contentType,
			ResponseHeaders: map[string]string{},
		}, nil

	case "url":
		timeout := time.Duration(src.TimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 15 * time.Second
		}

		client := &http.Client{
			Timeout: timeout,
		}

		req, err := buildHTTPRequestForSource(src)
		if err != nil {
			return nil, nil, fmt.Errorf("构造 URL 请求失败，source=%s: %w", src.Name, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("请求 URL 来源失败，source=%s: %w", src.Name, err)
		}
		defer resp.Body.Close()

		headers := make(map[string]string)
		for k, values := range resp.Header {
			if len(values) == 0 {
				continue
			}
			headers[k] = strings.Join(values, ", ")
		}

		meta := &probeFetchMeta{
			StatusText:      resp.Status,
			ContentType:     resp.Header.Get("Content-Type"),
			ResponseHeaders: headers,
		}

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, meta, fmt.Errorf("读取 URL 来源响应失败，source=%s: %w", src.Name, readErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return body, meta, fmt.Errorf("URL 来源返回非 2xx，source=%s status=%s", src.Name, resp.Status)
		}

		return body, meta, nil

	default:
		return nil, nil, fmt.Errorf("不支持的 source.type: %s", src.Type)
	}
}

func buildHTTPRequestForSource(src SourceConfig) (*http.Request, error) {
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

// ParseSourcePayloadForSource 根据 source 配置决定如何解析原始内容。
func ParseSourcePayloadForSource(src SourceConfig, data []byte) ([]SourceList, error) {
	format := normalizeSourceFormat(src.Format)

	switch format {
	case "json":
		return ParseSourcePayload(data)

	case "plain_cidr":
		return ParsePlainCIDRPayload(src, data)

	case "":
		// 未显式指定 format 时，先尝试 JSON。
		if lists, err := ParseSourcePayload(data); err == nil {
			return lists, nil
		}

		// 只有 target_list_name 已经明确时，才允许自动按 plain_cidr 正式落地成 SourceList。
		if strings.TrimSpace(src.TargetListName) != "" {
			if lists, err := ParsePlainCIDRPayload(src, data); err == nil {
				return lists, nil
			}
		}

		return nil, fmt.Errorf("无法识别 source 内容格式：既不是合法 JSON，也不能按 plain_cidr 正式落地")

	default:
		return nil, fmt.Errorf("不支持的 source.format: %s", src.Format)
	}
}

func normalizeSourceFormat(format string) string {
	return strings.ToLower(strings.TrimSpace(format))
}

// ParseSourcePayload 将原始 JSON 数据解析为统一的 []SourceList。
// 当前支持三种 JSON 格式：
// 1. 标准结构：{"lists":[...]}
// 2. 简写 map：{"toWanTelecom":["1.1.1.1"]}
// 3. 根数组：[{"name":"toWanTelecom","entries":[...]}]
func ParseSourcePayload(data []byte) ([]SourceList, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("source 数据为空")
	}

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

	{
		var arr []SourceList
		if err := json.Unmarshal(data, &arr); err == nil {
			normalizeSourceLists(arr)
			return arr, nil
		}
	}

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

// ParsePlainCIDRPayload 用于正式加载 plain_cidr 来源。
// 这里要求 target_list_name 必填，因为正式执行时必须知道条目归属哪个 list。
func ParsePlainCIDRPayload(src SourceConfig, data []byte) ([]SourceList, error) {
	targetListName := strings.TrimSpace(src.TargetListName)
	if targetListName == "" {
		return nil, fmt.Errorf("plain_cidr 来源必须配置 target_list_name")
	}

	entries, err := parsePlainCIDREntriesStrict(data, src.LineCommentPrefixes)
	if err != nil {
		return nil, err
	}

	lists := []SourceList{
		{
			Name:        targetListName,
			Entries:     entries,
			Description: strings.TrimSpace(src.Name),
			Family:      src.TargetListFamily,
		},
	}

	normalizeSourceLists(lists)
	return lists, nil
}

// ParsePlainCIDRPayloadForProbe 用于“Source 测试结果”场景。
// 这里允许 target_list_name 为空，因为测试阶段我们只需要知道：
// 1. 内容能不能被识别
// 2. 一共识别出多少条
// 3. 原始内容是不是 plain_cidr 风格
func ParsePlainCIDRPayloadForProbe(src SourceConfig, data []byte) ([]SourceList, []string, PlainCIDRProbeStats, error) {
	entries, warnings, stats, err := parsePlainCIDREntriesForProbe(data, src.LineCommentPrefixes)
	if err != nil {
		return nil, warnings, stats, err
	}

	listName := strings.TrimSpace(src.TargetListName)
	if listName == "" {
		listName = "(plain_cidr:auto)"
		warnings = append(warnings, "当前未填写 target_list_name；测试可识别条目，但正式保存后仍建议补齐目标 list。")
	}

	lists := []SourceList{
		{
			Name:        listName,
			Entries:     entries,
			Description: strings.TrimSpace(src.Name),
			Family:      src.TargetListFamily,
		},
	}

	normalizeSourceLists(lists)
	return lists, warnings, stats, nil
}

func parsePlainCIDREntriesStrict(data []byte, prefixes []string) ([]string, error) {
	prefixes = normalizeCommentPrefixes(prefixes)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var entries []string
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if lineNo == 1 {
			line = strings.TrimPrefix(line, "\uFEFF")
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if hasAnyPrefix(line, prefixes) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		token := strings.TrimSpace(fields[0])
		if token == "" {
			continue
		}

		if !isValidIPOrCIDR(token) {
			return nil, fmt.Errorf("plain_cidr 第 %d 行不是合法 IP/CIDR：%s", lineNo, token)
		}

		entries = append(entries, token)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("扫描 plain_cidr 来源失败: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("plain_cidr 来源没有解析出任何条目")
	}

	return entries, nil
}

func parsePlainCIDREntriesForProbe(data []byte, prefixes []string) ([]string, []string, PlainCIDRProbeStats, error) {
	prefixes = normalizeCommentPrefixes(prefixes)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var entries []string
	var warnings []string
	var stats PlainCIDRProbeStats

	lineNo := 0
	invalidCount := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if lineNo == 1 {
			line = strings.TrimPrefix(line, "\uFEFF")
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if hasAnyPrefix(line, prefixes) {
			continue
		}

		stats.LineCount++

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		token := strings.TrimSpace(fields[0])
		if token == "" {
			continue
		}

		if !isValidIPOrCIDR(token) {
			invalidCount++
			stats.InvalidLineCount++
			if len(warnings) < 8 {
				warnings = append(warnings, fmt.Sprintf("第 %d 行不是合法 IP/CIDR，测试阶段已忽略：%s", lineNo, token))
			}
			continue
		}

		stats.ValidLineCount++
		entries = append(entries, token)
	}

	if err := scanner.Err(); err != nil {
		return nil, warnings, stats, fmt.Errorf("扫描 plain_cidr 来源失败: %w", err)
	}

	if len(entries) == 0 {
		return nil, warnings, stats, fmt.Errorf("未识别出任何有效的 plain_cidr 条目")
	}

	if invalidCount > len(warnings) {
		warnings = append(warnings, fmt.Sprintf("另有 %d 行无效内容已省略。", invalidCount-len(warnings)))
	}

	return entries, warnings, stats, nil
}

func normalizeCommentPrefixes(prefixes []string) []string {
	if len(prefixes) == 0 {
		return []string{"#", "//", ";"}
	}
	return prefixes
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		p := strings.TrimSpace(prefix)
		if p == "" {
			continue
		}
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func isValidIPOrCIDR(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}

	if strings.Contains(token, "/") {
		_, _, err := net.ParseCIDR(token)
		return err == nil
	}

	return net.ParseIP(token) != nil
}

func inspectJSON(data []byte) (bool, string) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false, ""
	}

	var v any
	if err := json.Unmarshal(trimmed, &v); err != nil {
		return false, ""
	}

	switch v.(type) {
	case map[string]any:
		return true, "object"
	case []any:
		return true, "array"
	case string:
		return true, "string"
	case float64:
		return true, "number"
	case bool:
		return true, "boolean"
	case nil:
		return true, "null"
	default:
		return true, "unknown"
	}
}

func sourceLocation(src SourceConfig) string {
	switch src.Type {
	case "file":
		return strings.TrimSpace(src.Path)
	case "url":
		return strings.TrimSpace(src.URL)
	default:
		return ""
	}
}

func buildRawPreview(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	text := string(data)
	text = strings.ReplaceAll(text, "\r\n", "\n")

	const maxRunes = 4000
	runes := []rune(text)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "\n\n...(preview truncated)"
	}
	return text
}

// normalizeSourceLists 对解析后的来源列表做轻量整理。
func normalizeSourceLists(lists []SourceList) {
	for i := range lists {
		lists[i].Name = strings.TrimSpace(lists[i].Name)
		lists[i].Description = strings.TrimSpace(lists[i].Description)
	}

	sort.Slice(lists, func(i, j int) bool {
		return lists[i].Name < lists[j].Name
	})
}
