package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOut {
			raw, err := client.GetRaw("/api/v1/auth/me")
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var me struct {
			Email      string `json:"email"`
			GivenName  string `json:"given_name"`
			FamilyName string `json:"family_name"`
			Superadmin bool   `json:"superadmin"`
			Organizations []struct {
				Name string `json:"name"`
				Slug any    `json:"slug"`
			} `json:"organizations"`
		}
		if err := client.Get("/api/v1/auth/me", &me); err != nil {
			return err
		}

		fmt.Printf("Name:  %s %s\n", me.GivenName, me.FamilyName)
		fmt.Printf("Email: %s\n", me.Email)
		if me.Superadmin {
			fmt.Println("Role:  superadmin")
		}
		if len(me.Organizations) > 0 {
			fmt.Println("\nOrganizations:")
			w := newTabWriter()
			fmt.Fprintln(w, "  SLUG\tNAME")
			for _, org := range me.Organizations {
				fmt.Fprintf(w, "  %s\t%s\n", org.Slug, org.Name)
			}
			w.Flush()
		}
		return nil
	},
}
