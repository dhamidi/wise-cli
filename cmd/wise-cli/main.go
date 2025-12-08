package main

import (
	"fmt"
	"os"

	"github.com/dhamidi/wise-cli/queries"
	"github.com/spf13/cobra"
)

var (
	apiToken string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "wise-cli",
		Short: "Wise CLI tool",
		Long:  "A command-line interface for Wise API",
	}

	rootCmd.PersistentFlags().StringVar(&apiToken, "token", os.Getenv("WISE_API_TOKEN"), "Wise API token (or set WISE_API_TOKEN env var)")

	rootCmd.AddCommand(recipientsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var recipientsCmd = &cobra.Command{
	Use:   "recipients",
	Short: "List recipient accounts",
	Long:  "Fetch a list of your recipient accounts from Wise",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			return fmt.Errorf("API token required: set --token flag or WISE_API_TOKEN env var")
		}

		profileID, _ := cmd.Flags().GetInt("profile-id")
		currency, _ := cmd.Flags().GetString("currency")
		typeFilter, _ := cmd.Flags().GetString("type")
		size, _ := cmd.Flags().GetInt("size")

		req := queries.ListRecipientsRequest{
			ProfileID: profileID,
			Currency:  currency,
			Type:      typeFilter,
			Size:      size,
		}

		recipients, err := queries.ListRecipients(apiToken, req)
		if err != nil {
			return fmt.Errorf("failed to list recipients: %w", err)
		}

		if len(recipients) == 0 {
			fmt.Println("No recipients found")
			return nil
		}

		// Format output
		fmt.Printf("%-10s %-30s %-10s %-15s %-10s\n", "ID", "Name", "Currency", "Country", "Type")
		fmt.Println(string(make([]byte, 80)))

		for _, r := range recipients {
			name := r.Name.FullName
			if name == "" {
				name = "N/A"
			}
			fmt.Printf("%-10d %-30s %-10s %-15s %-10s\n",
				r.ID,
				name,
				r.Currency,
				r.Country,
				r.Type,
			)
		}

		return nil
	},
}

func init() {
	recipientsCmd.Flags().IntP("profile-id", "p", 0, "Profile ID to filter by")
	recipientsCmd.Flags().StringP("currency", "c", "", "Filter by currency (e.g. USD,GBP)")
	recipientsCmd.Flags().StringP("type", "t", "", "Filter by account type (e.g. iban,swift_code)")
	recipientsCmd.Flags().IntP("size", "s", 20, "Number of results to return")
}
