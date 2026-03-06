package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Boardwise via browser authorization",
	Annotations: map[string]string{"skipAuth": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		unauthClient := api.NewClient(baseURL, "")

		// Step 1: start device flow
		var deviceResp struct {
			DeviceCode      string `json:"device_code"`
			UserCode        string `json:"user_code"`
			VerificationURI string `json:"verification_uri"`
			ExpiresIn       int    `json:"expires_in"`
			Interval        int    `json:"interval"`
		}
		if err := unauthClient.Post("/api/v1/auth/device", map[string]string{
			"client_name": "Boardwise CLI",
		}, &deviceResp); err != nil {
			return fmt.Errorf("failed to start device authorization: %w", err)
		}

		fmt.Printf("\nOpen this URL in your browser:\n\n  %s\n\n", deviceResp.VerificationURI)
		fmt.Printf("Then enter this code when prompted:\n\n  %s\n\n", deviceResp.UserCode)
		fmt.Println("Waiting for authorization...")

		interval := time.Duration(deviceResp.Interval) * time.Second
		if interval == 0 {
			interval = 5 * time.Second
		}
		timeout := time.Duration(deviceResp.ExpiresIn) * time.Second
		if timeout == 0 {
			timeout = 15 * time.Minute
		}

		deadline := time.Now().Add(timeout)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if time.Now().After(deadline) {
				return fmt.Errorf("authorization timed out")
			}

			var tokenResp struct {
				AccessToken string `json:"access_token"`
				TokenType   string `json:"token_type"`
				Error       string `json:"error"`
			}

			err := unauthClient.Post("/api/v1/auth/device/token", map[string]string{
				"device_code": deviceResp.DeviceCode,
			}, &tokenResp)

			if err != nil {
				var apiErr *api.APIError
				if errors.As(err, &apiErr) {
					switch apiErr.StatusCode {
					case 428: // authorization_pending
						continue
					case 403: // access_denied
						return fmt.Errorf("authorization denied")
					case 410: // expired
						return fmt.Errorf("authorization code expired")
					}
				}
				return fmt.Errorf("polling error: %w", err)
			}

			if tokenResp.AccessToken == "" {
				continue
			}

			// Got the token — fetch user info
			meClient := api.NewClient(baseURL, tokenResp.AccessToken)
			var me struct {
				Email      string `json:"email"`
				GivenName  string `json:"given_name"`
				FamilyName string `json:"family_name"`
				Superadmin bool   `json:"superadmin"`
				Organizations []struct {
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"organizations"`
			}
			if err := meClient.Get("/api/v1/auth/me", &me); err != nil {
				return fmt.Errorf("failed to fetch user info: %w", err)
			}

			cfg.Token = tokenResp.AccessToken
			cfg.Superadmin = me.Superadmin
			cfg.URL = baseURL

			// Auto-set default org if user only belongs to one
			if len(me.Organizations) == 1 {
				cfg.DefaultOrg = me.Organizations[0].Slug
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("\nLogged in as %s %s (%s)\n", me.GivenName, me.FamilyName, me.Email)
			if me.Superadmin {
				fmt.Println("Superadmin commands are now available.")
			}
			if cfg.DefaultOrg != "" {
				fmt.Printf("Default org set to: %s\n", cfg.DefaultOrg)
			} else if len(me.Organizations) > 1 {
				fmt.Println("\nYour organizations:")
				for _, org := range me.Organizations {
					fmt.Printf("  %s  %s\n", org.Slug, org.Name)
				}
				fmt.Println("\nSet a default with: bw --org <slug> (or use --org on each command)")
			}

			os.Exit(0)
		}
		return nil
	},
}
