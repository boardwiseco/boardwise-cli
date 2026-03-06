package cmd

import (
	"fmt"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var agendaCmd = &cobra.Command{
	Use:   "agenda",
	Short: "Manage agenda items",
}

var (
	agendaTitle       string
	agendaDuration    int
	agendaDescription string
	agendaPosition    int
)

var agendaAddCmd = &cobra.Command{
	Use:   "add <meeting-id>",
	Short: "Add an agenda item to a meeting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		body := map[string]any{
			"title":            agendaTitle,
			"duration_minutes": agendaDuration,
		}
		if agendaDescription != "" {
			body["description"] = agendaDescription
		}
		if agendaPosition > 0 {
			body["position"] = agendaPosition
		}

		path := api.BuildPath(orgSlug, "/meetings/"+args[0]+"/agenda_items.json")

		if jsonOut {
			raw, err := client.PostRaw(path, body)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			ID   any    `json:"id"`
			Title    string `json:"title"`
			Position int    `json:"position"`
		}
		if err := client.Post(path, body, &result); err != nil {
			return err
		}

		fmt.Printf("Agenda item added: %s (position %d)\n", result.Title, result.Position)
		return nil
	},
}

func init() {
	agendaAddCmd.Flags().StringVar(&agendaTitle, "title", "", "Agenda item title (required)")
	agendaAddCmd.Flags().IntVar(&agendaDuration, "duration", 0, "Duration in minutes (required)")
	agendaAddCmd.Flags().StringVar(&agendaDescription, "description", "", "Agenda item description")
	agendaAddCmd.Flags().IntVar(&agendaPosition, "position", 0, "Position in agenda (auto-assigned if omitted)")
	agendaAddCmd.MarkFlagRequired("title")
	agendaAddCmd.MarkFlagRequired("duration")

	agendaCmd.AddCommand(agendaAddCmd)
}
