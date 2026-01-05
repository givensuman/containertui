// Package colors provides color management for the application.
package colors

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
)

// ANSI color constants
const (
	ColorBlack ANSIColor = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorBrightBlack
	ColorBrightRed
	ColorBrightGreen
	ColorBrightYellow
	ColorBrightBlue
	ColorBrightMagenta
	ColorBrightCyan
	ColorBrightWhite
)

type ANSIColor int

// String returns the ANSI color code
func (c ANSIColor) String() string {
	return ansiColorMap[c]
}

var ansiColorMap = map[ANSIColor]string{
	ColorBlack:         "0",
	ColorRed:           "1",
	ColorGreen:         "2",
	ColorYellow:        "3",
	ColorBlue:          "4",
	ColorMagenta:       "5",
	ColorCyan:          "6",
	ColorWhite:         "7",
	ColorBrightBlack:   "8",
	ColorBrightRed:     "9",
	ColorBrightGreen:   "10",
	ColorBrightYellow:  "11",
	ColorBrightBlue:    "12",
	ColorBrightMagenta: "13",
	ColorBrightCyan:    "14",
	ColorBrightWhite:   "15",
}

// Yellow returns the yellow color, using config override if available
func Yellow() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Yellow.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Yellow)
	}
	return lipgloss.Color(ColorBrightYellow.String())
}

// Green returns the green color, using config override if available
func Green() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Green.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Green)
	}
	return lipgloss.Color(ColorBrightGreen.String())
}

// Red returns the red color, using config override if available
func Red() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Red.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Red)
	}
	return lipgloss.Color(ColorBrightRed.String())
}

// Blue returns the blue color, using config override if available
func Blue() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Blue.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Blue)
	}
	return lipgloss.Color(ColorBrightBlue.String())
}

// Primary returns the primary color, using config override if available, defaults to Blue
func Primary() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Primary.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Primary)
	}
	return Blue() // Default to blue for primary
}
