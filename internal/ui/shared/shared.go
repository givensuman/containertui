// Package shared defines shared UI logic
package shared

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SimpleViewModel is a simple tea.Model that just renders a string.
// It is useful for wrapping a rendered string to pass as a model (e.g. for backgrounds).
type SimpleViewModel struct {
	Content string
}

func (m SimpleViewModel) Init() tea.Cmd {
	return nil
}

func (m SimpleViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m SimpleViewModel) View() string {
	return m.Content
}
