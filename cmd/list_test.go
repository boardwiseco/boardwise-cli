package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// servePages answers with pages[page-1] for ?page=N (page 1 when absent),
// linking each page to the next with an absolute URL as Rails does.
func servePages(t *testing.T, wantPath string, pages ...string) *[]string {
	t.Helper()
	var queries []string
	var base string
	server := useServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
		}
		queries = append(queries, r.URL.RawQuery)
		n := 1
		if p := r.URL.Query().Get("page"); p != "" {
			fmt.Sscan(p, &n)
		}
		if n < len(pages) {
			q := r.URL.Query()
			q.Set("page", fmt.Sprint(n+1))
			w.Header().Set("Link", fmt.Sprintf(`<%s%s?%s>; rel="next"`, base, r.URL.Path, q.Encode()))
		}
		fmt.Fprint(w, pages[n-1])
	})
	base = server.URL
	return &queries
}

func TestListAllCollectsAKeyedListFromEveryPage(t *testing.T) {
	queries := servePages(t, "/123456/people.json",
		`{"people":[{"name":"Ada"},{"name":"Grace"}]}`,
		`{"people":[{"name":"Edsger"}]}`)

	var people []struct{ Name string }
	printed, err := listAll("/123456/people.json", "people", &people)

	if err != nil || printed {
		t.Fatalf("printed = %v, err = %v", printed, err)
	}
	var names []string
	for _, p := range people {
		names = append(names, p.Name)
	}
	if strings.Join(names, ",") != "Ada,Grace,Edsger" {
		t.Fatalf("names = %v, want every page's people", names)
	}
	if len(*queries) != 2 || !strings.Contains((*queries)[0], "per_page=100") {
		t.Fatalf("queries = %v, want two requests asking for 100 per page", *queries)
	}
}

func TestListAllCollectsABareArray(t *testing.T) {
	servePages(t, "/123456/action_items.json",
		`[{"title":"One"}]`,
		`[{"title":"Two"}]`,
		`[{"title":"Three"}]`)

	var items []struct{ Title string }
	if _, err := listAll("/123456/action_items.json?status=all", "", &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[2].Title != "Three" {
		t.Fatalf("items = %+v, want all three", items)
	}
}

func TestListAllPrintsEveryItemInTheAPIShapeWithJSON(t *testing.T) {
	cases := []struct {
		name, key string
		pages     []string
		want      any
	}{
		{"keyed", "people", []string{`{"people":[{"name":"Ada"}]}`, `{"people":[{"name":"Grace"}]}`},
			map[string]any{"people": []any{map[string]any{"name": "Ada"}, map[string]any{"name": "Grace"}}}},
		{"bare", "", []string{`[{"title":"One"}]`, `[{"title":"Two"}]`},
			[]any{map[string]any{"title": "One"}, map[string]any{"title": "Two"}}},
		{"empty keyed", "people", []string{`{"people":[]}`}, map[string]any{"people": []any{}}},
		{"empty bare", "", []string{`[]`}, []any{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			servePages(t, "/123456/list.json", c.pages...)
			jsonOut = true

			var items []map[string]any
			var printed bool
			out, err := captureStdout(t, func() error {
				var err error
				printed, err = listAll("/123456/list.json", c.key, &items)
				return err
			})

			if err != nil || !printed {
				t.Fatalf("printed = %v, err = %v", printed, err)
			}
			var got any
			if err := json.Unmarshal([]byte(out), &got); err != nil {
				t.Fatalf("output is not JSON: %v\n%s", err, out)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("printed %s\nwant %v", out, c.want)
			}
		})
	}
}

func TestTasksListKeepsTheStatusFilterAcrossPages(t *testing.T) {
	queries := servePages(t, "/123456/action_items.json", `[{"title":"One"}]`, `[{"title":"Two"}]`)
	old := tasksListStatus
	tasksListStatus = "all"
	t.Cleanup(func() { tasksListStatus = old })

	out, err := captureStdout(t, func() error { return tasksListCmd.RunE(tasksListCmd, nil) })

	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "One") || !strings.Contains(out, "Two") {
		t.Fatalf("output lacks a page:\n%s", out)
	}
	for _, q := range *queries {
		if !strings.Contains(q, "status=all") || !strings.Contains(q, "per_page=100") {
			t.Fatalf("query %q lost status=all or per_page=100", q)
		}
	}
}

func TestMyScheduleCollectsBothListsFromEveryPage(t *testing.T) {
	servePages(t, "/my/schedule/list.json",
		`{"upcoming_meetings":[{"title":"Next AGM"}],"past_meetings":[{"title":"Last AGM"}]}`,
		`{"upcoming_meetings":[],"past_meetings":[{"title":"Older AGM"}]}`)

	out, err := captureStdout(t, func() error { return myScheduleCmd.RunE(myScheduleCmd, nil) })
	if err != nil {
		t.Fatal(err)
	}
	for _, title := range []string{"Next AGM", "Last AGM", "Older AGM"} {
		if !strings.Contains(out, title) {
			t.Errorf("output lacks %q:\n%s", title, out)
		}
	}

	jsonOut = true
	out, err = captureStdout(t, func() error { return myScheduleCmd.RunE(myScheduleCmd, nil) })
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Upcoming []map[string]any `json:"upcoming_meetings"`
		Past     []map[string]any `json:"past_meetings"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	if len(got.Upcoming) != 1 || len(got.Past) != 2 {
		t.Fatalf("JSON has %d upcoming and %d past, want 1 and 2:\n%s", len(got.Upcoming), len(got.Past), out)
	}
}
