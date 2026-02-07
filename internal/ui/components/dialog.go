package components

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/layout"
)

type DialogButton struct {
	Label  string
	Action any
}

type DialogType int

const (
	DialogTypeInfo DialogType = iota
	DialogTypeSuccess
	DialogTypeWarning
	DialogTypeError
)

type DialogSize int

const (
	DialogSizeSmall DialogSize = iota
	DialogSizeMedium
	DialogSizeLarge
)

type Dialog struct {
	base.Component
	style          lipgloss.Style
	message        string
	buttons        []DialogButton
	selectedButton int
	width          int
	height         int
	dialogType     DialogType
	dialogSize     DialogSize
}

var (
	_ base.ComponentModel = (*Dialog)(nil)
	_ fmt.Stringer        = (*Dialog)(nil)
)

// NewDialog creates a new dialog with default medium size and info type
func NewDialog(message string, buttons []DialogButton) Dialog {
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeInfo)
}

// NewDialogWithOptions creates a new dialog with custom size and type
func NewDialogWithOptions(message string, buttons []DialogButton, size DialogSize, dialogType DialogType) Dialog {
	width, height := state.GetWindowSize()

	if len(buttons) == 0 {
		buttons = []DialogButton{{Label: "OK", Action: base.NewDialogAction[any](base.CloseDialog, nil)}}
	}

	dialog := Dialog{
		message:        message,
		buttons:        buttons,
		selectedButton: 0,
		width:          width,
		height:         height,
		dialogType:     dialogType,
		dialogSize:     size,
	}

	dialog.updateStyle()
	return dialog
}

// Helper constructors for common dialog types
func NewInfoDialog(message string, buttons []DialogButton) Dialog {
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeInfo)
}

func NewSuccessDialog(message string) Dialog {
	buttons := []DialogButton{{Label: "OK", Action: base.NewDialogAction[any](base.CloseDialog, nil)}}
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeSuccess)
}

func NewWarningDialog(message string, buttons []DialogButton) Dialog {
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeWarning)
}

func NewErrorDialog(message string) Dialog {
	buttons := []DialogButton{{Label: "OK", Action: base.NewDialogAction[any](base.CloseDialog, nil)}}
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeError)
}

func NewConfirmDialog(message string, onConfirm any) Dialog {
	buttons := []DialogButton{
		{Label: "Cancel", Action: base.NewDialogAction[any](base.CloseDialog, nil)},
		{Label: "Confirm", Action: onConfirm},
	}
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeInfo)
}

func NewDeleteDialog(message string, onDelete any) Dialog {
	buttons := []DialogButton{
		{Label: "Cancel", Action: base.NewDialogAction[any](base.CloseDialog, nil)},
		{Label: "Delete", Action: onDelete},
	}
	return NewDialogWithOptions(message, buttons, DialogSizeMedium, DialogTypeWarning)
}

// NewProgressDialog creates a dialog for showing progress operations
// It has no buttons and is typically closed programmatically when the operation completes
func NewProgressDialog(message string) Dialog {
	return NewDialogWithOptions(message, []DialogButton{}, DialogSizeSmall, DialogTypeInfo)
}

// updateStyle applies the current dialog type and size to the style
func (dialog *Dialog) updateStyle() {
	// Get border color based on dialog type
	var borderColor = colors.Primary()
	switch dialog.dialogType {
	case DialogTypeSuccess:
		borderColor = colors.Success()
	case DialogTypeWarning:
		borderColor = colors.Warning()
	case DialogTypeError:
		borderColor = colors.Error()
	case DialogTypeInfo:
		borderColor = colors.Primary()
	}

	// Create base style
	dialog.style = lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(borderColor)

	// Calculate dimensions based on size
	layoutManager := layout.NewLayoutManager(dialog.width, dialog.height)
	var dimensions layout.Dimensions
	switch dialog.dialogSize {
	case DialogSizeSmall:
		dimensions = layoutManager.CalculateSmall(dialog.style)
	case DialogSizeMedium:
		dimensions = layoutManager.CalculateMedium(dialog.style)
	case DialogSizeLarge:
		dimensions = layoutManager.CalculateLarge(dialog.style)
	default:
		dimensions = layoutManager.CalculateMedium(dialog.style)
	}

	dialog.style = dialog.style.Width(dimensions.Width).Height(dimensions.Height)
}

func (dialog *Dialog) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dialog.width = msg.Width
	dialog.height = msg.Height
	dialog.updateStyle()
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

// renderButtons renders the dialog buttons
func (dialog Dialog) renderButtons() string {
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

	return lipgloss.JoinHorizontal(lipgloss.Center, buttonViews...)
}

func (dialog Dialog) View() tea.View {
	buttonsView := dialog.renderButtons()

	// Get the content width (dialog width minus borders and padding)
	contentWidth := dialog.style.GetWidth()
	if contentWidth > 0 {
		frameSize := dialog.style.GetHorizontalFrameSize()
		contentWidth = contentWidth - frameSize
	}

	// Center the message and buttons within the content width
	messageStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
	buttonsStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)

	content := dialog.style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		messageStyle.Render(dialog.message),
		"",
		buttonsStyle.Render(buttonsView),
	))

	return tea.NewView(content)
}

func (dialog Dialog) String() string {
	buttonsView := dialog.renderButtons()

	// Get the content width (dialog width minus borders and padding)
	contentWidth := dialog.style.GetWidth()
	if contentWidth > 0 {
		frameSize := dialog.style.GetHorizontalFrameSize()
		contentWidth = contentWidth - frameSize
	}

	// Center the message and buttons within the content width
	messageStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
	buttonsStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)

	return dialog.style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		messageStyle.Render(dialog.message),
		"",
		buttonsStyle.Render(buttonsView),
	))
}
