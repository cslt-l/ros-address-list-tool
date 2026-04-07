package app

import "testing"

func TestValidateConfigRequiresTokenWhenListenExposed(t *testing.T) {
	cfg := AppConfig{
		AutoCreateLists: true,
		LogFile:         "./logs/app.log",
		Output: OutputConfig{
			Path:           "./output/routeros-address-list.rsc",
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{Listen: ":8090"},
	}

	if err := ValidateConfig(cfg); err == nil {
		t.Fatalf("expected exposed listen without token to fail validation")
	}

	cfg.Server.AuthToken = "secret"
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected exposed listen with token to pass, got %v", err)
	}
}

func TestValidateConfigAllowsDisabledSourceWithoutLocation(t *testing.T) {
	cfg := AppConfig{
		AutoCreateLists: true,
		LogFile:         "./logs/app.log",
		Output: OutputConfig{
			Path:           "./output/routeros-address-list.rsc",
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{Listen: "127.0.0.1:8090"},
		DesiredSources: []SourceConfig{{
			Name:    "disabled-source",
			Type:    "url",
			Enabled: false,
		}},
	}

	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected disabled source without url/path to pass, got %v", err)
	}
}
