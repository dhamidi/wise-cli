package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// NewRecipientRequest holds parameters for creating a recipient account
type NewRecipientRequest struct {
	ProfileID         int
	Currency          string
	Type              string
	AccountHolderName string
	OwnedByCustomer   *bool
	Details           map[string]interface{}
}

// RecipientName holds recipient name information
type RecipientName struct {
	FullName   string `json:"fullName"`
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
}

// Recipient represents a Wise recipient account
type Recipient struct {
	ID              int                    `json:"id"`
	CreatorID       int                    `json:"creatorId"`
	ProfileID       int                    `json:"profileId"`
	Name            RecipientName          `json:"name"`
	Currency        string                 `json:"currency"`
	Country         string                 `json:"country"`
	Type            string                 `json:"type"`
	LegalEntityType string                 `json:"legalEntityType"`
	Active          bool                   `json:"active"`
	Details         map[string]interface{} `json:"details"`
	AccountSummary  string                 `json:"accountSummary"`
	Hash            string                 `json:"hash"`
	IsInternal      bool                   `json:"isInternal"`
	OwnedByCustomer bool                   `json:"ownedByCustomer"`
}

// NewRecipient creates a recipient account using the Wise API
func NewRecipient(apiToken string, req NewRecipientRequest) (*Recipient, error) {
	payload := map[string]interface{}{
		"currency":          req.Currency,
		"type":              req.Type,
		"profile":           req.ProfileID,
		"accountHolderName": req.AccountHolderName,
	}

	if req.OwnedByCustomer != nil {
		payload["ownedByCustomer"] = *req.OwnedByCustomer
	}

	if req.Details != nil {
		payload["details"] = req.Details
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "https://api.wise.com/v1/accounts"

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipient: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var recipient Recipient
	if err := json.Unmarshal(body, &recipient); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &recipient, nil
}
