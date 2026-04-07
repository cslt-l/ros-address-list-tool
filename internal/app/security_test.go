package app

import "testing"

func TestValidateSourceURLStringRejectsPrivateTargets(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "public https", input: "https://example.com/list.json", wantErr: false},
		{name: "loopback ipv4", input: "http://127.0.0.1/data.json", wantErr: true},
		{name: "private ipv4", input: "http://192.168.1.10/data.json", wantErr: true},
		{name: "loopback ipv6", input: "http://[::1]/data.json", wantErr: true},
		{name: "userinfo", input: "https://user:pass@example.com/data.json", wantErr: true},
		{name: "invalid scheme", input: "ftp://example.com/data.json", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSourceURLString(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestResolveRenderOutputPath(t *testing.T) {
	base := "./output/routeros-address-list.rsc"

	ok, err := resolveRenderOutputPath(base, "./output/custom.rsc")
	if err != nil {
		t.Fatalf("expected override inside output dir to pass, got %v", err)
	}
	if ok == "" {
		t.Fatalf("expected resolved path, got empty")
	}

	if _, err := resolveRenderOutputPath(base, "./logs/app.log"); err == nil {
		t.Fatalf("expected non-.rsc or out-of-dir override to fail")
	}
	if _, err := resolveRenderOutputPath(base, "../escape.rsc"); err == nil {
		t.Fatalf("expected parent escape override to fail")
	}
}
