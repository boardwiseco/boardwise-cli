package config

import (
	"os"
	"path/filepath"
	"testing"
)

// useTempConfigDir makes os.UserConfigDir resolve inside a temp dir, on
// Linux (XDG_CONFIG_HOME) and macOS (HOME), and returns the config path.
func useTempConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	path, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	if rel, err := filepath.Rel(dir, path); err != nil || filepath.IsAbs(rel) || rel[:2] == ".." {
		t.Fatalf("config path %s is outside the temp dir %s", path, dir)
	}
	return path
}

func TestLoadIgnoresTheRetiredSuperadminKey(t *testing.T) {
	path := useTempConfigDir(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	old := `{"token":"bw_old","default_org":"123456","superadmin":true,"url":"https://app.boardwise.co"}`
	if err := os.WriteFile(path, []byte(old), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()

	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Token != "bw_old" || cfg.DefaultOrg != "123456" || cfg.URL != "https://app.boardwise.co" {
		t.Fatalf("Load() = %+v, want the old file's other fields", cfg)
	}
}
