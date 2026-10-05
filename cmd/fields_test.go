package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func serveJSON(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}
}

func TestTasksListShowsAssignees(t *testing.T) {
	useServer(t, serveJSON(`[{"id":"a1","title":"Sign the lease","status":"pending","due_by":"2026-11-15T00:00:00Z",
		"assignees":[{"person_id":"p1","name":"Ada Lovelace"},{"person_id":"p2","name":"Grace Hopper"}]}]`))

	out, err := captureStdout(t, func() error { return tasksListCmd.RunE(tasksListCmd, nil) })

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ASSIGNEES") || !strings.Contains(out, "Ada Lovelace, Grace Hopper") {
		t.Fatalf("output lacks the assignees:\n%s", out)
	}
}

func TestMyComplianceListsShowNestedTitles(t *testing.T) {
	cases := []struct {
		name string
		run  func() error
		body string
		want []string
	}{
		{"consents", func() error { return myConsentsCmd.RunE(myConsentsCmd, nil) },
			`{"consent_packets":[{"signer_id":"s1","state":"pending","url":"https://app.boardwise.co/my/consents/c1",
			  "consent_packet":{"id":"c1","title":"Approve the budget","status":"partially_signed","organization_id":"o1","organization_name":"Acme"}}]}`,
			[]string{"Approve the budget", "Acme", "https://app.boardwise.co/my/consents/c1"}},
		{"declarations", func() error { return myDeclarationsCmd.RunE(myDeclarationsCmd, nil) },
			`{"declarations":[{"respondent_id":"r1","state":"pending","url":"https://app.boardwise.co/my/declarations/d1",
			  "declaration":{"id":"d1","title":"Conflicts 2026","status":"issued","organization_id":"o1","organization_name":"Acme"}}]}`,
			[]string{"Conflicts 2026", "https://app.boardwise.co/my/declarations/d1"}},
		{"surveys", func() error { return mySurveysCmd.RunE(mySurveysCmd, nil) },
			`{"surveys":[{"respondent_id":"r1","state":"pending",
			  "survey":{"id":"s1","title":"Board evaluation","status":"issued","organization_id":"o1","organization_name":"Acme"}}]}`,
			[]string{"Board evaluation", "Acme"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			useServer(t, serveJSON(c.body))

			out, err := captureStdout(t, c.run)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, want := range c.want {
				if !strings.Contains(out, want) {
					t.Errorf("output lacks %q:\n%s", want, out)
				}
			}
		})
	}
}
