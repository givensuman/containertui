package containers

import (
	"os/exec"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

func (containerList *ContainerList) getSelectedContainerIDs() []string {
	selectedContainerIDs := make([]string, 0, len(containerList.selectedContainers.selections))
	for containerID := range containerList.selectedContainers.selections {
		selectedContainerIDs = append(selectedContainerIDs, containerID)
	}

	return selectedContainerIDs
}

func (containerList *ContainerList) getSelectedContainerIndices() []int {
	selectedContainerIndices := make([]int, 0, len(containerList.selectedContainers.selections))
	for _, index := range containerList.selectedContainers.selections {
		selectedContainerIndices = append(selectedContainerIndices, index)
	}

	return selectedContainerIndices
}

func (containerList *ContainerList) setWorkingState(containerIDs []string, working bool) {
	items := containerList.list.Items()
	for index, item := range items {
		if container, ok := item.(ContainerItem); ok && slices.Contains(containerIDs, container.ID) {
			container.isWorking = working
			if working {
				container.spinner = newSpinner()
			}
			containerList.list.SetItem(index, container)
		}
	}
}

func (containerList *ContainerList) anySelectedWorking() bool {
	for containerID := range containerList.selectedContainers.selections {
		if item := containerList.findItemByID(containerID); item != nil && item.isWorking {
			return true
		}
	}
	return false
}

func (containerList *ContainerList) findItemByID(containerID string) *ContainerItem {
	items := containerList.list.Items()
	for _, item := range items {
		if container, ok := item.(ContainerItem); ok && container.ID == containerID {
			return &container
		}
	}
	return nil
}

func (containerList *ContainerList) handlePauseContainers() tea.Cmd {
	if len(containerList.selectedContainers.selections) > 0 {
		selectedContainerIDs := containerList.getSelectedContainerIDs()
		if containerList.anySelectedWorking() {
			return nil
		}
		containerList.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Pause, selectedContainerIDs)
	} else {
		selectedItem, ok := containerList.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			containerList.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Pause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (containerList *ContainerList) handleUnpauseContainers() tea.Cmd {
	if len(containerList.selectedContainers.selections) > 0 {
		selectedContainerIDs := containerList.getSelectedContainerIDs()
		if containerList.anySelectedWorking() {
			return nil
		}
		containerList.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Unpause, selectedContainerIDs)
	} else {
		selectedItem, ok := containerList.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			containerList.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Unpause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (containerList *ContainerList) handleStartContainers() tea.Cmd {
	if len(containerList.selectedContainers.selections) > 0 {
		selectedContainerIDs := containerList.getSelectedContainerIDs()
		if containerList.anySelectedWorking() {
			return nil
		}
		containerList.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Start, selectedContainerIDs)
	} else {
		selectedItem, ok := containerList.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			containerList.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Start, []string{selectedItem.ID})
		}
	}

	return nil
}

func (containerList *ContainerList) handleStopContainers() tea.Cmd {
	if len(containerList.selectedContainers.selections) > 0 {
		selectedContainerIDs := containerList.getSelectedContainerIDs()
		if containerList.anySelectedWorking() {
			return nil
		}
		containerList.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Stop, selectedContainerIDs)
	} else {
		selectedItem, ok := containerList.list.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			containerList.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Stop, []string{selectedItem.ID})
		}
	}

	return nil
}

func (containerList *ContainerList) handleRemoveContainers() tea.Cmd {
	if len(containerList.selectedContainers.selections) > 0 {
		if containerList.anySelectedWorking() {
			return nil
		}
		selectedContainerIndices := containerList.getSelectedContainerIndices()

		var requestedContainersToDelete []*ContainerItem
		items := containerList.list.Items()

		for _, index := range selectedContainerIndices {
			requestedContainer := items[index].(ContainerItem)
			requestedContainersToDelete = append(requestedContainersToDelete, &requestedContainer)
		}

		return func() tea.Msg {
			return MessageOpenDeleteConfirmationDialog{requestedContainersToDelete}
		}
	} else {
		item, ok := containerList.list.SelectedItem().(ContainerItem)
		if ok && !item.isWorking {
			return func() tea.Msg {
				return MessageOpenDeleteConfirmationDialog{[]*ContainerItem{&item}}
			}
		}
	}

	return nil
}

func (containerList *ContainerList) handleShowLogs() tea.Cmd {
	item, ok := containerList.list.SelectedItem().(ContainerItem)
	if !ok || item.isWorking {
		return nil
	}

	if item.State != "running" {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	command := exec.Command("sh", "-c", "docker logs \"$0\" 2>&1 | less", item.ID)
	return tea.ExecProcess(command, func(err error) tea.Msg {
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

func (containerList *ContainerList) handleExecShell() tea.Cmd {
	item, ok := containerList.list.SelectedItem().(ContainerItem)
	if !ok || item.isWorking {
		return nil
	}

	if item.State != "running" {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	// We'll use tea.ExecProcess to run `docker exec -it <id> /bin/sh`.
	// This suspends the Bubbletea UI and lets the subprocess take over TTY.
	// Note: We are using "sh" as a generic shell, but some containers might only have "bash" or "ash".
	// Ideally we could probe or let user choose, but "sh" is safest default.
	command := exec.Command("sh", "-c", "exec docker exec -it \"$0\" /bin/sh", item.ID)
	return tea.ExecProcess(command, func(err error) tea.Msg {
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
		// Refresh container state after coming back, just in case.
		// Note: We might want a specific message type for this.
		return nil
	})
}

func (containerList *ContainerList) handleConfirmationOfRemoveContainers() tea.Cmd {
	if len(containerList.selectedContainers.selections) > 0 {
		selectedContainerIDs := containerList.getSelectedContainerIDs()
		containerList.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Remove, selectedContainerIDs)
	} else {
		item, ok := containerList.list.SelectedItem().(ContainerItem)
		if ok {
			containerList.setWorkingState([]string{item.ID}, true)
			return PerformContainerOperation(Remove, []string{item.ID})
		}
	}

	return nil
}

func (containerList *ContainerList) handleToggleSelection() {
	index := containerList.list.Index()
	selectedItem, ok := containerList.list.SelectedItem().(ContainerItem)
	if ok && !selectedItem.isWorking {
		isSelected := selectedItem.isSelected

		if isSelected {
			containerList.selectedContainers.unselectContainerInList(selectedItem.ID)
		} else {
			containerList.selectedContainers.selectContainerInList(selectedItem.ID, index)
		}

		selectedItem.isSelected = !isSelected
		containerList.list.SetItem(index, selectedItem)
	}
}

func (containerList *ContainerList) handleToggleSelectionOfAll() {
	allNonWorkingAlreadySelected := true
	items := containerList.list.Items()

	for _, item := range items {
		if container, ok := item.(ContainerItem); ok && !container.isWorking {
			if _, selected := containerList.selectedContainers.selections[container.ID]; !selected {
				allNonWorkingAlreadySelected = false
				break
			}
		}
	}

	if allNonWorkingAlreadySelected {
		// Unselect all items.
		containerList.selectedContainers = newSelectedContainers()

		for index, item := range containerList.list.Items() {
			container, ok := item.(ContainerItem)
			if ok {
				container.isSelected = false
				containerList.list.SetItem(index, container)
			}
		}
	} else {
		// Select all non-working items.
		containerList.selectedContainers = newSelectedContainers()

		for index, item := range containerList.list.Items() {
			container, ok := item.(ContainerItem)
			if ok && !container.isWorking {
				container.isSelected = true
				containerList.list.SetItem(index, container)
				containerList.selectedContainers.selectContainerInList(container.ID, index)
			}
		}
	}
}

func (containerList *ContainerList) handleContainerOperationResult(msg MessageContainerOperationResult) tea.Cmd {
	containerList.setWorkingState(msg.IDs, false)

	if msg.Error != nil {
		return notifications.ShowError(msg.Error)
	}

	if msg.Operation == Remove {
		items := containerList.list.Items()

		var indicesToRemove []int
		for index, item := range items {
			if container, ok := item.(ContainerItem); ok {
				for _, containerID := range msg.IDs {
					if container.ID == containerID {
						indicesToRemove = append([]int{index}, indicesToRemove...)
						break
					}
				}
			}
		}
		for _, index := range indicesToRemove {
			containerList.list.RemoveItem(index)
		}

		return notifications.ShowSuccess("Container(s) removed successfully")
	}

	var newState string
	switch msg.Operation {
	case Pause:
		newState = "paused"
	case Unpause, Start:
		newState = "running"
	case Stop:
		newState = "exited"
	default:
		return nil
	}

	items := containerList.list.Items()
	for index, item := range items {
		if container, ok := item.(ContainerItem); ok {
			for _, containerID := range msg.IDs {
				if container.ID == containerID {
					container.State = newState
					containerList.list.SetItem(index, container)
					break
				}
			}
		}
	}
	return nil
}
