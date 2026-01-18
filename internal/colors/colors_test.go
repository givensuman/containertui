package colors

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/context"
)

func TestANSIColors(t *testing.T) {
	if ColorBlack.String() != "0" {
		t.Errorf("expected ColorBlack to be '0', got %s", ColorBlack.String())
	}
	if ColorBrightYellow.String() != "11" {
		t.Errorf("expected ColorBrightYellow to be '11', got %s", ColorBrightYellow.String())
	}
}

func TestColorFunctions(t *testing.T) {
	context.SetConfig(config.DefaultConfig())

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
}

func TestColorOverrides(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Theme.Primary = "#89b4fa"
	cfg.Theme.Border = "#585b70"
	cfg.Theme.Text = "#cdd6f4"
	cfg.Theme.Muted = "#585b70"
	cfg.Theme.Selected = "#89b4fa"
	cfg.Theme.Success = "#a6e3a1"
	cfg.Theme.Warning = "#f9e2af"
	cfg.Theme.Error = "#f38ba8"
	context.SetConfig(cfg)

	primary := Primary()
	if primary != lipgloss.Color("#89b4fa") {
		t.Errorf("expected Primary to be '#89b4fa', got %s", primary)
	}

	border := Border()
	if border != lipgloss.Color("#585b70") {
		t.Errorf("expected Border to be '#585b70', got %s", border)
	}

	text := Text()
	if text != lipgloss.Color("#cdd6f4") {
		t.Errorf("expected White/Text to be '#cdd6f4', got %s", text)
	}

	muted := Muted()
	if muted != lipgloss.Color("#585b70") {
		t.Errorf("expected Gray/Muted to be '#585b70', got %s", muted)
	}

	selected := Selected()
	if selected != lipgloss.Color("#89b4fa") {
		t.Errorf("expected Selected to be '#89b4fa', got %s", selected)
	}

	success := Success()
	if success != lipgloss.Color("#a6e3a1") {
		t.Errorf("expected Green/Success to be '#a6e3a1', got %s", success)
	}

	warning := Warning()
	if warning != lipgloss.Color("#f9e2af") {
		t.Errorf("expected Yellow/Warning to be '#f9e2af', got %s", warning)
	}

	errColor := Error()
	if errColor != lipgloss.Color("#f38ba8") {
		t.Errorf("expected Red/Error to be '#f38ba8', got %s", errColor)
	}
}

func TestParseColors(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		expected    *config.ThemeConfig
		expectError bool
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: &config.ThemeConfig{},
		},
		{
			name:  "single color",
			input: []string{"primary=#89b4fa"},
			expected: &config.ThemeConfig{
				Primary: config.ConfigString("#89b4fa"),
			},
		},
		{
			name:  "multiple colors in one string",
			input: []string{"primary=#89b4fa,warning=#f9e2af,success=#a6e3a1"},
			expected: &config.ThemeConfig{
				Primary: config.ConfigString("#89b4fa"),
				Warning: config.ConfigString("#f9e2af"),
				Success: config.ConfigString("#a6e3a1"),
			},
		},
		{
			name:  "multiple strings",
			input: []string{"primary=#89b4fa", "warning=#f9e2af", "success=#a6e3a1"},
			expected: &config.ThemeConfig{
				Primary: config.ConfigString("#89b4fa"),
				Warning: config.ConfigString("#f9e2af"),
				Success: config.ConfigString("#a6e3a1"),
			},
		},
		{
			name:  "mixed format",
			input: []string{"primary=#89b4fa,warning=#f9e2af", "success=#a6e3a1"},
			expected: &config.ThemeConfig{
				Primary: config.ConfigString("#89b4fa"),
				Warning: config.ConfigString("#f9e2af"),
				Success: config.ConfigString("#a6e3a1"),
			},
		},
		{
			name:  "with spaces",
			input: []string{"primary=#89b4fa, warning=#f9e2af"},
			expected: &config.ThemeConfig{
				Primary: config.ConfigString("#89b4fa"),
				Warning: config.ConfigString("#f9e2af"),
			},
		},
		{
			name:        "invalid format - no equals",
			input:       []string{"primary#89b4fa"},
			expectError: true,
		},
		{
			name:        "invalid format - too many equals",
			input:       []string{"primary=#89b4fa=000"},
			expectError: true,
		},
		{
			name:        "unknown color key",
			input:       []string{"unknown=#89b4fa"},
			expectError: true,
		},
		{
			name:  "all colors",
			input: []string{"primary=#89b4fa,border=#585b70,text=#cdd6f4,muted=#585b70,selected=#89b4fa,success=#a6e3a1,warning=#f9e2af,error=#f38ba8"},
			expected: &config.ThemeConfig{
				Primary:  config.ConfigString("#89b4fa"),
				Border:   config.ConfigString("#585b70"),
				Text:     config.ConfigString("#cdd6f4"),
				Muted:    config.ConfigString("#585b70"),
				Selected: config.ConfigString("#89b4fa"),
				Success:  config.ConfigString("#a6e3a1"),
				Warning:  config.ConfigString("#f9e2af"),
				Error:    config.ConfigString("#f38ba8"),
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
			if result.Border != tt.expected.Border {
				t.Errorf("expected Border=%s, got %s", tt.expected.Border, result.Border)
			}
			if result.Text != tt.expected.Text {
				t.Errorf("expected Text=%s, got %s", tt.expected.Text, result.Text)
			}
			if result.Muted != tt.expected.Muted {
				t.Errorf("expected Muted=%s, got %s", tt.expected.Muted, result.Muted)
			}
			if result.Selected != tt.expected.Selected {
				t.Errorf("expected Selected=%s, got %s", tt.expected.Selected, result.Selected)
			}
			if result.Success != tt.expected.Success {
				t.Errorf("expected Success=%s, got %s", tt.expected.Success, result.Success)
			}
			if result.Warning != tt.expected.Warning {
				t.Errorf("expected Warning=%s, got %s", tt.expected.Warning, result.Warning)
			}
			if result.Error != tt.expected.Error {
				t.Errorf("expected Error=%s, got %s", tt.expected.Error, result.Error)
			}
		})
	}
}
