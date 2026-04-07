package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCurrentSnapshotIgnoresUnknownListWhenAutoCreateDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "current.json")
	payload := `{"unknownList": ["1.1.1.1"]}`
	if err := os.WriteFile(srcPath, []byte(payload), 0o644); err != nil {
		t.Fatalf("write temp source: %v", err)
	}

	cfg := AppConfig{
		AutoCreateLists: false,
		LogFile:         filepath.Join(tmpDir, "app.log"),
		Lists: []ListDefinition{{
			Name:    "knownList",
			Family:  FamilyIPv4,
			Enabled: true,
		}},
		CurrentStateSources: []SourceConfig{{
			Name:    "current-file",
			Type:    "file",
			Path:    srcPath,
			Enabled: true,
			Format:  "json",
		}},
		Output: OutputConfig{
			Path:           filepath.Join(tmpDir, "out.rsc"),
			Mode:           RenderModeReplaceAll,
			ManagedComment: "managed-by-go",
		},
		Server: ServerConfig{Listen: "127.0.0.1:8090"},
	}

	snap, err := BuildCurrentSnapshot(cfg)
	if err != nil {
		t.Fatalf("expected unknown current list to be ignored, got %v", err)
	}
	if len(snap.Entries) != 0 {
		t.Fatalf("expected no tracked entries, got %+v", snap.Entries)
	}
}
