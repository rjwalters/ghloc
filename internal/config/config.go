package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// GitHub App credentials
	AppID          int64
	PrivateKeyPath string
	WebhookSecret  string

	// Server
	Port string

	// Storage
	DBPath string

	// Optional: commit badges/charts back to repos
	CommitArtifacts bool
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		return nil, fmt.Errorf("GITHUB_APP_ID is required")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("GITHUB_APP_ID must be a number: %w", err)
	}

	privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	if privateKeyPath == "" {
		return nil, fmt.Errorf("GITHUB_PRIVATE_KEY_PATH is required")
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "ghloc.db"
	}

	commitArtifacts, _ := strconv.ParseBool(os.Getenv("COMMIT_ARTIFACTS"))

	return &Config{
		AppID:           appID,
		PrivateKeyPath:  privateKeyPath,
		WebhookSecret:   webhookSecret,
		Port:            port,
		DBPath:          dbPath,
		CommitArtifacts: commitArtifacts,
	}, nil
}
