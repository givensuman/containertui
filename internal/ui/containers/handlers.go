package containers

import (
	"github.com/givensuman/containertui/internal/context"
	"github.com/moby/moby/api/types/container"
)

func (m *Model) getSelectedContainerIDs() []string {
	selectedContainerIDs := make([]string, len(m.selectedContainers.selections))
	for id := range m.selectedContainers.selections {
		selectedContainerIDs = append(selectedContainerIDs, id)
	}

	return selectedContainerIDs
}

func (m *Model) getSelectedContainerIndices() []int {
	selectedContainerIndices := make([]int, len(m.selectedContainers.selections))
	for _, index := range m.selectedContainers.selections {
		selectedContainerIndices = append(selectedContainerIndices, index)
	}

	return selectedContainerIndices
}

func (m *Model) handlePauseContainers() {
	if len(m.selectedContainers.selections) > 0 {
		selectedContainerIDs := m.getSelectedContainerIDs()
		selectedContainerIndices := m.getSelectedContainerIndices()

		context.GetClient().PauseContainers(selectedContainerIDs)
		items := m.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StatePaused
			m.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := m.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().PauseContainer(selectedItem.ID)
			selectedItem.State = container.StatePaused
			m.list.SetItem(m.list.Index(), selectedItem)
		}
	}
}

func (m *Model) handleUnpauseContainers() {
	if len(m.selectedContainers.selections) > 0 {
		selectedContainerIDs := m.getSelectedContainerIDs()
		selectedContainerIndices := m.getSelectedContainerIndices()

		context.GetClient().UnpauseContainers(selectedContainerIDs)
		items := m.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StateRunning
			m.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := m.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().UnpauseContainer(selectedItem.ID)
			selectedItem.State = container.StateRunning
			m.list.SetItem(m.list.Index(), selectedItem)
		}
	}
}

func (m *Model) handleStartContainers() {
	if len(m.selectedContainers.selections) > 0 {
		selectedContainerIDs := m.getSelectedContainerIDs()
		selectedContainerIndices := m.getSelectedContainerIndices()

		context.GetClient().StartContainers(selectedContainerIDs)
		items := m.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StateRunning
			m.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := m.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().StartContainer(selectedItem.ID)
			selectedItem.State = container.StateRunning
			m.list.SetItem(m.list.Index(), selectedItem)
		}
	}
}

func (m *Model) handleStopContainers() {
	if len(m.selectedContainers.selections) > 0 {
		selectedContainerIDs := m.getSelectedContainerIDs()
		selectedContainerIndices := m.getSelectedContainerIndices()

		context.GetClient().StopContainers(selectedContainerIDs)
		items := m.list.Items()
		for _, index := range selectedContainerIndices {
			item := items[index].(ContainerItem)
			item.State = container.StateExited
			m.list.SetItem(index, item)
		}
	} else {
		selectedItem, ok := m.list.SelectedItem().(ContainerItem)
		if ok {
			context.GetClient().StopContainer(selectedItem.ID)
			selectedItem.State = container.StateExited
			m.list.SetItem(m.list.Index(), selectedItem)
		}
	}
}

func (m *Model) handleToggleSelection() {
	index := m.list.Index()
	selectedItem, ok := m.list.SelectedItem().(ContainerItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			m.selectedContainers.unselectContainerInList(selectedItem.ID)
		} else {
			m.selectedContainers.selectContainerInList(selectedItem.ID, index)
		}

		selectedItem.isSelected = !isSelected
		m.list.SetItem(index, selectedItem)
	}
}

func (m *Model) handleToggleSelectionOfAll() {
	allAlreadySelected := true
	items := m.list.Items()

	for _, item := range items {
		if c, ok := item.(ContainerItem); ok {
			if _, selected := m.selectedContainers.selections[c.ID]; !selected {
				allAlreadySelected = false
				break
			}
		}
	}

	if allAlreadySelected {
		// Unselect all items
		m.selectedContainers = newSelectedContainers()

		for index, item := range m.list.Items() {
			item, ok := item.(ContainerItem)
			if ok {
				item.isSelected = false
				m.list.SetItem(index, item)
			}
		}
	} else {
		// Select all items
		m.selectedContainers = newSelectedContainers()

		for index, item := range m.list.Items() {
			item, ok := item.(ContainerItem)
			if ok {
				item.isSelected = true
				m.list.SetItem(index, item)
				m.selectedContainers.selectContainerInList(item.ID, index)
			}
		}
	}
}
