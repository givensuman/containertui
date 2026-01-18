// Package colors provides color management for the application.
package colors

import (
	"github.com/charmbracelet/lipgloss"
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

func Black() lipgloss.Color {
	return lipgloss.Color(ColorBlack.String())
}

func Red() lipgloss.Color {
	return lipgloss.Color(ColorRed.String())
}

func Green() lipgloss.Color {
	return lipgloss.Color(ColorBrightGreen.String())
}

func Yellow() lipgloss.Color {
	return lipgloss.Color(ColorBrightYellow.String())
}

func Blue() lipgloss.Color {
	return lipgloss.Color(ColorBlue.String())
}

func Magenta() lipgloss.Color {
	return lipgloss.Color(ColorMagenta.String())
}

func Cyan() lipgloss.Color {
	return lipgloss.Color(ColorCyan.String())
}

func White() lipgloss.Color {
	return lipgloss.Color(ColorWhite.String())
}

func BrightBlack() lipgloss.Color {
	return lipgloss.Color(ColorBrightBlack.String())
}

func BrightRed() lipgloss.Color {
	return lipgloss.Color(ColorBrightRed.String())
}

func BrightGreen() lipgloss.Color {
	return lipgloss.Color(ColorBrightGreen.String())
}

func BrightYellow() lipgloss.Color {
	return lipgloss.Color(ColorBrightYellow.String())
}

func BrightBlue() lipgloss.Color {
	return lipgloss.Color(ColorBrightBlue.String())
}

func BrightMagenta() lipgloss.Color {
	return lipgloss.Color(ColorBrightMagenta.String())
}

func BrightCyan() lipgloss.Color {
	return lipgloss.Color(ColorBrightCyan.String())
}

func BrightWhite() lipgloss.Color {
	return lipgloss.Color(ColorBrightWhite.String())
}

func Gray() lipgloss.Color {
	return lipgloss.Color(ColorBrightBlack.String())
}
