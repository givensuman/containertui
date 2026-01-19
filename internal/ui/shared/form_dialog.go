package shared

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
)

// FormField represents a single field in a form.
type FormField struct {
	Label       string
	Placeholder string
	Value       string
	Validator   func(string) error // nil = no validation, returns error if invalid
	Required    bool
}

// FormDialog is a modal dialog with multiple text input fields.
type FormDialog struct {
	Component
	title          string
	fields         []FormField
	textInputs     []textinput.Model
	focusIndex     int
	submitLabel    string
	actionType     string // Type for ConfirmationMessage
	actionMetadata map[string]interface{}
	style          lipgloss.Style
	width          int
	height         int
	errorMessage   string
}

// FormDialog implements StringViewModel and ComponentModel
// var _ StringViewModel = (*FormDialog)(nil)
// var _ ComponentModel = (*FormDialog)(nil)

// NewFormDialog creates a new form dialog with the specified fields.
func NewFormDialog(title string, fields []FormField, actionType string, metadata map[string]interface{}) FormDialog {
	width, height := context.GetWindowSize()

	style := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary()).
		Align(lipgloss.Left)

	layoutManager := NewLayoutManager(width, height)
	modalDimensions := layoutManager.CalculateModal(style)
	style = style.Width(modalDimensions.Width).Height(modalDimensions.Height)

	// Initialize text inputs
	textInputs := make([]textinput.Model, len(fields))
	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.SetValue(field.Value)
		ti.CharLimit = 256
		ti.SetWidth(modalDimensions.Width - 6) // Account for padding and borders

		if i == 0 {
			ti.Focus()
		}

		textInputs[i] = ti
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return FormDialog{
		title:          title,
		fields:         fields,
		textInputs:     textInputs,
		focusIndex:     0,
		submitLabel:    "Submit",
		actionType:     actionType,
		actionMetadata: metadata,
		style:          style,
		width:          width,
		height:         height,
		errorMessage:   "",
	}
}

func (dialog *FormDialog) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dialog.width = msg.Width
	dialog.height = msg.Height

	layoutManager := NewLayoutManager(msg.Width, msg.Height)
	modalDimensions := layoutManager.CalculateModal(dialog.style)
	dialog.style = dialog.style.Width(modalDimensions.Width).Height(modalDimensions.Height)

	// Update text input widths
	for i := range dialog.textInputs {
		dialog.textInputs[i].SetWidth(modalDimensions.Width - 6)
	}
}

func (dialog FormDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (dialog FormDialog) Update(msg tea.Msg) (FormDialog, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dialog.UpdateWindowDimensions(msg)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return dialog, func() tea.Msg { return CloseDialogMessage{} }

		case "tab", "down":
			// Move to next field
			dialog.textInputs[dialog.focusIndex].Blur()
			dialog.focusIndex = (dialog.focusIndex + 1) % len(dialog.textInputs)
			dialog.textInputs[dialog.focusIndex].Focus()
			dialog.errorMessage = "" // Clear error when changing fields
			return dialog, textinput.Blink

		case "shift+tab", "up":
			// Move to previous field
			dialog.textInputs[dialog.focusIndex].Blur()
			dialog.focusIndex = (dialog.focusIndex - 1 + len(dialog.textInputs)) % len(dialog.textInputs)
			dialog.textInputs[dialog.focusIndex].Focus()
			dialog.errorMessage = "" // Clear error when changing fields
			return dialog, textinput.Blink

		case "enter":
			// Validate and submit
			if err := dialog.validate(); err != nil {
				dialog.errorMessage = err.Error()
				return dialog, nil
			}

			// Collect field values
			values := make(map[string]string)
			for i, field := range dialog.fields {
				values[field.Label] = dialog.textInputs[i].Value()
			}

			// Add values to metadata
			dialog.actionMetadata["values"] = values

			return dialog, func() tea.Msg {
				return ConfirmationMessage{
					Action: SmartDialogAction{
						Type:    dialog.actionType,
						Payload: dialog.actionMetadata,
					},
				}
			}
		}
	}

	// Update the focused text input
	var cmd tea.Cmd
	dialog.textInputs[dialog.focusIndex], cmd = dialog.textInputs[dialog.focusIndex].Update(msg)
	cmds = append(cmds, cmd)

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
				return fmt.Errorf("%s: %v", field.Label, err)
			}
		}
	}
	return nil
}

func (dialog FormDialog) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Primary()).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(dialog.title))
	b.WriteString("\n\n")

	// Fields
	labelStyle := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Bold(true)

	for i, field := range dialog.fields {
		// Label
		label := field.Label
		if field.Required {
			label += " *"
		}
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")

		// Text input
		inputStyle := lipgloss.NewStyle()
		if i == dialog.focusIndex {
			inputStyle = inputStyle.Foreground(colors.Primary())
		} else {
			inputStyle = inputStyle.Foreground(colors.Muted())
		}
		b.WriteString(inputStyle.Render(dialog.textInputs[i].View()))
		b.WriteString("\n\n")
	}

	// Error message
	if dialog.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(colors.Error()).
			Bold(true)
		b.WriteString(errorStyle.Render("Error: " + dialog.errorMessage))
		b.WriteString("\n\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(colors.Muted()).
		Italic(true)
	b.WriteString(helpStyle.Render("Tab: next field • Enter: submit • Esc: cancel"))

	return dialog.style.Render(b.String())
}
