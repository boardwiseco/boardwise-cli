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
	Use:         "login",
	Short:       "Log in to Boardwise via browser authorization",
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
		if err := unauthClient.Post("/api/v1/auth/device", deviceRequest(loginReadOnly, orgSlug), &deviceResp); err != nil {
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

		accessToken, err := pollForToken(unauthClient, deviceResp.DeviceCode, interval, time.Now().Add(timeout))
		if err != nil {
			return err
		}

		// Got the token: fetch user info
		meClient := api.NewClient(baseURL, accessToken)
		var me struct {
			Email         string `json:"email"`
			GivenName     string `json:"given_name"`
			FamilyName    string `json:"family_name"`
			Organizations []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"organizations"`
		}
		if err := meClient.Get("/api/v1/auth/me", &me); err != nil {
			return fmt.Errorf("failed to fetch user info: %w", err)
		}

		cfg.Token = accessToken
		cfg.URL = baseURL

		// A token bound with --org makes that org the default; otherwise
		// auto-set it if the user only belongs to one.
		if orgSlug != "" {
			cfg.DefaultOrg = orgSlug
		} else if len(me.Organizations) == 1 {
			cfg.DefaultOrg = me.Organizations[0].Slug
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("\nLogged in as %s %s (%s)\n", me.GivenName, me.FamilyName, me.Email)
		if loginReadOnly {
			fmt.Println("This token is read-only: it can view but not change anything.")
		}
		if orgSlug != "" {
			fmt.Printf("This token is limited to organization %s; the bw my commands, which span organizations, are refused for it.\n", orgSlug)
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
		return nil
	},
}

var loginReadOnly bool

func init() {
	loginCmd.Flags().BoolVar(&loginReadOnly, "read-only", false, "ask for a read-only token (it can view but not change anything)")
	// --org is the global flag: on login it limits the token to that
	// organization and makes it the default.
}

// deviceRequest is the body that starts the device flow. A read-only
// request asks for scope "read"; an organization binds the token to it.
func deviceRequest(readOnly bool, org string) map[string]string {
	body := map[string]string{"client_name": "Boardwise CLI"}
	if readOnly {
		body["scope"] = "read"
	}
	if org != "" {
		body["organization_slug"] = org
	}
	return body
}

// deviceCodeGrant is the OAuth grant type for collecting a device-flow token
// (RFC 8628).
const deviceCodeGrant = "urn:ietf:params:oauth:grant-type:device_code"

// pollForToken polls the OAuth token endpoint until the person approves in
// the browser, and returns the access token. Every refusal is a 400 whose
// "error" says what happened: authorization_pending means keep polling;
// anything else ends the attempt.
func pollForToken(client *api.Client, deviceCode string, interval time.Duration, deadline time.Time) (string, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if time.Now().After(deadline) {
			return "", fmt.Errorf("authorization timed out")
		}

		var tokenResp struct {
			AccessToken string `json:"access_token"`
		}
		err := client.Post("/api/v1/auth/token", map[string]string{
			"grant_type":  deviceCodeGrant,
			"device_code": deviceCode,
		}, &tokenResp)

		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) {
				switch apiErr.Code() {
				case "authorization_pending":
					continue
				case "slow_down": // RFC 8628: poll 5 seconds less often from now on
					interval += 5 * time.Second
					ticker.Reset(interval)
					continue
				case "access_denied":
					return "", fmt.Errorf("authorization denied")
				case "expired_token":
					return "", fmt.Errorf("authorization code expired, run bw login again")
				}
				if apiErr.StatusCode == 429 {
					continue
				}
			}
			return "", fmt.Errorf("polling error: %w", err)
		}

		if tokenResp.AccessToken != "" {
			return tokenResp.AccessToken, nil
		}
	}
	return "", fmt.Errorf("authorization timed out")
}
