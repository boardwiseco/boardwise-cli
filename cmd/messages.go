package cmd

import (
	"fmt"

	"github.com/boardwise/cli/internal/api"
	"github.com/spf13/cobra"
)

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Read and send board messages",
}

var messagesListCmd = &cobra.Command{
	Use:   "list <group-slug>",
	Short: "List messages for a board or committee",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		groupSlug := args[0]
		path := api.BuildPath(orgSlug, "/groups/"+groupSlug+"/messages.json")

		if jsonOut {
			raw, err := client.GetRaw(path)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Messages []struct {
				Subject   string `json:"subject"`
				CreatedAt string `json:"created_at"`
			} `json:"messages"`
		}
		if err := client.Get(path, &result); err != nil {
			return err
		}

		if len(result.Messages) == 0 {
			fmt.Println("No messages.")
			return nil
		}

		w := newTabWriter()
		fmt.Fprintln(w, "SUBJECT\tDATE")
		for _, m := range result.Messages {
			fmt.Fprintf(w, "%s\t%s\n", m.Subject, m.CreatedAt)
		}
		w.Flush()
		return nil
	},
}

var messagesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get message details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		path := api.BuildPath(orgSlug, "/messages/"+args[0]+".json")

		if jsonOut {
			raw, err := client.GetRaw(path)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var m struct {
			Subject   string `json:"subject"`
			Body      string `json:"body"`
			CreatedAt string `json:"created_at"`
		}
		if err := client.Get(path, &m); err != nil {
			return err
		}

		fmt.Printf("Subject: %s\n", m.Subject)
		fmt.Printf("Date:    %s\n", m.CreatedAt)
		fmt.Printf("\n%s\n", m.Body)
		return nil
	},
}

var (
	msgSubject string
	msgBody    string
)

var messagesSendCmd = &cobra.Command{
	Use:   "send <group-slug>",
	Short: "Send a message to a board or committee",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireOrg(); err != nil {
			return err
		}

		groupSlug := args[0]
		path := api.BuildPath(orgSlug, "/groups/"+groupSlug+"/messages.json")

		body := map[string]any{
			"message": map[string]string{
				"subject": msgSubject,
				"body":    msgBody,
			},
		}

		if jsonOut {
			raw, err := client.PostRaw(path, body)
			if err != nil {
				return err
			}
			fmt.Println(string(raw))
			return nil
		}

		var result struct {
			Subject string `json:"subject"`
		}
		if err := client.Post(path, body, &result); err != nil {
			return err
		}

		fmt.Printf("Message sent: %s\n", result.Subject)
		return nil
	},
}

func init() {
	messagesSendCmd.Flags().StringVar(&msgSubject, "subject", "", "Message subject (required)")
	messagesSendCmd.Flags().StringVar(&msgBody, "body", "", "Message body (HTML supported, required)")
	messagesSendCmd.MarkFlagRequired("subject")
	messagesSendCmd.MarkFlagRequired("body")

	messagesCmd.AddCommand(messagesListCmd)
	messagesCmd.AddCommand(messagesGetCmd)
	messagesCmd.AddCommand(messagesSendCmd)
}
