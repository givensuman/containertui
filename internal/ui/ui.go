// Package ui implements the terminal user interface.
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/containers"
)

type Model struct {
	width           int
	height          int
	containersModel containers.Model
}

func NewModel() Model {
	width, height := context.GetWindowSize()

	return Model{
		width:           width,
		height:          height,
		containersModel: containers.NewContainersList(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		context.SetWindowSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "ctrl+d":
			return m, tea.Quit
		}
	}

	containersModel, cmd := m.containersModel.Update(msg)
	m.containersModel = containersModel.(containers.Model)

	return m, cmd
}

func (m Model) View() string {
	return m.containersModel.View()
}

// Start the UI rendering loop.
func Start() error {
	model := NewModel()

	file, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		panic(err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
