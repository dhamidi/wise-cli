package queries

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UserDetails represents user details
type UserDetails struct {
	FirstName      string   `json:"firstName"`
	LastName       string   `json:"lastName"`
	PhoneNumber    string   `json:"phoneNumber"`
	DateOfBirth    *string  `json:"dateOfBirth"`
	Occupation     *string  `json:"occupation"`
	Avatar         *string  `json:"avatar"`
	PrimaryAddress *int     `json:"primaryAddress"`
	Address        *Address `json:"address"`
}

// User represents the authenticated user from /v1/me endpoint
type User struct {
	ID      int         `json:"id"`
	Name    string      `json:"name"`
	Email   string      `json:"email"`
	Active  bool        `json:"active"`
	Details UserDetails `json:"details"`
}

// GetMe fetches the authenticated user's details from the Wise API
func GetMe(apiToken string) (*User, error) {
	endpoint := "https://api.wise.com/v1/me"

	httpReq, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &user, nil
}
