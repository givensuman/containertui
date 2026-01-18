package colors

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
)

func Primary() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Primary.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Primary)
	}

	return Blue()
}

func Border() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Border.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Border)
	}

	return Gray()
}

func Text() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Text.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Text)
	}

	return White()
}

func Muted() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Muted.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Muted)
	}

	return Gray()
}

func Selected() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Selected.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Selected)
	}

	return Primary()
}

func Success() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Success.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Success)
	}

	return Green()
}

func Warning() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Warning.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Warning)
	}

	return Yellow()
}

func Error() lipgloss.Color {
	cfg := context.GetConfig()
	if cfg != nil && cfg.Theme.Error.IsAssigned() {
		return lipgloss.Color(cfg.Theme.Error)
	}

	return Red()
}
