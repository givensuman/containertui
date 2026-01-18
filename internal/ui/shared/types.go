package shared

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Component struct {
	WindowWidth  int
	WindowHeight int
}

type ComponentModel interface {
	tea.Model
	UpdateWindowDimensions(msg tea.WindowSizeMsg)
}

// Dialog-related messages

// SmartDialogAction defines the action to take upon confirmation
type SmartDialogAction struct {
	Type    string // "DeleteContainer", "DeleteImage", "NavigateToContainer"
	Payload any    // Data required for the action (e.g. IDs)
}

// ConfirmationMessage is sent when an action is confirmed
type ConfirmationMessage struct {
	Action SmartDialogAction
}

// CloseDialogMessage is sent when the dialog is cancelled
type CloseDialogMessage struct{}
