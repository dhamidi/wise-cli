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
	rootCmd.AddCommand(quoteCmd)

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

var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Get exchange quote",
	Long:  "Create a quote for a currency conversion",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			return fmt.Errorf("API token required: set --token flag or WISE_API_TOKEN env var")
		}

		profileID, _ := cmd.Flags().GetInt("profile-id")
		sourceCurrency, _ := cmd.Flags().GetString("source-currency")
		targetCurrency, _ := cmd.Flags().GetString("target-currency")
		sourceAmount, _ := cmd.Flags().GetFloat64("source-amount")
		targetAmount, _ := cmd.Flags().GetFloat64("target-amount")

		if profileID == 0 {
			return fmt.Errorf("profile-id is required")
		}
		if sourceCurrency == "" {
			return fmt.Errorf("source-currency is required")
		}
		if targetCurrency == "" {
			return fmt.Errorf("target-currency is required")
		}
		if sourceAmount == 0 && targetAmount == 0 {
			return fmt.Errorf("either source-amount or target-amount is required")
		}
		if sourceAmount != 0 && targetAmount != 0 {
			return fmt.Errorf("only one of source-amount or target-amount can be specified")
		}

		req := queries.GetQuoteRequest{
			ProfileID:      profileID,
			SourceCurrency: sourceCurrency,
			TargetCurrency: targetCurrency,
		}

		if sourceAmount != 0 {
			req.SourceAmount = &sourceAmount
		}
		if targetAmount != 0 {
			req.TargetAmount = &targetAmount
		}

		quote, err := queries.GetQuote(apiToken, req)
		if err != nil {
			return fmt.Errorf("failed to get quote: %w", err)
		}

		// Format output
		fmt.Println("Quote Details:")
		fmt.Println("==============")
		fmt.Printf("Quote ID:           %s\n", quote.ID)
		fmt.Printf("Status:             %s\n", quote.Status)
		fmt.Printf("Source:             %.2f %s\n", quote.SourceAmount, quote.SourceCurrency)
		fmt.Printf("Target:             %.2f %s\n", quote.TargetAmount, quote.TargetCurrency)
		fmt.Printf("Exchange Rate:      %.6f\n", quote.Rate)
		fmt.Printf("Rate Type:          %s\n", quote.RateType)
		fmt.Printf("Created:            %s\n", quote.CreatedTime)
		fmt.Printf("Rate Expires:       %s\n", quote.RateExpirationTime)
		fmt.Printf("Quote Expires:      %s\n", quote.ExpirationTime)

		if len(quote.PaymentOptions) > 0 {
			fmt.Println("\nPayment Options:")
			for i, opt := range quote.PaymentOptions {
				if opt.Disabled {
					fmt.Printf("\n[%d] %s → %s (DISABLED)\n", i+1, opt.PayIn, opt.PayOut)
					if opt.DisabledReason != nil {
						fmt.Printf("    Reason: %s\n", opt.DisabledReason.Message)
					}
				} else {
					fmt.Printf("\n[%d] %s → %s\n", i+1, opt.PayIn, opt.PayOut)
					fmt.Printf("    Source:     %.2f %s\n", opt.SourceAmount, sourceCurrency)
					fmt.Printf("    Target:     %.2f %s\n", opt.TargetAmount, targetCurrency)
					fmt.Printf("    Fee:        %.2f %s\n", opt.Fee.Total, sourceCurrency)
					if opt.FormattedEstimatedDelivery != "" {
						fmt.Printf("    Delivery:   %s\n", opt.FormattedEstimatedDelivery)
					}
				}
			}
		}

		if len(quote.Notices) > 0 {
			fmt.Println("\nNotices:")
			for _, notice := range quote.Notices {
				fmt.Printf("[%s] %s\n", notice.Type, notice.Text)
				if notice.Link != nil {
					fmt.Printf("    Link: %s\n", *notice.Link)
				}
			}
		}

		return nil
	},
}

func init() {
	recipientsCmd.Flags().IntP("profile-id", "p", 0, "Profile ID to filter by")
	recipientsCmd.Flags().StringP("currency", "c", "", "Filter by currency (e.g. USD,GBP)")
	recipientsCmd.Flags().StringP("type", "t", "", "Filter by account type (e.g. iban,swift_code)")
	recipientsCmd.Flags().IntP("size", "s", 20, "Number of results to return")

	quoteCmd.Flags().IntP("profile-id", "p", 0, "Profile ID (required)")
	quoteCmd.MarkFlagRequired("profile-id")
	quoteCmd.Flags().StringP("source-currency", "s", "", "Source currency code (required)")
	quoteCmd.MarkFlagRequired("source-currency")
	quoteCmd.Flags().StringP("target-currency", "t", "", "Target currency code (required)")
	quoteCmd.MarkFlagRequired("target-currency")
	quoteCmd.Flags().Float64("source-amount", 0, "Amount in source currency (either this or target-amount)")
	quoteCmd.Flags().Float64("target-amount", 0, "Amount in target currency (either this or source-amount)")
}
