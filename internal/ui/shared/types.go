package shared

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Component struct {
	WindowWidth  int
	WindowHeight int
}

type ComponentModel interface {
	tea.Model
	UpdateWindowDimensions(msg tea.WindowSizeMsg)
}
