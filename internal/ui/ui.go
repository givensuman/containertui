// Package ui implements the terminal user interface
package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/containers"
)

// Model represents the top-level Bubbletea UI model.
// Contains terminal dimensions and the containers model (main TUI view).
type Model struct {
	width           int              // current terminal width
	height          int              // current terminal height
	containersModel containers.Model // main containers view/model
}

func NewModel() Model {
	width, height := context.GetWindowSize()

	return Model{
		width:           width,
		height:          height,
		containersModel: containers.New(),
	}
}

// Init performs any initial commands for the Bubbletea app
// (no async initialization needed here)
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all Bubbletea update messages dispatched by the tea runtime.
// Manages window resizing and quit keys, delegates other updates to containersModel.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update local size and global context
		m.width = msg.Width
		m.height = msg.Height
		context.SetWindowSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Handle quit signals (Ctrl-C, Ctrl-D)
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit
		}
	}

	// Forward non-window/non-quit messages to containers view model
	containersModel, cmd := m.containersModel.Update(msg)
	m.containersModel = containersModel.(containers.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the terminal as a string (delegated to containersModel).
func (m Model) View() string {
	return m.containersModel.View()
}

// Start the Bubbletea UI main loop.
// Returns error if Bubbletea program terminates abnormally.
func Start() error {
	model := NewModel()

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
