package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

// recordCreate fakes a create endpoint: it records the method, path and
// decoded JSON body, and answers 201 with a minimal record.
func recordCreate(t *testing.T) *struct {
	method, path string
	body         map[string]any
} {
	t.Helper()
	got := &struct {
		method, path string
		body         map[string]any
	}{}
	useServer(t, func(w http.ResponseWriter, r *http.Request) {
		got.method, got.path = r.Method, r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&got.body); err != nil {
			t.Errorf("request body is not JSON: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"id":"x1","title":"Created","subject":"Created","position":1}`)
	})
	return got
}

// setString sets a flag-backed global for one test and restores it afterwards.
func setString(t *testing.T, p *string, v string) {
	old := *p
	*p = v
	t.Cleanup(func() { *p = old })
}

func TestTasksCreateWrapsTheBodyInActionItem(t *testing.T) {
	got := recordCreate(t)
	setString(t, &taskTitle, "File the annual return")
	setString(t, &taskDueBy, "2026-11-15")
	setString(t, &taskGroupID, "g1")
	oldAssign := taskAssign
	taskAssign = []string{"p1", "p2"}
	t.Cleanup(func() { taskAssign = oldAssign })

	if _, err := captureStdout(t, func() error { return tasksCreateCmd.RunE(tasksCreateCmd, nil) }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{"action_item": map[string]any{
		"title":               "File the annual return",
		"due_by":              "2026-11-15",
		"group_id":            "g1",
		"assigned_person_ids": []any{"p1", "p2"},
	}}
	if got.method != "POST" || got.path != "/123456/action_items.json" {
		t.Fatalf("request = %s %s", got.method, got.path)
	}
	if !reflect.DeepEqual(got.body, want) {
		t.Fatalf("body = %v\nwant %v", got.body, want)
	}
}

func TestMeetingsCreateWrapsTheBodyInMeeting(t *testing.T) {
	got := recordCreate(t)
	setString(t, &meetingTitle, "Q4 board meeting")
	setString(t, &meetingStartsAt, "2026-12-01T15:00:00Z")
	setString(t, &meetingEndsAt, "2026-12-01T17:00:00Z")
	setString(t, &meetingTimeZone, "America/Toronto")

	if _, err := captureStdout(t, func() error { return meetingsCreateCmd.RunE(meetingsCreateCmd, nil) }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{"meeting": map[string]any{
		"title":     "Q4 board meeting",
		"starts_at": "2026-12-01T15:00:00Z",
		"ends_at":   "2026-12-01T17:00:00Z",
		"time_zone": "America/Toronto",
	}}
	if got.path != "/123456/meetings.json" || !reflect.DeepEqual(got.body, want) {
		t.Fatalf("%s body = %v\nwant %v", got.path, got.body, want)
	}
}

func TestAgendaAddWrapsTheBodyInAgendaItem(t *testing.T) {
	got := recordCreate(t)
	setString(t, &agendaTitle, "Approve minutes")
	oldDuration := agendaDuration
	agendaDuration = 10
	t.Cleanup(func() { agendaDuration = oldDuration })

	if _, err := captureStdout(t, func() error { return agendaAddCmd.RunE(agendaAddCmd, []string{"m1"}) }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{"agenda_item": map[string]any{"title": "Approve minutes", "duration_minutes": float64(10)}}
	if got.path != "/123456/meetings/m1/agenda_items.json" || !reflect.DeepEqual(got.body, want) {
		t.Fatalf("%s body = %v\nwant %v", got.path, got.body, want)
	}
}

func TestMessagesSendWrapsTheBodyInMessage(t *testing.T) {
	got := recordCreate(t)
	setString(t, &msgSubject, "Materials are up")
	setString(t, &msgBody, "<p>See the board book.</p>")

	if _, err := captureStdout(t, func() error { return messagesSendCmd.RunE(messagesSendCmd, []string{"board"}) }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]any{"message": map[string]any{"subject": "Materials are up", "body": "<p>See the board book.</p>"}}
	if got.path != "/123456/groups/board/messages.json" || !reflect.DeepEqual(got.body, want) {
		t.Fatalf("%s body = %v\nwant %v", got.path, got.body, want)
	}
}
