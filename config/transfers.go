package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TransferData represents stored transfer information
type TransferData struct {
	ID                    int     `json:"id"`
	Status                string  `json:"status"`
	SourceValue           float64 `json:"sourceValue"`
	SourceCurrency        string  `json:"sourceCurrency"`
	TargetValue           float64 `json:"targetValue"`
	TargetCurrency        string  `json:"targetCurrency"`
	Rate                  float64 `json:"rate"`
	Created               string  `json:"created"`
	QuoteUUID             string  `json:"quoteUuid"`
	CustomerTransactionID string  `json:"customerTransactionId"`
	TargetAccount         int     `json:"targetAccount"`
	Reference             *string `json:"reference,omitempty"`
	SourceAccount         *int    `json:"sourceAccount,omitempty"`
	PayinSessionID        *string `json:"payinSessionId,omitempty"`
	HasActiveIssues       bool    `json:"hasActiveIssues"`
}

// SaveTransfer saves transfer data indexed by customer transaction ID (UUID)
func SaveTransfer(customerTxID string, data TransferData) error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	transfersDir := filepath.Join(cacheDir, "transfers")
	if err := os.MkdirAll(transfersDir, 0755); err != nil {
		return fmt.Errorf("failed to create transfers directory: %w", err)
	}

	transferPath := filepath.Join(transfersDir, customerTxID+".json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transfer data: %w", err)
	}

	if err := os.WriteFile(transferPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save transfer: %w", err)
	}

	return nil
}

// LoadTransfer loads transfer data by customer transaction ID
func LoadTransfer(customerTxID string) (TransferData, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return TransferData{}, err
	}

	transferPath := filepath.Join(cacheDir, "transfers", customerTxID+".json")
	jsonData, err := os.ReadFile(transferPath)
	if err != nil {
		if os.IsNotExist(err) {
			return TransferData{}, fmt.Errorf("transfer not found: %s", customerTxID)
		}
		return TransferData{}, fmt.Errorf("failed to read transfer: %w", err)
	}

	var data TransferData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return TransferData{}, fmt.Errorf("failed to parse transfer: %w", err)
	}

	return data, nil
}
