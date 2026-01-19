package colors

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/context"
)

func Primary() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Primary.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Primary))
	}

	return Blue()
}

func Border() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Border.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Border))
	}

	return Gray()
}

func Text() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Text.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Text))
	}

	return White()
}

func Muted() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Muted.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Muted))
	}

	return Gray()
}

func Selected() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Selected.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Selected))
	}

	return Primary()
}

func Success() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Success.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Success))
	}

	return Green()
}

func Warning() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Warning.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Warning))
	}

	return Yellow()
}

func Error() color.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Error.IsAssigned() {
		return lipgloss.Color(string(cfg.Theme.Error))
	}

	return Red()
}
