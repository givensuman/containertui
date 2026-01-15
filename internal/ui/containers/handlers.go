package containers

import (
	"os/exec"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/moby/moby/api/types/container"
)

func (cl *ContainerList) getSelectedContainerIDs() []string {
	selectedContainerIDs := make([]string, 0, len(cl.selectedContainers.selections))
	for id := range cl.selectedContainers.selections {
		selectedContainerIDs = append(selectedContainerIDs, id)
	}

	return selectedContainerIDs
}

func (cl *ContainerList) getSelectedContainerIndices() []int {
	selectedContainerIndices := make([]int, 0, len(cl.selectedContainers.selections))
	for _, index := range cl.selectedContainers.selections {
		selectedContainerIndices = append(selectedContainerIndices, index)
	}

	return selectedContainerIndices
}

func (cl *ContainerList) setWorkingState(ids []string, working bool) {
	items := cl.list.Items()
	for i, item := range items {
		if c, ok := item.(ContainerItem); ok && slices.Contains(ids, c.ID) {
			c.isWorking = working
			if working {
				c.spinner = newSpinner()
			}
			cl.list.SetItem(i, c)
		}
	}
}

func (cl *ContainerList) anySelectedWorking() bool {
	for id := range cl.selectedContainers.selections {
		if item := cl.findItemByID(id); item != nil && item.isWorking {
			return true
		}
	}
	return false
}

func (cl *ContainerList) findItemByID(id string) *ContainerItem {
	items := cl.list.Items()
	for _, item := range items {
		if c, ok := item.(ContainerItem); ok && c.ID == id {
			return &c
		}
	}
	return nil
}

func (cl *ContainerList) handlePauseContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		if cl.anySelectedWorking() {
			return nil
		}
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Pause, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Pause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (cl *ContainerList) handleUnpauseContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		if cl.anySelectedWorking() {
			return nil
		}
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Unpause, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Unpause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (cl *ContainerList) handleStartContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		if cl.anySelectedWorking() {
			return nil
		}
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Start, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Start, []string{selectedItem.ID})
		}
	}

	return nil
}

func (cl *ContainerList) handleStopContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		if cl.anySelectedWorking() {
			return nil
		}
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Stop, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Stop, []string{selectedItem.ID})
		}
	}

	return nil
}

func (cl *ContainerList) handleRemoveContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		if cl.anySelectedWorking() {
			return nil
		}
		selectedContainerIndices := cl.getSelectedContainerIndices()

		var requestedContainersToDelete []*ContainerItem
		items := cl.list.Items()

		for _, index := range selectedContainerIndices {
			requestedContainer := items[index].(ContainerItem)
			requestedContainersToDelete = append(requestedContainersToDelete, &requestedContainer)
		}

		return func() tea.Msg {
			return MessageOpenDeleteConfirmationDialog{requestedContainersToDelete}
		}
	} else {
		item, ok := cl.list.SelectedItem().(ContainerItem)
		if ok && !item.isWorking {
			return func() tea.Msg {
				return MessageOpenDeleteConfirmationDialog{[]*ContainerItem{&item}}
			}
		}
	}

	return nil
}

func (cl *ContainerList) handleShowLogs() tea.Cmd {
	item, ok := cl.list.SelectedItem().(ContainerItem)
	if !ok || item.isWorking {
		return nil
	}

	if item.State != container.StateRunning {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	c := exec.Command("sh", "-c", "docker logs \"$0\" 2>&1 | less", item.ID)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return notifications.AddNotificationMsg{
				Message:  err.Error(),
				Level:    notifications.Error,
				Duration: 10 * 1000 * 1000 * 1000, // 10s
			}
		}
		return nil
	})
}

func (cl *ContainerList) handleExecShell() tea.Cmd {
	item, ok := cl.list.SelectedItem().(ContainerItem)
	if !ok || item.isWorking {
		return nil
	}

	if item.State != container.StateRunning {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	// We'll use tea.ExecProcess to run `docker exec -it <id> /bin/sh`
	// This suspends the Bubbletea UI and lets the subprocess take over TTY
	// Note: We are using "sh" as a generic shell, but some containers might only have "bash" or "ash".
	// Ideally we could probe or let user choose, but "sh" is safest default.
	c := exec.Command("sh", "-c", "exec docker exec -it \"$0\" /bin/sh", item.ID)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			// tea.ExecProcess callback returns a Msg, not a Cmd.
			// So we need to construct the Msg manually or change how notifications work.
			// But notifications.ShowError returns a Cmd.
			// Let's just create the message directly.
			return notifications.AddNotificationMsg{
				Message:  err.Error(),
				Level:    notifications.Error,
				Duration: 10 * 1000 * 1000 * 1000, // 10s
			}
		}
		// Refresh container state after coming back, just in case
		// Note: We might want a specific message type for this
		return nil
	})
}

func (cl *ContainerList) handleConfirmationOfRemoveContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Remove, selectedContainerIDs)
	} else {
		item, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			cl.setWorkingState([]string{item.ID}, true)
			return PerformContainerOperation(Remove, []string{item.ID})
		}
	}

	return nil
}

func (cl *ContainerList) handleToggleSelection() {
	index := cl.list.Index()
	selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
	if ok && !selectedItem.isWorking {
		isSelected := selectedItem.isSelected

		if isSelected {
			cl.selectedContainers.unselectContainerInList(selectedItem.ID)
		} else {
			cl.selectedContainers.selectContainerInList(selectedItem.ID, index)
		}

		selectedItem.isSelected = !isSelected
		cl.list.SetItem(index, selectedItem)
	}
}

func (cl *ContainerList) handleToggleSelectionOfAll() {
	allNonWorkingAlreadySelected := true
	items := cl.list.Items()

	for _, item := range items {
		if c, ok := item.(ContainerItem); ok && !c.isWorking {
			if _, selected := cl.selectedContainers.selections[c.ID]; !selected {
				allNonWorkingAlreadySelected = false
				break
			}
		}
	}

	if allNonWorkingAlreadySelected {
		// Unselect all items
		cl.selectedContainers = newSelectedContainers()

		for index, item := range cl.list.Items() {
			item, ok := item.(ContainerItem)
			if ok {
				item.isSelected = false
				cl.list.SetItem(index, item)
			}
		}
	} else {
		// Select all non-working items
		cl.selectedContainers = newSelectedContainers()

		for index, item := range cl.list.Items() {
			item, ok := item.(ContainerItem)
			if ok && !item.isWorking {
				item.isSelected = true
				cl.list.SetItem(index, item)
				cl.selectedContainers.selectContainerInList(item.ID, index)
			}
		}
	}
}

func (cl *ContainerList) handleContainerOperationResult(msg MessageContainerOperationResult) tea.Cmd {
	cl.setWorkingState(msg.IDs, false)

	if msg.Error != nil {
		return notifications.ShowError(msg.Error)
	}

	if msg.Operation == Remove {
		items := cl.list.Items()

		var indicesToRemove []int
		for i, item := range items {
			if c, ok := item.(ContainerItem); ok {
				for _, id := range msg.IDs {
					if c.ID == id {
						indicesToRemove = append([]int{i}, indicesToRemove...)
						break
					}
				}
			}
		}
		for _, index := range indicesToRemove {
			cl.list.RemoveItem(index)
		}

		return notifications.ShowSuccess("Container(s) removed successfully")
	}

	var newState container.ContainerState
	switch msg.Operation {
	case Pause:
		newState = container.StatePaused
	case Unpause, Start:
		newState = container.StateRunning
	case Stop:
		newState = container.StateExited
	default:
		return nil
	}

	items := cl.list.Items()
	for i, item := range items {
		if c, ok := item.(ContainerItem); ok {
			for _, id := range msg.IDs {
				if c.ID == id {
					c.State = newState
					cl.list.SetItem(i, c)
					break
				}
			}
		}
	}
	return nil
}
