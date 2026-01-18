package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if cfg.NoNerdFonts != false {
		t.Errorf("expected NoNerdFonts to be false, got %v", cfg.NoNerdFonts)
	}
}

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir failed: %v", err)
	}
	expectedSuffix := "containertui"
	if !strings.HasSuffix(dir, expectedSuffix) {
		t.Errorf("expected ConfigDir to end with 'containertui', got %s", dir)
	}
}

func TestConfigFilePath(t *testing.T) {
	path, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath failed: %v", err)
	}
	expectedSuffix := filepath.Join("containertui", "config.yaml")
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("expected ConfigFilePath to end with 'containertui/config.yaml', got %s", path)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Test loading non-existent file returns default config
	cfg, err := LoadFromFile("/non/existent/path")
	if err != nil {
		t.Fatalf("LoadFromFile with non-existent path failed: %v", err)
	}
	if cfg.NoNerdFonts != false {
		t.Errorf("expected default config, got NoNerdFonts %v", cfg.NoNerdFonts)
	}

	// Test loading from temp file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "config.yaml")

	// Write a test config
	testConfig := `no-nerd-fonts: true`
	err = os.WriteFile(tempFile, []byte(testConfig), 0o600)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err = LoadFromFile(tempFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}
	if cfg.NoNerdFonts != true {
		t.Errorf("expected NoNerdFonts true, got %v", cfg.NoNerdFonts)
	}

	// Test loading invalid YAML
	err = os.WriteFile(tempFile, []byte("invalid: yaml: :"), 0o600)
	if err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}
	_, err = LoadFromFile(tempFile)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
