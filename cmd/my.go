package cmd

import (
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
		if jsonOut {
			raw, err := client.GetRaw("/my/notifications.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Notifications []struct {
				Message string `json:"message"`
			} `json:"notifications"`
		}
		if err := client.Get("/my/notifications.json", &result); err != nil {
			return err
		}

		if len(result.Notifications) == 0 {
			fmt.Println("No unread notifications.")
			return nil
		}
		for _, n := range result.Notifications {
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

var myScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "List your upcoming and past meetings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOut {
			raw, err := client.GetRaw("/my/schedule/list.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		type meetingRow struct {
			Title            string `json:"title"`
			StartsAt         string `json:"starts_at"`
			VisibilityState  string `json:"visibility_state"`
			OrganizationName string `json:"organization_name"`
		}
		var result struct {
			UpcomingMeetings []meetingRow `json:"upcoming_meetings"`
			PastMeetings     []meetingRow `json:"past_meetings"`
		}
		if err := client.Get("/my/schedule/list.json", &result); err != nil {
			return err
		}

		if len(result.UpcomingMeetings)+len(result.PastMeetings) == 0 {
			fmt.Println("No meetings.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tSTARTS AT\tSTATUS\tORG")
		for _, m := range result.UpcomingMeetings {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Title, m.StartsAt, m.VisibilityState, m.OrganizationName)
		}
		for _, m := range result.PastMeetings {
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
		if jsonOut {
			raw, err := client.GetRaw("/my/schedule/pending_rsvps.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			PendingRSVPs []struct {
				Meeting struct {
					Title            string `json:"title"`
					StartsAt         string `json:"starts_at"`
					OrganizationName string `json:"organization_name"`
				} `json:"meeting"`
			} `json:"pending_rsvps"`
		}
		if err := client.Get("/my/schedule/pending_rsvps.json", &result); err != nil {
			return err
		}

		if len(result.PendingRSVPs) == 0 {
			fmt.Println("No pending RSVPs.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tSTARTS AT\tORG")
		for _, r := range result.PendingRSVPs {
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
		if jsonOut {
			raw, err := client.GetRaw("/my/assignments.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Assignments []struct {
				ActionItem struct {
					Title            string `json:"title"`
					DueBy            string `json:"due_by"`
					OrganizationName string `json:"organization_name"`
				} `json:"action_item"`
			} `json:"assignments"`
		}
		if err := client.Get("/my/assignments.json", &result); err != nil {
			return err
		}

		if len(result.Assignments) == 0 {
			fmt.Println("No pending tasks.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG")
		for _, a := range result.Assignments {
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
		if jsonOut {
			raw, err := client.GetRaw("/my/messages.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Messages []struct {
				Subject          string `json:"subject"`
				OrganizationName string `json:"organization_name"`
			} `json:"messages"`
		}
		if err := client.Get("/my/messages.json", &result); err != nil {
			return err
		}

		if len(result.Messages) == 0 {
			fmt.Println("No unread messages.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "SUBJECT\tORG")
		for _, m := range result.Messages {
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
		if jsonOut {
			raw, err := client.GetRaw("/my/consents.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			ConsentPackets []struct {
				URL           string `json:"url"`
				ConsentPacket struct {
					Title            string `json:"title"`
					DueBy            string `json:"due_by"`
					OrganizationName string `json:"organization_name"`
				} `json:"consent_packet"`
			} `json:"consent_packets"`
		}
		if err := client.Get("/my/consents.json", &result); err != nil {
			return err
		}

		if len(result.ConsentPackets) == 0 {
			fmt.Println("No pending consents.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG\tSIGN AT")
		for _, c := range result.ConsentPackets {
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
		if jsonOut {
			raw, err := client.GetRaw("/my/declarations.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Declarations []struct {
				URL         string `json:"url"`
				Declaration struct {
					Title            string `json:"title"`
					DueBy            string `json:"due_by"`
					OrganizationName string `json:"organization_name"`
				} `json:"declaration"`
			} `json:"declarations"`
		}
		if err := client.Get("/my/declarations.json", &result); err != nil {
			return err
		}

		if len(result.Declarations) == 0 {
			fmt.Println("No pending declarations.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG\tCOMPLETE AT")
		for _, d := range result.Declarations {
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
		if jsonOut {
			raw, err := client.GetRaw("/my/surveys.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Surveys []struct {
				Survey struct {
					Title            string `json:"title"`
					DueBy            string `json:"due_by"`
					OrganizationName string `json:"organization_name"`
				} `json:"survey"`
			} `json:"surveys"`
		}
		if err := client.Get("/my/surveys.json", &result); err != nil {
			return err
		}

		if len(result.Surveys) == 0 {
			fmt.Println("No pending surveys.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tDUE\tORG")
		for _, s := range result.Surveys {
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
