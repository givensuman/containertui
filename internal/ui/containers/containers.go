// Package containers defines the containers component.
package containers

import (
	stdcontext "context"
	"fmt"
	"os/exec"
	"slices"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/utils"
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

// MsgPruneComplete is sent when the prune operation completes
type MsgPruneComplete struct {
	SpaceReclaimed uint64
	Err            error
}

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
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab/shift+tab", "switch focus"),
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
	forceRemoveContainer key.Binding
	pruneContainers      key.Binding
	showLogs             key.Binding
	execShell            key.Binding
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	renameContainer      key.Binding
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
		forceRemoveContainer: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "force remove"),
		),
		pruneContainers: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "prune stopped"),
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
		renameContainer: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "rename container"),
		),
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5", "6"),
			key.WithHelp("1-6", "switch tab"),
		),
	}
}

// Model represents the containers component state.
type Model struct {
	components.ResourceView[string, ContainerItem]
	keybindings *keybindings

	inspection         types.ContainerJSON
	detailsKeybindings detailsKeybindings
	detailsPanel       components.DetailsPanel

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

	// Disable filtering for containers tab to preserve color rendering
	resourceView.SplitView.List.SetFilteringEnabled(false)

	// Set detail panel title
	resourceView.SplitView.SetDetailTitle("Inspect")

	configureContainersSplitView(resourceView)

	// Set the custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        containerKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		detailsPanel:       components.NewDetailsPanel(),
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		containerKeybindings.switchTab,
		containerKeybindings.pauseContainer,
		containerKeybindings.unpauseContainer,
		containerKeybindings.startContainer,
		containerKeybindings.stopContainer,
		containerKeybindings.restartContainer,
		containerKeybindings.removeContainer,
		containerKeybindings.forceRemoveContainer,
		containerKeybindings.pruneContainers,
		containerKeybindings.renameContainer,
		containerKeybindings.showLogs,
		containerKeybindings.execShell,
		containerKeybindings.toggleSelection,
		containerKeybindings.toggleSelectionOfAll,
	}

	return model
}

func (model *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.ResourceView.UpdateWindowDimensions(msg)
}

func configureContainersSplitView(resourceView *components.ResourceView[string, ContainerItem]) {
	// Containers intentionally uses list + inspect split without extra stats pane.
}

// refreshWithState refreshes the container list while preserving isWorking and isSelected states
func (model *Model) refreshWithState() tea.Cmd {
	return func() tea.Msg {
		// Fetch fresh container data from Docker
		containers, err := state.GetClient().GetContainers(stdcontext.Background())
		if err != nil {
			return MsgContainersRefreshed{Err: err}
		}

		// Get current items to preserve state
		currentItems := model.GetItems()
		stateMap := make(map[string]struct {
			isWorking  bool
			isSelected bool
			spinner    spinner.Model
		})
		for _, item := range currentItems {
			stateMap[item.ID] = struct {
				isWorking  bool
				isSelected bool
				spinner    spinner.Model
			}{
				isWorking:  item.isWorking,
				isSelected: item.isSelected,
				spinner:    item.spinner,
			}
		}

		// Create new items, preserving state where applicable
		items := make([]ContainerItem, 0, len(containers))
		for _, container := range containers {
			item := ContainerItem{
				Container:  container,
				isSelected: false,
				isWorking:  false,
				spinner:    newSpinner(),
			}
			// Restore state if this container existed before
			if state, exists := stateMap[container.ID]; exists {
				item.isSelected = state.isSelected
				item.isWorking = state.isWorking
				item.spinner = state.spinner
			}
			items = append(items, item)
		}

		return MsgContainersRefreshed{Items: items}
	}
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
		// Refresh the containers list via custom refresh that preserves state
		cmds = append(cmds, model.refreshWithState())

	case MsgContainerOperationResult:
		if cmd := model.handleContainerOperationResult(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case MsgContainersRefreshed:
		if msg.Err == nil {
			listItems := make([]list.Item, len(msg.Items))
			for i, item := range msg.Items {
				listItems[i] = item
			}
			cmds = append(cmds, model.SplitView.List.SetItems(listItems))
		}

	case base.MsgContainerCreated:
		cmds = append(cmds, model.refreshWithState())

	case base.MsgResourceChanged:
		if msg.Resource == base.ResourceContainer {
			cmds = append(cmds, model.refreshWithState())
		}

	case MsgContainerInspection:
		if msg.ID == model.detailsPanel.GetCurrentID() && msg.Err == nil {
			model.inspection = msg.Container
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.detailsPanel.RestoreScrollPosition(model.getViewport())

	case MsgPruneComplete:
		if progressDialog, ok := model.Foreground.(components.ProgressDialog); ok {
			_ = progressDialog.SetPercent(1.0)
		}
		model.CloseOverlay()
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		successMsg := fmt.Sprintf("Pruned stopped containers, freed %s", utils.HumanizeBytes(msg.SpaceReclaimed))
		if msg.SpaceReclaimed == 0 {
			successMsg = "No stopped containers to prune"
		}
		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceContainer,
					Operation: base.OperationPruned,
					IDs:       nil, // Prune affects multiple containers
					Metadata: map[string]any{
						"spaceReclaimed": msg.SpaceReclaimed,
					},
				}
			},
		)

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
				return model, tea.Batch(spinnerCmd, PerformContainerOperations(Remove, containerIDs, false))
			}
			if confirmMsg.Action.Type == "ForceDeleteContainer" {
				containerIDs, ok := confirmMsg.Action.Payload.([]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for ForceDeleteContainer"))
				}
				spinnerCmd := model.setWorkingState(containerIDs, true)
				model.CloseOverlay()
				return model, tea.Batch(
					spinnerCmd,
					PerformContainerOperations(Remove, containerIDs, true),
				)
			}
			if confirmMsg.Action.Type == "RenameContainer" {
				// Extract form values and container ID
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				formValues, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}
				containerID, ok := payload["containerID"].(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid container ID"))
				}

				newName := formValues["New Name"]
				if newName == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("name cannot be empty"))
				}

				model.CloseOverlay()
				return model, model.performRenameContainer(containerID, newName)
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
				if cmd := model.handleRemoveContainers(false); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, model.keybindings.forceRemoveContainer):
				if cmd := model.handleRemoveContainers(true); cmd != nil {
					cmds = append(cmds, cmd)
				}
			case key.Matches(msg, model.keybindings.pruneContainers):
				if !model.hasPrunableContainers() {
					cmds = append(cmds, notifications.ShowSuccess("No stopped containers to prune"))
					break
				}
				cmds = append(cmds, model.handlePruneContainers())
			case key.Matches(msg, model.keybindings.renameContainer):
				model.handleRenameContainer()
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
			if model.IsDetailFocused() {
				if key.Matches(msg, model.detailsKeybindings.ToggleJSON) {
					newFormat, cmd := model.detailsPanel.HandleToggleFormat()
					_ = newFormat // format is tracked internally
					model.refreshInspectionContent()
					cmds = append(cmds, cmd)
				}
				if key.Matches(msg, model.detailsKeybindings.CopyOutput) {
					cmd := model.detailsPanel.HandleCopyToClipboard(model.inspection)
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// 6. Update Detail Content if selection changes
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil {
		if selectedItem.ID != model.detailsPanel.GetCurrentID() {
			// SetCurrentID will save scroll position for previous ID
			model.detailsPanel.SetCurrentID(selectedItem.ID, model.getViewport())

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
	return lipgloss.NewStyle().MarginTop(1).Render(model.ResourceView.View())
}

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model Model) ShortHelp() []key.Binding {
	// If detail pane is focused, show detail keybindings
	if model.IsDetailFocused() {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.ToggleJSON,
			model.detailsKeybindings.CopyOutput,
			model.detailsKeybindings.Switch,
		}
	}

	if model.IsExtraFocused() {
		return model.ResourceView.ShortHelp()
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

// refreshInspectionContent regenerates and sets the inspection content
func (model *Model) refreshInspectionContent() {
	if model.inspection.ID == "" {
		return
	}

	// Use DetailsPanel to get the current format
	format := model.detailsPanel.GetFormatForDisplay()

	content := builders.BuildContainerPanel(model.inspection, model.GetContentWidth(), false, format)
	model.SetContent(content)
}

func (model Model) FullHelp() [][]key.Binding {
	// If detail pane is focused, show detail keybindings
	if model.IsDetailFocused() {
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

	if model.IsExtraFocused() {
		return model.ResourceView.FullHelp()
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
		return tea.Batch(spinnerCmd, PerformContainerOperations(Pause, selectedIDs, false))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Pause, selectedItem.ID, false))
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
		return tea.Batch(spinnerCmd, PerformContainerOperations(Unpause, selectedIDs, false))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Unpause, selectedItem.ID, false))
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
		return tea.Batch(spinnerCmd, PerformContainerOperations(Start, selectedIDs, false))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Start, selectedItem.ID, false))
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
		return tea.Batch(spinnerCmd, PerformContainerOperations(Stop, selectedIDs, false))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Stop, selectedItem.ID, false))
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
		return tea.Batch(spinnerCmd, PerformContainerOperations(Restart, selectedIDs, false))
	} else {
		selectedItem := model.GetSelectedItem()
		if selectedItem != nil && !selectedItem.isWorking {
			spinnerCmd := model.setWorkingState([]string{selectedItem.ID}, true)
			return tea.Batch(spinnerCmd, PerformContainerOperation(Restart, selectedItem.ID, false))
		}
	}
	return nil
}

func (model *Model) handleRemoveContainers(force bool) tea.Cmd {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		if model.anySelectedWorking() {
			return nil
		}

		// Build container names list for display
		var containerNames []string
		items := model.GetItems()
		for _, item := range items {
			if slices.Contains(selectedIDs, item.ID) {
				containerNames = append(containerNames, item.Name)
			}
		}

		// If force, show confirmation with warning
		if force {
			message := fmt.Sprintf("⚠️  FORCE DELETE %d container(s)?", len(containerNames))
			if len(containerNames) <= 5 {
				message += "\n\n"
				for _, name := range containerNames {
					message += fmt.Sprintf("• %s\n", name)
				}
			}
			message += "\nThis will forcefully stop and remove containers."

			confirmDialog := components.NewDialog(
				message,
				[]components.DialogButton{
					{Label: "Cancel"},
					{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteContainer", Payload: selectedIDs}},
				},
			)
			model.SetOverlay(confirmDialog)
			return nil
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
			if force {
				// Force delete single container with confirmation
				confirmDialog := components.NewDialog(
					fmt.Sprintf("⚠️  FORCE DELETE container %s?\n\nThis will forcefully stop and remove the container.", item.Name),
					[]components.DialogButton{
						{Label: "Cancel"},
						{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteContainer", Payload: []string{item.ID}}},
					},
				)
				model.SetOverlay(confirmDialog)
				return nil
			}

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
	return nil
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

func (model *Model) handleContainerOperationResult(msg MsgContainerOperationResult) tea.Cmd {
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

func isPruneEligibleContainerState(state string) bool {
	switch state {
	case "exited", "created", "dead":
		return true
	default:
		return false
	}
}

func (model Model) hasPrunableContainers() bool {
	for _, item := range model.GetItems() {
		if isPruneEligibleContainerState(item.State) {
			return true
		}
	}

	return false
}

// handlePruneContainers prunes all stopped containers
func (model *Model) handlePruneContainers() tea.Cmd {
	progressDialog := components.NewProgressDialogWithBar("Pruning stopped containers")
	progressDialog.EnableAutoAdvance(0.95, 0.04)
	progressDialog.SetStatus("Discovering stopped containers to prune...")
	model.SetOverlay(progressDialog)

	// Start async prune operation
	return func() tea.Msg {
		ctx := stdcontext.Background()
		spaceReclaimed, err := state.GetClient().PruneContainers(ctx)
		return MsgPruneComplete{
			SpaceReclaimed: spaceReclaimed,
			Err:            err,
		}
	}
}

// handleRenameContainer shows a dialog to rename the selected container
func (model *Model) handleRenameContainer() {
	item := model.GetSelectedItem()
	if item == nil || item.isWorking {
		return
	}

	fields := []components.FormField{
		{
			Label:       "New Name",
			Placeholder: "my-container",
			Required:    true,
		},
	}

	metadata := map[string]any{
		"containerID": item.ID,
	}

	dialog := components.NewFormDialog(
		"Rename Container",
		fields,
		base.SmartDialogAction{Type: "RenameContainer"},
		metadata,
	)

	model.SetOverlay(dialog)
}

// performRenameContainer renames a container
func (model *Model) performRenameContainer(containerID, newName string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		err := state.GetClient().RenameContainer(ctx, containerID, newName)
		if err != nil {
			return notifications.ShowError(err)
		}
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Renamed to: %s", newName)),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceContainer,
					Operation: base.OperationUpdated,
					IDs:       []string{containerID},
				}
			},
		)
	}
}
