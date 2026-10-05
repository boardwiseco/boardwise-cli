package cmd

import (
	"fmt"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var boardsCmd = &cobra.Command{
	Use:   "boards",
	Short: "Manage boards and committees",
}

var boardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List boards and committees",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := api.BuildPath(orgSlug, "/groups.json")

		var groups []struct {
			Name        string `json:"name"`
			Type        string `json:"type"`
			MemberCount int    `json:"member_count"`
		}
		if printed, err := listAll(path, "groups", &groups); err != nil || printed {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "NAME\tTYPE\tMEMBERS")
		for _, g := range groups {
			fmt.Fprintf(w, "%s\t%s\t%d\n", g.Name, g.Type, g.MemberCount)
		}
		w.Flush()
		return nil
	},
}

var boardsGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get details of a board or committee",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := api.BuildPath(orgSlug, "/groups/"+args[0]+".json")

		if jsonOut {
			raw, err := client.GetRaw(path)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Members []struct {
				Name             string `json:"name"`
				GroupRole        string `json:"group_role"`
				OrganizationRole string `json:"organization_role"`
			} `json:"members"`
		}
		if err := client.Get(path, &result); err != nil {
			return err
		}

		fmt.Printf("Name: %s\nType: %s\n", result.Name, result.Type)
		if len(result.Members) > 0 {
			fmt.Println("\nMembers:")
			w := newTabWriter()
			fmt.Fprintln(w, "  NAME\tROLE\tORG ROLE")
			for _, m := range result.Members {
				fmt.Fprintf(w, "  %s\t%s\t%s\n", m.Name, m.GroupRole, m.OrganizationRole)
			}
			w.Flush()
		}
		return nil
	},
}

func init() {
	boardsCmd.AddCommand(boardsListCmd)
	boardsCmd.AddCommand(boardsGetCmd)
}
