package containers

import (
	"log"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
		if c, ok := item.(ContainerItem); ok {
			for _, id := range ids {
				if c.ID == id {
					c.isWorking = working
					if working {
						c.spinner = spinner.New()
					}
					cl.list.SetItem(i, c)
					break
				}
			}
		}
	}
}

func (cl *ContainerList) handlePauseContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Pause, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Pause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (cl *ContainerList) handleUnpauseContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Unpause, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Unpause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (cl *ContainerList) handleStartContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Start, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Start, []string{selectedItem.ID})
		}
	}

	return nil
}

func (cl *ContainerList) handleStopContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		cl.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Stop, selectedContainerIDs)
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			cl.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Stop, []string{selectedItem.ID})
		}
	}

	return nil
}

func (cl *ContainerList) handleRemoveContainers() tea.Cmd {
	if len(cl.selectedContainers.selections) > 0 {
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
		if ok {
			return func() tea.Msg {
				return MessageOpenDeleteConfirmationDialog{[]*ContainerItem{&item}}
			}
		}
	}

	return nil
}

func (cl *ContainerList) handleShowLogs() tea.Cmd {
	item, ok := cl.list.SelectedItem().(ContainerItem)
	if !ok {
		log.Print()
		return nil
	}

	// TODO: Replace with notification
	if item.State != container.StateRunning {
		log.Printf("%s is not running...", item.Name)
		return nil
	}

	return OpenContainerLogs(&item)
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
	if ok {
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
	allAlreadySelected := true
	items := cl.list.Items()

	for _, item := range items {
		if c, ok := item.(ContainerItem); ok {
			if _, selected := cl.selectedContainers.selections[c.ID]; !selected {
				allAlreadySelected = false
				break
			}
		}
	}

	if allAlreadySelected {
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
		// Select all items
		cl.selectedContainers = newSelectedContainers()

		for index, item := range cl.list.Items() {
			item, ok := item.(ContainerItem)
			if ok {
				item.isSelected = true
				cl.list.SetItem(index, item)
				cl.selectedContainers.selectContainerInList(item.ID, index)
			}
		}
	}
}

func (cl *ContainerList) handleContainerOperationResult(msg MessageContainerOperationResult) {
	// Always set working state to false, even on error
	cl.setWorkingState(msg.IDs, false)

	if msg.Error != nil {
		// TODO: Replace with notification
		return
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

		return
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
		return
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
}
