package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/boardwise/cli/internal/api"
)

// tokenServer answers POST /api/v1/auth/token with the given responses in
// order (the last one repeats) and records each request body.
func tokenServer(t *testing.T, responses ...func(w http.ResponseWriter)) (*httptest.Server, *[]map[string]string) {
	t.Helper()
	var requests []map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/auth/token" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		requests = append(requests, body)

		i := len(requests) - 1
		if i >= len(responses) {
			i = len(responses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		responses[i](w)
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func refusal(code string) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": code, "error_description": code})
	}
}

func token(value string) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": value, "token_type": "Bearer"})
	}
}

func poll(server *httptest.Server) (string, error) {
	client := api.NewClient(server.URL, "")
	return pollForToken(client, "device-123", 5*time.Millisecond, time.Now().Add(5*time.Second))
}

func TestPollForTokenKeepsPollingWhilePending(t *testing.T) {
	server, requests := tokenServer(t, refusal("authorization_pending"), refusal("authorization_pending"), token("bw_abc"))

	got, err := poll(server)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "bw_abc" {
		t.Fatalf("token = %q, want bw_abc", got)
	}
	if len(*requests) != 3 {
		t.Fatalf("polled %d times, want 3", len(*requests))
	}
	for _, body := range *requests {
		if body["grant_type"] != deviceCodeGrant || body["device_code"] != "device-123" {
			t.Fatalf("request body = %v, want the device-code grant and the device code", body)
		}
	}
}

func TestPollForTokenStopsOnTerminalErrors(t *testing.T) {
	cases := map[string]string{
		"access_denied":          "denied",
		"expired_token":          "expired",
		"invalid_grant":          "polling error",
		"unsupported_grant_type": "polling error",
	}
	for code, want := range cases {
		t.Run(code, func(t *testing.T) {
			server, requests := tokenServer(t, refusal(code))

			_, err := poll(server)

			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("error = %v, want one mentioning %q", err, want)
			}
			if len(*requests) != 1 {
				t.Fatalf("polled %d times after %s, want 1", len(*requests), code)
			}
		})
	}
}

func TestPollForTokenTimesOut(t *testing.T) {
	server, _ := tokenServer(t, refusal("authorization_pending"))
	client := api.NewClient(server.URL, "")

	_, err := pollForToken(client, "device-123", 5*time.Millisecond, time.Now().Add(30*time.Millisecond))

	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("error = %v, want a timeout", err)
	}
}

func TestAPIErrorCode(t *testing.T) {
	cases := map[string]string{
		`{"error":"authorization_pending","message":"Waiting"}`: "authorization_pending",
		`not json`: "",
		`{}`:       "",
	}
	for body, want := range cases {
		if got := (&api.APIError{StatusCode: 400, Body: body}).Code(); got != want {
			t.Errorf("Code() for %s = %q, want %q", body, got, want)
		}
	}
}

func TestDeviceRequest(t *testing.T) {
	cases := []struct {
		readOnly bool
		org      string
		want     map[string]string
	}{
		{false, "", map[string]string{"client_name": "Boardwise CLI"}},
		{true, "", map[string]string{"client_name": "Boardwise CLI", "scope": "read"}},
		{false, "304923", map[string]string{"client_name": "Boardwise CLI", "organization_slug": "304923"}},
		{true, "304923", map[string]string{"client_name": "Boardwise CLI", "scope": "read", "organization_slug": "304923"}},
	}
	for _, c := range cases {
		if got := deviceRequest(c.readOnly, c.org); !reflect.DeepEqual(got, c.want) {
			t.Errorf("deviceRequest(%v, %q) = %v, want %v", c.readOnly, c.org, got, c.want)
		}
	}
}
