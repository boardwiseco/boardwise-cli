package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:         "logout",
	Short:       "Revoke your token and clear stored credentials",
	Annotations: map[string]string{"skipAuth": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Token == "" {
			if err := cfg.Clear(); err != nil {
				return fmt.Errorf("failed to clear config: %w", err)
			}
			fmt.Println("Logged out.")
			return nil
		}

		// Revoke first, while we still have the token; clear local
		// credentials whatever the outcome.
		revokeErr := client.Delete("/api/v1/auth")

		if err := cfg.Clear(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}

		var apiErr *api.APIError
		switch {
		case revokeErr == nil:
			fmt.Println("Logged out. Your token has been revoked.")
		case errors.As(revokeErr, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized:
			fmt.Println("Logged out. Your token had already been revoked or had expired.")
		default:
			fmt.Fprintf(os.Stderr, "Warning: could not revoke your token on the server: %v\n", revokeErr)
			fmt.Fprintf(os.Stderr, "It may still be valid. Revoke it from your API tokens page: %s/my/api_tokens\n",
				strings.TrimRight(client.BaseURL, "/"))
			fmt.Println("Logged out on this computer.")
		}
		return nil
	},
}
