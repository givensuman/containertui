package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/layout"
)

// DialogButton defines a button in the dialog
type DialogButton struct {
	Label  string
	Action base.SmartDialogAction // Empty action means Cancel/Close
	IsSafe bool                   // True = Primary/Safe color, False = Danger/Red
}

type SmartDialog struct {
	base.Component
	style          lipgloss.Style
	message        string
	buttons        []DialogButton
	selectedButton int
	width          int
	height         int
}

// SmartDialog implements StringViewModel and ComponentModel
// var _ StringViewModel = (*SmartDialog)(nil)
// var _ ComponentModel = (*SmartDialog)(nil)

// NewSmartDialog creates a generic confirmation or warning dialog.
func NewSmartDialog(message string, buttons []DialogButton) SmartDialog {
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
		buttons = []DialogButton{{Label: "Cancel", IsSafe: true}}
	}

	return SmartDialog{
		style:          style,
		message:        message,
		buttons:        buttons,
		selectedButton: 0,
		width:          width,
		height:         height,
	}
}

func (dialog *SmartDialog) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dialog.width = msg.Width
	dialog.height = msg.Height

	layoutManager := layout.NewLayoutManager(msg.Width, msg.Height)
	modalDimensions := layoutManager.CalculateModal(dialog.style)
	dialog.style = dialog.style.Width(modalDimensions.Width).Height(modalDimensions.Height)
}

func (dialog SmartDialog) Init() tea.Cmd {
	return nil
}

func (dialog SmartDialog) Update(msg tea.Msg) (SmartDialog, tea.Cmd) {
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
			if selectedButton.Action.Type == "" {
				return dialog, func() tea.Msg { return base.CloseDialogMessage{} }
			}
			return dialog, func() tea.Msg { return base.ConfirmationMessage{Action: selectedButton.Action} }
		}
	}

	return dialog, nil
}

func (dialog SmartDialog) View() string {
	buttonViews := make([]string, 0, len(dialog.buttons))

	defaultButtonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Foreground(colors.Text()).
		Background(colors.Muted())

	activeSafeStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Bold(true).
		Foreground(colors.Text()).
		Background(colors.Primary())

	activeDangerStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1).
		Bold(true).
		Foreground(colors.Text()).
		Background(colors.Error())

	for index, button := range dialog.buttons {
		var buttonStyle lipgloss.Style

		if index == dialog.selectedButton {
			if button.IsSafe {
				buttonStyle = activeSafeStyle
			} else {
				buttonStyle = activeDangerStyle
			}
		} else {
			buttonStyle = defaultButtonStyle
		}

		buttonViews = append(buttonViews, buttonStyle.Render(button.Label))
	}

	buttonsView := lipgloss.JoinHorizontal(lipgloss.Center, buttonViews...)

	currentButton := dialog.buttons[dialog.selectedButton]
	renderStyle := dialog.style
	if !currentButton.IsSafe {
		renderStyle = renderStyle.BorderForeground(colors.Error())
	}

	return renderStyle.Render(lipgloss.JoinVertical(
		lipgloss.Center,
		dialog.message,
		"",
		buttonsView,
	))
}
