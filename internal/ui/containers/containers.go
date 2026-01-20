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
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/notifications"
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

// Model represents the containers component state.
type Model struct {
	components.ResourceView[string, ContainerItem]
	keybindings  *keybindings
	sessionState sessionState

	foreground interface{} // Can be DeleteConfirmation, *ContainerLogs, or FormDialog

	currentContainerID string
	cpuHistory         []float64
	lastStats          client.ContainerStats

	inspection         types.ContainerJSON
	detailsKeybindings detailsKeybindings

	WindowWidth  int
	WindowHeight int
}

// Ensure Model satisfies base.Component but we cannot directly assign (*Model)(nil) if Model has embedded fields that complicate it?
// Actually base.Component is struct { WindowWidth, WindowHeight int }.
// Model embeds ResourceView which embeds base.Component.
// So Model HAS WindowWidth/WindowHeight.
// BUT `var _ base.Component = (*Model)(nil)` tries to assign *Model to base.Component (struct).
// This is invalid Go. You can assign to interface, but base.Component is a struct.
// You cannot say "Model implements struct".
// If base.Component was an interface it would be fine.
// Since base.Component is a struct, we don't need this check.

var (
	_ tea.Model = (*Model)(nil)
)

func New() Model {
	containerKeybindings := newKeybindings()

	// Initialize ResourceView
	fetchContainers := func() ([]ContainerItem, error) {
		containers, err := context.GetClient().GetContainers()
		if err != nil {
			return nil, err
		}
		items := make([]ContainerItem, 0, len(containers))
		for _, container := range containers {
			items = append(items, ContainerItem{
				Container:  container,
				isSelected: false,
				isWorking:  false,
				spinner:    newSpinner(),
			})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, ContainerItem](
		"Containers",
		fetchContainers,
		func(item ContainerItem) string { return item.ID },
		func(item ContainerItem) string { return item.Name },
		func(w, h int) {
			// Window resize handled by base component
		},
	)

	// Set the custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        containerKeybindings,
		sessionState:       viewMain,
		cpuHistory:         make([]float64, 0),
		detailsKeybindings: newDetailsKeybindings(),
	}

	// Add custom keybindings to help
	model.ResourceView.AdditionalHelp = []key.Binding{
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
	model.ResourceView.UpdateWindowDimensions(msg)

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
	return tea.Batch(
		model.ResourceView.Init(),
		tickCmd(),
	)
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch model.sessionState {
	case viewMain:
		// Forward messages to ResourceView first
		updatedView, viewCmd := model.ResourceView.Update(msg)
		model.ResourceView = updatedView
		cmds = append(cmds, viewCmd)

		// Handle keybindings when list is focused
		if model.ResourceView.IsListFocused() {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				if model.ResourceView.IsFiltering() {
					break
				}

				// Don't intercept global navigation keys
				if key.Matches(msg, model.keybindings.switchTab) {
					return model, nil
				}

				switch {
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
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil {
			if selectedItem.ID != model.currentContainerID {
				model.currentContainerID = selectedItem.ID
				model.cpuHistory = make([]float64, 0)

				// Capture ID for closure
				id := selectedItem.ID
				cmds = append(cmds, func() tea.Msg {
					containerInfo, err := context.GetClient().InspectContainer(id)
					return MsgContainerInspection{ID: id, Container: containerInfo, Err: err}
				})
			}
		}

	case viewOverlay:
		// Update ResourceView for background resize but don't process keys
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			updatedView, viewCmd := model.ResourceView.Update(msg)
			model.ResourceView = updatedView
			cmds = append(cmds, viewCmd)
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
			case components.SmartDialog:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			case components.FormDialog:
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
			inspectionContent := formatInspection(model.inspection, model.lastStats, model.cpuHistory, model.ResourceView.GetContentWidth())
			model.ResourceView.SetContent(inspectionContent)
		}

	case MsgContainerStats:
		if msg.ID == model.currentContainerID && msg.Err == nil {
			model.lastStats = msg.Stats
			model.cpuHistory = append(model.cpuHistory, msg.Stats.CPUPercent)
			if len(model.cpuHistory) > 30 {
				model.cpuHistory = model.cpuHistory[1:]
			}
			inspectionContent := formatInspection(model.inspection, model.lastStats, model.cpuHistory, model.ResourceView.GetContentWidth())
			model.ResourceView.SetContent(inspectionContent)
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
		case components.SmartDialog:
			fgView = fg.View()
		case components.FormDialog:
			fgView = fg.View()
		}

		return components.RenderOverlay(
			model.ResourceView.View(),
			fgView,
			model.WindowWidth,
			model.WindowHeight,
		)
	}

	return tea.NewView(model.ResourceView.View())
}

func (model *Model) handleStatsTick() []tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tickCmd())

	if model.sessionState == viewMain {
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil && selectedItem.State == "running" {
			// Capture ID for closure
			id := selectedItem.ID
			cmds = append(cmds, func() tea.Msg {
				containerStats, err := context.GetClient().GetContainerStats(id)
				return MsgContainerStats{ID: id, Stats: containerStats, Err: err}
			})
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
	return model.ResourceView.ShortHelp()
}

func (model Model) FullHelp() [][]key.Binding {
	if model.sessionState == viewOverlay {
		return nil
	}
	return model.ResourceView.FullHelp()
}

// Handler Functions (Moved from list.go/handlers.go)

func (model *Model) getSelectedContainerIDs() []string {
	// Use ResourceView's SelectionManager to get selections
	return model.ResourceView.GetSelectedIDs()
}

func (model *Model) setWorkingState(containerIDs []string, working bool) {
	items := model.ResourceView.GetItems()
	var updatedItems []ContainerItem

	for _, item := range items {
		if slices.Contains(containerIDs, item.ID) {
			item.isWorking = working
			if working {
				item.spinner = newSpinner()
			}
			updatedItems = append(updatedItems, item)
		} else {
			updatedItems = append(updatedItems, item)
		}
	}

	// We need to update items in the list.
	// Since ResourceView doesn't expose partial updates easily yet,
	// we'll update the whole list or individual items if we can match indices.
	// However, ResourceView is generic. Let's iterate and update.

	currentItems := model.ResourceView.GetItems()
	for i, item := range currentItems {
		if slices.Contains(containerIDs, item.ID) {
			item.isWorking = working
			if working {
				item.spinner = newSpinner()
			}
			model.ResourceView.SetItem(i, item)
		}
	}
}

func (model *Model) anySelectedWorking() bool {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	items := model.ResourceView.GetItems()

	for _, item := range items {
		if slices.Contains(selectedIDs, item.ID) {
			if item.isWorking {
				return true
			}
		}
	}
	return false
}

func (model *Model) handlePauseContainers() tea.Cmd {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedIDs, true)
		return PerformContainerOperation(Pause, selectedIDs)
	} else {
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Pause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleUnpauseContainers() tea.Cmd {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedIDs, true)
		return PerformContainerOperation(Unpause, selectedIDs)
	} else {
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Unpause, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleStartContainers() tea.Cmd {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedIDs, true)
		return PerformContainerOperation(Start, selectedIDs)
	} else {
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Start, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleStopContainers() tea.Cmd {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedIDs, true)
		return PerformContainerOperation(Stop, selectedIDs)
	} else {
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Stop, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleRestartContainers() tea.Cmd {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		model.setWorkingState(selectedIDs, true)
		return PerformContainerOperation(Restart, selectedIDs)
	} else {
		selectedItem := model.ResourceView.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			model.setWorkingState([]string{selectedItem.ID}, true)
			return PerformContainerOperation(Restart, []string{selectedItem.ID})
		}
	}
	return nil
}

func (model *Model) handleRemoveContainers() tea.Cmd {
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}

		var requestedContainersToDelete []*ContainerItem
		items := model.ResourceView.GetItems()

		for _, item := range items {
			if slices.Contains(selectedIDs, item.ID) {
				// Create a copy to take address safely
				itm := item
				requestedContainersToDelete = append(requestedContainersToDelete, &itm)
			}
		}

		return func() tea.Msg {
			return MessageOpenDeleteConfirmationDialog{requestedContainersToDelete}
		}
	} else {
		item := model.ResourceView.GetSelectedItem()
		if item != nil && !item.isWorking {
			// Create a copy to take address safely
			itm := *item
			return func() tea.Msg {
				return MessageOpenDeleteConfirmationDialog{[]*ContainerItem{&itm}}
			}
		}
	}
	return nil
}

func (model *Model) handleShowLogs() tea.Cmd {
	item := model.ResourceView.GetSelectedItem()
	if item == nil || item.isWorking {
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
	item := model.ResourceView.GetSelectedItem()
	if item == nil || item.isWorking {
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
	selectedIDs := model.ResourceView.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		model.setWorkingState(selectedIDs, true)
		return PerformContainerOperation(Remove, selectedIDs)
	} else {
		item := model.ResourceView.GetSelectedItem()
		if item != nil {
			model.setWorkingState([]string{item.ID}, true)
			return PerformContainerOperation(Remove, []string{item.ID})
		}
	}
	return nil
}

func (model *Model) handleToggleSelection() {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem != nil && !selectedItem.isWorking {
		model.ResourceView.ToggleSelection(selectedItem.ID)

		// Update the visual state of the item
		index := model.ResourceView.GetSelectedIndex()
		selectedItem.isSelected = !selectedItem.isSelected
		model.ResourceView.SetItem(index, *selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	// First check if we need to select all or deselect all
	// Logic: If any non-working item is unselected, select all. Otherwise deselect all.

	items := model.ResourceView.GetItems()
	selectedIDs := model.ResourceView.GetSelectedIDs()

	shouldSelectAll := false
	for _, item := range items {
		if !item.isWorking {
			if !slices.Contains(selectedIDs, item.ID) {
				shouldSelectAll = true
				break
			}
		}
	}

	if shouldSelectAll {
		// Select all non-working items
		for i, item := range items {
			if !item.isWorking {
				if !slices.Contains(selectedIDs, item.ID) {
					model.ResourceView.ToggleSelection(item.ID)
				}
				// Visual update
				item.isSelected = true
				model.ResourceView.SetItem(i, item)
			}
		}
	} else {
		// Deselect all
		for i, item := range items {
			if slices.Contains(selectedIDs, item.ID) {
				model.ResourceView.ToggleSelection(item.ID)
			}
			// Visual update
			item.isSelected = false
			model.ResourceView.SetItem(i, item)
		}
	}
}

func (model *Model) handleContainerOperationResult(msg MessageContainerOperationResult) tea.Cmd {
	model.setWorkingState(msg.IDs, false)

	if msg.Error != nil {
		return notifications.ShowError(msg.Error)
	}

	if msg.Operation == Remove {
		// Remove items from list
		// Since ResourceView doesn't have RemoveItem by ID easily, we need to find indices.
		// NOTE: Removing items while iterating or by index requires care as indices shift.
		// However, standard bubbletea list.Model handles RemoveItem safely if we do it one by one?
		// Actually, let's refresh the list from source or filter it locally.

		// Ideally we'd just re-fetch, but for instant feedback let's filter the current items.
		currentItems := model.ResourceView.GetItems()
		var newItems []ContainerItem
		for _, item := range currentItems {
			if !slices.Contains(msg.IDs, item.ID) {
				newItems = append(newItems, item)
			}
		}

		// This is a bit of a hack since ResourceView encapsulates the list model.
		// But ResourceView usually exposes SetItems.
		// Let's assume ResourceView allows full replacement or we trigger a refresh.
		// Since ResourceView is generic, we can just trigger a refresh if we had a Reload command.
		// But here we want to update the local state.

		// Let's assume for now we just trigger a refresh.
		return model.ResourceView.Refresh()
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

	// Update states locally
	items := model.ResourceView.GetItems()
	for i, item := range items {
		if slices.Contains(msg.IDs, item.ID) {
			item.State = newState
			model.ResourceView.SetItem(i, item)
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
