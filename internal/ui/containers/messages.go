package containers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/context"
)

// MessageCloseOverlay indicates the overlay should display
// its background
type MessageCloseOverlay struct{}

func CloseOverlay() tea.Cmd {
	return func() tea.Msg {
		return MessageCloseOverlay{}
	}
}

// MessageOpenContainerLogs indicates the user
// has requested to view the logs of the selected container
type MessageOpenContainerLogs struct {
	container *ContainerItem
}

func OpenContainerLogs(container *ContainerItem) tea.Cmd {
	return func() tea.Msg {
		return MessageOpenContainerLogs{container}
	}
}

// MessageOpenDeleteConfirmationDialog indicates the user
// has requested to delete an item in the ContainerList
type MessageOpenDeleteConfirmationDialog struct {
	requestedContainersToDelete []*ContainerItem
}

// MessageConfirmDelete indicates the user confirmed
// they wish to delete an item in the ContainerList
type MessageConfirmDelete struct{}

// MessageContainerOperationResult indicates the result of a container operation
type MessageContainerOperationResult struct {
	Operation string // "pause", "unpause", "start", "stop"
	IDs       []string
	Error     error
}

type Operation int // TODO: Better implementation
const (
	Pause Operation = iota
	Unpause
	Start
	Stop
	Remove
)

// PerformContainerOperation performs the specified operation on the given container IDs asynchronously
func PerformContainerOperation(operation string, ids []string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch operation {
		case "pause":
			err = context.GetClient().PauseContainers(ids)
		case "unpause":
			err = context.GetClient().UnpauseContainers(ids)
		case "start":
			err = context.GetClient().StartContainers(ids)
		case "stop":
			err = context.GetClient().StopContainers(ids)
		case "remove":
			err = context.GetClient().RemoveContainers(ids)
		}
		return MessageContainerOperationResult{Operation: operation, IDs: ids, Error: err}
	}
}
