package queries

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dhamidi/wise-cli/config"
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

// GetAccountNumber extracts account number from recipient details
func (r *Recipient) GetAccountNumber() string {
	if r.Details == nil {
		return ""
	}

	detailsMap, ok := r.Details.(map[string]interface{})
	if !ok {
		return ""
	}

	// Try IBAN first (most common)
	if iban, exists := detailsMap["iban"]; exists {
		if str, ok := iban.(string); ok && str != "" {
			return str
		}
	}

	// Try other field names that might contain account number
	for _, field := range []string{"accountNumber", "number", "accountId", "id", "bic"} {
		if val, exists := detailsMap[field]; exists {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}

	return ""
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

// ListRecipients queries the Wise API for recipient accounts with caching
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
	queryStr := params.Encode()
	if queryStr != "" {
		endpoint += "?" + queryStr
	}

	// Generate cache key based on query parameters
	cacheKey := generateCacheKey("recipients", queryStr)

	// Check cache first
	if cached, err := config.GetCacheEntry(cacheKey); err == nil && cached != "" {
		var apiResp ListRecipientsResponse
		if err := json.Unmarshal([]byte(cached), &apiResp); err == nil {
			return apiResp.Content, nil
		}
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

	// Store in cache with HTTP headers
	if err := config.SetCacheEntry(cacheKey, string(body), httpResp.Header); err != nil {
		// Log error but don't fail the request
		fmt.Fprintf(os.Stderr, "Warning: failed to cache recipients: %v\n", err)
	}

	return apiResp.Content, nil
}

// generateCacheKey creates a cache key from endpoint and query string
func generateCacheKey(endpoint, queryStr string) string {
	hash := md5.Sum([]byte(queryStr))
	return fmt.Sprintf("%s-%x.json", strings.TrimSuffix(endpoint, ".json"), hash)
}
