package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// NewQuoteRequest holds parameters for creating an authenticated quote
type NewQuoteRequest struct {
	ProfileID      int
	SourceCurrency string
	TargetCurrency string
	SourceAmount   *float64
	TargetAmount   *float64
}

// NewQuote creates an authenticated quote for a currency conversion
func NewQuote(apiToken string, req NewQuoteRequest) (*Quote, error) {
	payload := map[string]interface{}{
		"sourceCurrency": req.SourceCurrency,
		"targetCurrency": req.TargetCurrency,
	}

	if req.SourceAmount != nil {
		payload["sourceAmount"] = *req.SourceAmount
	}
	if req.TargetAmount != nil {
		payload["targetAmount"] = *req.TargetAmount
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("https://api.wise.com/v3/profiles/%d/quotes", req.ProfileID)

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(body))
	}

	var quote Quote
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &quote, nil
}

// Quote represents a Wise exchange quote
type Quote struct {
	ID                   string                 `json:"id"`
	SourceAmount         float64                `json:"sourceAmount"`
	SourceCurrency       string                 `json:"sourceCurrency"`
	TargetAmount         float64                `json:"targetAmount"`
	TargetCurrency       string                 `json:"targetCurrency"`
	Rate                 float64                `json:"rate"`
	CreatedTime          string                 `json:"createdTime"`
	RateExpirationTime   string                 `json:"rateExpirationTime"`
	RateType             string                 `json:"rateType"`
	PayOut               string                 `json:"payOut"`
	Profile              int                    `json:"profile"`
	User                 int                    `json:"user"`
	ProvidedAmountType   string                 `json:"providedAmountType"`
	Status               string                 `json:"status"`
	ExpirationTime       string                 `json:"expirationTime"`
	PaymentOptions       []PaymentOption        `json:"paymentOptions"`
	Notices              []Notice               `json:"notices"`
	PricingConfiguration map[string]interface{} `json:"pricingConfiguration"`
}

type PaymentOption struct {
	ID                         string          `json:"id"`
	PayIn                      string          `json:"payIn"`
	PayOut                     string          `json:"payOut"`
	SourceAmount               float64         `json:"sourceAmount"`
	TargetAmount               float64         `json:"targetAmount"`
	Fee                        Fee             `json:"fee"`
	EstimatedDelivery          string          `json:"estimatedDelivery"`
	FormattedEstimatedDelivery string          `json:"formattedEstimatedDelivery"`
	Disabled                   bool            `json:"disabled"`
	DisabledReason             *DisabledReason `json:"disabledReason"`
}

type Fee struct {
	TransferWise float64 `json:"transferwise"`
	PayIn        float64 `json:"payIn"`
	Discount     float64 `json:"discount"`
	Partner      float64 `json:"partner"`
	Total        float64 `json:"total"`
}

type DisabledReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Notice struct {
	Text string  `json:"text"`
	Link *string `json:"link"`
	Type string  `json:"type"`
}
