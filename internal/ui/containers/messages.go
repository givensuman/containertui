package containers

import (
	tea "charm.land/bubbletea/v2"
	stdcontext "context"
	"github.com/givensuman/containertui/internal/state"
)

// MsgContainerOperationResult indicates the result of a container operation.
type MsgContainerOperationResult struct {
	Operation Operation
	ID        string
	Error     error
}

// MsgContainersRefreshed carries a full refreshed container item list.
type MsgContainersRefreshed struct {
	Items []ContainerItem
	Err   error
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
func PerformContainerOperation(operation Operation, containerID string, force bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		client := state.GetBackend()
		switch operation {
		case Pause:
			err = client.PauseContainer(stdcontext.Background(), containerID)
		case Unpause:
			err = client.UnpauseContainer(stdcontext.Background(), containerID)
		case Start:
			err = client.StartContainer(stdcontext.Background(), containerID)
		case Stop:
			err = client.StopContainer(stdcontext.Background(), containerID)
		case Restart:
			err = client.RestartContainer(stdcontext.Background(), containerID)
		case Remove:
			err = client.RemoveContainer(stdcontext.Background(), containerID, force)
		}
		return MsgContainerOperationResult{Operation: operation, ID: containerID, Error: err}
	}
}

// PerformContainerOperations performs the specified operation on multiple containers asynchronously.
// Returns a batch of commands, one for each container.
func PerformContainerOperations(operation Operation, containerIDs []string, force bool) tea.Cmd {
	var cmds []tea.Cmd
	for _, id := range containerIDs {
		cmds = append(cmds, PerformContainerOperation(operation, id, force))
	}
	return tea.Batch(cmds...)
}
