package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/boardwise/cli/internal/api"
	"github.com/boardwise/cli/internal/config"
)

// loggedIn stores a config with a token in a temp config dir, points the
// globals at it and at a client for baseURL.
func loggedIn(t *testing.T, baseURL string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir) // os.UserConfigDir on Linux
	t.Setenv("HOME", dir)            // and on macOS

	saved := &config.Config{Token: "bw_secret", DefaultOrg: "123456", URL: baseURL}
	if err := saved.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load()
	if err != nil || loaded.Token != "bw_secret" {
		t.Fatalf("config did not round-trip: %+v, %v", loaded, err)
	}

	oldClient, oldCfg := client, cfg
	t.Cleanup(func() { client, cfg = oldClient, oldCfg })
	cfg = loaded
	client = api.NewClient(baseURL, loaded.Token)
}

func assertCleared(t *testing.T) {
	t.Helper()
	stored, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if stored.Token != "" || stored.DefaultOrg != "" || stored.URL != "" {
		t.Fatalf("stored config after logout = %+v, want it cleared", stored)
	}
}

func TestLogoutRevokesTheTokenAndClearsConfig(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		wantStdout string
	}{
		{"revoked", http.StatusOK, "token has been revoked"},
		{"already revoked", http.StatusUnauthorized, "already been revoked or had expired"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var method, path, auth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method, path, auth = r.Method, r.URL.Path, r.Header.Get("Authorization")
				w.WriteHeader(c.status)
				if c.status == http.StatusUnauthorized {
					_, _ = w.Write([]byte(`{"error":"unauthorized","message":"Sign in again."}`))
				} else {
					_, _ = w.Write([]byte(`{}`))
				}
			}))
			defer server.Close()
			loggedIn(t, server.URL)

			stdout, stderr, err := captureOutput(t, func() error { return logoutCmd.RunE(logoutCmd, nil) })

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if method != "DELETE" || path != "/api/v1/auth" || auth != "Bearer bw_secret" {
				t.Fatalf("request = %s %s (Authorization %q), want DELETE /api/v1/auth with the token", method, path, auth)
			}
			if !strings.Contains(stdout, c.wantStdout) || stderr != "" {
				t.Fatalf("stdout = %q, stderr = %q", stdout, stderr)
			}
			assertCleared(t)
		})
	}
}

func TestLogoutClearsConfigAndWarnsWhenTheServerIsUnreachable(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	unreachable := server.URL
	server.Close()
	loggedIn(t, unreachable)

	stdout, stderr, err := captureOutput(t, func() error { return logoutCmd.RunE(logoutCmd, nil) })

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "may still be valid") || !strings.Contains(stderr, unreachable+"/my/api_tokens") {
		t.Fatalf("stderr = %q, want a warning pointing at the API tokens page", stderr)
	}
	if !strings.Contains(stdout, "Logged out") {
		t.Fatalf("stdout = %q", stdout)
	}
	assertCleared(t)
}

func TestLogoutWithoutATokenJustClears(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++ }))
	defer server.Close()
	loggedIn(t, server.URL)
	cfg.Token = ""

	if _, err := captureStdout(t, func() error { return logoutCmd.RunE(logoutCmd, nil) }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("sent %d requests without a token, want 0", requests)
	}
	assertCleared(t)
}
