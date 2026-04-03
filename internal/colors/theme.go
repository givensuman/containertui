package colors

import (
	"image/color"
	"os"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/state"
)

var (
	backgroundOnce   sync.Once
	detectedDarkMode bool
)

func hasDarkBackground() bool {
	backgroundOnce.Do(func() {
		detectedDarkMode = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	})

	return detectedDarkMode
}

func adaptiveColor(light, dark color.Color) color.Color {
	return lipgloss.LightDark(hasDarkBackground())(light, dark)
}

func Primary() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Primary.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Primary))
	}

	return adaptiveColor(lipgloss.Color(ColorBlue.String()), lipgloss.Color(ColorBrightBlue.String()))
}

func Border() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Border.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Border))
	}

	return adaptiveColor(lipgloss.Color(ColorBlack.String()), lipgloss.Color(ColorBrightBlack.String()))
}

func Text() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Text.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Text))
	}

	return adaptiveColor(lipgloss.Color(ColorBlack.String()), lipgloss.Color(ColorBrightWhite.String()))
}

func Muted() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Muted.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Muted))
	}

	return adaptiveColor(lipgloss.Color(ColorBrightBlack.String()), lipgloss.Color(ColorBrightBlack.String()))
}

// PrimaryText returns a high-contrast foreground for primary backgrounds.
func PrimaryText() color.Color {
	return adaptiveColor(lipgloss.Color(ColorBrightWhite.String()), lipgloss.Color(ColorBlack.String()))
}

func Selected() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Selected.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Selected))
	}

	return Primary()
}

func Success() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Success.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Success))
	}

	return adaptiveColor(lipgloss.Color(ColorGreen.String()), lipgloss.Color(ColorBrightGreen.String()))
}

func Warning() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Warning.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Warning))
	}

	return adaptiveColor(lipgloss.Color(ColorYellow.String()), lipgloss.Color(ColorBrightYellow.String()))
}

func Error() color.Color {
	cfg := state.GetConfig()
	if cfg != nil && cfg.Theme.Error.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Error))
	}

	return adaptiveColor(lipgloss.Color(ColorRed.String()), lipgloss.Color(ColorBrightRed.String()))
}
