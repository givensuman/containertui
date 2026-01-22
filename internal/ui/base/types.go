package base

import (
	tea "charm.land/bubbletea/v2"
)

type Component struct {
	WindowWidth  int
	WindowHeight int
}

type ComponentModel interface {
	tea.Model
	UpdateWindowDimensions(msg tea.WindowSizeMsg)
}

type StringViewModel interface {
	View() string
}

type Action int

const (
	CloseDialog Action = iota
	DeleteContainer
	DeleteImage
	NavigateToContainer
)

// DialogAction defines the action to take upon confirmation
type DialogAction[T any] struct {
	Type    Action
	Payload T // Data required for the action
}

func NewDialogAction[T any](actionType Action, payload T) DialogAction[T] {
	return DialogAction[T]{
		Type:    actionType,
		Payload: payload,
	}
}

// SmartDialogAction uses string-based action types for flexible dialog handling
type SmartDialogAction struct {
	Type    string
	Payload any
}

// ConfirmationMessage is sent when an action is confirmed
type ConfirmationMessage[T any] struct {
	Action DialogAction[T]
}

// SmartConfirmationMessage is sent when a SmartDialog action is confirmed
type SmartConfirmationMessage struct {
	Action SmartDialogAction
}

// CloseDialogMessage is sent when the dialog is cancelled
type CloseDialogMessage struct{}

// MsgFocusChanged is sent when focus switches between list and detail panes
type MsgFocusChanged struct {
	IsDetailsFocused bool
}
