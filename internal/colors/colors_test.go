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

	black := Black()
	if black == "" {
		t.Error("Black() returned empty color")
	}

	red := Red()
	if red == "" {
		t.Error("Red() returned empty color")
	}

	magenta := Magenta()
	if magenta == "" {
		t.Error("Magenta() returned empty color")
	}

	cyan := Cyan()
	if cyan == "" {
		t.Error("Cyan() returned empty color")
	}

	brightBlack := BrightBlack()
	if brightBlack == "" {
		t.Error("BrightBlack() returned empty color")
	}

	brightRed := BrightRed()
	if brightRed == "" {
		t.Error("BrightRed() returned empty color")
	}

	brightGreen := BrightGreen()
	if brightGreen == "" {
		t.Error("BrightGreen() returned empty color")
	}

	brightYellow := BrightYellow()
	if brightYellow == "" {
		t.Error("BrightYellow() returned empty color")
	}

	brightBlue := BrightBlue()
	if brightBlue == "" {
		t.Error("BrightBlue() returned empty color")
	}

	brightMagenta := BrightMagenta()
	if brightMagenta == "" {
		t.Error("BrightMagenta() returned empty color")
	}

	brightCyan := BrightCyan()
	if brightCyan == "" {
		t.Error("BrightCyan() returned empty color")
	}

	brightWhite := BrightWhite()
	if brightWhite == "" {
		t.Error("BrightWhite() returned empty color")
	}

	// Test that Primary defaults to Gray
	if primary != blue {
		t.Error("Primary() should default to Blue()")
	}
}

func TestColorOverrides(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Colors.Yellow = "#FFFF00"
	cfg.Colors.Green = "#00FF00"
	cfg.Colors.Blue = "#0000FF"
	cfg.Colors.White = "#FFFFFF"
	cfg.Colors.Black = "#000000"
	cfg.Colors.Red = "#FF0000"
	cfg.Colors.Magenta = "#FF00FF"
	cfg.Colors.Cyan = "#00FFFF"
	cfg.Colors.BrightBlack = "#808080"
	cfg.Colors.BrightRed = "#FF8080"
	cfg.Colors.BrightGreen = "#80FF80"
	cfg.Colors.BrightYellow = "#FFFF80"
	cfg.Colors.BrightBlue = "#8080FF"
	cfg.Colors.BrightMagenta = "#FF80FF"
	cfg.Colors.BrightCyan = "#80FFFF"
	cfg.Colors.BrightWhite = "#FFFFFF"
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

	black := Black()
	if black != lipgloss.Color("#000000") {
		t.Errorf("expected Black to be '#000000', got %s", black)
	}

	red := Red()
	if red != lipgloss.Color("#FF0000") {
		t.Errorf("expected Red to be '#FF0000', got %s", red)
	}

	magenta := Magenta()
	if magenta != lipgloss.Color("#FF00FF") {
		t.Errorf("expected Magenta to be '#FF00FF', got %s", magenta)
	}

	cyan := Cyan()
	if cyan != lipgloss.Color("#00FFFF") {
		t.Errorf("expected Cyan to be '#00FFFF', got %s", cyan)
	}

	brightBlack := BrightBlack()
	if brightBlack != lipgloss.Color("#808080") {
		t.Errorf("expected BrightBlack to be '#808080', got %s", brightBlack)
	}

	brightRed := BrightRed()
	if brightRed != lipgloss.Color("#FF8080") {
		t.Errorf("expected BrightRed to be '#FF8080', got %s", brightRed)
	}

	brightGreen := BrightGreen()
	if brightGreen != lipgloss.Color("#80FF80") {
		t.Errorf("expected BrightGreen to be '#80FF80', got %s", brightGreen)
	}

	brightYellow := BrightYellow()
	if brightYellow != lipgloss.Color("#FFFF80") {
		t.Errorf("expected BrightYellow to be '#FFFF80', got %s", brightYellow)
	}

	brightBlue := BrightBlue()
	if brightBlue != lipgloss.Color("#8080FF") {
		t.Errorf("expected BrightBlue to be '#8080FF', got %s", brightBlue)
	}

	brightMagenta := BrightMagenta()
	if brightMagenta != lipgloss.Color("#FF80FF") {
		t.Errorf("expected BrightMagenta to be '#FF80FF', got %s", brightMagenta)
	}

	brightCyan := BrightCyan()
	if brightCyan != lipgloss.Color("#80FFFF") {
		t.Errorf("expected BrightCyan to be '#80FFFF', got %s", brightCyan)
	}

	brightWhite := BrightWhite()
	if brightWhite != lipgloss.Color("#FFFFFF") {
		t.Errorf("expected BrightWhite to be '#FFFFFF', got %s", brightWhite)
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
			input: []string{"primary=#FF0000,yellow=#FFFF00,green=#00FF00,blue=#0000FF,black=#000000,red=#FF0000,magenta=#FF00FF,cyan=#00FFFF,bright-black=#808080,bright-red=#FF8080,bright-green=#80FF80,bright-yellow=#FFFF80,bright-blue=#8080FF,bright-magenta=#FF80FF,bright-cyan=#80FFFF,bright-white=#FFFFFF"},
			expected: &config.ColorConfig{
				Primary:       config.ConfigString("#FF0000"),
				Yellow:        config.ConfigString("#FFFF00"),
				Green:         config.ConfigString("#00FF00"),
				Blue:          config.ConfigString("#0000FF"),
				Black:         config.ConfigString("#000000"),
				Red:           config.ConfigString("#FF0000"),
				Magenta:       config.ConfigString("#FF00FF"),
				Cyan:          config.ConfigString("#00FFFF"),
				BrightBlack:   config.ConfigString("#808080"),
				BrightRed:     config.ConfigString("#FF8080"),
				BrightGreen:   config.ConfigString("#80FF80"),
				BrightYellow:  config.ConfigString("#FFFF80"),
				BrightBlue:    config.ConfigString("#8080FF"),
				BrightMagenta: config.ConfigString("#FF80FF"),
				BrightCyan:    config.ConfigString("#80FFFF"),
				BrightWhite:   config.ConfigString("#FFFFFF"),
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
			if result.Gray != tt.expected.Gray {
				t.Errorf("expected Gray=%s, got %s", tt.expected.Gray, result.Gray)
			}
			if result.Blue != tt.expected.Blue {
				t.Errorf("expected Blue=%s, got %s", tt.expected.Blue, result.Blue)
			}
			if result.White != tt.expected.White {
				t.Errorf("expected White=%s, got %s", tt.expected.White, result.White)
			}
			if result.Black != tt.expected.Black {
				t.Errorf("expected Black=%s, got %s", tt.expected.Black, result.Black)
			}
			if result.Red != tt.expected.Red {
				t.Errorf("expected Red=%s, got %s", tt.expected.Red, result.Red)
			}
			if result.Magenta != tt.expected.Magenta {
				t.Errorf("expected Magenta=%s, got %s", tt.expected.Magenta, result.Magenta)
			}
			if result.Cyan != tt.expected.Cyan {
				t.Errorf("expected Cyan=%s, got %s", tt.expected.Cyan, result.Cyan)
			}
			if result.BrightBlack != tt.expected.BrightBlack {
				t.Errorf("expected BrightBlack=%s, got %s", tt.expected.BrightBlack, result.BrightBlack)
			}
			if result.BrightRed != tt.expected.BrightRed {
				t.Errorf("expected BrightRed=%s, got %s", tt.expected.BrightRed, result.BrightRed)
			}
			if result.BrightGreen != tt.expected.BrightGreen {
				t.Errorf("expected BrightGreen=%s, got %s", tt.expected.BrightGreen, result.BrightGreen)
			}
			if result.BrightYellow != tt.expected.BrightYellow {
				t.Errorf("expected BrightYellow=%s, got %s", tt.expected.BrightYellow, result.BrightYellow)
			}
			if result.BrightBlue != tt.expected.BrightBlue {
				t.Errorf("expected BrightBlue=%s, got %s", tt.expected.BrightBlue, result.BrightBlue)
			}
			if result.BrightMagenta != tt.expected.BrightMagenta {
				t.Errorf("expected BrightMagenta=%s, got %s", tt.expected.BrightMagenta, result.BrightMagenta)
			}
			if result.BrightCyan != tt.expected.BrightCyan {
				t.Errorf("expected BrightCyan=%s, got %s", tt.expected.BrightCyan, result.BrightCyan)
			}
			if result.BrightWhite != tt.expected.BrightWhite {
				t.Errorf("expected BrightWhite=%s, got %s", tt.expected.BrightWhite, result.BrightWhite)
			}
		})
	}
}
