package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var superadminCmd = &cobra.Command{
	Use:    "superadmin",
	Short:  "Superadmin operations",
	Hidden: true, // unhidden at startup if cfg.Superadmin == true
}

var superadminOrgsCmd = &cobra.Command{
	Use:   "orgs",
	Short: "Manage all organizations",
}

var superadminOrgsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all organizations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOut {
			raw, err := client.GetRaw("/superadmin/organizations.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result []struct {
			ID   any    `json:"id"`
			Name      string `json:"name"`
			Slug      string `json:"slug"`
			CreatedAt string `json:"created_at"`
		}
		if err := client.Get("/superadmin/organizations.json", &result); err != nil {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "ID\tSLUG\tNAME\tCREATED")
		for _, o := range result {
			fmt.Fprintf(w, "%v\t%s\t%s\t%s\n", o.ID, o.Slug, o.Name, o.CreatedAt)
		}
		w.Flush()
		return nil
	},
}

var superadminOrgsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get organization details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/superadmin/organizations/" + args[0] + ".json"

		if jsonOut {
			raw, err := client.GetRaw(path)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var o struct {
			ID   any    `json:"id"`
			Name      string `json:"name"`
			Slug      string `json:"slug"`
			CreatedAt string `json:"created_at"`
			PeopleCount int  `json:"people_count"`
		}
		if err := client.Get(path, &o); err != nil {
			return err
		}

		fmt.Printf("ID:      %v\n", o.ID)
		fmt.Printf("Name:    %s\n", o.Name)
		fmt.Printf("Slug:    %s\n", o.Slug)
		fmt.Printf("Created: %s\n", o.CreatedAt)
		fmt.Printf("People:  %%v\n", o.PeopleCount)
		return nil
	},
}

var superadminUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOut {
			raw, err := client.GetRaw("/superadmin/users.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result []struct {
			ID   any    `json:"id"`
			Email     string `json:"email"`
			GivenName string `json:"given_name"`
			FamilyName string `json:"family_name"`
			CreatedAt string `json:"created_at"`
		}
		if err := client.Get("/superadmin/users.json", &result); err != nil {
			return err
		}

		w := newTabWriter()
		fmt.Fprintln(w, "ID\tNAME\tEMAIL\tCREATED")
		for _, u := range result {
			fmt.Fprintf(w, "%v\t%s %s\t%s\t%s\n", u.ID, u.GivenName, u.FamilyName, u.Email, u.CreatedAt)
		}
		w.Flush()
		return nil
	},
}

var superadminSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Show system stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOut {
			raw, err := client.GetRaw("/superadmin/system.json")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var s struct {
			Database struct {
				Size string `json:"size"`
			} `json:"database"`
			Jobs struct {
				Queued    int `json:"queued"`
				Running   int `json:"running"`
				Failed    int `json:"failed"`
			} `json:"jobs"`
		}
		if err := client.Get("/superadmin/system.json", &s); err != nil {
			return err
		}

		fmt.Printf("Database size: %s\n", s.Database.Size)
		fmt.Printf("Jobs queued:   %%v\n", s.Jobs.Queued)
		fmt.Printf("Jobs running:  %%v\n", s.Jobs.Running)
		fmt.Printf("Jobs failed:   %%v\n", s.Jobs.Failed)
		return nil
	},
}

func init() {
	superadminOrgsCmd.AddCommand(superadminOrgsListCmd)
	superadminOrgsCmd.AddCommand(superadminOrgsGetCmd)

	superadminCmd.AddCommand(superadminOrgsCmd)
	superadminCmd.AddCommand(superadminUsersCmd)
	superadminCmd.AddCommand(superadminSystemCmd)
}
