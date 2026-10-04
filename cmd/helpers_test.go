package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/boardwise/cli/internal/api"
	"github.com/boardwise/cli/internal/config"
)

// useServer points the package-level globals at a fake API for one test and
// restores them afterwards.
func useServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	oldClient, oldOrg, oldJSON, oldCfg := client, orgSlug, jsonOut, cfg
	t.Cleanup(func() { client, orgSlug, jsonOut, cfg = oldClient, oldOrg, oldJSON, oldCfg })

	client = api.NewClient(server.URL, "bw_test")
	orgSlug = "123456"
	jsonOut = false
	cfg = &config.Config{Token: "bw_test"}
	return server
}

// captureStdout runs fn and returns what it wrote to standard output.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	runErr := fn()
	os.Stdout = old
	w.Close()
	return <-done, runErr
}
