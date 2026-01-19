package containers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/context"
)

// MessageCloseOverlay indicates the overlay should display its background.
type MessageCloseOverlay struct{}

func CloseOverlay() tea.Cmd {
	return func() tea.Msg {
		return MessageCloseOverlay{}
	}
}

// MessageOpenDeleteConfirmationDialog indicates the user
// has requested to delete an item in the ContainerList.
type MessageOpenDeleteConfirmationDialog struct {
	requestedContainersToDelete []*ContainerItem
}

// MessageConfirmDelete indicates the user confirmed
// they wish to delete an item in the ContainerList.
type MessageConfirmDelete struct{}

// MessageContainerOperationResult indicates the result of a container operation.
type MessageContainerOperationResult struct {
	Operation Operation
	IDs       []string
	Error     error
}

type Operation int

const (
	Pause Operation = iota
	Unpause
	Start
	Stop
	Restart
	Remove
)

// PerformContainerOperation performs the specified operation on the given container IDs asynchronously.
func PerformContainerOperation(operation Operation, containerIDs []string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch operation {
		case Pause:
			err = context.GetClient().PauseContainers(containerIDs)
		case Unpause:
			err = context.GetClient().UnpauseContainers(containerIDs)
		case Start:
			err = context.GetClient().StartContainers(containerIDs)
		case Stop:
			err = context.GetClient().StopContainers(containerIDs)
		case Restart:
			err = context.GetClient().RestartContainers(containerIDs)
		case Remove:
			err = context.GetClient().RemoveContainers(containerIDs)
		}
		return MessageContainerOperationResult{Operation: operation, IDs: containerIDs, Error: err}
	}
}
