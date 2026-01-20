// Package config provides functionality to load and manage application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	NoNerdFonts ConfigBool  `yaml:"no-nerd-fonts"`
	Theme       ThemeConfig `yaml:"colors,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		NoNerdFonts: false,
		Theme:       emptyThemeConfig(),
	}
}

// LoadFromFile loads configuration from a YAML file.
// If path is empty, uses the default config file path.
func LoadFromFile(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = ConfigFilePath()
		if err != nil {
			return nil, fmt.Errorf("failed to get config file path: %w", err)
		}
	}

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	file, err := os.Open(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}

		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to close file reader: %w", err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}

// ConfigDir returns the default configuration directory.
func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	return filepath.Join(configDir, "containertui"), nil
}

// ConfigFilePath returns the default configuration file path.
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "config.yaml"), nil
}
