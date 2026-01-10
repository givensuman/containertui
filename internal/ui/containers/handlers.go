package containers

import (
	"github.com/givensuman/containertui/internal/context"
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

func (cl *ContainerList) handlePauseContainers() {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		selectedContainerIndices := cl.getSelectedContainerIndices()

		context.GetClient().PauseContainers(selectedContainerIDs)
		items := cl.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StatePaused
			cl.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().PauseContainer(selectedItem.ID)
			selectedItem.State = container.StatePaused
			cl.list.SetItem(cl.list.Index(), selectedItem)
		}
	}
}

func (cl *ContainerList) handleUnpauseContainers() {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		selectedContainerIndices := cl.getSelectedContainerIndices()

		context.GetClient().UnpauseContainers(selectedContainerIDs)
		items := cl.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StateRunning
			cl.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().UnpauseContainer(selectedItem.ID)
			selectedItem.State = container.StateRunning
			cl.list.SetItem(cl.list.Index(), selectedItem)
		}
	}
}

func (cl *ContainerList) handleStartContainers() {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		selectedContainerIndices := cl.getSelectedContainerIndices()

		context.GetClient().StartContainers(selectedContainerIDs)
		items := cl.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StateRunning
			cl.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().StartContainer(selectedItem.ID)
			selectedItem.State = container.StateRunning
			cl.list.SetItem(cl.list.Index(), selectedItem)
		}
	}
}

func (cl *ContainerList) handleStopContainers() {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		selectedContainerIndices := cl.getSelectedContainerIndices()

		context.GetClient().StopContainers(selectedContainerIDs)
		items := cl.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StateExited
			cl.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().StopContainer(selectedItem.ID)
			selectedItem.State = container.StateExited
			cl.list.SetItem(cl.list.Index(), selectedItem)
		}
	}
}

func (cl *ContainerList) handleConfirmationOfRemoveContainers() {
	if len(cl.selectedContainers.selections) > 0 {
		selectedContainerIDs := cl.getSelectedContainerIDs()
		selectedContainerIndices := cl.getSelectedContainerIndices()

		context.GetClient().RemoveContainers(selectedContainerIDs)
		for _, index := range selectedContainerIndices {
			cl.list.RemoveItem(index)
		}
	} else {
		item, ok := cl.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().RemoveContainer(item.ID)
			cl.list.RemoveItem(cl.list.Index())
		}
	}
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
