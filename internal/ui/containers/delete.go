package containers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/moby/moby/api/types/container"
)

type buttonOption int

const (
	confirm buttonOption = iota
	decline
)

func (bo buttonOption) String() string {
	switch bo {
	case confirm:
		return "Confirm"
	case decline:
		return "Decline"
	}

	return "Unknown"
}

type DeleteConfirmation struct {
	style lipgloss.Style
	item  *ContainerItem
}

func NewDeleteConfirmation(item *ContainerItem) DeleteConfirmation {
	var style lipgloss.Style = lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary())

	return DeleteConfirmation{style, item}
}

func (dc *DeleteConfirmation) Delete() {
	if dc.item.State == container.StateRunning {
		context.GetClient().StopContainer(dc.item.ID)
	}

	context.GetClient().RemoveContainer(dc.item.ID)
}

func (dc DeleteConfirmation) Init() tea.Cmd {
	return nil
}

func (dc DeleteConfirmation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return dc, nil
}

func (dc DeleteConfirmation) View() string {
	return dc.style.Render("Hello World!")
}
