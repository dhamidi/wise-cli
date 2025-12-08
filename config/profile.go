package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultProfileFileName = "default-profile"

// SaveDefaultProfile saves the default profile ID
func SaveDefaultProfile(profileID int) error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	profilePath := filepath.Join(cacheDir, defaultProfileFileName)
	if err := os.WriteFile(profilePath, []byte(strconv.Itoa(profileID)), 0644); err != nil {
		return fmt.Errorf("failed to save default profile: %w", err)
	}

	return nil
}

// LoadDefaultProfile loads the default profile ID
func LoadDefaultProfile() (int, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return 0, err
	}

	profilePath := filepath.Join(cacheDir, defaultProfileFileName)
	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read default profile: %w", err)
	}

	profileID, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid profile ID in config: %w", err)
	}

	return profileID, nil
}
