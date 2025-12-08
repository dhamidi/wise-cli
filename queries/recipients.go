package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Recipient represents a Wise recipient account
type Recipient struct {
	ID              int    `json:"id"`
	CreatorID       int    `json:"creatorId"`
	ProfileID       int    `json:"profileId"`
	Name            Name   `json:"name"`
	Currency        string `json:"currency"`
	Country         string `json:"country"`
	Type            string `json:"type"`
	LegalEntityType string `json:"legalEntityType"`
	Status          bool   `json:"status"`
	Details         any    `json:"details"`
	Hash            string `json:"hash"`
}

type Name struct {
	FullName   string `json:"fullName"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	MiddleName string `json:"middleName"`
}

// ListRecipientsRequest holds parameters for listing recipients
type ListRecipientsRequest struct {
	ProfileID int
	Currency  string
	Active    *bool
	Type      string
	Size      int
	SeekPos   int
	Sort      string
}

// ListRecipientsResponse holds the response from the Wise API
type ListRecipientsResponse struct {
	Content     []Recipient `json:"content"`
	Size        int         `json:"size"`
	SeekPos     int         `json:"seekPosition"`
	SeekNext    int         `json:"seekPositionForNext"`
	SeekCurrent int         `json:"seekPositionForCurrent"`
}

// ListRecipients queries the Wise API for recipient accounts
func ListRecipients(apiToken string, req ListRecipientsRequest) ([]Recipient, error) {
	params := url.Values{}
	if req.ProfileID != 0 {
		params.Set("profileId", fmt.Sprintf("%d", req.ProfileID))
	}
	if req.Currency != "" {
		params.Set("currency", req.Currency)
	}
	if req.Active != nil {
		params.Set("active", fmt.Sprintf("%v", *req.Active))
	}
	if req.Type != "" {
		params.Set("type", req.Type)
	}
	if req.Size > 0 {
		params.Set("size", fmt.Sprintf("%d", req.Size))
	}
	if req.SeekPos > 0 {
		params.Set("seekPosition", fmt.Sprintf("%d", req.SeekPos))
	}
	if req.Sort != "" {
		params.Set("sort", req.Sort)
	}

	endpoint := "https://api.wise.com/v2/accounts"
	if params.Encode() != "" {
		endpoint += "?" + params.Encode()
	}

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipients: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var apiResp ListRecipientsResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.Content, nil
}
