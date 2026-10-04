package cmd

import (
	"fmt"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var peopleCmd = &cobra.Command{
	Use:   "people",
	Short: "Manage people in the organization",
}

var peopleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active people in the organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := api.BuildPath(orgSlug, "/people.json")

		var people []struct {
			Name      string `json:"name"`
			Email     string `json:"email"`
			RoleLevel string `json:"role_level"`
		}
		if printed, err := listAll(path, "people", &people); err != nil || printed {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "NAME\tEMAIL\tROLE")
		for _, p := range people {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Email, p.RoleLevel)
		}
		w.Flush()
		return nil
	},
}

func init() {
	peopleCmd.AddCommand(peopleListCmd)
}
