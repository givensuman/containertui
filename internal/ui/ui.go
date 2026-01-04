package ui

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	listModel ListModel
}

func (m Model) Init() tea.Cmd {
	return m.listModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	v := tea.NewView(docStyle.Render(m.listModel.View()))
	v.AltScreen = true
	return v
}

func Start() {
	listKeys := newListKeyMap()
	l := NewList()
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.pauseContainer,
			listKeys.unpauseContainer,
			listKeys.startContainer,
			listKeys.stopContainer,
			listKeys.toggleSelect,
		}
	}
	lm := ListModel{
		list: l,
		keys: listKeys,
	}
	m := Model{
		listModel: lm,
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
