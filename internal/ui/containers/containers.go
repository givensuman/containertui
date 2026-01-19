// Package containers defines the containers component.
package containers

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/shared"
	"github.com/guptarohit/asciigraph"
)

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

// MsgStatsTick is sent to trigger a stats refresh.
type MsgStatsTick time.Time

// MsgContainerStats contains the stats for a container.
type MsgContainerStats struct {
	ID    string
	Stats client.ContainerStats
	Err   error
}

// MsgContainerInspection contains the inspection data for a container.
type MsgContainerInspection struct {
	ID        string
	Container types.ContainerJSON
	Err       error
}

type detailsKeybindings struct {
	Up     key.Binding
	Down   key.Binding
	Switch key.Binding
}

func newDetailsKeybindings() detailsKeybindings {
	return detailsKeybindings{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Switch: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
	}
}

type keybindings struct {
	pauseContainer       key.Binding
	unpauseContainer     key.Binding
	startContainer       key.Binding
	stopContainer        key.Binding
	restartContainer     key.Binding
	removeContainer      key.Binding
	showLogs             key.Binding
	execShell            key.Binding
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	switchTab            key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		pauseContainer: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause container"),
		),
		unpauseContainer: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "unpause container"),
		),
		startContainer: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start container"),
		),
		stopContainer: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stop container"),
		),
		restartContainer: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "restart container"),
		),
		removeContainer: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove container"),
		),
		showLogs: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "show container logs"),
		),
		execShell: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "exec shell"),
		),
		toggleSelection: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle selection"),
		),
		toggleSelectionOfAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle selection of all"),
		),
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "tab", "shift+tab"),
			key.WithHelp("1-4/tab", "switch tab"),
		),
	}
}

// selectedContainers maps a container's ID to its index in the list.
type selectedContainers struct {
	selections map[string]int
}

func newSelectedContainers() *selectedContainers {
	return &selectedContainers{
		selections: make(map[string]int),
	}
}

func (selectedContainers *selectedContainers) selectContainerInList(id string, index int) {
	selectedContainers.selections[id] = index
}

func (selectedContainers selectedContainers) unselectContainerInList(id string) {
	delete(selectedContainers.selections, id)
}

// Model represents the containers component state.
type Model struct {
	shared.Component
	splitView          shared.SplitView
	selectedContainers *selectedContainers
	keybindings        *keybindings
	sessionState       sessionState

	foreground interface{} // Can be DeleteConfirmation, *ContainerLogs, or FormDialog

	currentContainerID string
	cpuHistory         []float64
	lastStats          client.ContainerStats

	inspection         types.ContainerJSON
	detailsKeybindings detailsKeybindings
}

var (
	_ tea.Model             = (*Model)(nil)
	_ shared.ComponentModel = (*Model)(nil)
)

func New() Model {
	containers, err := context.GetClient().GetContainers()
	if err != nil {
		containers = []client.Container{}
	}
	containerItems := make([]list.Item, 0, len(containers))
	for _, container := range containers {
		containerItems = append(
			containerItems,
			ContainerItem{
				Container:  container,
				isSelected: false,
				isWorking:  false,
				spinner:    newSpinner(),
			},
		)
	}

	width, height := context.GetWindowSize()

	delegate := newDefaultDelegate()
	listModel := list.New(containerItems, delegate, width, height)

	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.Filter.Focused.Prompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.Filter.Cursor.Color = colors.Primary()

	containerKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			containerKeybindings.pauseContainer,
			containerKeybindings.unpauseContainer,
			containerKeybindings.startContainer,
			containerKeybindings.stopContainer,
			containerKeybindings.restartContainer,
			containerKeybindings.removeContainer,
			containerKeybindings.showLogs,
			containerKeybindings.execShell,
			containerKeybindings.toggleSelection,
			containerKeybindings.toggleSelectionOfAll,
			containerKeybindings.switchTab,
		}
	}

	splitView := shared.NewSplitView(listModel, shared.NewViewportPane())

	model := Model{
		splitView:          splitView,
		selectedContainers: newSelectedContainers(),
		keybindings:        containerKeybindings,
		sessionState:       viewMain,
		cpuHistory:         make([]float64, 0),
		detailsKeybindings: newDetailsKeybindings(),
	}

	return model
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return MsgStatsTick(t)
	})
}

func (model *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.WindowWidth = msg.Width
	model.WindowHeight = msg.Height
	model.splitView.SetSize(msg.Width, msg.Height)

	switch model.sessionState {
	case viewOverlay:
		switch foregroundModel := model.foreground.(type) {
		case *ContainerLogs:
			foregroundModel.setDimensions(msg.Width, msg.Height)
			model.foreground = foregroundModel
		case DeleteConfirmation:
			foregroundModel.UpdateWindowDimensions(msg)
			model.foreground = foregroundModel
		}
	}
}

func (model Model) Init() tea.Cmd {
	return tickCmd()
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch model.sessionState {
	case viewMain:
		// Forward messages to SplitView first
		updatedSplitView, splitCmd := model.splitView.Update(msg)
		model.splitView = updatedSplitView
		cmds = append(cmds, splitCmd)

		// Handle keybindings when list is focused
		if model.splitView.Focus == shared.FocusList {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				if model.splitView.List.FilterState() == list.Filtering {
					break
				}

				if msg.String() == "q" {
					return model, tea.Quit
				}

				switch {
				case key.Matches(msg, model.keybindings.switchTab):
					// Handled by parent
					return model, nil
				case key.Matches(msg, model.keybindings.pauseContainer):
					cmds = append(cmds, model.handlePauseContainers())
				case key.Matches(msg, model.keybindings.unpauseContainer):
					cmds = append(cmds, model.handleUnpauseContainers())
				case key.Matches(msg, model.keybindings.startContainer):
					cmds = append(cmds, model.handleStartContainers())
				case key.Matches(msg, model.keybindings.stopContainer):
					cmds = append(cmds, model.handleStopContainers())
				case key.Matches(msg, model.keybindings.restartContainer):
					cmds = append(cmds, model.handleRestartContainers())
				case key.Matches(msg, model.keybindings.removeContainer):
					cmds = append(cmds, model.handleRemoveContainers())
				case key.Matches(msg, model.keybindings.showLogs):
					if cmd := model.handleShowLogs(); cmd != nil {
						cmds = append(cmds, cmd)
					}
				case key.Matches(msg, model.keybindings.execShell):
					if cmd := model.handleExecShell(); cmd != nil {
						cmds = append(cmds, cmd)
					}
				case key.Matches(msg, model.keybindings.toggleSelection):
					model.handleToggleSelection()
				case key.Matches(msg, model.keybindings.toggleSelectionOfAll):
					model.handleToggleSelectionOfAll()
				}
			}
		}

		// Update Detail Content if selection changes
		selectedItem := model.splitView.List.SelectedItem()
		if selectedItem != nil {
			if containerItem, ok := selectedItem.(ContainerItem); ok {
				if containerItem.ID != model.currentContainerID {
					model.currentContainerID = containerItem.ID
					model.cpuHistory = make([]float64, 0)
					cmds = append(cmds, func() tea.Msg {
						containerInfo, err := context.GetClient().InspectContainer(containerItem.ID)
						return MsgContainerInspection{ID: containerItem.ID, Container: containerInfo, Err: err}
					})
				}
			}
		}

	case viewOverlay:
		// Update SplitView for background resize but don't process keys
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			updatedSplitView, splitCmd := model.splitView.Update(msg)
			model.splitView = updatedSplitView
			cmds = append(cmds, splitCmd)
		}

		if model.foreground != nil {
			// Type switch to call Update on different foreground types
			switch fg := model.foreground.(type) {
			case DeleteConfirmation:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			case *ContainerLogs:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			case shared.SmartDialog:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			case shared.FormDialog:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.UpdateWindowDimensions(msg)

	case MessageCloseOverlay:
		model.sessionState = viewMain

	case MessageOpenDeleteConfirmationDialog:
		fg := newDeleteConfirmation(msg.requestedContainersToDelete...)
		model.foreground = fg
		model.sessionState = viewOverlay
		cmds = append(cmds, fg.Init())

	case MessageConfirmDelete:
		cmds = append(cmds, model.handleConfirmationOfRemoveContainers())

	case MessageContainerOperationResult:
		if cmd := model.handleContainerOperationResult(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case MsgStatsTick:
		cmds = append(cmds, model.handleStatsTick()...)

	case MsgContainerInspection:
		if msg.ID == model.currentContainerID && msg.Err == nil {
			model.inspection = msg.Container
			if pane, ok := model.splitView.Detail.(*shared.ViewportPane); ok {
				inspectionContent := formatInspection(model.inspection, model.lastStats, model.cpuHistory, pane.Viewport.Width())
				pane.SetContent(inspectionContent)
			}
		}

	case MsgContainerStats:
		if msg.ID == model.currentContainerID && msg.Err == nil {
			model.lastStats = msg.Stats
			model.cpuHistory = append(model.cpuHistory, msg.Stats.CPUPercent)
			if len(model.cpuHistory) > 30 {
				model.cpuHistory = model.cpuHistory[1:]
			}
			if pane, ok := model.splitView.Detail.(*shared.ViewportPane); ok {
				inspectionContent := formatInspection(model.inspection, model.lastStats, model.cpuHistory, pane.Viewport.Width())
				pane.SetContent(inspectionContent)
			}
		}

	// Forward spinner ticks to list items
	case spinner.TickMsg:
		var batchCmds []tea.Cmd
		for index, item := range model.splitView.List.Items() {
			if container, ok := item.(ContainerItem); ok && container.isWorking {
				var cmd tea.Cmd
				container.spinner, cmd = container.spinner.Update(msg)
				model.splitView.List.SetItem(index, container)
				batchCmds = append(batchCmds, cmd)
			}
		}
		if len(batchCmds) > 0 {
			cmds = append(cmds, tea.Batch(batchCmds...))
		}
	}

	return model, tea.Batch(cmds...)
}

func (model Model) View() tea.View {
	if model.sessionState == viewOverlay && model.foreground != nil {
		// Type switch to call View on different foreground types
		var fgView string
		switch fg := model.foreground.(type) {
		case DeleteConfirmation:
			fgView = fg.View()
		case *ContainerLogs:
			fgView = fg.View()
		case shared.SmartDialog:
			fgView = fg.View()
		case shared.FormDialog:
			fgView = fg.View()
		}

		return shared.RenderOverlay(
			model.splitView.View(),
			fgView,
			model.WindowWidth,
			model.WindowHeight,
		)
	}

	return tea.NewView(model.splitView.View())
}

func (model *Model) handleStatsTick() []tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tickCmd())

	if model.sessionState == viewMain {
		selectedItem := model.splitView.List.SelectedItem()
		if selectedItem != nil {
			if containerItem, ok := selectedItem.(ContainerItem); ok && containerItem.State == "running" {
				cmds = append(cmds, func() tea.Msg {
					containerStats, err := context.GetClient().GetContainerStats(containerItem.ID)
					return MsgContainerStats{ID: containerItem.ID, Stats: containerStats, Err: err}
				})
			}
		}
	}
	return cmds
}

func (model Model) ShortHelp() []key.Binding {
	if model.sessionState == viewOverlay {
		if helpKeyMap, ok := model.foreground.(help.KeyMap); ok {
			return helpKeyMap.ShortHelp()
		}
		return nil
	}

	switch model.splitView.Focus {
	case shared.FocusList:
		return model.splitView.List.ShortHelp()
	case shared.FocusDetail:
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
		}
	}

	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	if model.sessionState == viewOverlay {
		return nil
	}

	switch model.splitView.Focus {
	case shared.FocusList:
		return model.splitView.List.FullHelp()
	case shared.FocusDetail:
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
		}
	}
	return nil
}

// Handler Functions (Moved from list.go/handlers.go)

func (model *Model) getSelectedContainerIDs() []string {
	selectedContainerIDs := make([]string, 0, len(model.selectedContainers.selections))
	for containerID := range model.selectedContainers.selections {
		selectedContainerIDs = append(selectedContainerIDs, containerID)
	}
	return selectedContainerIDs
}

func (model *Model) getSelectedContainerIndices() []int {
	selectedContainerIndices := make([]int, 0, len(model.selectedContainers.selections))
	for _, index := range model.selectedContainers.selections {
		selectedContainerIndices = append(selectedContainerIndices, index)
	}
	return selectedContainerIndices
}

func (model *Model) setWorkingState(containerIDs []string, working bool) {
	items := model.splitView.List.Items()
	for index, item := range items {
		if container, ok := item.(ContainerItem); ok && slices.Contains(containerIDs, container.ID) {
			container.isWorking = working
			if working {
				container.spinner = newSpinner()
			}
			model.splitView.List.SetItem(index, container)
		}
	}
}

func (model *Model) anySelectedWorking() bool {
	for containerID := range model.selectedContainers.selections {
		if item := model.findItemByID(containerID); item != nil && item.isWorking {
			return true
		}
	}
	return false
}

func (model *Model) findItemByID(containerID string) *ContainerItem {
	items := model.splitView.List.Items()
	for _, item := range items {
		if container, ok := item.(ContainerItem); ok && container.ID == containerID {
			return &container
		}
	}
	return nil
}

func (model *Model) handlePauseContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		selectedContainerIDs := model.getSelectedContainerIDs()
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Pause, selectedContainerIDs)
	} else {
		selectedItem, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Pause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleUnpauseContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		selectedContainerIDs := model.getSelectedContainerIDs()
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Unpause, selectedContainerIDs)
	} else {
		selectedItem, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Unpause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleStartContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		selectedContainerIDs := model.getSelectedContainerIDs()
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Start, selectedContainerIDs)
	} else {
		selectedItem, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Start, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleStopContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		selectedContainerIDs := model.getSelectedContainerIDs()
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Stop, selectedContainerIDs)
	} else {
		selectedItem, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Stop, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleRestartContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		selectedContainerIDs := model.getSelectedContainerIDs()
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Restart, selectedContainerIDs)
	} else {
		selectedItem, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Restart, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleRemoveContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		selectedContainerIndices := model.getSelectedContainerIndices()

		var requestedContainersToDelete []*ContainerItem
		items := model.splitView.List.Items()

		for _, index := range selectedContainerIndices {
			requestedContainer := items[index].(ContainerItem)
			requestedContainersToDelete = append(requestedContainersToDelete, &requestedContainer)
		}

		return func() tea.Msg {
			return MessageOpenDeleteConfirmationDialog{requestedContainersToDelete}
		}
	} else {
		item, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok && !item.isWorking {
			return func() tea.Msg {
				return MessageOpenDeleteConfirmationDialog{[]*ContainerItem{&item}}
			}
		}
	}
	return nil
}

func (model *Model) handleShowLogs() tea.Cmd {
	item, ok := model.splitView.List.SelectedItem().(ContainerItem)
	if !ok || item.isWorking {
		return nil
	}

	if item.State != "running" {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	command := exec.Command("sh", "-c", "docker logs \"$0\" 2>&1 | less", item.ID) //nolint:gosec
	return tea.ExecProcess(command, func(err error) tea.Msg {
		if err != nil {
			return notifications.AddNotificationMsg{
				Message:  err.Error(),
				Level:    notifications.Error,
				Duration: 10 * 1000 * 1000 * 1000,
			}
		}
		return nil
	})
}

func (model *Model) handleExecShell() tea.Cmd {
	item, ok := model.splitView.List.SelectedItem().(ContainerItem)
	if !ok || item.isWorking {
		return nil
	}

	if item.State != "running" {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	command := exec.Command("sh", "-c", "exec docker exec -it \"$0\" /bin/sh", item.ID) //nolint:gosec
	return tea.ExecProcess(command, func(err error) tea.Msg {
		if err != nil {
			return notifications.AddNotificationMsg{
				Message:  err.Error(),
				Level:    notifications.Error,
				Duration: 10 * 1000 * 1000 * 1000,
			}
		}
		return nil
	})
}

func (model *Model) handleConfirmationOfRemoveContainers() tea.Cmd {
	if len(model.selectedContainers.selections) > 0 {
		selectedContainerIDs := model.getSelectedContainerIDs()
		model.setWorkingState(selectedContainerIDs, true)
		return PerformContainerOperation(Remove, selectedContainerIDs)
	} else {
		item, ok := model.splitView.List.SelectedItem().(ContainerItem)
		if ok {
			model.setWorkingState([]string{item.ID}, true)
			return PerformContainerOperation(Remove, []string{item.ID})
		}
	}
	return nil
}

func (model *Model) handleToggleSelection() {
	index := model.splitView.List.Index()
	selectedItem, ok := model.splitView.List.SelectedItem().(ContainerItem)
	if ok && !selectedItem.isWorking {
		isSelected := selectedItem.isSelected

		if isSelected {
			model.selectedContainers.unselectContainerInList(selectedItem.ID)
		} else {
			model.selectedContainers.selectContainerInList(selectedItem.ID, index)
		}

		selectedItem.isSelected = !isSelected
		model.splitView.List.SetItem(index, selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	allNonWorkingAlreadySelected := true
	items := model.splitView.List.Items()

	for _, item := range items {
		if container, ok := item.(ContainerItem); ok && !container.isWorking {
			if _, selected := model.selectedContainers.selections[container.ID]; !selected {
				allNonWorkingAlreadySelected = false
				break
			}
		}
	}

	if allNonWorkingAlreadySelected {
		model.selectedContainers = newSelectedContainers()
		for index, item := range model.splitView.List.Items() {
			container, ok := item.(ContainerItem)
			if ok {
				container.isSelected = false
				model.splitView.List.SetItem(index, container)
			}
		}
	} else {
		model.selectedContainers = newSelectedContainers()
		for index, item := range model.splitView.List.Items() {
			container, ok := item.(ContainerItem)
			if ok && !container.isWorking {
				container.isSelected = true
				model.splitView.List.SetItem(index, container)
				model.selectedContainers.selectContainerInList(container.ID, index)
			}
		}
	}
}

func (model *Model) handleContainerOperationResult(msg MessageContainerOperationResult) tea.Cmd {
	model.setWorkingState(msg.IDs, false)

	if msg.Error != nil {
		return notifications.ShowError(msg.Error)
	}

	if msg.Operation == Remove {
		items := model.splitView.List.Items()
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
			model.splitView.List.RemoveItem(index)
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
	case Remove:
		return nil
	default:
		return nil
	}

	items := model.splitView.List.Items()
	for index, item := range items {
		if container, ok := item.(ContainerItem); ok {
			for _, containerID := range msg.IDs {
				if container.ID == containerID {
					container.State = newState
					model.splitView.List.SetItem(index, container)
					break
				}
			}
		}
	}
	return nil
}

func formatInspection(container types.ContainerJSON, containerStats client.ContainerStats, cpuHistory []float64, viewportWidth int) string {
	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).MarginBottom(1)
	builder.WriteString(headerStyle.Render(fmt.Sprintf("%s (%s)", container.Name, container.ID[:12])) + "\n")

	infoStyle := lipgloss.NewStyle().Foreground(colors.Text()).MarginBottom(1)
	stateColor := colors.Success()
	if !container.State.Running {
		stateColor = colors.Muted()
	}
	stateString := lipgloss.NewStyle().Foreground(stateColor).Render(container.State.Status)

	builder.WriteString(infoStyle.Render("Image: "+container.Config.Image) + "\n")
	builder.WriteString(fmt.Sprintf("State: %s\n\n", stateString))

	if container.State.Running && len(cpuHistory) > 0 {
		graphHeight := 8
		graphWidth := viewportWidth - 10
		if graphWidth > 10 {
			usageGraph := asciigraph.Plot(cpuHistory,
				asciigraph.Height(graphHeight),
				asciigraph.Width(graphWidth),
				asciigraph.Caption("CPU Usage (%)"),
				asciigraph.SeriesColors(
					asciigraph.Blue,
				),
			)
			builder.WriteString(usageGraph + "\n\n")

			statsStyle := lipgloss.NewStyle().Foreground(colors.Primary())
			memUsageMB := containerStats.MemUsage / 1024 / 1024
			memLimitMB := containerStats.MemLimit / 1024 / 1024
			builder.WriteString(statsStyle.Render(fmt.Sprintf("CPU: %.2f%% | Mem: %.0fMB / %.0fMB",
				containerStats.CPUPercent, memUsageMB, memLimitMB)) + "\n\n")
		}
	}

	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).Underline(true).MarginTop(1).MarginBottom(0)

	builder.WriteString(sectionHeader.Render("Configuration") + "\n")
	builder.WriteString(fmt.Sprintf("Cmd: %v\n", container.Config.Cmd))
	builder.WriteString(fmt.Sprintf("Entrypoint: %v\n", container.Config.Entrypoint))
	builder.WriteString(fmt.Sprintf("WorkingDir: %s\n", container.Config.WorkingDir))

	if len(container.Config.Env) > 0 {
		builder.WriteString("\n" + sectionHeader.Render("Environment Variables") + "\n")
		for _, envVar := range container.Config.Env {
			builder.WriteString(envVar + "\n")
		}
	}

	if len(container.Mounts) > 0 {
		builder.WriteString("\n" + sectionHeader.Render("Mounts") + "\n")
		for _, mount := range container.Mounts {
			builder.WriteString(fmt.Sprintf("%s -> %s (%s)\n", mount.Source, mount.Destination, mount.Type))
		}
	}

	if container.NetworkSettings != nil && len(container.NetworkSettings.Ports) > 0 {
		builder.WriteString("\n" + sectionHeader.Render("Ports") + "\n")
		for port, bindings := range container.NetworkSettings.Ports {
			portBindings := make([]string, 0, len(bindings))
			for _, portBinding := range bindings {
				portBindings = append(portBindings, fmt.Sprintf("%s:%s", portBinding.HostIP, portBinding.HostPort))
			}
			builder.WriteString(fmt.Sprintf("%s -> %v\n", port, portBindings))
		}
	}

	return builder.String()
}
