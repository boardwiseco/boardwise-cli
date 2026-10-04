package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// pagedServer serves three pages of /items.json the way Rails does: absolute
// Link URLs that carry the request's other query parameters.
func pagedServer(t *testing.T, queries *[]url.Values) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/123456/items.json" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer bw_test" {
			t.Errorf("Authorization = %q", got)
		}
		q := r.URL.Query()
		*queries = append(*queries, q)

		page := q.Get("page")
		if page == "" {
			page = "1"
		}
		link := func(n int, rel string) string {
			next := url.Values{}
			for k, v := range q {
				next[k] = v
			}
			next.Set("page", fmt.Sprint(n))
			return fmt.Sprintf(`<%s/123456/items.json?%s>; rel="%s"`, server.URL, next.Encode(), rel)
		}
		var links []string
		switch page {
		case "1":
			links = []string{link(1, "first"), link(2, "next"), link(3, "last")}
		case "2":
			links = []string{link(1, "first"), link(1, "prev"), link(3, "next"), link(3, "last")}
		case "3":
			links = []string{link(1, "first"), link(2, "prev"), link(3, "last")}
		}
		w.Header().Set("Link", strings.Join(links, ", "))
		w.Header().Set("X-Total-Count", "3")
		fmt.Fprintf(w, `{"items":[{"page":%s}]}`, page)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestEachPageFollowsNextLinks(t *testing.T) {
	var queries []url.Values
	server := pagedServer(t, &queries)
	client := NewClient(server.URL, "bw_test")

	var pages []string
	err := client.EachPage("/123456/items.json?status=all", func(page []byte) error {
		pages = append(pages, string(page))
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{`{"items":[{"page":1}]}`, `{"items":[{"page":2}]}`, `{"items":[{"page":3}]}`}
	if strings.Join(pages, "|") != strings.Join(want, "|") {
		t.Fatalf("pages = %v, want %v", pages, want)
	}
	if len(queries) != 3 {
		t.Fatalf("sent %d requests, want 3", len(queries))
	}
	for i, q := range queries {
		if q.Get("per_page") != "100" {
			t.Errorf("request %d per_page = %q, want 100", i+1, q.Get("per_page"))
		}
		if q.Get("status") != "all" {
			t.Errorf("request %d status = %q, want the caller's status=all kept", i+1, q.Get("status"))
		}
	}
}

func TestEachPageKeepsACallersPerPage(t *testing.T) {
	var queries []url.Values
	server := pagedServer(t, &queries)

	err := NewClient(server.URL, "bw_test").EachPage("/123456/items.json?per_page=10", func([]byte) error { return nil })

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := queries[0].Get("per_page"); got != "10" {
		t.Fatalf("per_page = %q, want the caller's 10", got)
	}
}

func TestEachPageStopsWithoutALinkHeader(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		fmt.Fprint(w, `{"people":[]}`)
	}))
	defer server.Close()

	err := NewClient(server.URL, "bw_test").EachPage("/123456/people.json", func([]byte) error { return nil })

	if err != nil || requests != 1 {
		t.Fatalf("err = %v, requests = %d; want one request and no error", err, requests)
	}
}

func TestEachPageRefusesANextLinkToAnotherHost(t *testing.T) {
	elsewhere := 0
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		elsewhere++
		t.Errorf("request sent to another host with Authorization %q", r.Header.Get("Authorization"))
	}))
	defer other.Close()

	for _, next := range []string{
		other.URL + "/123456/items.json?page=2",                 // same IP, different port
		strings.Replace(other.URL, "127.0.0.1", "localhost", 1), // different host name
		"https://evil.example/123456/items.json?page=2",
	} {
		t.Run(next, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, next))
				fmt.Fprint(w, `{"items":[]}`)
			}))
			defer server.Close()

			pages := 0
			err := NewClient(server.URL, "bw_test").EachPage("/123456/items.json", func([]byte) error {
				pages++
				return nil
			})

			if err == nil || !strings.Contains(err.Error(), "refusing") {
				t.Fatalf("error = %v, want a refusal", err)
			}
			if pages != 1 {
				t.Fatalf("read %d pages, want 1", pages)
			}
		})
	}
	if elsewhere != 0 {
		t.Fatalf("%d requests reached the other host", elsewhere)
	}
}

func TestNextLink(t *testing.T) {
	cases := map[string]string{
		`<https://a/x?page=1>; rel="first", <https://a/x?page=2>; rel="next", <https://a/x?page=9>; rel="last"`: "https://a/x?page=2",
		`<https://a/x?page=2>;rel=next`:                  "https://a/x?page=2",
		`<https://a/x?page=2>; rel="next last"`:          "https://a/x?page=2",
		`<https://a/x?page=1>; rel="prev"`:               "",
		`<https://a/x?page=3>; title="next"; rel="last"`: "",
		``: "",
	}
	for header, want := range cases {
		if got := nextLink([]string{header}); got != want {
			t.Errorf("nextLink(%q) = %q, want %q", header, got, want)
		}
	}
}
