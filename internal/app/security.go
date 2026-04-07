package app

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxSourceReadBytes int64 = 10 << 20 // 10 MiB

var (
	sharedCGNATRange = mustParseCIDR("100.64.0.0/10")
	benchmarkRange4  = mustParseCIDR("198.18.0.0/15")
	benchmarkRange6  = mustParseCIDR("2001:2::/48")
)

func mustParseCIDR(raw string) *net.IPNet {
	_, network, err := net.ParseCIDR(raw)
	if err != nil {
		panic(err)
	}
	return network
}

func buildSourceHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	dialer := &net.Dialer{Timeout: timeout}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := splitHostPort(addr)
		if err != nil {
			return nil, err
		}

		dialAddr, err := resolveAllowedDialAddress(ctx, host, port)
		if err != nil {
			return nil, err
		}
		return dialer.DialContext(ctx, network, dialAddr)
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("redirect too many times")
			}
			return validateSourceURLString(req.URL.String())
		},
	}
}

func splitHostPort(addr string) (string, string, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err == nil {
		return host, port, nil
	}
	return "", "", fmt.Errorf("解析目标地址失败: %w", err)
}

func resolveAllowedDialAddress(ctx context.Context, host, port string) (string, error) {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" || port == "" {
		return "", fmt.Errorf("无效的目标地址")
	}

	if ip := net.ParseIP(host); ip != nil {
		if isForbiddenSourceIP(ip) {
			return "", fmt.Errorf("禁止访问内网或保留地址: %s", ip.String())
		}
		return net.JoinHostPort(ip.String(), port), nil
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return "", fmt.Errorf("解析 source 主机失败: %w", err)
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("source 主机未解析到任何 IP")
	}

	for _, item := range ips {
		if item.IP == nil || isForbiddenSourceIP(item.IP) {
			continue
		}
		return net.JoinHostPort(item.IP.String(), port), nil
	}
	return "", fmt.Errorf("source 主机仅解析到内网或保留地址")
}

func validateSourceURLString(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("只允许 http 或 https")
	}
	if parsed.User != nil {
		return fmt.Errorf("url 不允许包含 userinfo")
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("url 缺少 host")
	}
	if ip := net.ParseIP(host); ip != nil && isForbiddenSourceIP(ip) {
		return fmt.Errorf("禁止访问内网或保留地址")
	}
	return nil
}

func isForbiddenSourceIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsMulticast() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if sharedCGNATRange.Contains(ip) || benchmarkRange4.Contains(ip) || benchmarkRange6.Contains(ip) {
		return true
	}
	return false
}

func readBodyWithLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("内容超过大小限制（%d bytes）", maxBytes)
	}
	return body, nil
}

func resolveRenderOutputPath(configuredPath, requestedPath string) (string, error) {
	configuredPath = strings.TrimSpace(configuredPath)
	if configuredPath == "" {
		return "", fmt.Errorf("output.path 不能为空")
	}

	configuredAbs, err := filepath.Abs(configuredPath)
	if err != nil {
		return "", fmt.Errorf("解析 output.path 失败: %w", err)
	}
	baseDir := filepath.Dir(configuredAbs)
	target := configuredAbs
	if strings.TrimSpace(requestedPath) != "" {
		target, err = filepath.Abs(strings.TrimSpace(requestedPath))
		if err != nil {
			return "", fmt.Errorf("解析 output_path 失败: %w", err)
		}
	}

	rel, err := filepath.Rel(baseDir, target)
	if err != nil {
		return "", fmt.Errorf("校验输出路径失败: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("output_path 必须位于已配置输出目录下：%s", baseDir)
	}
	if !strings.EqualFold(filepath.Ext(target), ".rsc") {
		return "", fmt.Errorf("output_path 只允许写入 .rsc 文件")
	}
	return target, nil
}

func extractRequestToken(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(strings.ToLower(authorization), strings.ToLower(prefix)) {
			return strings.TrimSpace(authorization[len(prefix):])
		}
	}
	return strings.TrimSpace(r.Header.Get("X-API-Token"))
}

func isAuthorizedAPIRequest(r *http.Request, configuredToken string) bool {
	configuredToken = strings.TrimSpace(configuredToken)
	if configuredToken == "" {
		return isLoopbackRequest(r)
	}
	provided := extractRequestToken(r)
	if provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(configuredToken)) == 1
}

func requiresAuthTokenForListen(addr string) bool {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return false
	}

	host := addr
	if strings.HasPrefix(host, ":") {
		return true
	}
	if strings.Contains(host, ":") {
		parsedHost, _, err := net.SplitHostPort(host)
		if err == nil {
			host = parsedHost
		}
	}
	host = strings.Trim(host, "[]")
	if host == "" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return true
	}
	return !ip.IsLoopback()
}

func sanitizeConfigForAPI(cfg AppConfig) AppConfig {
	cfg.Server.AuthToken = ""
	return cfg
}
