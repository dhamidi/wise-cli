package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const tokenFileName = "token"

// SaveToken saves the API token to the cache directory
func SaveToken(token string) error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	tokenPath := filepath.Join(cacheDir, tokenFileName)
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// LoadToken loads the API token from the cache directory
func LoadToken() (string, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	tokenPath := filepath.Join(cacheDir, tokenFileName)
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	return string(token), nil
}
