package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
	Use:   "my",
	Short: "Your personal dashboard across all organizations",
}

var myNotificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "List your unread notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		var notifications []struct {
			Message string `json:"message"`
		}
		if printed, err := listAll("/my/notifications.json", "notifications", &notifications); err != nil || printed {
			return err
		}

		if len(notifications) == 0 {
			fmt.Println("No unread notifications.")
			return nil
		}
		for _, n := range notifications {
			fmt.Printf("• %s\n", n.Message)
		}
		return nil
	},
}

var myNotificationsMarkReadCmd = &cobra.Command{
	Use:   "mark-read",
	Short: "Mark all notifications as read",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Post("/my/notifications/mark_all_read.json", nil, nil); err != nil {
			return err
		}
		fmt.Println("All notifications marked as read.")
		return nil
	},
}

type scheduleMeeting struct {
	Title            string `json:"title"`
	StartsAt         string `json:"starts_at"`
	VisibilityState  string `json:"visibility_state"`
	OrganizationName string `json:"organization_name"`
}

// mySchedule reads every page of /my/schedule/list.json. Its two lists page
// in step, so each page adds to both.
func mySchedule() (upcoming, past []json.RawMessage, err error) {
	upcoming, past = []json.RawMessage{}, []json.RawMessage{}
	err = client.EachPage("/my/schedule/list.json", func(page []byte) error {
		var body struct {
			UpcomingMeetings []json.RawMessage `json:"upcoming_meetings"`
			PastMeetings     []json.RawMessage `json:"past_meetings"`
		}
		if err := json.Unmarshal(page, &body); err != nil {
			return fmt.Errorf("unexpected response from /my/schedule/list.json: %w", err)
		}
		upcoming = append(upcoming, body.UpcomingMeetings...)
		past = append(past, body.PastMeetings...)
		return nil
	})
	return upcoming, past, err
}

var myScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "List your upcoming and past meetings",
	RunE: func(cmd *cobra.Command, args []string) error {
		upcomingRaw, pastRaw, err := mySchedule()
		if err != nil {
			return err
		}

		if jsonOut {
			return printJSON(map[string]any{"upcoming_meetings": upcomingRaw, "past_meetings": pastRaw})
		}

		var upcoming, past []scheduleMeeting
		for _, list := range []struct {
			raw  []json.RawMessage
			into *[]scheduleMeeting
		}{{upcomingRaw, &upcoming}, {pastRaw, &past}} {
			data, err := json.Marshal(list.raw)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(data, list.into); err != nil {
				return err
			}
		}

		if len(upcoming)+len(past) == 0 {
			fmt.Println("No meetings.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tSTARTS AT\tSTATUS\tORG")
		for _, m := range append(upcoming, past...) {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Title, m.StartsAt, m.VisibilityState, m.OrganizationName)
		}
		w.Flush()
		return nil
	},
}

var myRSVPsCmd = &cobra.Command{
	Use:   "rsvps",
	Short: "List meetings with pending RSVPs",
	RunE: func(cmd *cobra.Command, args []string) error {
		var rsvps []struct {
			Meeting struct {
				Title            string `json:"title"`
				StartsAt         string `json:"starts_at"`
				OrganizationName string `json:"organization_name"`
			} `json:"meeting"`
		}
		if printed, err := listAll("/my/schedule/pending_rsvps.json", "pending_rsvps", &rsvps); err != nil || printed {
			return err
		}

		if len(rsvps) == 0 {
			fmt.Println("No pending RSVPs.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tSTARTS AT\tORG")
		for _, r := range rsvps {
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Meeting.Title, r.Meeting.StartsAt, r.Meeting.OrganizationName)
		}
		w.Flush()
		return nil
	},
}

var myTasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List pending tasks assigned to you",
	RunE: func(cmd *cobra.Command, args []string) error {
		var assignments []struct {
			ActionItem struct {
				Title            string `json:"title"`
				DueBy            string `json:"due_by"`
				OrganizationName string `json:"organization_name"`
			} `json:"action_item"`
		}
		if printed, err := listAll("/my/assignments.json", "assignments", &assignments); err != nil || printed {
			return err
		}

		if len(assignments) == 0 {
			fmt.Println("No pending tasks.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG")
		for _, a := range assignments {
			fmt.Fprintf(w, "%s\t%s\t%s\n", a.ActionItem.Title, a.ActionItem.DueBy, a.ActionItem.OrganizationName)
		}
		w.Flush()
		return nil
	},
}

var myMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "List your unread messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		var messages []struct {
			Subject          string `json:"subject"`
			OrganizationName string `json:"organization_name"`
		}
		if printed, err := listAll("/my/messages.json", "messages", &messages); err != nil || printed {
			return err
		}

		if len(messages) == 0 {
			fmt.Println("No unread messages.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "SUBJECT\tORG")
		for _, m := range messages {
			fmt.Fprintf(w, "%s\t%s\n", m.Subject, m.OrganizationName)
		}
		w.Flush()
		return nil
	},
}

var myConsentsCmd = &cobra.Command{
	Use:   "consents",
	Short: "List consent packets awaiting your signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		var consents []struct {
			URL           string `json:"url"`
			ConsentPacket struct {
				Title            string `json:"title"`
				DueBy            string `json:"due_by"`
				OrganizationName string `json:"organization_name"`
			} `json:"consent_packet"`
		}
		if printed, err := listAll("/my/consents.json", "consent_packets", &consents); err != nil || printed {
			return err
		}

		if len(consents) == 0 {
			fmt.Println("No pending consents.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG\tSIGN AT")
		for _, c := range consents {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ConsentPacket.Title, c.ConsentPacket.DueBy, c.ConsentPacket.OrganizationName, c.URL)
		}
		w.Flush()
		return nil
	},
}

var myDeclarationsCmd = &cobra.Command{
	Use:   "declarations",
	Short: "List declarations awaiting your response",
	RunE: func(cmd *cobra.Command, args []string) error {
		var declarations []struct {
			URL         string `json:"url"`
			Declaration struct {
				Title            string `json:"title"`
				DueBy            string `json:"due_by"`
				OrganizationName string `json:"organization_name"`
			} `json:"declaration"`
		}
		if printed, err := listAll("/my/declarations.json", "declarations", &declarations); err != nil || printed {
			return err
		}

		if len(declarations) == 0 {
			fmt.Println("No pending declarations.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG\tCOMPLETE AT")
		for _, d := range declarations {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", d.Declaration.Title, d.Declaration.DueBy, d.Declaration.OrganizationName, d.URL)
		}
		w.Flush()
		return nil
	},
}

var mySurveysCmd = &cobra.Command{
	Use:   "surveys",
	Short: "List surveys awaiting your response",
	RunE: func(cmd *cobra.Command, args []string) error {
		var surveys []struct {
			Survey struct {
				Title            string `json:"title"`
				DueBy            string `json:"due_by"`
				OrganizationName string `json:"organization_name"`
			} `json:"survey"`
		}
		if printed, err := listAll("/my/surveys.json", "surveys", &surveys); err != nil || printed {
			return err
		}

		if len(surveys) == 0 {
			fmt.Println("No pending surveys.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG")
		for _, s := range surveys {
			fmt.Fprintf(w, "%s\t%s\t%s\n", s.Survey.Title, s.Survey.DueBy, s.Survey.OrganizationName)
		}
		w.Flush()
		return nil
	},
}

func init() {
	myNotificationsCmd.AddCommand(myNotificationsMarkReadCmd)

	myCmd.AddCommand(myNotificationsCmd)
	myCmd.AddCommand(myScheduleCmd)
	myCmd.AddCommand(myRSVPsCmd)
	myCmd.AddCommand(myTasksCmd)
	myCmd.AddCommand(myMessagesCmd)
	myCmd.AddCommand(myConsentsCmd)
	myCmd.AddCommand(myDeclarationsCmd)
	myCmd.AddCommand(mySurveysCmd)
}
