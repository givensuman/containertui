package containers

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/types"
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
	types.Component
	style               lipgloss.Style
	requestedContainers []*ContainerItem
	hoveredButtonOption buttonOption
}

var (
	_ tea.Model            = (*DeleteConfirmation)(nil)
	_ types.ComponentModel = (*DeleteConfirmation)(nil)
)

func newDeleteConfirmation(requestedContainers ...*ContainerItem) DeleteConfirmation {
	width, height := context.GetWindowSize()

	style := lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary()).
		MaxHeight(height / 2).
		MaxWidth(width / 2)

	return DeleteConfirmation{
		style:               style,
		requestedContainers: requestedContainers,
		hoveredButtonOption: decline,
	}
}

func (dc *DeleteConfirmation) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dc.WindowWidth = msg.Width
	dc.WindowHeight = msg.Height
}

func (dc DeleteConfirmation) Init() tea.Cmd {
	return nil
}

func (dc DeleteConfirmation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dc.UpdateWindowDimensions(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEscape.String(), tea.KeyEsc.String():
			cmds = append(cmds, func() tea.Msg { return MessageCloseOverlay{} })

		case tea.KeyTab.String(), tea.KeyShiftTab.String():
			switch dc.hoveredButtonOption {
			case confirm:
				dc.hoveredButtonOption = decline
			case decline:
				dc.hoveredButtonOption = confirm
			}

		case "l", tea.KeyRight.String():
			if dc.hoveredButtonOption == decline {
				dc.hoveredButtonOption = confirm
			}

		case "h", tea.KeyLeft.String():
			if dc.hoveredButtonOption == confirm {
				dc.hoveredButtonOption = decline
			}

		case tea.KeyEnter.String():
			if dc.hoveredButtonOption == confirm {
				cmds = append(cmds, func() tea.Msg { return MessageConfirmDelete{} })
			}

			cmds = append(cmds, func() tea.Msg { return MessageCloseOverlay{} })
		}
	}

	return dc, tea.Batch(cmds...)
}

func (dc DeleteConfirmation) View() string {
	hoveredButtonStyle := lipgloss.NewStyle().
		Background(colors.Primary()).
		Bold(true).
		Foreground(colors.Black()).
		Padding(0, 1)

	defaultButtonStyle := lipgloss.NewStyle().
		Background(colors.Gray()).
		Foreground(colors.White()).
		Padding(0, 1)

	confirmButton := confirm.String()
	declineButton := decline.String()

	if dc.hoveredButtonOption == confirm {
		confirmButton = hoveredButtonStyle.Background(colors.Red()).Render(confirmButton)
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
	if len(dc.requestedContainers) == 1 {
		message = fmt.Sprintf("Are you sure you want to delete %s?", dc.requestedContainers[0].Name)
	} else {
		message = fmt.Sprintf("Are you sure you want to delete the %d selected containers?", len(dc.requestedContainers))
	}

	if dc.hoveredButtonOption == confirm {
		dc.style = dc.style.BorderForeground(colors.Red())
	}

	return dc.style.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		message,
		"   ",
		buttons,
	))
}
