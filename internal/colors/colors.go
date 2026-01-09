// Package colors provides color management for the application.
package colors

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
)

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

func Yellow() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Yellow.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Yellow)
	}

	return lipgloss.Color(ColorBrightYellow.String())
}

func Green() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Green.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Green)
	}

	return lipgloss.Color(ColorBrightGreen.String())
}

func Gray() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Gray.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Gray)
	}

	return lipgloss.Color(ColorBrightBlack.String())
}

func Blue() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Blue.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Blue)
	}

	return lipgloss.Color(ColorBlue.String())
}

func White() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.White.IsAssigned() {
		return lipgloss.Color(cfg.Colors.White)
	}

	return lipgloss.Color(ColorWhite.String())
}

func Primary() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Primary.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Primary)
	}

	return Blue()
}

func Black() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Black.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Black)
	}

	return lipgloss.Color(ColorBlack.String())
}

func Red() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Red.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Red)
	}

	return lipgloss.Color(ColorRed.String())
}

func Magenta() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Magenta.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Magenta)
	}

	return lipgloss.Color(ColorMagenta.String())
}

func Cyan() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.Cyan.IsAssigned() {
		return lipgloss.Color(cfg.Colors.Cyan)
	}

	return lipgloss.Color(ColorCyan.String())
}

func BrightBlack() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightBlack.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightBlack)
	}

	return lipgloss.Color(ColorBrightBlack.String())
}

func BrightRed() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightRed.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightRed)
	}

	return lipgloss.Color(ColorBrightRed.String())
}

func BrightGreen() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightGreen.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightGreen)
	}

	return lipgloss.Color(ColorBrightGreen.String())
}

func BrightYellow() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightYellow.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightYellow)
	}

	return lipgloss.Color(ColorBrightYellow.String())
}

func BrightBlue() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightBlue.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightBlue)
	}

	return lipgloss.Color(ColorBrightBlue.String())
}

func BrightMagenta() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightMagenta.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightMagenta)
	}

	return lipgloss.Color(ColorBrightMagenta.String())
}

func BrightCyan() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightCyan.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightCyan)
	}

	return lipgloss.Color(ColorBrightCyan.String())
}

func BrightWhite() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg.Colors.BrightWhite.IsAssigned() {
		return lipgloss.Color(cfg.Colors.BrightWhite)
	}

	return lipgloss.Color(ColorBrightWhite.String())
}
