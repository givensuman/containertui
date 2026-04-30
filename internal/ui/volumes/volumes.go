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
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/safety"
	"github.com/givensuman/containertui/internal/ui/utils"
)

// MsgVolumeInspection contains the inspection data for a volume.
type MsgVolumeInspection struct {
	Name   string
	Volume backend.VolumeDetail
	Err    error
}

// MsgPruneComplete is sent when the prune operation completes
type MsgPruneComplete struct {
	SpaceReclaimed uint64
	Err            error
}

// MsgCreateVolumeComplete indicates volume creation finished.
type MsgCreateVolumeComplete struct {
	VolumeName string
	Err        error
}

type MsgAttachVolumeComplete struct {
	VolumeName  string
	ContainerID string
	Err         error
}

type MsgDetachVolumeComplete struct {
	VolumeName  string
	ContainerID string
	Err         error
}

type keybindings struct {
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	remove               key.Binding
	pruneVolumes         key.Binding
	createVolume         key.Binding
	attachVolume         key.Binding
	detachVolume         key.Binding
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
		pruneVolumes: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prune unused"),
		),
		createVolume: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "create volume"),
		),
		attachVolume: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "attach volume"),
		),
		detachVolume: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "detach volume"),
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
	detailsKeybindings components.DetailsKeybindings
	inspection         backend.VolumeDetail
	detailsPanel       components.DetailsPanel
}

func New() Model {
	volumeKeybindings := newKeybindings()

	fetchVolumes := func() ([]VolumeItem, error) {
		volumeList, err := state.GetBackend().ListVolumes(stdcontext.Background())
		if err != nil {
			return nil, err
		}

		// Get volume usage map (single API call for all volumes)
		mountedVolumes, err := state.GetBackend().GetAllVolumeUsage(stdcontext.Background())
		if err != nil {
			return nil, err
		}

		items := make([]VolumeItem, 0, len(volumeList))
		for _, volume := range volumeList {
			// Check if this volume is in the mounted map
			isMounted := mountedVolumes[volume.Name]

			items = append(items, VolumeItem{
				Volume:    volume,
				IsMounted: isMounted,
			})
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
		detailsKeybindings: components.NewDetailsKeybindings(),
		inspection:         backend.VolumeDetail{},
		detailsPanel:       components.NewDetailsPanel(),
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		volumeKeybindings.switchTab,
		volumeKeybindings.remove,
		volumeKeybindings.pruneVolumes,
		volumeKeybindings.createVolume,
		volumeKeybindings.attachVolume,
		volumeKeybindings.detachVolume,
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
			cmds = append(cmds, func() tea.Msg { return base.MsgRestoreScroll{} })
		}

	case base.MsgRestoreScroll:
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
		successMsg := fmt.Sprintf("Pruned unused volumes, freed %s", utils.HumanizeBytes(msg.SpaceReclaimed))
		if msg.SpaceReclaimed == 0 {
			successMsg = "No unused volumes to prune"
		}
		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceVolume,
					Operation: base.OperationPruned,
					IDs:       nil,
					Metadata: map[string]any{
						"spaceReclaimed": msg.SpaceReclaimed,
					},
				}
			},
		)

	case MsgCreateVolumeComplete:
		return model.handleCreateVolumeComplete(msg)

	case MsgAttachVolumeComplete:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Attached volume %s to container %s", msg.VolumeName, msg.ContainerID)),
			model.Refresh(),
		)

	case MsgDetachVolumeComplete:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Detached volume %s from container %s", msg.VolumeName, msg.ContainerID)),
			model.Refresh(),
		)
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			if confirmMsg.Action.Type == "PruneVolumes" {
				model.CloseOverlay()
				if cmd := model.handlePruneVolumes(); cmd != nil {
					return model, cmd
				}
				return model, nil
			}
			if confirmMsg.Action.Type == "DeleteVolume" {
				volumeName, ok := confirmMsg.Action.Payload.(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for DeleteVolume"))
				}
				err := state.GetBackend().RemoveVolume(stdcontext.Background(), volumeName)
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
				err := state.GetBackend().RemoveVolume(stdcontext.Background(), volumeName)
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

				name := formValues["Name"]
				driver := formValues["Driver"]
				labels := formValues["Labels"]

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
			} else if confirmMsg.Action.Type == "AttachVolumeAction" {
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				values, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}
				containerID := strings.TrimSpace(values["Container ID"])
				volumeName, _ := payload["volumeName"].(string)
				if containerID == "" || volumeName == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("container ID and volume name are required"))
				}
				model.CloseOverlay()
				return model, model.performAttachVolume(volumeName, containerID)
			} else if confirmMsg.Action.Type == "DetachVolumeAction" {
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				values, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}
				containerID := strings.TrimSpace(values["Container ID"])
				volumeName, _ := payload["volumeName"].(string)
				if containerID == "" || volumeName == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("container ID and volume name are required"))
				}
				model.CloseOverlay()
				return model, model.performDetachVolume(volumeName, containerID)
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

			case key.Matches(msg, model.keybindings.remove):
				model.handleRemove()
				return model, nil

			case key.Matches(msg, model.keybindings.pruneVolumes):
				if !model.hasPrunableVolumes() {
					return model, notifications.ShowSuccess("No unused volumes to prune")
				}
				model.showPruneVolumesConfirmation()
				return model, tea.Batch(cmds...)

			case key.Matches(msg, model.keybindings.createVolume):
				model = model.withCreateVolumeDialog()
				return model, nil
			case key.Matches(msg, model.keybindings.attachVolume):
				model = model.withAttachVolumeDialog()
				return model, nil
			case key.Matches(msg, model.keybindings.detachVolume):
				model = model.withDetachVolumeDialog()
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

func (model Model) handleCreateVolumeComplete(msg MsgCreateVolumeComplete) (Model, tea.Cmd) {
	if msg.Err != nil {
		return model, notifications.ShowError(msg.Err)
	}

	return model, tea.Batch(
		notifications.ShowSuccess(fmt.Sprintf("Created volume: %s", msg.VolumeName)),
		model.Refresh(),
		func() tea.Msg {
			return base.MsgResourceChanged{
				Resource:  base.ResourceVolume,
				Operation: base.OperationCreated,
				IDs:       []string{msg.VolumeName},
			}
		},
	)
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

func (model Model) handleRemove() {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	containersUsingVolume, err := state.GetBackend().GetContainersUsingVolume(stdcontext.Background(), selectedItem.Volume.Name)
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
		confirmationDialog := components.NewDialog(
			safety.ForceDeleteInUseConfirmation("Volume", selectedItem.Volume.Name, len(containersUsingVolume), containersUsingVolume),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteVolume", Payload: selectedItem.Volume.Name}},
			},
		)
		model.SetOverlay(confirmationDialog)
	} else {
		confirmationDialog := components.NewDialog(
			safety.DeleteConfirmation("volume", selectedItem.Volume.Name),
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

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No volume selected"))
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
			volumeInfo, err := state.GetBackend().InspectVolume(stdcontext.Background(), volumeName)
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
func (model *Model) refreshInspectionContent() {
	// Use DetailsPanel to get the current format
	format := model.detailsPanel.GetFormatForDisplay()

	// Build content with current format
	content := builders.BuildVolumePanel(model.inspection, model.GetContentWidth(), format)
	model.SetContent(content)

	// Update "Used By" panel
	model.updateUsedByPanel()
}

// updateUsedByPanel updates the extra pane with containers using this volume
func (model *Model) updateUsedByPanel() {
	if model.inspection.Name == "" {
		model.SetExtraContent("")
		return
	}

	// Fetch containers using this volume
	usedBy, err := state.GetBackend().GetContainersUsingVolume(stdcontext.Background(), model.inspection.Name)
	if err != nil {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	model.SetExtraContent(buildVolumeUsageContent(model.inspection, usedBy))
}

func buildVolumeUsageContent(inspection backend.VolumeDetail, usedBy []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Driver: %s\n", inspection.Driver))
	b.WriteString(fmt.Sprintf("Mountpoint: %s\n\n", inspection.Mountpoint))

	b.WriteString("Dependency Trace\n")
	if len(usedBy) == 0 {
		b.WriteString("No containers currently depend on this volume.")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%d containers depend on this volume:\n", len(usedBy)))
	for _, name := range usedBy {
		b.WriteString("• ")
		b.WriteString(name)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
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

func (model Model) hasPrunableVolumes() bool {
	for _, item := range model.GetItems() {
		if !item.IsMounted {
			return true
		}
	}

	return false
}

// handlePruneVolumes prunes unused volumes
func (model *Model) handlePruneVolumes() tea.Cmd {
	progressDialog := components.NewProgressDialogWithBar("Pruning unused volumes")
	progressDialog.EnableAutoAdvance(0.95, 0.04)
	progressDialog.SetStatus("Discovering unused volumes to prune...")
	model.SetOverlay(progressDialog)

	// Start async prune operation
	return func() tea.Msg {
		ctx := stdcontext.Background()
		spaceReclaimed, err := state.GetBackend().PruneVolumes(ctx)
		return MsgPruneComplete{
			SpaceReclaimed: spaceReclaimed,
			Err:            err,
		}
	}
}

func (model *Model) showPruneVolumesConfirmation() {
	candidates := model.pruneVolumeCandidates()
	if len(candidates) == 0 {
		return
	}

	samples := candidates
	if len(samples) > 3 {
		samples = samples[:3]
	}

	confirmDialog := components.NewDialog(
		safety.PruneConfirmation("volumes", len(candidates), samples),
		[]components.DialogButton{
			{Label: "Cancel"},
			{Label: "Prune", Action: base.SmartDialogAction{Type: "PruneVolumes"}},
		},
	)
	model.SetOverlay(confirmDialog)
}

func (model Model) pruneVolumeCandidates() []string {
	items := model.GetItems()
	candidates := make([]string, 0, len(items))
	for _, item := range items {
		if !item.IsMounted {
			candidates = append(candidates, item.Volume.Name)
		}
	}

	return candidates
}

// withCreateVolumeDialog returns model with create-volume dialog shown.
func (model Model) withCreateVolumeDialog() Model {
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
	return model
}

// performCreateVolume creates a volume
func (model Model) performCreateVolume(name, driver string, labels map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()

		// Use "local" as default driver if not specified
		if driver == "" {
			driver = "local"
		}

		volumeName, err := state.GetBackend().CreateVolume(ctx, name, driver, labels)
		if err != nil {
			return MsgCreateVolumeComplete{Err: fmt.Errorf("failed to create volume: %w", err)}
		}

		return MsgCreateVolumeComplete{VolumeName: volumeName}
	}
}

func (model Model) withAttachVolumeDialog() Model {
	selected := model.GetSelectedItem()
	if selected == nil {
		return model
	}

	fields := []components.FormField{{
		Label:       "Container ID",
		Placeholder: "container-id-or-name",
		Required:    true,
	}}

	payload := map[string]any{"volumeName": selected.Volume.Name}
	dialog := components.NewFormDialog(
		"Attach Volume",
		fields,
		base.SmartDialogAction{Type: "AttachVolumeAction", Payload: payload},
		nil,
	)
	model.SetOverlay(dialog)
	return model
}

func (model Model) withDetachVolumeDialog() Model {
	selected := model.GetSelectedItem()
	if selected == nil {
		return model
	}

	fields := []components.FormField{{
		Label:       "Container ID",
		Placeholder: "container-id-or-name",
		Required:    true,
	}}

	payload := map[string]any{"volumeName": selected.Volume.Name}
	dialog := components.NewFormDialog(
		"Detach Volume",
		fields,
		base.SmartDialogAction{Type: "DetachVolumeAction", Payload: payload},
		nil,
	)
	model.SetOverlay(dialog)
	return model
}

func (model Model) performAttachVolume(volumeName, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := fmt.Errorf("attach volume is not yet supported by Docker API directly; use bind mounts or recreate container")
		return MsgAttachVolumeComplete{VolumeName: volumeName, ContainerID: containerID, Err: err}
	}
}

func (model Model) performDetachVolume(volumeName, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := fmt.Errorf("detach volume is not yet supported by Docker API directly; stop/recreate container without the mount")
		return MsgDetachVolumeComplete{VolumeName: volumeName, ContainerID: containerID, Err: err}
	}
}
