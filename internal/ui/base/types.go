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

// MsgContainerCreated is sent when a container is created from another tab
// to trigger a refresh of the Containers tab
type MsgContainerCreated struct {
	ContainerID string
}

// MsgImagePulled is sent when an image is pulled from another tab
// to trigger a refresh of the Images tab
type MsgImagePulled struct {
	ImageName string
}

// ResourceType represents the type of Docker resource
type ResourceType string

const (
	ResourceContainer ResourceType = "container"
	ResourceImage     ResourceType = "image"
	ResourceVolume    ResourceType = "volume"
	ResourceNetwork   ResourceType = "network"
)

// OperationType represents the type of operation performed on a resource
type OperationType string

const (
	OperationCreated OperationType = "created"
	OperationDeleted OperationType = "deleted"
	OperationUpdated OperationType = "updated"
	OperationPruned  OperationType = "pruned"
)

// MsgResourceChanged is a universal message for resource state changes
// that require cross-tab UI updates. This message is sent when any operation
// modifies resources (create, delete, update, prune) so that all relevant tabs
// can refresh their views to reflect the current state.
type MsgResourceChanged struct {
	Resource  ResourceType   // Type of resource that changed
	Operation OperationType  // Type of operation performed
	IDs       []string       // IDs of affected resources
	Metadata  map[string]any // Optional operation-specific data
}
