package containers

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type buttonOption int

const (
	confirm buttonOption = iota
	decline
)

func (bo buttonOption) String() string {
	switch bo {
	case confirm:
		return "Delete"
	case decline:
		return "Cancel"
	}

	return "Unknown"
}

type DeleteConfirmation struct {
	shared.Component
	style               lipgloss.Style
	requestedContainers []*ContainerItem
	hoveredButtonOption buttonOption
}

var (
	_ tea.Model             = (*DeleteConfirmation)(nil)
	_ shared.ComponentModel = (*DeleteConfirmation)(nil)
)

func newDeleteConfirmation(requestedContainers ...*ContainerItem) DeleteConfirmation {
	width, height := context.GetWindowSize()

	style := lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary()).
		Align(lipgloss.Center)

	lm := shared.NewLayoutManager(width, height)
	dims := lm.CalculateModal(style)

	style = style.Width(dims.Width).Height(dims.Height)

	return DeleteConfirmation{
		style:               style,
		requestedContainers: requestedContainers,
		hoveredButtonOption: decline,
	}
}

func (model *DeleteConfirmation) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.WindowWidth = msg.Width
	model.WindowHeight = msg.Height

	layoutManager := shared.NewLayoutManager(msg.Width, msg.Height)
	dimensions := layoutManager.CalculateModal(model.style)

	model.style = model.style.Width(dimensions.Width).Height(dimensions.Height)
}

func (model DeleteConfirmation) Init() tea.Cmd {
	return nil
}

func (model DeleteConfirmation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.UpdateWindowDimensions(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEscape.String(), tea.KeyEsc.String():
			cmds = append(cmds, CloseOverlay())

		case tea.KeyTab.String(), tea.KeyShiftTab.String():
			switch model.hoveredButtonOption {
			case confirm:
				model.hoveredButtonOption = decline
			case decline:
				model.hoveredButtonOption = confirm
			}

		case "l", tea.KeyRight.String():
			if model.hoveredButtonOption == decline {
				model.hoveredButtonOption = confirm
			}

		case "h", tea.KeyLeft.String():
			if model.hoveredButtonOption == confirm {
				model.hoveredButtonOption = decline
			}

		case tea.KeyEnter.String():
			if model.hoveredButtonOption == confirm {
				cmds = append(cmds, func() tea.Msg { return MessageConfirmDelete{} })
			}

			cmds = append(cmds, func() tea.Msg { return MessageCloseOverlay{} })
		}
	}

	return model, tea.Batch(cmds...)
}

func (model DeleteConfirmation) View() string {
	hoveredButtonStyle := lipgloss.NewStyle().
		Background(colors.Primary()).
		Bold(true).
		Foreground(colors.Text()).
		Padding(0, 1)

	defaultButtonStyle := lipgloss.NewStyle().
		Background(colors.Muted()).
		Foreground(colors.Text()).
		Padding(0, 1)

	confirmButton := confirm.String()
	declineButton := decline.String()

	if model.hoveredButtonOption == confirm {
		confirmButton = hoveredButtonStyle.Background(colors.Error()).Render(confirmButton)
		declineButton = defaultButtonStyle.Render(declineButton)
	} else {
		confirmButton = defaultButtonStyle.Render(confirmButton)
		declineButton = hoveredButtonStyle.Render(declineButton)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		declineButton,
		"   ",
		confirmButton,
	)

	var message string
	if len(model.requestedContainers) == 1 {
		message = fmt.Sprintf("Are you sure you want to delete %s?", model.requestedContainers[0].Name)
	} else {
		message = fmt.Sprintf("Are you sure you want to delete the %d selected containers?", len(model.requestedContainers))
	}

	if model.hoveredButtonOption == confirm {
		model.style = model.style.BorderForeground(colors.Error())
	}

	return model.style.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		message,
		"",
		buttons,
	))
}
