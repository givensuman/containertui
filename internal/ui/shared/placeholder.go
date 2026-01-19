package shared

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
)

type Placeholder struct {
	Title string
}

func (p Placeholder) Init() tea.Cmd {
	return nil
}

func (p Placeholder) Update(msg tea.Msg) (Placeholder, tea.Cmd) {
	return p, nil
}

func (p Placeholder) View() string {
	style := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Align(lipgloss.Center, lipgloss.Center).
		PaddingTop(2)

	return style.Render("Placeholder: " + p.Title)
}
