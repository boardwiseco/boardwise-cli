package cmd

import (
	"fmt"
	"strings"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var meetingsCmd = &cobra.Command{
	Use:   "meetings",
	Short: "Manage meetings",
}

var meetingsListStatus string

var meetingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List meetings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := withQuery(api.BuildPath(orgSlug, "/meetings.json"), map[string]string{"status": meetingsListStatus})

		var meetings []struct {
			Title           string `json:"title"`
			StartsAt        string `json:"starts_at"`
			VisibilityState string `json:"visibility_state"`
			GroupName       string `json:"group_name"`
		}
		if printed, err := listAll(path, "meetings", &meetings); err != nil || printed {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "TITLE\tSTARTS AT\tSTATUS\tBOARD")
		for _, m := range meetings {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Title, m.StartsAt, m.VisibilityState, m.GroupName)
		}
		w.Flush()
		return nil
	},
}

var meetingsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get meeting details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := api.BuildPath(orgSlug, "/meetings/"+args[0]+".json")

		if jsonOut {
			raw, err := client.GetRaw(path)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var m struct {
			Title           string `json:"title"`
			StartsAt        string `json:"starts_at"`
			EndsAt          string `json:"ends_at"`
			VisibilityState string `json:"visibility_state"`
			Location        string `json:"location"`
			GroupName       string `json:"group_name"`
			AgendaItems     []struct {
				Position        int    `json:"position"`
				Title           string `json:"title"`
				DurationMinutes int    `json:"duration_minutes"`
			} `json:"agenda_items"`
			Attendees []struct {
				Name string `json:"name"`
				RSVP string `json:"rsvp"`
			} `json:"attendees"`
			Documents []struct {
				Name string `json:"name"`
			} `json:"documents"`
		}
		if err := client.Get(path, &m); err != nil {
			return err
		}

		fmt.Printf("Title:    %s\n", m.Title)
		if m.GroupName != "" {
			fmt.Printf("Board:    %s\n", m.GroupName)
		}
		fmt.Printf("Starts:   %s\n", m.StartsAt)
		fmt.Printf("Ends:     %s\n", m.EndsAt)
		fmt.Printf("Status:   %s\n", m.VisibilityState)
		if m.Location != "" {
			fmt.Printf("Location: %s\n", m.Location)
		}

		if len(m.AgendaItems) > 0 {
			fmt.Println("\nAgenda:")
			for _, item := range m.AgendaItems {
				fmt.Printf("  %d. %s (%d min)\n", item.Position, item.Title, item.DurationMinutes)
			}
		}

		if len(m.Attendees) > 0 {
			names := make([]string, len(m.Attendees))
			for i, a := range m.Attendees {
				names[i] = a.Name
			}
			fmt.Printf("\nAttendees: %s\n", strings.Join(names, ", "))
		}

		if len(m.Documents) > 0 {
			fmt.Println("\nDocuments:")
			for _, d := range m.Documents {
				fmt.Printf("  • %s\n", d.Name)
			}
		}
		return nil
	},
}

var (
	meetingTitle    string
	meetingStartsAt string
	meetingEndsAt   string
	meetingGroupID  string
	meetingLocation string
	meetingTimeZone string
)

var meetingsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new meeting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		meeting := map[string]any{
			"title":     meetingTitle,
			"starts_at": meetingStartsAt,
			"ends_at":   meetingEndsAt,
		}
		if meetingGroupID != "" {
			meeting["group_id"] = meetingGroupID
		}
		if meetingLocation != "" {
			meeting["location"] = meetingLocation
		}
		if meetingTimeZone != "" {
			meeting["time_zone"] = meetingTimeZone
		}

		body := map[string]any{"meeting": meeting}
		path := api.BuildPath(orgSlug, "/meetings.json")

		if jsonOut {
			raw, err := client.PostRaw(path, body)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Title string `json:"title"`
		}
		if err := client.Post(path, body, &result); err != nil {
			return err
		}

		fmt.Printf("Meeting created: %s\n", result.Title)
		return nil
	},
}

func init() {
	meetingsListCmd.Flags().StringVar(&meetingsListStatus, "status", "", "Filter: upcoming, past, or all")

	meetingsCreateCmd.Flags().StringVar(&meetingTitle, "title", "", "Meeting title (required)")
	meetingsCreateCmd.Flags().StringVar(&meetingStartsAt, "starts-at", "", "Start time ISO 8601 (required)")
	meetingsCreateCmd.Flags().StringVar(&meetingEndsAt, "ends-at", "", "End time ISO 8601 (required)")
	meetingsCreateCmd.Flags().StringVar(&meetingGroupID, "group", "", "Board or committee ID")
	meetingsCreateCmd.Flags().StringVar(&meetingLocation, "location", "", "Meeting location")
	meetingsCreateCmd.Flags().StringVar(&meetingTimeZone, "timezone", "", "IANA time zone (default: UTC)")
	meetingsCreateCmd.MarkFlagRequired("title")
	meetingsCreateCmd.MarkFlagRequired("starts-at")
	meetingsCreateCmd.MarkFlagRequired("ends-at")

	meetingsCmd.AddCommand(meetingsListCmd)
	meetingsCmd.AddCommand(meetingsGetCmd)
	meetingsCmd.AddCommand(meetingsCreateCmd)
}
