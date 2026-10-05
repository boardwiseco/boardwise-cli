package cmd

import (
	"fmt"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Browse documents",
}

var docsGroupID string

var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := withQuery(api.BuildPath(orgSlug, "/documents.json"), map[string]string{"group_id": docsGroupID})

		var result []struct {
			ID          any    `json:"id"`
			Name        string `json:"name"`
			ContentType string `json:"content_type"`
			CreatedAt   string `json:"created_at"`
		}
		if printed, err := listAll(path, "", &result); err != nil || printed {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tCREATED")
		for _, d := range result {
			fmt.Fprintf(w, "%v\t%s\t%s\t%s\n", d.ID, d.Name, d.ContentType, d.CreatedAt)
		}
		w.Flush()
		return nil
	},
}

func init() {
	docsListCmd.Flags().StringVar(&docsGroupID, "group", "", "Filter by board or committee ID")
	docsCmd.AddCommand(docsListCmd)
}
