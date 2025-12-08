package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// NewTransferRequest holds parameters for creating a transfer
type NewTransferRequest struct {
	TargetAccount         int
	QuoteUUID             string
	CustomerTransactionID string
	Reference             *string
	SourceAccount         *int
}

// TransferDetails represents the details field in a transfer
type TransferDetails struct {
	Reference string `json:"reference,omitempty"`
}

// Transfer represents a Wise transfer
type Transfer struct {
	ID                    int             `json:"id"`
	User                  int             `json:"user"`
	TargetAccount         int             `json:"targetAccount"`
	SourceAccount         *int            `json:"sourceAccount"`
	Quote                 *int            `json:"quote"`
	QuoteUUID             string          `json:"quoteUuid"`
	Status                string          `json:"status"`
	Reference             *string         `json:"reference"`
	Rate                  float64         `json:"rate"`
	Created               string          `json:"created"`
	Business              *int            `json:"business"`
	Details               TransferDetails `json:"details"`
	HasActiveIssues       bool            `json:"hasActiveIssues"`
	SourceCurrency        string          `json:"sourceCurrency"`
	SourceValue           float64         `json:"sourceValue"`
	TargetCurrency        string          `json:"targetCurrency"`
	TargetValue           float64         `json:"targetValue"`
	CustomerTransactionID string          `json:"customerTransactionId"`
	PayinSessionID        *string         `json:"payinSessionId"`
}

// NewTransfer creates a transfer using the Wise API
func NewTransfer(apiToken string, req NewTransferRequest) (*Transfer, error) {
	payload := map[string]interface{}{
		"targetAccount":         req.TargetAccount,
		"quoteUuid":             req.QuoteUUID,
		"customerTransactionId": req.CustomerTransactionID,
	}

	if req.SourceAccount != nil {
		payload["sourceAccount"] = *req.SourceAccount
	}

	if req.Reference != nil {
		details := map[string]interface{}{
			"reference": *req.Reference,
		}
		payload["details"] = details
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "https://api.wise.com/v1/transfers"

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var transfer Transfer
	if err := json.Unmarshal(body, &transfer); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &transfer, nil
}
