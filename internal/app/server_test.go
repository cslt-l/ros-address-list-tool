package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func newTestHTTPHandler(t *testing.T, cfg AppConfig) http.Handler {
	t.Helper()

	t.Setenv("ROS_LIST_API_TOKEN", "")
	cfg.ApplyDefaults()

	tmpDir := t.TempDir()
	cfg.LogFile = filepath.Join(tmpDir, "app.log")
	cfg.Output.Path = filepath.Join(tmpDir, "routeros-address-list.rsc")

	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}

	store := &ConfigStore{
		path: filepath.Join(tmpDir, "config.json"),
		cfg:  cfg,
	}

	logger := log.New(io.Discard, "", 0)
	return NewHTTPHandler(store, logger)
}

func newSecuredTestConfig() AppConfig {
	return AppConfig{
		AutoCreateLists: false,
		LogFile:         "./logs/app.log",
		Lists: []ListDefinition{
			{
				Name:        "toWanTelecom4",
				Family:      FamilyIPv4,
				Enabled:     true,
				Description: "中国电信 IPv4",
			},
		},
		DesiredSources: []SourceConfig{
			{
				Name:             "china-telecom-ipv4",
				Type:             "url",
				URL:              "https://example.com/chinanet.txt",
				Headers:          map[string]string{"Authorization": "Bearer upstream-secret", "X-Api-Key": "abc123", "User-Agent": "ros-list-test"},
				Format:           "plain_cidr",
				TargetListName:   "toWanTelecom4",
				TargetListFamily: FamilyIPv4,
				Enabled:          true,
				Priority:         100,
				TimeoutSeconds:   30,
			},
		},
		Output: OutputConfig{
			Path:           "./output/routeros-address-list.rsc",
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{
			Listen:    ":9000",
			EnableWeb: false,
			WebDir:    "./web/src",
			AuthToken: "test-token",
		},
	}
}

func TestHealthzAllowsAnonymousAccess(t *testing.T) {
	handler := newTestHTTPHandler(t, newSecuredTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %#v", body)
	}
}

func TestConfigRequiresBearerToken(t *testing.T) {
	handler := newTestHTTPHandler(t, newSecuredTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["error"] != "unauthorized" {
		t.Fatalf("expected unauthorized body, got %#v", body)
	}
}

func TestConfigAllowsAuthorizedAccess(t *testing.T) {
	handler := newTestHTTPHandler(t, newSecuredTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestConfigDoesNotLeakAuthTokenAndRedactsHeaders(t *testing.T) {
	handler := newTestHTTPHandler(t, newSecuredTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
	}

	raw := rec.Body.String()

	if strings.Contains(raw, "test-token") {
		t.Fatalf("response leaked server auth token: %s", raw)
	}
	if strings.Contains(raw, "upstream-secret") {
		t.Fatalf("response leaked upstream authorization header: %s", raw)
	}
	if strings.Contains(raw, "abc123") {
		t.Fatalf("response leaked api key header: %s", raw)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	serverRaw, ok := body["server"].(map[string]any)
	if !ok {
		t.Fatalf("response missing server object: %#v", body)
	}

	if v, exists := serverRaw["auth_token"]; exists {
		if s, _ := v.(string); strings.TrimSpace(s) != "" {
			t.Fatalf("expected auth_token to be absent or empty, got %#v", v)
		}
	}

	sourcesRaw, ok := body["desired_sources"].([]any)
	if !ok || len(sourcesRaw) == 0 {
		t.Fatalf("response missing desired_sources: %#v", body["desired_sources"])
	}

	firstSource, ok := sourcesRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("invalid desired_sources[0]: %#v", sourcesRaw[0])
	}

	headersRaw, ok := firstSource["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected redacted headers in response, got %#v", firstSource["headers"])
	}

	if got := headersRaw["Authorization"]; got != "***redacted***" {
		t.Fatalf("expected Authorization to be redacted, got %#v", got)
	}
	if got := headersRaw["X-Api-Key"]; got != "***redacted***" {
		t.Fatalf("expected X-Api-Key to be redacted, got %#v", got)
	}
	if got := headersRaw["User-Agent"]; got != "ros-list-test" {
		t.Fatalf("expected non-sensitive header to be preserved, got %#v", got)
	}
}

func TestRenderRejectsOutputPathOverride(t *testing.T) {
	handler := newTestHTTPHandler(t, newSecuredTestConfig())

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/render",
		strings.NewReader(`{"mode":"replace_all","output_path":"D:/tmp/hack.rsc"}`),
	)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v, raw=%s", err, rec.Body.String())
	}

	msg := body["error"]
	if !strings.Contains(msg, `unknown field "output_path"`) {
		t.Fatalf("expected unknown field output_path error, got %q", msg)
	}
}
