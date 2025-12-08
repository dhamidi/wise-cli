package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dhamidi/wise-cli/commands"
	"github.com/dhamidi/wise-cli/config"
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

	// Try to load token from environment, then from cache
	tokenDefault := os.Getenv("WISE_API_TOKEN")
	if tokenDefault == "" {
		if cached, err := config.LoadToken(); err == nil && cached != "" {
			tokenDefault = cached
		}
	}

	rootCmd.PersistentFlags().StringVar(&apiToken, "token", tokenDefault, "Wise API token (or set WISE_API_TOKEN env var)")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(profilesCmd)
	rootCmd.AddCommand(selectProfileCmd)
	rootCmd.AddCommand(recipientsCmd)
	rootCmd.AddCommand(quoteCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(sendToCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List profiles",
	Long:  "Fetch a list of all profiles belonging to your Wise account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			return fmt.Errorf("API token required: set --token flag or WISE_API_TOKEN env var")
		}

		profiles, err := queries.ListProfiles(apiToken)
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles found")
			return nil
		}

		// Format output
		fmt.Printf("%-10s %-15s %-10s %-25s %-20s\n", "ID", "Type", "State", "Email", "Name")
		fmt.Println(string(make([]byte, 85)))

		for _, p := range profiles {
			name := ""
			if p.FirstName != nil && p.LastName != nil {
				name = *p.FirstName + " " + *p.LastName
			} else if p.BusinessName != nil {
				name = *p.BusinessName
			}
			if name == "" {
				name = "N/A"
			}

			fmt.Printf("%-10d %-15s %-10s %-25s %-20s\n",
				p.ID,
				p.Type,
				p.CurrentState,
				p.Email,
				name,
			)
		}

		return nil
	},
}

var selectProfileCmd = &cobra.Command{
	Use:   "select-profile <profile-id-or-name>",
	Short: "Select default profile",
	Long:  "Select a profile by ID or name to use as the default profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			return fmt.Errorf("API token required: set --token flag or WISE_API_TOKEN env var")
		}

		profileIdOrName := args[0]

		profiles, err := queries.ListProfiles(apiToken)
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		var selectedProfile *queries.Profile

		// Try to parse as ID first
		profileID := 0
		if _, err := fmt.Sscanf(profileIdOrName, "%d", &profileID); err == nil && profileID > 0 {
			for i := range profiles {
				if profiles[i].ID == profileID {
					selectedProfile = &profiles[i]
					break
				}
			}
		}

		// If not found by ID, try by exact name match
		if selectedProfile == nil {
			for i := range profiles {
				name := ""
				if profiles[i].FirstName != nil && profiles[i].LastName != nil {
					name = *profiles[i].FirstName + " " + *profiles[i].LastName
				} else if profiles[i].BusinessName != nil {
					name = *profiles[i].BusinessName
				}

				if name == profileIdOrName {
					selectedProfile = &profiles[i]
					break
				}
			}
		}

		// If still not found, try by substring match
		if selectedProfile == nil {
			for i := range profiles {
				name := ""
				if profiles[i].FirstName != nil && profiles[i].LastName != nil {
					name = *profiles[i].FirstName + " " + *profiles[i].LastName
				} else if profiles[i].BusinessName != nil {
					name = *profiles[i].BusinessName
				}

				if strings.Contains(strings.ToLower(name), strings.ToLower(profileIdOrName)) {
					selectedProfile = &profiles[i]
					break
				}
			}
		}

		if selectedProfile == nil {
			return fmt.Errorf("profile not found: %s", profileIdOrName)
		}

		// Determine profile name for display
		displayName := ""
		if selectedProfile.FirstName != nil && selectedProfile.LastName != nil {
			displayName = *selectedProfile.FirstName + " " + *selectedProfile.LastName
		} else if selectedProfile.BusinessName != nil {
			displayName = *selectedProfile.BusinessName
		} else {
			displayName = selectedProfile.Email
		}

		// Save selected profile
		selectedData := config.SelectedProfile{
			ID:   selectedProfile.ID,
			Name: displayName,
			Type: selectedProfile.Type,
		}

		if err := config.SaveSelectedProfile(selectedData); err != nil {
			return fmt.Errorf("failed to save selected profile: %w", err)
		}

		fmt.Printf("✓ Selected profile: %s (ID: %d, Type: %s)\n", displayName, selectedProfile.ID, selectedProfile.Type)
		return nil
	},
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

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new resources",
	Long:  "Create new resources like quotes and transfers",
}

var newTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Create transfer",
	Long:  "Create a transfer based on a quote",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			return fmt.Errorf("API token required: set --token flag or WISE_API_TOKEN env var")
		}

		targetAccount, _ := cmd.Flags().GetInt("target-account")
		quoteUUID, _ := cmd.Flags().GetString("quote-uuid")
		customerTransactionID, _ := cmd.Flags().GetString("customer-transaction-id")
		reference, _ := cmd.Flags().GetString("reference")
		sourceAccount, _ := cmd.Flags().GetInt("source-account")

		if targetAccount == 0 {
			return fmt.Errorf("target-account is required")
		}
		if quoteUUID == "" {
			return fmt.Errorf("quote-uuid is required")
		}
		if customerTransactionID == "" {
			return fmt.Errorf("customer-transaction-id is required")
		}

		req := commands.NewTransferRequest{
			TargetAccount:         targetAccount,
			QuoteUUID:             quoteUUID,
			CustomerTransactionID: customerTransactionID,
		}

		if reference != "" {
			req.Reference = &reference
		}
		if sourceAccount != 0 {
			req.SourceAccount = &sourceAccount
		}

		transfer, err := commands.NewTransfer(apiToken, req)
		if err != nil {
			return fmt.Errorf("failed to create transfer: %w", err)
		}

		// Format output
		fmt.Println("Transfer Created:")
		fmt.Println("=================")
		fmt.Printf("ID:                      %d\n", transfer.ID)
		fmt.Printf("Status:                  %s\n", transfer.Status)
		fmt.Printf("Source:                  %.2f %s\n", transfer.SourceValue, transfer.SourceCurrency)
		fmt.Printf("Target:                  %.2f %s\n", transfer.TargetValue, transfer.TargetCurrency)
		fmt.Printf("Exchange Rate:           %.6f\n", transfer.Rate)
		fmt.Printf("Target Account:          %d\n", transfer.TargetAccount)
		fmt.Printf("Quote ID:                %s\n", transfer.QuoteUUID)
		fmt.Printf("Customer Transaction ID: %s\n", transfer.CustomerTransactionID)
		fmt.Printf("Created:                 %s\n", transfer.Created)

		if transfer.SourceAccount != nil {
			fmt.Printf("Source Account:          %d\n", *transfer.SourceAccount)
		}

		if transfer.Reference != nil && *transfer.Reference != "" {
			fmt.Printf("Reference:               %s\n", *transfer.Reference)
		}

		if transfer.PayinSessionID != nil && *transfer.PayinSessionID != "" {
			fmt.Printf("Payin Session ID:        %s\n", *transfer.PayinSessionID)
		}

		fmt.Printf("Has Active Issues:       %v\n", transfer.HasActiveIssues)

		return nil
	},
}

var newQuoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Create authenticated quote",
	Long:  "Create an authenticated quote for a currency conversion",
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

		req := commands.NewQuoteRequest{
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

		quote, err := commands.NewQuote(apiToken, req)
		if err != nil {
			return fmt.Errorf("failed to create quote: %w", err)
		}

		// Format output
		fmt.Println("Quote Details:")
		fmt.Println("==============")
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

var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Get exchange quote",
	Long:  "Create an authenticated quote for a currency conversion",
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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with Wise API token",
	Long:  "Save your Wise API token for future use (reads from stdin)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter your Wise API token: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("failed to read token from stdin")
		}

		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		if err := config.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		// Update the global apiToken for this session
		apiToken = token

		cacheDir, err := config.CacheDir()
		if err == nil {
			fmt.Printf("✓ Token saved to %s\n", cacheDir)
		} else {
			fmt.Println("✓ Token saved")
		}

		return nil
	},
}

var sendToCmd = &cobra.Command{
	Use:   "send-to <recipient-name> <amount> <currency>",
	Short: "Send money to a recipient",
	Long:  "Send money to a recipient by name, creating a quote and transfer automatically",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiToken == "" {
			return fmt.Errorf("API token required: set --token flag or WISE_API_TOKEN env var")
		}

		recipientName := args[0]
		amount, err := cmd.Flags().GetFloat64("amount")
		if err == nil && amount == 0 {
			// Try to parse from args if not provided as flag
			fmt.Sscanf(args[1], "%f", &amount)
		}
		currency := args[2]
		profileID, _ := cmd.Flags().GetInt("profile-id")
		sourceAccount, _ := cmd.Flags().GetInt("source-account")
		reference, _ := cmd.Flags().GetString("reference")
		customerTxID, _ := cmd.Flags().GetString("customer-transaction-id")

		if profileID == 0 {
			return fmt.Errorf("profile-id is required")
		}
		if amount == 0 {
			return fmt.Errorf("amount is required and must be greater than 0")
		}
		if currency == "" {
			return fmt.Errorf("currency is required")
		}

		// Step 1: Find the recipient by name
		fmt.Printf("Finding recipient: %s\n", recipientName)
		recipients, err := queries.ListRecipients(apiToken, queries.ListRecipientsRequest{
			ProfileID: profileID,
			Currency:  currency,
		})
		if err != nil {
			return fmt.Errorf("failed to list recipients: %w", err)
		}

		var targetRecipient *queries.Recipient
		for i := range recipients {
			if recipients[i].Name.FullName == recipientName {
				targetRecipient = &recipients[i]
				break
			}
		}

		if targetRecipient == nil {
			return fmt.Errorf("recipient not found: %s", recipientName)
		}
		fmt.Printf("Found recipient: %s (ID: %d)\n", targetRecipient.Name.FullName, targetRecipient.ID)

		// Step 2: Create a quote
		fmt.Printf("Creating quote: %.2f %s\n", amount, currency)
		quoteReq := commands.NewQuoteRequest{
			ProfileID:      profileID,
			SourceCurrency: currency,
			TargetCurrency: targetRecipient.Currency,
			SourceAmount:   &amount,
		}

		quote, err := commands.NewQuote(apiToken, quoteReq)
		if err != nil {
			return fmt.Errorf("failed to create quote: %w", err)
		}
		fmt.Printf("Quote created: %s\n", quote.ID)

		// Step 3: Create a transfer
		fmt.Println("Creating transfer...")
		if customerTxID == "" {
			return fmt.Errorf("customer-transaction-id is required for transfer")
		}

		transferReq := commands.NewTransferRequest{
			TargetAccount:         targetRecipient.ID,
			QuoteUUID:             quote.ID,
			CustomerTransactionID: customerTxID,
		}

		if reference != "" {
			transferReq.Reference = &reference
		}
		if sourceAccount != 0 {
			transferReq.SourceAccount = &sourceAccount
		}

		transfer, err := commands.NewTransfer(apiToken, transferReq)
		if err != nil {
			return fmt.Errorf("failed to create transfer: %w", err)
		}

		// Save transfer to cache
		transferData := config.TransferData{
			ID:                    transfer.ID,
			Status:                transfer.Status,
			SourceValue:           transfer.SourceValue,
			SourceCurrency:        transfer.SourceCurrency,
			TargetValue:           transfer.TargetValue,
			TargetCurrency:        transfer.TargetCurrency,
			Rate:                  transfer.Rate,
			Created:               transfer.Created,
			QuoteUUID:             transfer.QuoteUUID,
			CustomerTransactionID: transfer.CustomerTransactionID,
			TargetAccount:         transfer.TargetAccount,
			Reference:             transfer.Reference,
			SourceAccount:         transfer.SourceAccount,
			PayinSessionID:        transfer.PayinSessionID,
			HasActiveIssues:       transfer.HasActiveIssues,
		}
		if err := config.SaveTransfer(customerTxID, transferData); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save transfer to cache: %v\n", err)
		}

		// Format output
		fmt.Println("\n✓ Transfer Created Successfully:")
		fmt.Println("================================")
		fmt.Printf("Transfer ID:             %d\n", transfer.ID)
		fmt.Printf("Status:                  %s\n", transfer.Status)
		fmt.Printf("Recipient:               %s\n", recipientName)
		fmt.Printf("Source:                  %.2f %s\n", transfer.SourceValue, transfer.SourceCurrency)
		fmt.Printf("Target:                  %.2f %s\n", transfer.TargetValue, transfer.TargetCurrency)
		fmt.Printf("Exchange Rate:           %.6f\n", transfer.Rate)
		fmt.Printf("Quote ID:                %s\n", transfer.QuoteUUID)
		fmt.Printf("Customer Transaction ID: %s\n", transfer.CustomerTransactionID)
		fmt.Printf("Created:                 %s\n", transfer.Created)

		if transfer.Reference != nil && *transfer.Reference != "" {
			fmt.Printf("Reference:               %s\n", *transfer.Reference)
		}

		return nil
	},
}

func init() {
	recipientsCmd.Flags().IntP("profile-id", "p", 0, "Profile ID to filter by")
	recipientsCmd.Flags().StringP("currency", "c", "", "Filter by currency (e.g. USD,GBP)")
	recipientsCmd.Flags().StringP("type", "t", "", "Filter by account type (e.g. iban,swift_code)")
	recipientsCmd.Flags().IntP("size", "s", 20, "Number of results to return")

	newCmd.AddCommand(newQuoteCmd)
	newCmd.AddCommand(newTransferCmd)

	newQuoteCmd.Flags().IntP("profile-id", "p", 0, "Profile ID (required)")
	newQuoteCmd.MarkFlagRequired("profile-id")
	newQuoteCmd.Flags().StringP("source-currency", "s", "", "Source currency code (required)")
	newQuoteCmd.MarkFlagRequired("source-currency")
	newQuoteCmd.Flags().StringP("target-currency", "t", "", "Target currency code (required)")
	newQuoteCmd.MarkFlagRequired("target-currency")
	newQuoteCmd.Flags().Float64("source-amount", 0, "Amount in source currency (either this or target-amount)")
	newQuoteCmd.Flags().Float64("target-amount", 0, "Amount in target currency (either this or source-amount)")

	quoteCmd.Flags().IntP("profile-id", "p", 0, "Profile ID (required)")
	quoteCmd.MarkFlagRequired("profile-id")
	quoteCmd.Flags().StringP("source-currency", "s", "", "Source currency code (required)")
	quoteCmd.MarkFlagRequired("source-currency")
	quoteCmd.Flags().StringP("target-currency", "t", "", "Target currency code (required)")
	quoteCmd.MarkFlagRequired("target-currency")
	quoteCmd.Flags().Float64("source-amount", 0, "Amount in source currency (either this or target-amount)")
	quoteCmd.Flags().Float64("target-amount", 0, "Amount in target currency (either this or source-amount)")

	newTransferCmd.Flags().IntP("target-account", "a", 0, "Target account ID (required)")
	newTransferCmd.MarkFlagRequired("target-account")
	newTransferCmd.Flags().StringP("quote-uuid", "q", "", "Quote UUID (required)")
	newTransferCmd.MarkFlagRequired("quote-uuid")
	newTransferCmd.Flags().StringP("customer-transaction-id", "c", "", "Customer transaction ID for idempotency (required)")
	newTransferCmd.MarkFlagRequired("customer-transaction-id")
	newTransferCmd.Flags().StringP("reference", "r", "", "Payment reference (optional)")
	newTransferCmd.Flags().IntP("source-account", "s", 0, "Source account ID (optional)")

	sendToCmd.Flags().IntP("profile-id", "p", 0, "Profile ID (required)")
	sendToCmd.MarkFlagRequired("profile-id")
	sendToCmd.Flags().StringP("customer-transaction-id", "c", "", "Customer transaction ID for idempotency (required)")
	sendToCmd.MarkFlagRequired("customer-transaction-id")
	sendToCmd.Flags().StringP("reference", "r", "", "Payment reference (optional)")
	sendToCmd.Flags().IntP("source-account", "s", 0, "Source account ID (optional)")
}
