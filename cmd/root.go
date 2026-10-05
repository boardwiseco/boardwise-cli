package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/boardwise/cli/internal/api"
	"github.com/boardwise/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	client  *api.Client
	orgSlug string
	jsonOut bool
	baseURL string
)

var rootCmd = &cobra.Command{
	Use:   "bw",
	Short: "Boardwise CLI",
	Long:  "Interact with your Boardwise boards, meetings, tasks, and more.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Commands that don't need auth skip this
		if cmd.Annotations["skipAuth"] == "true" {
			return nil
		}
		if cfg.Token == "" {
			return fmt.Errorf("not logged in — run: bw login")
		}
		return nil
	},
}

func Execute() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Resolve base URL: flag > env > config > default
	if baseURL == "" {
		baseURL = os.Getenv("BW_API_URL")
	}
	if baseURL == "" {
		baseURL = cfg.URL
	}
	client = api.NewClient(baseURL, cfg.Token)

	// Set default org from config if not provided via flag
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Annotations["skipAuth"] == "true" {
			return nil
		}
		if cfg.Token == "" {
			return fmt.Errorf("not logged in — run: bw login")
		}
		if orgSlug == "" {
			orgSlug = cfg.DefaultOrg
		}
		return nil
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&orgSlug, "org", "o", "", "organization slug (or set default with BW_ORG)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output as JSON")
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "", "API base URL (default https://app.boardwise.co)")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(meCmd)
	rootCmd.AddCommand(boardsCmd)
	rootCmd.AddCommand(meetingsCmd)
	rootCmd.AddCommand(agendaCmd)
	rootCmd.AddCommand(peopleCmd)
	rootCmd.AddCommand(tasksCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(messagesCmd)
	rootCmd.AddCommand(myCmd)
}

// requireOrg validates that --org is set for org-scoped commands.
func requireOrg() error {
	if orgSlug == "" {
		return fmt.Errorf("--org <slug> is required (or set a default with: bw config set-org <slug>)")
	}
	return nil
}

// withQuery appends the non-empty params to path as a query string.
func withQuery(path string, params map[string]string) string {
	query := url.Values{}
	for name, value := range params {
		if value != "" {
			query.Set(name, value)
		}
	}
	if len(query) == 0 {
		return path
	}
	return path + "?" + query.Encode()
}

// listAll reads every page of a paginated list. key names the field that
// holds the list in each page ("" when the page is a bare array).
//
// With --json it prints every item in the API's own shape ({key: [...]} or
// [...]) and returns printed=true. Otherwise it decodes the items into
// items, a pointer to a slice.
func listAll(path, key string, items any) (printed bool, err error) {
	all := []json.RawMessage{}
	err = client.EachPage(path, func(page []byte) error {
		raw := page
		if key != "" {
			var body map[string]json.RawMessage
			if err := json.Unmarshal(page, &body); err != nil {
				return fmt.Errorf("unexpected response from %s: %w", path, err)
			}
			raw = body[key]
		}
		var pageItems []json.RawMessage
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &pageItems); err != nil {
				return fmt.Errorf("unexpected response from %s: %w", path, err)
			}
		}
		all = append(all, pageItems...)
		return nil
	})
	if err != nil {
		return false, err
	}

	if jsonOut {
		if key == "" {
			return true, printJSON(all)
		}
		return true, printJSON(map[string]any{key: all})
	}

	data, err := json.Marshal(all)
	if err != nil {
		return false, err
	}
	return false, json.Unmarshal(data, items)
}

// printJSON pretty-prints v as JSON.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// newTabWriter returns a tabwriter for aligned table output.
func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
}
