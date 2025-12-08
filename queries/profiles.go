package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dhamidi/wise-cli/config"
)

// Address represents an address object from Wise API
type Address struct {
	ID               int     `json:"id"`
	AddressFirstLine string  `json:"addressFirstLine"`
	City             string  `json:"city"`
	CountryIso2Code  string  `json:"countryIso2Code"`
	CountryIso3Code  string  `json:"countryIso3Code"`
	PostCode         string  `json:"postCode"`
	StateCode        *string `json:"stateCode"`
}

// ContactDetails represents contact information
type ContactDetails struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

// Profile represents a Wise user profile
type Profile struct {
	ID              int            `json:"id"`
	PublicID        string         `json:"publicId"`
	UserID          int            `json:"userId"`
	Type            string         `json:"type"` // PERSONAL or BUSINESS
	Address         Address        `json:"address"`
	Email           string         `json:"email"`
	CreatedAt       string         `json:"createdAt"`
	UpdatedAt       string         `json:"updatedAt"`
	Avatar          *string        `json:"avatar"`
	CurrentState    string         `json:"currentState"`
	ContactDetails  ContactDetails `json:"contactDetails"`
	FirstName       *string        `json:"firstName"`
	LastName        *string        `json:"lastName"`
	DateOfBirth     *string        `json:"dateOfBirth"`
	BusinessName    *string        `json:"businessName"`
	BusinessLogoUrl *string        `json:"businessLogoUrl"`
}

// ListProfiles queries the Wise API for all profiles belonging to the user
func ListProfiles(apiToken string) ([]Profile, error) {
	endpoint := "https://api.wise.com/v2/profiles"

	// Generate cache key
	cacheKey := generateCacheKey("profiles", "")

	// Check cache first
	if cached, err := config.GetCacheEntry(cacheKey); err == nil && cached != "" {
		var profiles []Profile
		if err := json.Unmarshal([]byte(cached), &profiles); err == nil {
			return profiles, nil
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
		return nil, fmt.Errorf("failed to fetch profiles: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var profiles []Profile
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Store in cache with HTTP headers
	if err := config.SetCacheEntry(cacheKey, string(body), httpResp.Header); err != nil {
		// Log error but don't fail the request
		fmt.Fprintf(os.Stderr, "Warning: failed to cache profiles: %v\n", err)
	}

	return profiles, nil
}
