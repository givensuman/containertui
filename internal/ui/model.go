package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	sidebar string
	main    string
	footer  string
	width   int
	height  int
}

func NewModel() Model {
	return Model{
		sidebar: "",
		main:    "",
		footer:  "",
		width:   80,
		height:  24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	sidebarWidth := m.width / 3
	mainWidth := m.width - sidebarWidth
	footerHeight := 3
	contentHeight := m.height - footerHeight

	sidebarBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(sidebarWidth - 2). // subtract for border
		Height(contentHeight - 2).
		Render(m.sidebar)

	mainBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(mainWidth - 2).
		Height(contentHeight - 2).
		Render(m.main)

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, mainBox)

	footerBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(m.width - 2).
		Height(footerHeight - 2).
		Render(m.footer)

	return lipgloss.JoinVertical(lipgloss.Top, content, footerBox)
}
