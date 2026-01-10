package containers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type ContainerLogs struct {
	shared.Component
	container *ContainerItem
}

var (
	_ tea.Model            = (*ContainerLogs)(nil)
	_ shared.ComponentModel = (*ContainerLogs)(nil)
)

func newContainerLogs(container *ContainerItem) ContainerLogs {
	return ContainerLogs{
		container: container,
	}
}

func (cl *ContainerLogs) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	cl.WindowWidth = msg.Width
	cl.WindowHeight = msg.Height
}

func (cl ContainerLogs) Init() tea.Cmd {
	return nil
}

func (cl ContainerLogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cl.UpdateWindowDimensions(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEscape.String(), tea.KeyEsc.String():
			cmds = append(cmds, CloseOverlay())
		}
	}

	return cl, tea.Batch(cmds...)
}

func (cl ContainerLogs) View() string {
	return "Hello, world!"
}
