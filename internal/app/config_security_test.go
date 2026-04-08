package app

import (
	"strings"
	"testing"
)

func TestValidateConfigRequiresTokenWhenListenExposed(t *testing.T) {
	t.Setenv("ROS_LIST_API_TOKEN", "")

	cfg := AppConfig{
		AutoCreateLists: true,
		LogFile:         "./logs/app.log",
		Output: OutputConfig{
			Path:           "./output/routeros-address-list.rsc",
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{
			Listen:    ":8090",
			EnableWeb: false,
			WebDir:    "./web/src",
			AuthToken: "",
		},
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatalf("expected exposed listen without token to fail validation")
	}

	if !strings.Contains(err.Error(), "必须配置 server.auth_token") {
		t.Fatalf("expected missing auth token error, got: %v", err)
	}

	cfg.Server.AuthToken = "secret"
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected exposed listen with token to pass, got %v", err)
	}
}

func TestValidateConfigAllowsDisabledSourceWithoutLocation(t *testing.T) {
	t.Setenv("ROS_LIST_API_TOKEN", "")

	cfg := AppConfig{
		AutoCreateLists: true,
		LogFile:         "./logs/app.log",
		Output: OutputConfig{
			Path:           "./output/routeros-address-list.rsc",
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{
			Listen:    "127.0.0.1:8090",
			EnableWeb: false,
			WebDir:    "./web/src",
		},
		DesiredSources: []SourceConfig{
			{
				Name:    "disabled-source",
				Type:    "url",
				Enabled: false,
			},
		},
	}

	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected disabled source without url/path to pass, got %v", err)
	}
}
