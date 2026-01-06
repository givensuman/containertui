package colors

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/context"
)

func TestANSIColors(t *testing.T) {
	// Test that ANSI color constants work
	if ColorBlack.String() != "0" {
		t.Errorf("expected ColorBlack to be '0', got %s", ColorBlack.String())
	}
	if ColorBrightYellow.String() != "11" {
		t.Errorf("expected ColorBrightYellow to be '11', got %s", ColorBrightYellow.String())
	}
}

func TestColorFunctions(t *testing.T) {
	// Set up default config
	context.SetConfig(config.DefaultConfig())

	// Test default colors
	yellow := Yellow()
	if yellow == "" {
		t.Error("Yellow() returned empty color")
	}

	green := Green()
	if green == "" {
		t.Error("Green() returned empty color")
	}

	gray := Gray()
	if gray == "" {
		t.Error("Gray() returned empty color")
	}

	primary := Primary()
	if primary == "" {
		t.Error("Primary() returned empty color")
	}

	blue := Blue()
	if blue == "" {
		t.Error("Blue() returned empty color")
	}

	white := White()
	if white == "" {
		t.Error("White() returned empty color")
	}

	// Test that Primary defaults to Gray
	if primary != gray {
		t.Error("Primary() should default to Gray()")
	}
}

func TestColorOverrides(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Colors.Yellow = "#FFFF00"
	cfg.Colors.Green = "#00FF00"
	cfg.Colors.Blue = "#0000FF"
	cfg.Colors.White = "#FFFFFF"
	context.SetConfig(cfg)

	yellow := Yellow()
	if yellow != lipgloss.Color("#FFFF00") {
		t.Errorf("expected Yellow to be '#FFFF00', got %s", yellow)
	}

	green := Green()
	if green != lipgloss.Color("#00FF00") {
		t.Errorf("expected Green to be '#00FF00', got %s", green)
	}

	blue := Blue()
	if blue != lipgloss.Color("#0000FF") {
		t.Errorf("expected Blue to be '#0000FF', got %s", blue)
	}

	white := White()
	if white != lipgloss.Color("#FFFFFF") {
		t.Errorf("expected White to be '#FFFFFF', got %s", white)
	}

	gray := Gray()
	if gray == "" {
		t.Error("Gray() should still return default color when not overridden")
	}
}

func TestParseColors(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expected    *config.ColorConfig
		expectError bool
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: &config.ColorConfig{},
		},
		{
			name:  "single color",
			input: []string{"primary=#FF0000"},
			expected: &config.ColorConfig{
				Primary: config.ConfigString("#FF0000"),
			},
		},
		{
			name:  "multiple colors in one string",
			input: []string{"primary=#FF0000,yellow=#FFFF00,green=#00FF00"},
			expected: &config.ColorConfig{
				Primary: config.ConfigString("#FF0000"),
				Yellow:  config.ConfigString("#FFFF00"),
				Green:   config.ConfigString("#00FF00"),
			},
		},
		{
			name:  "multiple strings",
			input: []string{"primary=#FF0000", "yellow=#FFFF00", "green=#00FF00"},
			expected: &config.ColorConfig{
				Primary: config.ConfigString("#FF0000"),
				Yellow:  config.ConfigString("#FFFF00"),
				Green:   config.ConfigString("#00FF00"),
			},
		},
		{
			name:  "mixed format",
			input: []string{"primary=#FF0000,yellow=#FFFF00", "green=#00FF00"},
			expected: &config.ColorConfig{
				Primary: config.ConfigString("#FF0000"),
				Yellow:  config.ConfigString("#FFFF00"),
				Green:   config.ConfigString("#00FF00"),
			},
		},
		{
			name:  "with spaces",
			input: []string{"primary=#FF0000, yellow=#FFFF00"},
			expected: &config.ColorConfig{
				Primary: config.ConfigString("#FF0000"),
				Yellow:  config.ConfigString("#FFFF00"),
			},
		},
		{
			name:        "invalid format - no equals",
			input:       []string{"primary#FF0000"},
			expectError: true,
		},
		{
			name:        "invalid format - too many equals",
			input:       []string{"primary=#FF=000"},
			expectError: true,
		},
		{
			name:        "unknown color key",
			input:       []string{"unknown=#FF0000"},
			expectError: true,
		},
		{
			name:  "all colors",
			input: []string{"primary=#FF0000,yellow=#FFFF00,green=#00FF00,blue=#0000FF"},
			expected: &config.ColorConfig{
				Primary: config.ConfigString("#FF0000"),
				Yellow:  config.ConfigString("#FFFF00"),
				Green:   config.ConfigString("#00FF00"),
				Blue:    config.ConfigString("#0000FF"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseColors(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Primary != tt.expected.Primary {
				t.Errorf("expected Primary=%s, got %s", tt.expected.Primary, result.Primary)
			}
			if result.Yellow != tt.expected.Yellow {
				t.Errorf("expected Yellow=%s, got %s", tt.expected.Yellow, result.Yellow)
			}
			if result.Green != tt.expected.Green {
				t.Errorf("expected Green=%s, got %s", tt.expected.Green, result.Green)
			}
			if result.Blue != tt.expected.Blue {
				t.Errorf("expected Blue=%s, got %s", tt.expected.Blue, result.Blue)
			}
			if result.White != tt.expected.White {
				t.Errorf("expected White=%s, got %s", tt.expected.White, result.White)
			}
		})
	}
}
