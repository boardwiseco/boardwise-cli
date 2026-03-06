package cmd

import (
	"fmt"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage action items (tasks)",
}

var tasksListStatus string

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List action items",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := api.BuildPath(orgSlug, "/action_items.json")
		if tasksListStatus != "" {
			path += "?status=" + tasksListStatus
		}

		if jsonOut {
			raw, err := client.GetRaw(path)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result []struct {
			ID   any    `json:"id"`
			Title   string `json:"title"`
			DueBy   string `json:"due_by"`
			Status  string `json:"status"`
			Assigned []struct {
				Name string `json:"name"`
			} `json:"assigned"`
		}
		if err := client.Get(path, &result); err != nil {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "ID\tTITLE\tDUE\tSTATUS")
		for _, t := range result {
			fmt.Fprintf(w, "%v\t%s\t%s\t%s\n", t.ID, t.Title, t.DueBy, t.Status)
		}
		w.Flush()
		return nil
	},
}

var (
	taskTitle   string
	taskDueBy   string
	taskGroupID string
	taskAssign  []string
)

var tasksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new action item",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		body := map[string]any{
			"title": taskTitle,
		}
		if taskDueBy != "" {
			body["due_by"] = taskDueBy
		}
		if taskGroupID != "" {
			body["group_id"] = taskGroupID
		}
		if len(taskAssign) > 0 {
			body["assigned_person_ids"] = taskAssign
		}

		path := api.BuildPath(orgSlug, "/action_items.json")

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
			Title string `json:"title"`
		}
		if err := client.Post(path, body, &result); err != nil {
			return err
		}

		fmt.Printf("Action item created: %s (ID: %v)\n", result.Title, result.ID)
		return nil
	},
}

func init() {
	tasksListCmd.Flags().StringVar(&tasksListStatus, "status", "pending", "Filter: pending, completed, or all")

	tasksCreateCmd.Flags().StringVar(&taskTitle, "title", "", "Action item title (required)")
	tasksCreateCmd.Flags().StringVar(&taskDueBy, "due", "", "Due date (ISO 8601)")
	tasksCreateCmd.Flags().StringVar(&taskGroupID, "group", "", "Board or committee ID")
	tasksCreateCmd.Flags().StringArrayVar(&taskAssign, "assign", nil, "Person IDs to assign (can repeat)")
	tasksCreateCmd.MarkFlagRequired("title")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksCreateCmd)
}
