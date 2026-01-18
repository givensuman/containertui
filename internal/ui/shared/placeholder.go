package shared

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/colors"
)

type Placeholder struct {
	Title string
}

func (p Placeholder) Init() tea.Cmd {
	return nil
}

func (p Placeholder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}

func (p Placeholder) View() string {
	style := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Align(lipgloss.Center, lipgloss.Center).
		PaddingTop(2)

	return style.Render("Placeholder: " + p.Title)
}
