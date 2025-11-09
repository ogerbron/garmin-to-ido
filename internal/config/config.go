package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	GarminUsername string
	GarminPassword string
	IdoUsername    string
	IdoPassword    string
}

// Load reads configuration from a file
func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	cfg := &Config{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		switch key {
		case "GARMIN_USERNAME":
			cfg.GarminUsername = value
		case "GARMIN_PASSWORD":
			cfg.GarminPassword = value
		case "IDO_USERNAME":
			cfg.IdoUsername = value
		case "IDO_PASSWORD":
			cfg.IdoPassword = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return cfg, nil
}

// Validate checks if all required configuration values are present
func (c *Config) Validate() error {
	if c.GarminUsername == "" {
		return fmt.Errorf("GARMIN_USERNAME is required")
	}
	if c.GarminPassword == "" {
		return fmt.Errorf("GARMIN_PASSWORD is required")
	}
	if c.IdoUsername == "" {
		return fmt.Errorf("IDO_USERNAME is required")
	}
	if c.IdoPassword == "" {
		return fmt.Errorf("IDO_PASSWORD is required")
	}
	return nil
}
