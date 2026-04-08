package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = ".config/zaim"
	configFileName = "config.json"
)

// Config represents the CLI configuration.
type Config struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

// GetConfigDir returns the configuration directory path.
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", configDirName)
	}

	return filepath.Join(homeDir, configDirName)
}

// GetConfigPath returns the full path to the config file.
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), configFileName)
}

// Load loads the configuration from the config file.
func Load() (*Config, error) {
	configPath := GetConfigPath()

	//nolint:gosec // Config file path is constructed internally, not from user input.
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to the config file.
func Save(cfg *Config) error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if err := os.WriteFile(GetConfigPath(), data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the config file exists.
func Exists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}
