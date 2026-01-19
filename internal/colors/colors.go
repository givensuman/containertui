// Package colors provides color management for the application.
package colors

import (
	"image/color"

	"charm.land/lipgloss/v2"
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

func Black() color.Color {
	return lipgloss.Color(ColorBlack.String())
}

func Red() color.Color {
	return lipgloss.Color(ColorRed.String())
}

func Green() color.Color {
	return lipgloss.Color(ColorBrightGreen.String())
}

func Yellow() color.Color {
	return lipgloss.Color(ColorBrightYellow.String())
}

func Blue() color.Color {
	return lipgloss.Color(ColorBlue.String())
}

func Magenta() color.Color {
	return lipgloss.Color(ColorMagenta.String())
}

func Cyan() color.Color {
	return lipgloss.Color(ColorCyan.String())
}

func White() color.Color {
	return lipgloss.Color(ColorWhite.String())
}

func BrightBlack() color.Color {
	return lipgloss.Color(ColorBrightBlack.String())
}

func BrightRed() color.Color {
	return lipgloss.Color(ColorBrightRed.String())
}

func BrightGreen() color.Color {
	return lipgloss.Color(ColorBrightGreen.String())
}

func BrightYellow() color.Color {
	return lipgloss.Color(ColorBrightYellow.String())
}

func BrightBlue() color.Color {
	return lipgloss.Color(ColorBrightBlue.String())
}

func BrightMagenta() color.Color {
	return lipgloss.Color(ColorBrightMagenta.String())
}

func BrightCyan() color.Color {
	return lipgloss.Color(ColorBrightCyan.String())
}

func BrightWhite() color.Color {
	return lipgloss.Color(ColorBrightWhite.String())
}

func Gray() color.Color {
	return lipgloss.Color(ColorBrightBlack.String())
}
