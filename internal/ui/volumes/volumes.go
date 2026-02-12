// Package volumes defines the volumes component.
package volumes

import (
	stdcontext "context"
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/docker/docker/api/types/volume"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

// MsgVolumeInspection contains the inspection data for a volume.
type MsgVolumeInspection struct {
	Name   string
	Volume volume.Volume
	Err    error
}

// MsgRestoreScroll is sent to restore scroll position after content is set.
type MsgRestoreScroll struct{}

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
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	remove               key.Binding
	forceRemove          key.Binding
	pruneVolumes         key.Binding
	createVolume         key.Binding
	switchTab            key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		toggleSelection: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle selection"),
		),
		toggleSelectionOfAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle selection of all"),
		),
		remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove"),
		),
		forceRemove: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "force remove"),
		),
		pruneVolumes: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prune unused"),
		),
		createVolume: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "create volume"),
		),
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "switch tab"),
		),
	}
}

// Model represents the volumes component state.
type Model struct {
	components.ResourceView[string, VolumeItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	inspection         volume.Volume
	detailsPanel       components.DetailsPanel
}

func New() Model {
	volumeKeybindings := newKeybindings()

	fetchVolumes := func() ([]VolumeItem, error) {
		volumeList, err := state.GetClient().GetVolumes(stdcontext.Background())
		if err != nil {
			return nil, err
		}
		items := make([]VolumeItem, 0, len(volumeList))
		for _, volume := range volumeList {
			items = append(items, VolumeItem{Volume: volume})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, VolumeItem](
		"Volumes",
		fetchVolumes,
		func(item VolumeItem) string { return item.Volume.Name },
		func(item VolumeItem) string { return item.Title() },
		func(w, h int) {
			// Window resize handled by base component
		},
	)

	// Add extra pane below detail pane
	extraPane := components.NewViewportPane()
	extraPane.SetContent("")                            // Will be populated when a volume is selected
	resourceView.SplitView.SetExtraPane(extraPane, 0.3) // 30% of height

	// Set titles for the panes
	resourceView.SplitView.SetDetailTitle("Inspect")
	resourceView.SplitView.SetExtraTitle("Used By")

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        volumeKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		inspection:         volume.Volume{},
		detailsPanel:       components.NewDetailsPanel(),
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		volumeKeybindings.toggleSelection,
		volumeKeybindings.toggleSelectionOfAll,
		volumeKeybindings.remove,
		volumeKeybindings.forceRemove,
		volumeKeybindings.pruneVolumes,
		volumeKeybindings.createVolume,
	}

	return model
}

func (model Model) Init() tea.Cmd {
	return model.ResourceView.Init()
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// 1. Try standard ResourceView updates first (resizing, dialog closing, basic navigation)
	updatedView, cmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// 2. Handle Messages
	switch msg := msg.(type) {
	case MsgVolumeInspection:
		if msg.Name == model.detailsPanel.GetCurrentID() && msg.Err == nil {
			model.inspection = msg.Volume
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.detailsPanel.RestoreScrollPosition(model.getViewport())
		return model, nil
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteVolume" {
				volumeName, ok := confirmMsg.Action.Payload.(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for DeleteVolume"))
				}
				err := state.GetClient().RemoveVolume(stdcontext.Background(), volumeName)
				if err == nil {
					// Close the overlay and refresh list
					model.CloseOverlay()
					return model, tea.Batch(
						notifications.ShowSuccess(fmt.Sprintf("Volume removed: %s", volumeName)),
						model.Refresh(),
					)
				} else {
					// Show error notification
					model.CloseOverlay()
					return model, notifications.ShowError(err)
				}
			} else if confirmMsg.Action.Type == "ForceDeleteVolume" {
				volumeName, ok := confirmMsg.Action.Payload.(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for ForceDeleteVolume"))
				}
				err := state.GetClient().RemoveVolume(stdcontext.Background(), volumeName)
				model.CloseOverlay()
				if err != nil {
					return model, notifications.ShowError(fmt.Errorf("failed to force delete volume: %w", err))
				}
				return model, tea.Batch(
					notifications.ShowSuccess(fmt.Sprintf("Force deleted volume: %s", volumeName)),
					model.Refresh(),
				)
			} else if confirmMsg.Action.Type == "CreateVolumeAction" {
				// Extract form values
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

				name := formValues["0"]   // First field
				driver := formValues["1"] // Second field
				labels := formValues["2"] // Third field

				if name == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("volume name is required"))
				}

				// Parse labels if provided
				var labelsMap map[string]string
				if labels != "" {
					labelsMap = make(map[string]string)
					pairs := strings.Split(labels, ",")
					for _, pair := range pairs {
						kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
						if len(kv) == 2 {
							labelsMap[kv[0]] = kv[1]
						}
					}
				}

				model.CloseOverlay()
				return model, model.performCreateVolume(name, driver, labelsMap)
			}
			model.CloseOverlay()
			return model, nil
		}

		// Let ResourceView handle forwarding to overlay
		return model, tea.Batch(cmds...)
	}

	// 3. Main View Logic
	if model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.IsFiltering() {
				break
			}

			switch {
			case key.Matches(msg, model.keybindings.switchTab):
				return model, tea.Batch(cmds...) // Handled by parent

			case key.Matches(msg, model.keybindings.toggleSelection):
				model.handleToggleSelection()
				return model, nil

			case key.Matches(msg, model.keybindings.toggleSelectionOfAll):
				model.handleToggleSelectionOfAll()
				return model, nil

			case key.Matches(msg, model.keybindings.remove):
				model.handleRemove(false)
				return model, nil

			case key.Matches(msg, model.keybindings.forceRemove):
				model.handleRemove(true)
				return model, nil

			case key.Matches(msg, model.keybindings.pruneVolumes):
				if cmd := model.handlePruneVolumes(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)

			case key.Matches(msg, model.keybindings.createVolume):
				model.handleCreateVolume()
				return model, nil
			}
		}
	} else {
		// Detail or extra pane is focused
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			// Only handle these actions when detail pane is focused (not extra)
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

	// Update Detail Content
	detailCmd := model.updateDetailContent()
	if detailCmd != nil {
		cmds = append(cmds, detailCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model Model) handleToggleSelection() {
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil {
		model.ToggleSelection(selectedItem.Volume.Name)

		// Update the visual state of the item
		index := model.GetSelectedIndex()
		selectedItem.isSelected = !selectedItem.isSelected
		model.SetItem(index, *selectedItem)
	}
}

func (model Model) handleToggleSelectionOfAll() {
	// Check if we need to select all or deselect all
	items := model.GetItems()
	selectedIDs := model.GetSelectedIDs()

	shouldSelectAll := false
	for _, item := range items {
		if !slices.Contains(selectedIDs, item.Volume.Name) {
			shouldSelectAll = true
			break
		}
	}

	if shouldSelectAll {
		// Select all
		for i, item := range items {
			if !slices.Contains(selectedIDs, item.Volume.Name) {
				model.ToggleSelection(item.Volume.Name)
			}
			// Visual update
			item.isSelected = true
			model.SetItem(i, item)
		}
	} else {
		// Deselect all
		for i, item := range items {
			model.ToggleSelection(item.Volume.Name)
			// Visual update
			item.isSelected = false
			model.SetItem(i, item)
		}
	}
}

func (model Model) handleRemove(force bool) {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	if force {
		// Force delete - show confirmation dialog first
		confirmationDialog := components.NewDialog(
			fmt.Sprintf("Force delete volume %s?\n\nThis will delete the volume even if containers are using it.", selectedItem.Volume.Name),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteVolume", Payload: selectedItem.Volume.Name}},
			},
		)
		model.SetOverlay(confirmationDialog)
		return
	}

	containersUsingVolume, err := state.GetClient().GetContainersUsingVolume(stdcontext.Background(), selectedItem.Volume.Name)
	if err != nil {
		// If we can't check usage, show error and don't proceed with deletion
		errorDialog := components.NewDialog(
			fmt.Sprintf("Failed to check volume usage: %v\nCannot safely delete volume.", err),
			[]components.DialogButton{
				{Label: "OK"},
			},
		)
		model.SetOverlay(errorDialog)
		return
	}
	if len(containersUsingVolume) > 0 {
		warningDialog := components.NewDialog(
			fmt.Sprintf("Volume %s is used by %d containers (%v).\nCannot delete.",
				selectedItem.Volume.Name, len(containersUsingVolume), containersUsingVolume),
			[]components.DialogButton{
				{Label: "OK"},
			},
		)
		model.SetOverlay(warningDialog)
	} else {
		confirmationDialog := components.NewDialog(
			fmt.Sprintf("Are you sure you want to delete volume %s?", selectedItem.Volume.Name),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Delete", Action: base.SmartDialogAction{
					Type:    "DeleteVolume",
					Payload: selectedItem.Volume.Name,
				}},
			},
		)
		model.SetOverlay(confirmationDialog)
	}
}

func (model Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No volume selected."))
		model.SetExtraContent("") // Clear extra pane when no volume selected
		return nil
	}

	volumeName := selectedItem.Volume.Name
	currentID := model.detailsPanel.GetCurrentID()

	// If we've switched to a different volume, OR we don't have inspection data yet, fetch it
	if volumeName != currentID || model.inspection.Name == "" {
		// SetCurrentID will save scroll position for previous ID
		model.detailsPanel.SetCurrentID(volumeName, model.getViewport())

		// Fetch inspection data asynchronously
		return func() tea.Msg {
			volumeInfo, err := state.GetClient().InspectVolume(stdcontext.Background(), volumeName)
			return MsgVolumeInspection{Name: volumeName, Volume: volumeInfo, Err: err}
		}
	}

	return nil
}

// getViewport returns the viewport from the detail pane if available
func (model Model) getViewport() *viewport.Model {
	if vp, ok := model.SplitView.Detail.(*components.ViewportPane); ok {
		return &vp.Viewport
	}
	return nil
}

// refreshInspectionContent refreshes the detail content with current inspection data
func (model Model) refreshInspectionContent() {
	// Use DetailsPanel to get the current format
	format := model.detailsPanel.GetFormatForDisplay()

	// Build content with current format
	content := builders.BuildVolumePanel(model.inspection, model.GetContentWidth(), format)
	model.SetContent(content)

	// Update "Used By" panel
	model.updateUsedByPanel()
}

// updateUsedByPanel updates the extra pane with containers using this volume
func (model Model) updateUsedByPanel() {
	if model.inspection.Name == "" {
		model.SetExtraContent("")
		return
	}

	// Fetch containers using this volume
	usedBy, err := state.GetClient().GetContainersUsingVolume(stdcontext.Background(), model.inspection.Name)
	if err != nil {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	if len(usedBy) == 0 {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No containers using this volume"))
		return
	}

	// Build a formatted list of containers
	var output strings.Builder
	for i, containerName := range usedBy {
		if i > 0 {
			output.WriteString("\n")
		}
		output.WriteString(fmt.Sprintf("• %s", containerName))
	}

	model.SetExtraContent(output.String())
}

func (model Model) View() string {
	return model.ResourceView.View()
}

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model Model) ShortHelp() []key.Binding {
	// If detail or extra pane is focused, show detail keybindings
	if model.IsDetailFocused() {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
			model.detailsKeybindings.ToggleJSON,
			model.detailsKeybindings.CopyOutput,
		}
	} else if model.IsExtraFocused() {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
		}
	}
	return model.ResourceView.ShortHelp()
}

func (model Model) FullHelp() [][]key.Binding {
	// If detail or extra pane is focused, show detail keybindings
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
	} else if model.IsExtraFocused() {
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
		}
	}
	return model.ResourceView.FullHelp()
}

// humanizeBytes converts bytes to human-readable format
func humanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handlePruneVolumes prunes unused volumes
func (model Model) handlePruneVolumes() tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		spaceReclaimed, err := state.GetClient().PruneVolumes(ctx)
		if err != nil {
			return notifications.ShowError(err)
		}
		msg := fmt.Sprintf("Pruned unused volumes, freed %s", humanizeBytes(spaceReclaimed))
		return tea.Batch(
			notifications.ShowSuccess(msg),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceVolume,
					Operation: base.OperationPruned,
					IDs:       nil,
					Metadata: map[string]any{
						"spaceReclaimed": spaceReclaimed,
					},
				}
			},
		)
	}
}

// handleCreateVolume shows dialog to create a volume
func (model Model) handleCreateVolume() {
	fields := []components.FormField{
		{
			Label:       "Name",
			Placeholder: "my-volume",
			Required:    true,
		},
		{
			Label:       "Driver",
			Placeholder: "local",
			Required:    false,
		},
		{
			Label:       "Labels",
			Placeholder: "KEY=value,FOO=bar",
			Required:    false,
		},
	}

	dialog := components.NewFormDialog(
		"Create Volume",
		fields,
		base.SmartDialogAction{Type: "CreateVolumeAction"},
		nil,
	)

	model.SetOverlay(dialog)
}

// performCreateVolume creates a volume
func (model Model) performCreateVolume(name, driver string, labels map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()

		// Use "local" as default driver if not specified
		if driver == "" {
			driver = "local"
		}

		volumeID, err := state.GetClient().CreateVolume(ctx, name, driver, labels)
		if err != nil {
			return notifications.ShowError(fmt.Errorf("failed to create volume: %w", err))
		}
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Created volume: %s", volumeID)),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceVolume,
					Operation: base.OperationCreated,
					IDs:       []string{volumeID},
				}
			},
		)
	}
}
