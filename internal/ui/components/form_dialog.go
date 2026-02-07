package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/layout"
)

type FormField struct {
	Label       string
	Placeholder string
	Value       string
	Validator   func(string) error
	Required    bool
}

type FormDialog struct {
	base.Component
	title          string
	fields         []FormField
	textInputs     []textinput.Model
	focusIndex     int
	action         base.SmartDialogAction
	style          lipgloss.Style
	width          int
	height         int
	errorMessage   string
	dialogSize     DialogSize
	selectedButton int
	onButtons      bool // true when navigating buttons, false when on form fields
}

var (
	_ base.ComponentModel = (*FormDialog)(nil)
	_ fmt.Stringer        = (*FormDialog)(nil)
)

// NewFormDialog creates a new form dialog with default medium size
func NewFormDialog(title string, fields []FormField, action base.SmartDialogAction, metadata map[string]any) FormDialog {
	return NewFormDialogWithSize(title, fields, action, metadata, DialogSizeMedium)
}

// NewFormDialogWithSize creates a new form dialog with custom size
func NewFormDialogWithSize(title string, fields []FormField, action base.SmartDialogAction, metadata map[string]any, size DialogSize) FormDialog {
	width, height := context.GetWindowSize()

	// Initialize text inputs
	textInputs := make([]textinput.Model, len(fields))

	// We'll set width after creating the style
	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.SetValue(field.Value)
		ti.CharLimit = 256

		if i == 0 {
			ti.Focus()
		}

		textInputs[i] = ti
	}

	if metadata == nil {
		metadata = make(map[string]any)
	}

	if action.Payload == nil {
		action.Payload = metadata
	}

	dialog := FormDialog{
		title:          title,
		fields:         fields,
		textInputs:     textInputs,
		focusIndex:     0,
		action:         action,
		width:          width,
		height:         height,
		errorMessage:   "",
		dialogSize:     size,
		selectedButton: 0,
		onButtons:      false,
	}

	dialog.updateStyle()
	return dialog
}

// updateStyle applies the current dialog size to the style
func (dialog *FormDialog) updateStyle() {
	// Form dialogs always use primary color border
	dialog.style = lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary())

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

	// Update text input widths based on content width
	for i := range dialog.textInputs {
		dialog.textInputs[i].SetWidth(dimensions.ContentWidth - 4) // Account for some padding
	}
}

func (dialog *FormDialog) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dialog.width = msg.Width
	dialog.height = msg.Height
	dialog.updateStyle()
}

func (dialog FormDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (dialog FormDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dialog.UpdateWindowDimensions(msg)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return dialog, func() tea.Msg { return base.CloseDialogMessage{} }

		case "tab", "down":
			if dialog.onButtons {
				// Cycle between Cancel and Submit buttons
				dialog.selectedButton = (dialog.selectedButton + 1) % 2
			} else {
				// Move to next field or to buttons if at last field
				if dialog.focusIndex < len(dialog.textInputs)-1 {
					dialog.textInputs[dialog.focusIndex].Blur()
					dialog.focusIndex++
					dialog.textInputs[dialog.focusIndex].Focus()
					dialog.errorMessage = ""
					return dialog, textinput.Blink
				} else {
					// Move to buttons
					dialog.textInputs[dialog.focusIndex].Blur()
					dialog.onButtons = true
					dialog.selectedButton = 0
					dialog.errorMessage = ""
				}
			}

		case "shift+tab", "up":
			if dialog.onButtons {
				if dialog.selectedButton > 0 {
					// Cycle buttons backwards
					dialog.selectedButton--
				} else {
					// Go back to last field
					dialog.onButtons = false
					dialog.focusIndex = len(dialog.textInputs) - 1
					dialog.textInputs[dialog.focusIndex].Focus()
					return dialog, textinput.Blink
				}
			} else {
				// Move to previous field
				if dialog.focusIndex > 0 {
					dialog.textInputs[dialog.focusIndex].Blur()
					dialog.focusIndex--
					dialog.textInputs[dialog.focusIndex].Focus()
					dialog.errorMessage = ""
					return dialog, textinput.Blink
				}
			}

		case "enter":
			if dialog.onButtons {
				// Handle button press
				if dialog.selectedButton == 0 {
					// Cancel
					return dialog, func() tea.Msg { return base.CloseDialogMessage{} }
				} else {
					// Submit
					if err := dialog.validate(); err != nil {
						dialog.errorMessage = err.Error()
						// Return to form fields
						dialog.onButtons = false
						dialog.focusIndex = 0
						dialog.textInputs[dialog.focusIndex].Focus()
						return dialog, textinput.Blink
					}

					values := make(map[string]string)
					for i, field := range dialog.fields {
						values[field.Label] = dialog.textInputs[i].Value()
					}

					action := dialog.action
					if payload, ok := action.Payload.(map[string]any); ok {
						payload["values"] = values
						action.Payload = payload
					} else {
						action.Payload = map[string]any{"values": values}
					}

					return dialog, func() tea.Msg {
						return base.SmartConfirmationMessage{Action: action}
					}
				}
			} else {
				// Enter on field moves to next field or buttons
				if dialog.focusIndex < len(dialog.textInputs)-1 {
					dialog.textInputs[dialog.focusIndex].Blur()
					dialog.focusIndex++
					dialog.textInputs[dialog.focusIndex].Focus()
					dialog.errorMessage = ""
					return dialog, textinput.Blink
				} else {
					// Move to buttons
					dialog.textInputs[dialog.focusIndex].Blur()
					dialog.onButtons = true
					dialog.selectedButton = 1 // Default to Submit
					dialog.errorMessage = ""
				}
			}

		case "left", "h":
			if dialog.onButtons {
				dialog.selectedButton = (dialog.selectedButton - 1 + 2) % 2
			}

		case "right", "l":
			if dialog.onButtons {
				dialog.selectedButton = (dialog.selectedButton + 1) % 2
			}
		}
	}

	// Update the focused text input only if we're on form fields
	if !dialog.onButtons {
		var cmd tea.Cmd
		dialog.textInputs[dialog.focusIndex], cmd = dialog.textInputs[dialog.focusIndex].Update(msg)
		cmds = append(cmds, cmd)
	}

	return dialog, tea.Batch(cmds...)
}

// validate checks all fields and returns the first error found.
func (dialog FormDialog) validate() error {
	for i, field := range dialog.fields {
		value := dialog.textInputs[i].Value()

		// Check required fields
		if field.Required && strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", field.Label)
		}

		// Run custom validator if present and value is not empty
		if field.Validator != nil && strings.TrimSpace(value) != "" {
			if err := field.Validator(value); err != nil {
				return fmt.Errorf("%s: %w", field.Label, err)
			}
		}
	}
	return nil
}

// renderButtons renders the form dialog buttons
func (dialog FormDialog) renderButtons() string {
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

	cancelStyle := defaultButtonStyle
	submitStyle := defaultButtonStyle

	if dialog.onButtons {
		if dialog.selectedButton == 0 {
			cancelStyle = hoveredButtonStyle
		} else {
			submitStyle = hoveredButtonStyle
		}
	}

	cancelButton := cancelStyle.Render("Cancel")
	submitButton := submitStyle.Render("Submit")

	return lipgloss.JoinHorizontal(lipgloss.Center, cancelButton, submitButton)
}

func (dialog FormDialog) View() tea.View {
	// Get the content width for centering
	contentWidth := dialog.style.GetWidth()
	if contentWidth > 0 {
		frameSize := dialog.style.GetHorizontalFrameSize()
		contentWidth = contentWidth - frameSize
	}

	var b strings.Builder

	// Center the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Primary()).
		Width(contentWidth).
		Align(lipgloss.Center).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(dialog.title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Bold(true)

	for i, field := range dialog.fields {
		label := field.Label
		if field.Required {
			label += " *"
		}
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")

		inputStyle := lipgloss.NewStyle()
		if i == dialog.focusIndex && !dialog.onButtons {
			inputStyle = inputStyle.Foreground(colors.Primary())
		} else {
			inputStyle = inputStyle.Foreground(colors.Muted())
		}
		b.WriteString(inputStyle.Render(dialog.textInputs[i].View()))
		b.WriteString("\n\n")
	}

	if dialog.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(colors.Error()).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: " + dialog.errorMessage))
		b.WriteString("\n\n")
	}

	// Center the buttons
	buttonsView := dialog.renderButtons()
	buttonsStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
	b.WriteString(buttonsStyle.Render(buttonsView))

	return tea.NewView(dialog.style.Render(b.String()))
}

func (dialog FormDialog) String() string {
	// Get the content width for centering
	contentWidth := dialog.style.GetWidth()
	if contentWidth > 0 {
		frameSize := dialog.style.GetHorizontalFrameSize()
		contentWidth = contentWidth - frameSize
	}

	var b strings.Builder

	// Center the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Primary()).
		Width(contentWidth).
		Align(lipgloss.Center).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(dialog.title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Bold(true)

	for i, field := range dialog.fields {
		label := field.Label
		if field.Required {
			label += " *"
		}
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")

		inputStyle := lipgloss.NewStyle()
		if i == dialog.focusIndex && !dialog.onButtons {
			inputStyle = inputStyle.Foreground(colors.Primary())
		} else {
			inputStyle = inputStyle.Foreground(colors.Muted())
		}
		b.WriteString(inputStyle.Render(dialog.textInputs[i].View()))
		b.WriteString("\n\n")
	}

	if dialog.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(colors.Error()).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: " + dialog.errorMessage))
		b.WriteString("\n\n")
	}

	// Center the buttons
	buttonsView := dialog.renderButtons()
	buttonsStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
	b.WriteString(buttonsStyle.Render(buttonsView))

	return dialog.style.Render(b.String())
}
