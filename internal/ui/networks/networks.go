// Package networks defines the networks component.
package networks

import (
	stdcontext "context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/safety"
)

// MsgNetworkInspection contains the inspection data for a network.
type MsgNetworkInspection struct {
	ID      string
	Network types.NetworkResource
	Err     error
}

// MsgRestoreScroll is sent to restore scroll position after content is set.
type MsgRestoreScroll struct{}

// MsgPruneComplete is sent when the prune operation completes
type MsgPruneComplete struct {
	NetworksDeleted int
	Err             error
}

// MsgCreateNetworkComplete indicates network creation finished.
type MsgCreateNetworkComplete struct {
	NetworkID string
	Err       error
}

type MsgAttachContainerComplete struct {
	NetworkID   string
	ContainerID string
	Err         error
}

type MsgDetachContainerComplete struct {
	NetworkID   string
	ContainerID string
	Err         error
}

// isSystemNetwork returns true if the network is a predefined system network
func isSystemNetwork(name string) bool {
	systemNetworks := []string{"bridge", "host", "none", "podman"}
	for _, sysNet := range systemNetworks {
		if name == sysNet {
			return true
		}
	}
	return false
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
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	remove               key.Binding
	pruneNetworks        key.Binding
	createNetwork        key.Binding
	attachContainer      key.Binding
	detachContainer      key.Binding
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
		pruneNetworks: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prune unused"),
		),
		createNetwork: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "create network"),
		),
		attachContainer: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "attach container"),
		),
		detachContainer: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "detach container"),
		),
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "switch tab"),
		),
	}
}

// Model represents the networks component state.
type Model struct {
	components.ResourceView[string, NetworkItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	inspection         types.NetworkResource
	detailsPanel       components.DetailsPanel
}

func New() Model {
	networkKeybindings := newKeybindings()

	fetchNetworks := func() ([]NetworkItem, error) {
		networkList, err := state.GetClient().GetNetworks(stdcontext.Background())
		if err != nil {
			return nil, err
		}

		// Get network usage map (single API call for all networks)
		activeNetworks, err := state.GetClient().GetAllNetworkUsage(stdcontext.Background())
		if err != nil {
			return nil, err
		}

		items := make([]NetworkItem, 0, len(networkList))
		for _, network := range networkList {
			// Check if this network is in the active map (by name or ID)
			isActive := activeNetworks[network.Name] || activeNetworks[network.ID]

			items = append(items, NetworkItem{
				Network:  network,
				IsActive: isActive,
			})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, NetworkItem](
		"Networks",
		fetchNetworks,
		func(item NetworkItem) string { return item.Network.ID },
		func(item NetworkItem) string { return item.Title() },
		func(w, h int) {
			// Window resize handled by base component
		},
	)

	// Add extra pane below detail pane
	extraPane := components.NewViewportPane()
	extraPane.SetContent("")                            // Will be populated when a network is selected
	resourceView.SplitView.SetExtraPane(extraPane, 0.3) // 30% of height

	// Set titles for the panes
	resourceView.SplitView.SetDetailTitle("Inspect")
	resourceView.SplitView.SetExtraTitle("Used By")

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        networkKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		inspection:         types.NetworkResource{},
		detailsPanel:       components.NewDetailsPanel(),
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		networkKeybindings.switchTab,
		networkKeybindings.remove,
		networkKeybindings.pruneNetworks,
		networkKeybindings.createNetwork,
		networkKeybindings.attachContainer,
		networkKeybindings.detachContainer,
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
	case MsgNetworkInspection:
		if msg.ID == model.detailsPanel.GetCurrentID() && msg.Err == nil {
			model.inspection = msg.Network
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

		successMsg := "No unused networks to prune"
		if msg.NetworksDeleted == 1 {
			successMsg = "Pruned 1 unused network"
		} else if msg.NetworksDeleted > 1 {
			successMsg = fmt.Sprintf("Pruned %d unused networks", msg.NetworksDeleted)
		}

		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceNetwork,
					Operation: base.OperationPruned,
					IDs:       nil,
				}
			},
		)

	case MsgCreateNetworkComplete:
		return model.handleCreateNetworkComplete(msg)

	case MsgAttachContainerComplete:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Attached container %s to network", shortID(msg.ContainerID))),
			model.Refresh(),
		)

	case MsgDetachContainerComplete:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Detached container %s from network", shortID(msg.ContainerID))),
			model.Refresh(),
		)
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			if confirmMsg.Action.Type == "PruneNetworks" {
				model.CloseOverlay()
				if cmd := model.handlePruneNetworks(); cmd != nil {
					return model, cmd
				}
				return model, nil
			}
			if confirmMsg.Action.Type == "DeleteNetwork" {
				networkID, ok := confirmMsg.Action.Payload.(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for DeleteNetwork"))
				}
				err := state.GetClient().RemoveNetwork(stdcontext.Background(), networkID)
				if err == nil {
					// Close the overlay and refresh list
					model.CloseOverlay()
					return model, tea.Batch(
						notifications.ShowSuccess(fmt.Sprintf("Network removed: %s", networkID[:12])),
						model.Refresh(),
					)
				} else {
					// Show error notification
					model.CloseOverlay()
					return model, notifications.ShowError(err)
				}
			} else if confirmMsg.Action.Type == "ForceDeleteNetwork" {
				networkID, ok := confirmMsg.Action.Payload.(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type for ForceDeleteNetwork"))
				}
				err := state.GetClient().RemoveNetwork(stdcontext.Background(), networkID)
				model.CloseOverlay()
				if err != nil {
					return model, notifications.ShowError(fmt.Errorf("failed to force delete network: %w", err))
				}
				return model, tea.Batch(
					notifications.ShowSuccess(fmt.Sprintf("Force deleted network: %s", networkID[:12])),
					model.Refresh(),
				)
			} else if confirmMsg.Action.Type == "CreateNetworkAction" {
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
				subnet := formValues["Subnet (CIDR)"]
				gateway := formValues["Gateway"]
				ipv6 := formValues["Enable IPv6"]
				labels := formValues["Labels"]

				if name == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("network name is required"))
				}

				// Parse IPv6 boolean
				enableIPv6 := (strings.ToLower(ipv6) == "yes" || strings.ToLower(ipv6) == "true")

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
				return model, model.performCreateNetwork(name, driver, subnet, gateway, enableIPv6, labelsMap)
			} else if confirmMsg.Action.Type == "AttachContainerAction" {
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
				networkID, _ := payload["networkID"].(string)
				if containerID == "" || networkID == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("container ID and network ID are required"))
				}
				model.CloseOverlay()
				return model, model.performAttachContainer(networkID, containerID)
			} else if confirmMsg.Action.Type == "DetachContainerAction" {
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
				networkID, _ := payload["networkID"].(string)
				if containerID == "" || networkID == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("container ID and network ID are required"))
				}
				model.CloseOverlay()
				return model, model.performDetachContainer(networkID, containerID)
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
				return model, model.handleRemove()

			case key.Matches(msg, model.keybindings.pruneNetworks):
				if !model.hasPrunableNetworks() {
					return model, notifications.ShowSuccess("No unused networks to prune")
				}
				model.showPruneNetworksConfirmation()
				return model, tea.Batch(cmds...)

			case key.Matches(msg, model.keybindings.createNetwork):
				model = model.withCreateNetworkDialog()
				return model, nil
			case key.Matches(msg, model.keybindings.attachContainer):
				model.handleAttachContainer()
				return model, nil
			case key.Matches(msg, model.keybindings.detachContainer):
				model.handleDetachContainer()
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

	// 4. Update Detail Content
	detailCmd := model.updateDetailContent()
	if detailCmd != nil {
		cmds = append(cmds, detailCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model Model) handleCreateNetworkComplete(msg MsgCreateNetworkComplete) (Model, tea.Cmd) {
	if msg.Err != nil {
		return model, notifications.ShowError(msg.Err)
	}

	return model, tea.Batch(
		notifications.ShowSuccess(fmt.Sprintf("Created network: %s", msg.NetworkID[:12])),
		model.Refresh(),
		func() tea.Msg {
			return base.MsgResourceChanged{
				Resource:  base.ResourceNetwork,
				Operation: base.OperationCreated,
				IDs:       []string{msg.NetworkID},
			}
		},
	)
}

func (model Model) View() string {
	return model.ResourceView.View()
}

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model Model) handleToggleSelection() {
	model.HandleToggleSelection()

	index := model.GetSelectedIndex()
	if selectedItem := model.GetSelectedItem(); selectedItem != nil {
		selectedItem.isSelected = model.Selections.IsSelected(selectedItem.Network.ID)
		model.SetItem(index, *selectedItem)
	}
}

func (model Model) handleToggleSelectionOfAll() {
	model.HandleToggleAll()

	items := model.GetItems()
	for i, item := range items {
		item.isSelected = model.Selections.IsSelected(item.Network.ID)
		model.SetItem(i, item)
	}
}

func (model *Model) handleRemove() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	// Check if this is a system network
	if isSystemNetwork(selectedItem.Network.Name) {
		return notifications.ShowInfo(
			fmt.Sprintf("Cannot delete system network: %s", selectedItem.Network.Name),
		)
	}

	containersUsingNetwork, err := state.GetClient().GetContainersUsingNetwork(stdcontext.Background(), selectedItem.Network.ID)
	if err != nil {
		// If we can't check usage, show error and don't proceed with deletion
		errorDialog := components.NewDialog(
			fmt.Sprintf("Failed to check network usage: %v\nCannot safely delete network.", err),
			[]components.DialogButton{
				{Label: "OK"},
			},
		)
		model.SetOverlay(errorDialog)
		return nil
	}
	if len(containersUsingNetwork) > 0 {
		confirmationDialog := components.NewDialog(
			safety.ForceDeleteInUseConfirmation("Network", selectedItem.Network.Name, len(containersUsingNetwork), containersUsingNetwork),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteNetwork", Payload: selectedItem.Network.ID}},
			},
		)
		model.SetOverlay(confirmationDialog)
	} else {
		confirmationDialog := components.NewDialog(
			safety.DeleteConfirmation("network", selectedItem.Network.Name),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Delete", Action: base.SmartDialogAction{Type: "DeleteNetwork", Payload: selectedItem.Network.ID}},
			},
		)
		model.SetOverlay(confirmationDialog)
	}
	return nil
}

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No network selected"))
		model.SetExtraContent("") // Clear extra pane when no network selected
		return nil
	}

	networkID := selectedItem.Network.ID
	currentID := model.detailsPanel.GetCurrentID()

	// If we've switched to a different network, OR we don't have inspection data yet, fetch it
	if networkID != currentID || model.inspection.ID == "" {
		// SetCurrentID will save scroll position for previous ID
		model.detailsPanel.SetCurrentID(networkID, model.getViewport())

		// Fetch inspection data asynchronously
		return func() tea.Msg {
			networkInfo, err := state.GetClient().InspectNetwork(stdcontext.Background(), networkID)
			return MsgNetworkInspection{ID: networkID, Network: networkInfo, Err: err}
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
	content := builders.BuildNetworkPanel(model.inspection, model.GetContentWidth(), format)
	model.SetContent(content)

	// Update "Used By" panel
	model.updateUsedByPanel()
}

// updateUsedByPanel updates the extra pane with containers using this network
func (model *Model) updateUsedByPanel() {
	if model.inspection.ID == "" {
		model.SetExtraContent("")
		return
	}

	// Fetch containers using this network
	usedBy, err := state.GetClient().GetContainersUsingNetwork(stdcontext.Background(), model.inspection.ID)
	if err != nil {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	model.SetExtraContent(buildNetworkConnectivityContent(model.inspection, usedBy))
}

func buildNetworkConnectivityContent(inspection types.NetworkResource, usedBy []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Driver: %s\n", inspection.Driver))
	b.WriteString(fmt.Sprintf("Scope: %s\n", inspection.Scope))
	b.WriteString(fmt.Sprintf("ID: %s\n\n", inspection.ID))

	b.WriteString("Connectivity\n")
	if len(inspection.Containers) > 0 {
		b.WriteString("Endpoints:\n")
		for _, endpoint := range inspection.Containers {
			name := strings.TrimSpace(endpoint.Name)
			if name == "" {
				name = shortID(endpoint.EndpointID)
			}
			b.WriteString("• ")
			b.WriteString(name)
			if strings.TrimSpace(endpoint.IPv4Address) != "" {
				b.WriteString(" (")
				b.WriteString(endpoint.IPv4Address)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(usedBy) == 0 {
		b.WriteString("No containers currently connected to this network.")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%d connected containers:\n", len(usedBy)))
	for _, name := range usedBy {
		b.WriteString("• ")
		b.WriteString(name)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func shortID(id string) string {
	trimmed := strings.TrimSpace(id)
	if len(trimmed) > 12 {
		return trimmed[:12]
	}

	return trimmed
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

func (model Model) hasPrunableNetworks() bool {
	for _, item := range model.GetItems() {
		if item.IsActive {
			continue
		}
		if isSystemNetwork(item.Network.Name) {
			continue
		}

		return true
	}

	return false
}

// handlePruneNetworks prunes unused networks
func (model *Model) handlePruneNetworks() tea.Cmd {
	progressDialog := components.NewProgressDialogWithBar("Pruning unused networks")
	progressDialog.EnableAutoAdvance(0.95, 0.04)
	progressDialog.SetStatus("Discovering unused networks to prune...")
	model.SetOverlay(progressDialog)

	// Start async prune operation
	return func() tea.Msg {
		ctx := stdcontext.Background()
		networksDeleted, err := state.GetClient().PruneNetworks(ctx)
		return MsgPruneComplete{
			NetworksDeleted: networksDeleted,
			Err:             err,
		}
	}
}

func (model *Model) showPruneNetworksConfirmation() {
	candidates := model.pruneNetworkCandidates()
	if len(candidates) == 0 {
		return
	}

	samples := candidates
	if len(samples) > 3 {
		samples = samples[:3]
	}

	confirmDialog := components.NewDialog(
		safety.PruneConfirmation("networks", len(candidates), samples),
		[]components.DialogButton{
			{Label: "Cancel"},
			{Label: "Prune", Action: base.SmartDialogAction{Type: "PruneNetworks"}},
		},
	)
	model.SetOverlay(confirmDialog)
}

func (model Model) pruneNetworkCandidates() []string {
	items := model.GetItems()
	candidates := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsActive || isSystemNetwork(item.Network.Name) {
			continue
		}
		candidates = append(candidates, item.Network.Name)
	}

	return candidates
}

// withCreateNetworkDialog returns model with create-network dialog shown.
func (model Model) withCreateNetworkDialog() Model {
	fields := []components.FormField{
		{
			Label:       "Name",
			Placeholder: "my-network",
			Required:    true,
		},
		{
			Label:       "Driver",
			Placeholder: "bridge",
			Required:    false,
		},
		{
			Label:       "Subnet (CIDR)",
			Placeholder: "172.20.0.0/16",
			Required:    false,
		},
		{
			Label:       "Gateway",
			Placeholder: "172.20.0.1",
			Required:    false,
		},
		{
			Label:       "Enable IPv6",
			Placeholder: "yes/no",
			Required:    false,
		},
		{
			Label:       "Labels",
			Placeholder: "KEY=value,FOO=bar",
			Required:    false,
		},
	}

	dialog := components.NewFormDialog(
		"Create Network",
		fields,
		base.SmartDialogAction{Type: "CreateNetworkAction"},
		nil,
	)

	model.SetOverlay(dialog)
	return model
}

// performCreateNetwork creates a network
func (model Model) performCreateNetwork(name, driver, subnet, gateway string, enableIPv6 bool, labels map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()

		// Use "bridge" as default driver if not specified
		if driver == "" {
			driver = "bridge"
		}

		networkID, err := state.GetClient().CreateNetwork(ctx, name, driver, subnet, gateway, enableIPv6, labels)
		if err != nil {
			return MsgCreateNetworkComplete{Err: fmt.Errorf("failed to create network: %w", err)}
		}

		return MsgCreateNetworkComplete{NetworkID: networkID}
	}
}

func (model *Model) handleAttachContainer() {
	selected := model.GetSelectedItem()
	if selected == nil {
		return
	}

	fields := []components.FormField{{
		Label:       "Container ID",
		Placeholder: "container-id-or-name",
		Required:    true,
	}}

	payload := map[string]any{"networkID": selected.Network.ID}
	dialog := components.NewFormDialog(
		"Attach Container",
		fields,
		base.SmartDialogAction{Type: "AttachContainerAction", Payload: payload},
		nil,
	)
	model.SetOverlay(dialog)
}

func (model *Model) handleDetachContainer() {
	selected := model.GetSelectedItem()
	if selected == nil {
		return
	}

	fields := []components.FormField{{
		Label:       "Container ID",
		Placeholder: "container-id-or-name",
		Required:    true,
	}}

	payload := map[string]any{"networkID": selected.Network.ID}
	dialog := components.NewFormDialog(
		"Detach Container",
		fields,
		base.SmartDialogAction{Type: "DetachContainerAction", Payload: payload},
		nil,
	)
	model.SetOverlay(dialog)
}

func (model Model) performAttachContainer(networkID, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := state.GetClient().ConnectContainerToNetwork(stdcontext.Background(), containerID, networkID)
		return MsgAttachContainerComplete{NetworkID: networkID, ContainerID: containerID, Err: err}
	}
}

func (model Model) performDetachContainer(networkID, containerID string) tea.Cmd {
	return func() tea.Msg {
		err := state.GetClient().DisconnectContainerFromNetwork(stdcontext.Background(), containerID, networkID, false)
		return MsgDetachContainerComplete{NetworkID: networkID, ContainerID: containerID, Err: err}
	}
}
