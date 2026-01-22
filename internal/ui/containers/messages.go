package containers

import (
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/context"
)

// MessageContainerOperationResult indicates the result of a container operation.
type MessageContainerOperationResult struct {
	Operation Operation
	ID        string
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

// PerformContainerOperation performs the specified operation on a single container asynchronously.
func PerformContainerOperation(operation Operation, containerID string) tea.Cmd {
	return func() tea.Msg {
		var err error
		client := context.GetClient()
		switch operation {
		case Pause:
			err = client.PauseContainer(containerID)
		case Unpause:
			err = client.UnpauseContainer(containerID)
		case Start:
			err = client.StartContainer(containerID)
		case Stop:
			err = client.StopContainer(containerID)
		case Restart:
			err = client.RestartContainer(containerID)
		case Remove:
			err = client.RemoveContainer(containerID)
		}
		return MessageContainerOperationResult{Operation: operation, ID: containerID, Error: err}
	}
}

// PerformContainerOperations performs the specified operation on multiple containers asynchronously.
// Returns a batch of commands, one for each container.
func PerformContainerOperations(operation Operation, containerIDs []string) tea.Cmd {
	var cmds []tea.Cmd
	for _, id := range containerIDs {
		cmds = append(cmds, PerformContainerOperation(operation, id))
	}
	return tea.Batch(cmds...)
}
