package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	Annotations: map[string]string{"skipAuth": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfg.Clear(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}
		fmt.Println("Logged out.")
		return nil
	},
}
