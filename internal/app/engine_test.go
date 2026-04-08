package app

import (
	"io"
	"log"
	"path/filepath"
	"strings"
	"testing"
)

func newEngineTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func newBaseExecuteTestConfig(t *testing.T) AppConfig {
	t.Helper()

	tmpDir := t.TempDir()

	cfg := AppConfig{
		AutoCreateLists: false,
		LogFile:         filepath.Join(tmpDir, "app.log"),
		Lists: []ListDefinition{
			{
				Name:        "toWanTelecom4",
				Family:      FamilyIPv4,
				Enabled:     true,
				Description: "中国电信 IPv4",
			},
		},
		Output: OutputConfig{
			Path:           filepath.Join(tmpDir, "routeros-address-list.rsc"),
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{
			Listen:    "127.0.0.1:8090",
			EnableWeb: false,
			WebDir:    "./web/src",
		},
	}

	cfg.ApplyDefaults()
	return cfg
}

func TestExecuteRejectsDiffWithoutCurrentStateSources(t *testing.T) {
	t.Setenv("ROS_LIST_API_TOKEN", "")

	cfg := newBaseExecuteTestConfig(t)
	cfg.Output.Mode = RenderModeDiff
	cfg.CurrentStateSources = nil

	_, err := Execute(cfg, newEngineTestLogger())
	if err == nil {
		t.Fatal("expected diff mode without current_state_sources to fail")
	}

	msg := err.Error()
	if !strings.Contains(msg, "current_state_sources") {
		t.Fatalf("expected current_state_sources error, got: %v", err)
	}
}

func TestExecuteAllowsReplaceAllWithoutCurrentStateSources(t *testing.T) {
	t.Setenv("ROS_LIST_API_TOKEN", "")

	cfg := newBaseExecuteTestConfig(t)
	cfg.Output.Mode = RenderModeReplaceAll
	cfg.CurrentStateSources = nil

	result, err := Execute(cfg, newEngineTestLogger())
	if err != nil {
		t.Fatalf("expected replace_all without current_state_sources to pass, got: %v", err)
	}

	if result.Mode != RenderModeReplaceAll {
		t.Fatalf("expected mode=%q, got %q", RenderModeReplaceAll, result.Mode)
	}

	if result.OutputPath == "" {
		t.Fatalf("expected output path to be set")
	}

	if result.ListCount != 1 {
		t.Fatalf("expected list_count=1, got %d", result.ListCount)
	}
}

func TestExecuteRejectsDiffWithoutCurrentStateSourcesEvenWhenDesiredListsExist(t *testing.T) {
	t.Setenv("ROS_LIST_API_TOKEN", "")

	cfg := newBaseExecuteTestConfig(t)
	cfg.Output.Mode = RenderModeDiff
	cfg.DesiredSources = []SourceConfig{
		{
			Name:             "desired-disabled-placeholder",
			Type:             "url",
			Enabled:          false,
			Format:           "plain_cidr",
			TargetListName:   "toWanTelecom4",
			TargetListFamily: FamilyIPv4,
		},
	}
	cfg.CurrentStateSources = nil

	_, err := Execute(cfg, newEngineTestLogger())
	if err == nil {
		t.Fatal("expected diff mode without current_state_sources to fail")
	}

	msg := err.Error()
	if !strings.Contains(msg, "current_state_sources") {
		t.Fatalf("expected current_state_sources error, got: %v", err)
	}
}
