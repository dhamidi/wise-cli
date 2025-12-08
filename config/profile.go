package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const profileFileName = "selected-profile"

// SelectedProfile represents the cached selected profile
type SelectedProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// SaveSelectedProfile saves the selected profile to the cache directory
func SaveSelectedProfile(profile SelectedProfile) error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	profilePath := filepath.Join(cacheDir, profileFileName)
	jsonData, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(profilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// LoadSelectedProfile loads the selected profile from the cache directory
func LoadSelectedProfile() (*SelectedProfile, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return nil, err
	}

	profilePath := filepath.Join(cacheDir, profileFileName)
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile SelectedProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}
