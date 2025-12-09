package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dhamidi/wise-cli/config"
)

// Transfer details from commands/transfer.go are reused here via type alias
// We'll import the type from commands package or redefine it locally

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

// ListTransfersRequest holds parameters for listing transfers
type ListTransfersRequest struct {
	ProfileID int
	Status    string
	Since     *time.Time
	Until     *time.Time
	Limit     int
	Offset    int
}

// ListTransfers queries the Wise API for transfers with caching
func ListTransfers(apiToken string, req ListTransfersRequest) ([]Transfer, error) {
	return ListTransfersWithRefresh(apiToken, req, false)
}

// ListTransfersWithRefresh queries the Wise API for transfers, optionally bypassing cache
func ListTransfersWithRefresh(apiToken string, req ListTransfersRequest, refresh bool) ([]Transfer, error) {
	// Build query parameters
	params := url.Values{}

	if req.ProfileID != 0 {
		params.Set("profile", fmt.Sprintf("%d", req.ProfileID))
	}

	if req.Status != "" {
		params.Set("status", req.Status)
	}

	if req.Since != nil {
		params.Set("createdDateStart", req.Since.Format("2006-01-02T15:04:05Z"))
	}

	if req.Until != nil {
		params.Set("createdDateEnd", req.Until.Format("2006-01-02T15:04:05Z"))
	}

	if req.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", req.Limit))
	} else {
		params.Set("limit", "100") // Default limit
	}

	if req.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", req.Offset))
	}

	endpoint := "https://api.wise.com/v1/transfers"
	queryStr := params.Encode()
	if queryStr != "" {
		endpoint += "?" + queryStr
	}

	// Generate cache key based on query parameters
	cacheKey := generateCacheKey("transfers", queryStr)

	// Check cache first
	if cached, err := config.GetCacheEntryWithRefresh(cacheKey, refresh); err == nil && cached != "" {
		var transfers []Transfer
		if err := json.Unmarshal([]byte(cached), &transfers); err == nil {
			return transfers, nil
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
		return nil, fmt.Errorf("failed to fetch transfers: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var transfers []Transfer
	if err := json.Unmarshal(body, &transfers); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Store in cache with HTTP headers
	if err := config.SetCacheEntry(cacheKey, string(body), httpResp.Header); err != nil {
		// Log error but don't fail the request
		fmt.Fprintf(os.Stderr, "Warning: failed to cache transfers: %v\n", err)
	}

	return transfers, nil
}
