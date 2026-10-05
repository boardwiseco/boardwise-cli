package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// testClient returns a client for server whose sleeps are recorded, not slept.
func testClient(server *httptest.Server, slept *[]time.Duration) *Client {
	c := NewClient(server.URL, "bw_test")
	c.sleep = func(d time.Duration) { *slept = append(*slept, d) }
	return c
}

func TestRateLimitedRequestIsRetriedOnceWithTheSameBody(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if len(bodies) == 1 {
			w.Header().Set("Retry-After", "7")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate_limited","message":"Too many requests.","retry_after":7}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	defer server.Close()
	var slept []time.Duration

	var out struct{ ID string }
	err := testClient(server, &slept).Post("/thing.json", map[string]string{"title": "Audit"}, &out)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != "1" {
		t.Fatalf("decoded id = %q, want 1", out.ID)
	}
	if len(bodies) != 2 {
		t.Fatalf("sent %d requests, want 2", len(bodies))
	}
	if bodies[0] != `{"title":"Audit"}` || bodies[1] != bodies[0] {
		t.Fatalf("bodies = %q, want the same JSON body twice", bodies)
	}
	if len(slept) != 1 || slept[0] != 7*time.Second {
		t.Fatalf("slept %v, want [7s]", slept)
	}
}

func TestRateLimitedTwiceIsAnError(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate_limited","message":"Too many requests. Wait 1 seconds, then try again.","retry_after":1}`))
	}))
	defer server.Close()
	var slept []time.Duration

	_, err := testClient(server, &slept).GetRaw("/thing.json")

	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("error = %v, want a 429 APIError", err)
	}
	if requests != 2 {
		t.Fatalf("sent %d requests, want 2", requests)
	}
}

func TestRetryAfter(t *testing.T) {
	cases := map[string]time.Duration{
		"3":    3 * time.Second,
		"0":    0,
		"600":  60 * time.Second,
		"":     5 * time.Second,
		"soon": 5 * time.Second,
		"-1":   5 * time.Second,
		" 12 ": 12 * time.Second,
	}
	for header, want := range cases {
		if got := retryAfter(header); got != want {
			t.Errorf("retryAfter(%q) = %v, want %v", header, got, want)
		}
	}
}

func TestAPIErrorMessage(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{404, `{"error":"not_found","message":"We couldn't find that."}`, "We couldn't find that. (not_found)"},
		{422, `{"error":"validation_failed","message":"Title can't be blank","errors":{"title":["can't be blank"],"due_by":["is invalid","is in the past"]}}`,
			"Title can't be blank (validation_failed): due_by is invalid, is in the past; title can't be blank"},
		{502, `<html>Bad gateway</html>`, "HTTP 502: <html>Bad gateway</html>"},
		{500, `{"status":500}`, `HTTP 500: {"status":500}`},
	}
	for _, c := range cases {
		if got := (&APIError{StatusCode: c.status, Body: c.body}).Error(); got != c.want {
			t.Errorf("Error() for %s\n got %q\nwant %q", c.body, got, c.want)
		}
	}
}
