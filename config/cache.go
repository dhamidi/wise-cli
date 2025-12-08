package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// CacheDir returns the cache directory for wise-cli
func CacheDir() (string, error) {
	var cacheHome string
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		cacheHome = xdgCache
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheHome = filepath.Join(home, ".cache")
	}

	cacheDir := filepath.Join(cacheHome, "wise-cli")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Data      string    `json:"data"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// GetCacheEntry retrieves a cached entry if it's still valid
func GetCacheEntry(cacheKey string) (string, error) {
	return GetCacheEntryWithRefresh(cacheKey, false)
}

// GetCacheEntryWithRefresh retrieves a cached entry if it's still valid, optionally bypassing cache
func GetCacheEntryWithRefresh(cacheKey string, refresh bool) (string, error) {
	if refresh {
		// Skip cache if refresh is requested
		return "", nil
	}

	cacheDir, err := CacheDir()
	if err != nil {
		return "", err
	}

	cachePath := filepath.Join(cacheDir, cacheKey)
	body, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read cache: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return "", fmt.Errorf("failed to parse cache: %w", err)
	}

	if time.Now().After(entry.ExpiresAt) {
		// Cache expired, delete it
		os.Remove(cachePath)
		return "", nil
	}

	return entry.Data, nil
}

// SetCacheEntry stores a cache entry with expiration based on HTTP headers
func SetCacheEntry(cacheKey, data string, headers http.Header) error {
	expiresAt := parseExpiration(headers)

	entry := CacheEntry{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	cachePath := filepath.Join(cacheDir, cacheKey)
	if err := os.WriteFile(cachePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// parseExpiration extracts expiration time from HTTP cache headers
func parseExpiration(headers http.Header) time.Time {
	// Try Cache-Control: max-age first
	if cacheControl := headers.Get("Cache-Control"); cacheControl != "" {
		// Simple parsing for max-age
		if maxAge := extractMaxAge(cacheControl); maxAge > 0 {
			return time.Now().Add(time.Duration(maxAge) * time.Second)
		}
	}

	// Try Expires header
	if expires := headers.Get("Expires"); expires != "" {
		if t, err := time.Parse(time.RFC1123, expires); err == nil {
			return t
		}
	}

	// Default to 1 hour if no cache headers
	return time.Now().Add(1 * time.Hour)
}

// extractMaxAge extracts max-age value from Cache-Control header
func extractMaxAge(cacheControl string) int64 {
	// Simple parser for "max-age=3600"
	parts := []byte(cacheControl)
	for i := 0; i < len(parts); {
		// Find "max-age="
		if i+8 <= len(parts) && string(parts[i:i+8]) == "max-age=" {
			// Extract the number
			i += 8
			var numStr string
			for i < len(parts) && parts[i] >= '0' && parts[i] <= '9' {
				numStr += string(parts[i])
				i++
			}
			if numStr != "" {
				val, _ := strconv.ParseInt(numStr, 10, 64)
				return val
			}
		}
		i++
	}
	return 0
}
