package components

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/layout"
)

type DialogButton struct {
	Label  string
	Action any
}

type Dialog struct {
	base.Component
	style          lipgloss.Style
	message        string
	buttons        []DialogButton
	selectedButton int
	width          int
	height         int
}

var (
	_ base.ComponentModel = (*Dialog)(nil)
	_ fmt.Stringer        = (*Dialog)(nil)
)

func NewDialog(message string, buttons []DialogButton) Dialog {
	width, height := context.GetWindowSize()

	style := lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary()).
		Align(lipgloss.Center)

	layoutManager := layout.NewLayoutManager(width, height)
	modalDimensions := layoutManager.CalculateModal(style)
	style = style.Width(modalDimensions.Width).Height(modalDimensions.Height)

	if len(buttons) == 0 {
		buttons = []DialogButton{{Label: "Cancel", Action: base.NewDialogAction[any](base.CloseDialog, nil)}}
	}

	return Dialog{
		style:          style,
		message:        message,
		buttons:        buttons,
		selectedButton: 0,
		width:          width,
		height:         height,
	}
}

func (dialog *Dialog) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dialog.width = msg.Width
	dialog.height = msg.Height

	layoutManager := layout.NewLayoutManager(msg.Width, msg.Height)
	modalDimensions := layoutManager.CalculateModal(dialog.style)
	dialog.style = dialog.style.Width(modalDimensions.Width).Height(modalDimensions.Height)
}

func (dialog Dialog) Init() tea.Cmd {
	return nil
}

func (dialog Dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dialog.UpdateWindowDimensions(msg)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return dialog, func() tea.Msg { return base.CloseDialogMessage{} }

		case "tab", "right", "l":
			dialog.selectedButton = (dialog.selectedButton + 1) % len(dialog.buttons)

		case "shift+tab", "left", "h":
			dialog.selectedButton = (dialog.selectedButton - 1 + len(dialog.buttons)) % len(dialog.buttons)

		case "enter":
			selectedButton := dialog.buttons[dialog.selectedButton]

			switch action := selectedButton.Action.(type) {
			case base.DialogAction[any]:
				if action.Type == base.CloseDialog {
					return dialog, func() tea.Msg { return base.CloseDialogMessage{} }
				}
				return dialog, func() tea.Msg { return base.ConfirmationMessage[any]{Action: action} }

			case base.SmartDialogAction:
				return dialog, func() tea.Msg { return base.SmartConfirmationMessage{Action: action} }

			default:
				return dialog, func() tea.Msg { return base.CloseDialogMessage{} }
			}
		}
	}

	return dialog, nil
}

func (dialog Dialog) View() tea.View {
	buttonViews := make([]string, 0, len(dialog.buttons))

	defaultButtonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Foreground(colors.Text()).
		Background(colors.Muted())

	hoveredButtonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Bold(true).
		Foreground(colors.Text()).
		Background(colors.Primary())

	for index, button := range dialog.buttons {
		buttonStyle := defaultButtonStyle

		if index == dialog.selectedButton {
			buttonStyle = hoveredButtonStyle
		}

		buttonViews = append(buttonViews, buttonStyle.Render(button.Label))
	}

	buttonsView := lipgloss.JoinHorizontal(lipgloss.Center, buttonViews...)

	content := dialog.style.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		dialog.message,
		"",
		buttonsView,
	))

	return tea.NewView(content)
}

func (dialog Dialog) String() string {
	buttonViews := make([]string, 0, len(dialog.buttons))

	defaultButtonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Foreground(colors.Text()).
		Background(colors.Muted())

	hoveredButtonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Bold(true).
		Foreground(colors.Text()).
		Background(colors.Primary())

	for index, button := range dialog.buttons {
		buttonStyle := defaultButtonStyle

		if index == dialog.selectedButton {
			buttonStyle = hoveredButtonStyle
		}

		buttonViews = append(buttonViews, buttonStyle.Render(button.Label))
	}

	buttonsView := lipgloss.JoinHorizontal(lipgloss.Center, buttonViews...)

	return dialog.style.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		dialog.message,
		"",
		buttonsView,
	))
}
