// Package containers defines the containers component.
package containers

import (
	stdcontext "context"
	"fmt"
	"os/exec"
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

// MsgContainerInspection contains the inspection data for a container.
type MsgContainerInspection struct {
	ID        string
	Container types.ContainerJSON
	Err       error
}

// MsgRestoreScroll is sent to restore scroll position after content is set
type MsgRestoreScroll struct{}

// MsgRefreshContainers is sent periodically to refresh the container list
type MsgRefreshContainers time.Time

type detailsKeybindings struct {
	Up         key.Binding
	Down       key.Binding
	Switch     key.Binding
	ToggleJSON key.Binding
	CopyOutput key.Binding
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
		ToggleJSON: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "toggle JSON/YAML"),
		),
		CopyOutput: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy to clipboard"),
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
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "switch tab"),
		),
	}
}

// Model represents the containers component state.
type Model struct {
	components.ResourceView[string, ContainerItem]
	keybindings *keybindings

	currentContainerID string
	inspection         types.ContainerJSON
	detailsKeybindings detailsKeybindings

	// Track scroll position per container ID
	scrollPositions map[string]int

	// Track current output format (can toggle with 'J')
	currentFormat string

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

func New() Model {
	containerKeybindings := newKeybindings()

	// Initialize ResourceView
	fetchContainers := func() ([]ContainerItem, error) {
		containers, err := state.GetClient().GetContainers(stdcontext.Background())
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
		detailsKeybindings: newDetailsKeybindings(),
		scrollPositions:    make(map[string]int),
		currentFormat:      "", // Empty means use config default
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
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

func (model *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.WindowWidth = msg.Width
	model.WindowHeight = msg.Height
	model.ResourceView.UpdateWindowDimensions(msg)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return MsgRefreshContainers(t)
	})
}

func (model Model) Init() tea.Cmd {
	return tea.Batch(model.ResourceView.Init(), tickCmd())
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// 1. Forward messages to ResourceView first (handles dialog closing, resizing, etc.)
	updatedView, cmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// 2. Handle custom messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.UpdateWindowDimensions(msg)

	case MsgRefreshContainers:
		// Schedule next refresh
		cmds = append(cmds, tickCmd())
		// Refresh the containers list via ResourceView
		cmds = append(cmds, model.Refresh())

	case MessageContainerOperationResult:
		if cmd := model.handleContainerOperationResult(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case MsgContainerInspection:
		if msg.ID == model.currentContainerID && msg.Err == nil {
			model.inspection = msg.Container
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.restoreScrollPosition()
	}

	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteContainer" {
				containerIDs, ok := confirmMsg.Action.Payload.([]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for DeleteContainer"))
				}
				spinnerCmd := model.setWorkingState(containerIDs, true)
				model.CloseOverlay()
				return model, tea.Batch(spinnerCmd, PerformContainerOperations(Remove, containerIDs))
			}
		}
		return model, tea.Batch(cmds...)
	}

	// 4. Handle keybindings when list is focused
	if model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.IsFiltering() {
				break
			}

			// Don't intercept global navigation keys
			if key.Matches(msg, model.keybindings.switchTab) {
				return model, tea.Batch(cmds...)
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
				model.handleRemoveContainers()
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

	// 5. Handle keybindings when detail pane is focused
	if !model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, model.detailsKeybindings.ToggleJSON):
				cmd := model.handleToggleFormat()
				cmds = append(cmds, cmd)
			case key.Matches(msg, model.detailsKeybindings.CopyOutput):
				cmd := model.handleCopyToClipboard()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// 6. Update Detail Content if selection changes
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil {
		if selectedItem.ID != model.currentContainerID {
			// Save scroll position of previous container
			model.saveScrollPosition()

			model.currentContainerID = selectedItem.ID

			// Capture ID for closure
			id := selectedItem.ID
			cmds = append(cmds, func() tea.Msg {
				containerInfo, err := state.GetClient().InspectContainer(stdcontext.Background(), id)
				return MsgContainerInspection{ID: id, Container: containerInfo, Err: err}
			})
		}
	}

	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	return model.ResourceView.View()
}

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model Model) ShortHelp() []key.Binding {
	// If detail pane is focused, show detail keybindings
	if !model.IsListFocused() {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.ToggleJSON,
			model.detailsKeybindings.CopyOutput,
			model.detailsKeybindings.Switch,
		}
	}

	return model.ResourceView.ShortHelp()
}

// getViewport returns the viewport from the detail pane if available
func (model *Model) getViewport() *viewport.Model {
	if vp, ok := model.SplitView.Detail.(*components.ViewportPane); ok {
		return &vp.Viewport
	}
	return nil
}

// saveScrollPosition saves the current scroll position for the current container
func (model *Model) saveScrollPosition() {
	if model.currentContainerID != "" {
		if vp := model.getViewport(); vp != nil {
			model.scrollPositions[model.currentContainerID] = vp.YOffset()
		}
	}
}

// restoreScrollPosition restores the scroll position for the current container
func (model *Model) restoreScrollPosition() {
	if model.currentContainerID != "" {
		if vp := model.getViewport(); vp != nil {
			if pos, exists := model.scrollPositions[model.currentContainerID]; exists {
				vp.SetYOffset(pos)
			} else {
				vp.SetYOffset(0) // Reset to top for new containers
			}
		}
	}
}

// refreshInspectionContent regenerates and sets the inspection content
func (model *Model) refreshInspectionContent() {
	if model.inspection.ID == "" {
		return
	}

	// Determine format to use
	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	content := builders.BuildContainerPanel(model.inspection, model.GetContentWidth(), false, format)
	model.SetContent(content)

	// Restore scroll position after content is set
	// Need to do this on next frame after viewport processes the content
	model.restoreScrollPosition()
}

// handleCopyToClipboard copies the current inspection output to clipboard
func (model *Model) handleCopyToClipboard() tea.Cmd {
	if model.inspection.ID == "" {
		return nil
	}

	// Determine which format to use
	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	// Marshal the data without syntax highlighting
	data, err := infopanel.MarshalToFormat(model.inspection, format)
	if err != nil {
		return notifications.ShowError(err)
	}

	// Copy to clipboard
	err = clipboard.WriteAll(data)
	if err != nil {
		return notifications.ShowError(err)
	}

	return notifications.ShowSuccess("Copied to clipboard")
}

// handleToggleFormat toggles between JSON and YAML format
func (model *Model) handleToggleFormat() tea.Cmd {
	// Determine current effective format
	currentFormat := model.currentFormat
	if currentFormat == "" {
		cfg := state.GetConfig()
		currentFormat = cfg.InspectionFormat
		if currentFormat == "" {
			currentFormat = "yaml"
		}
	}

	// Toggle to the opposite format
	if currentFormat == "json" {
		model.currentFormat = "yaml"
	} else {
		model.currentFormat = "json"
	}

	// Refresh content with new format
	model.refreshInspectionContent()

	return notifications.ShowSuccess("Switched to " + model.currentFormat)
}

func (model Model) FullHelp() [][]key.Binding {
	// If detail pane is focused, show detail keybindings
	if !model.IsListFocused() {
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
			{
				model.detailsKeybindings.ToggleJSON,
				model.detailsKeybindings.CopyOutput,
			},
		}
	}

	return model.ResourceView.FullHelp()
}

// Handler Functions (Moved from list.go/handlers.go)

func (model *Model) setWorkingState(containerIDs []string, working bool) tea.Cmd {
	var cmds []tea.Cmd

	currentItems := model.GetItems()
	for i, item := range currentItems {
		if slices.Contains(containerIDs, item.ID) {
			item.isWorking = working
			if working {
				item.spinner = newSpinner()
				// Capture the spinner in a local variable to avoid closure capture bug
				spinner := item.spinner
				cmds = append(cmds, func() tea.Msg {
					return spinner.Tick()
				})
			}
			model.SetItem(i, item)
		}
	}

	return tea.Batch(cmds...)
}

func (model *Model) anySelectedWorking() bool {
	selectedIDs := model.GetSelectedIDs()
	items := model.GetItems()

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
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		spinnerCmd := model.setWorkingState(selectedIDs, true)
		return tea.Batch(spinnerCmd, PerformContainerOperations(Pause, selectedIDs))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Pause, selectedItem.ID))
		}
	}
	return nil
}

func (model *Model) handleUnpauseContainers() tea.Cmd {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		spinnerCmd := model.setWorkingState(selectedIDs, true)
		return tea.Batch(spinnerCmd, PerformContainerOperations(Unpause, selectedIDs))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Unpause, selectedItem.ID))
		}
	}
	return nil
}

func (model *Model) handleStartContainers() tea.Cmd {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		spinnerCmd := model.setWorkingState(selectedIDs, true)
		return tea.Batch(spinnerCmd, PerformContainerOperations(Start, selectedIDs))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Start, selectedItem.ID))
		}
	}
	return nil
}

func (model *Model) handleStopContainers() tea.Cmd {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		spinnerCmd := model.setWorkingState(selectedIDs, true)
		return tea.Batch(spinnerCmd, PerformContainerOperations(Stop, selectedIDs))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Stop, selectedItem.ID))
		}
	}
	return nil
}

func (model *Model) handleRestartContainers() tea.Cmd {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}
		spinnerCmd := model.setWorkingState(selectedIDs, true)
		return tea.Batch(spinnerCmd, PerformContainerOperations(Restart, selectedIDs))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Restart, selectedItem.ID))
		}
	}
	return nil
}

func (model *Model) handleRemoveContainers() {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return
		}

		// Build confirmation message with container names
		var containerNames []string
		items := model.GetItems()
		for _, item := range items {
			if slices.Contains(selectedIDs, item.ID) {
				containerNames = append(containerNames, item.Name)
			}
		}

		message := fmt.Sprintf("Are you sure you want to delete %d container(s)?", len(containerNames))
		if len(containerNames) <= 5 {
			message += "\n\n"
			for _, name := range containerNames {
				message += fmt.Sprintf("• %s\n", name)
			}
		}

		confirmDialog := components.NewDialog(
			message,
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Delete", Action: base.SmartDialogAction{Type: "DeleteContainer", Payload: selectedIDs}},
			},
		)
		model.SetOverlay(confirmDialog)
	} else {
		item := model.GetSelectedItem()
		if item != nil && !item.isWorking {
			confirmDialog := components.NewDialog(
				fmt.Sprintf("Are you sure you want to delete container %s?", item.Name),
				[]components.DialogButton{
					{Label: "Cancel"},
					{Label: "Delete", Action: base.SmartDialogAction{Type: "DeleteContainer", Payload: []string{item.ID}}},
				},
			)
			model.SetOverlay(confirmDialog)
		}
	}
}

func (model *Model) handleShowLogs() tea.Cmd {
	item := model.GetSelectedItem()
	if item == nil || item.isWorking {
		return nil
	}

	if item.State != "running" {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	// Use exec.Command with proper argument passing to avoid shell injection
	// docker logs <container-id> 2>&1 | less
	dockerCmd := exec.Command("docker", "logs", item.ID)
	lessCmd := exec.Command("less")

	// Pipe docker output to less
	pipe, err := dockerCmd.StdoutPipe()
	if err != nil {
		return notifications.ShowError(err)
	}
	dockerCmd.Stderr = dockerCmd.Stdout // Redirect stderr to stdout
	lessCmd.Stdin = pipe

	// Start both commands
	if err := dockerCmd.Start(); err != nil {
		return notifications.ShowError(err)
	}

	command := lessCmd
	return tea.ExecProcess(command, func(err error) tea.Msg {
		// Kill the docker logs command when less exits
		// This is important because docker logs -f runs forever
		if dockerCmd.Process != nil {
			_ = dockerCmd.Process.Kill()
		}
		// Wait for docker command to clean up
		_ = dockerCmd.Wait()
		if err != nil {
			return notifications.AddNotificationMsg{
				Message:  err.Error(),
				Level:    notifications.Error,
				Duration: 10 * time.Second,
			}
		}
		return nil
	})
}

func (model *Model) handleExecShell() tea.Cmd {
	item := model.GetSelectedItem()
	if item == nil || item.isWorking {
		return nil
	}

	if item.State != "running" {
		return notifications.ShowInfo(item.Name + " is not running")
	}

	// Use proper argument passing - no shell metacharacters possible
	command := exec.Command("docker", "exec", "-it", item.ID, "/bin/sh")
	return tea.ExecProcess(command, func(err error) tea.Msg {
		if err != nil {
			return notifications.AddNotificationMsg{
				Message:  err.Error(),
				Level:    notifications.Error,
				Duration: 10 * time.Second,
			}
		}
		return nil
	})
}

func (model *Model) handleToggleSelection() {
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil && !selectedItem.isWorking {
		model.HandleToggleSelection()

		// Update the visual state of the item
		index := model.GetSelectedIndex()
		if item := model.GetSelectedItem(); item != nil {
			item.isSelected = model.Selections.IsSelected(item.ID)
			model.SetItem(index, *item)
		}
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	// Check if any non-working items exist that are not selected
	items := model.GetItems()
	selectedIDs := model.GetSelectedIDs()

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
			if !item.isWorking && !slices.Contains(selectedIDs, item.ID) {
				model.ToggleSelection(item.ID)
			}
			item.isSelected = model.Selections.IsSelected(item.ID)
			model.SetItem(i, item)
		}
	} else {
		// Deselect all
		for i, item := range items {
			if slices.Contains(selectedIDs, item.ID) {
				model.ToggleSelection(item.ID)
			}
			item.isSelected = false
			model.SetItem(i, item)
		}
	}
}

func (model *Model) handleContainerOperationResult(msg MessageContainerOperationResult) tea.Cmd {
	// Stop spinner for this container
	model.setWorkingState([]string{msg.ID}, false)

	if msg.Error != nil {
		return notifications.ShowError(msg.Error)
	}

	// Show success notification
	var successMsg string
	switch msg.Operation {
	case Remove:
		successMsg = fmt.Sprintf("Container removed: %s", msg.ID[:12])
	case Start:
		successMsg = fmt.Sprintf("Container started: %s", msg.ID[:12])
	case Stop:
		successMsg = fmt.Sprintf("Container stopped: %s", msg.ID[:12])
	case Restart:
		successMsg = fmt.Sprintf("Container restarted: %s", msg.ID[:12])
	case Pause:
		successMsg = fmt.Sprintf("Container paused: %s", msg.ID[:12])
	case Unpause:
		successMsg = fmt.Sprintf("Container unpaused: %s", msg.ID[:12])
	}

	if msg.Operation == Remove {
		// Trigger a refresh to get updated container list
		return tea.Batch(
			notifications.ShowSuccess(successMsg),
			model.Refresh(),
		)
	}

	var newState string
	switch msg.Operation {
	case Pause:
		newState = "paused"
	case Unpause, Start:
		newState = "running"
	case Stop:
		newState = "exited"
	case Restart:
		newState = "running"
	case Remove:
		return nil
	default:
		return nil
	}

	// Update state locally for this container
	items := model.GetItems()
	for i, item := range items {
		if item.ID == msg.ID {
			item.State = newState
			model.SetItem(i, item)
			break
		}
	}

	return notifications.ShowSuccess(successMsg)
}
