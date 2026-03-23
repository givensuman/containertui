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
	Err error
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
	pruneNetworks        key.Binding
	createNetwork        key.Binding
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
		pruneNetworks: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prune unused"),
		),
		createNetwork: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "create network"),
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
		networkKeybindings.toggleSelection,
		networkKeybindings.toggleSelectionOfAll,
		networkKeybindings.remove,
		networkKeybindings.forceRemove,
		networkKeybindings.pruneNetworks,
		networkKeybindings.createNetwork,
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
		model.CloseOverlay()
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		return model, tea.Batch(
			notifications.ShowSuccess("Pruned unused networks"),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceNetwork,
					Operation: base.OperationPruned,
					IDs:       nil,
				}
			},
		)
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
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

				name := formValues["0"]    // First field
				driver := formValues["1"]  // Second field
				subnet := formValues["2"]  // Third field
				gateway := formValues["3"] // Fourth field
				ipv6 := formValues["4"]    // Fifth field
				labels := formValues["5"]  // Sixth field

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
				return model, model.handleRemove(false)

			case key.Matches(msg, model.keybindings.forceRemove):
				return model, model.handleRemove(true)

			case key.Matches(msg, model.keybindings.pruneNetworks):
				if cmd := model.handlePruneNetworks(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)

			case key.Matches(msg, model.keybindings.createNetwork):
				model.handleCreateNetwork()
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

func (model *Model) handleRemove(force bool) tea.Cmd {
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

	if force {
		// Force delete - show confirmation dialog first
		confirmationDialog := components.NewDialog(
			fmt.Sprintf("Force delete network %s?\n\nThis will delete the network even if containers are using it.", selectedItem.Network.Name),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteNetwork", Payload: selectedItem.Network.ID}},
			},
		)
		model.SetOverlay(confirmationDialog)
		return nil
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
		warningDialog := components.NewDialog(
			fmt.Sprintf("Network %s is used by %d containers (%v).\nCannot delete.", selectedItem.Network.Name, len(containersUsingNetwork), containersUsingNetwork),
			[]components.DialogButton{
				{Label: "OK"},
			},
		)
		model.SetOverlay(warningDialog)
	} else {
		confirmationDialog := components.NewDialog(
			fmt.Sprintf("Are you sure you want to delete network %s?", selectedItem.Network.Name),
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

	if len(usedBy) == 0 {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No containers using this network"))
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

// handlePruneNetworks prunes unused networks
func (model *Model) handlePruneNetworks() tea.Cmd {
	// Show progress dialog
	progressDialog := components.NewProgressDialog(
		"Pruning unused networks...\n\nThis may take a few moments...",
	)
	model.SetOverlay(progressDialog)

	// Start async prune operation
	return func() tea.Msg {
		ctx := stdcontext.Background()
		err := state.GetClient().PruneNetworks(ctx)
		return MsgPruneComplete{
			Err: err,
		}
	}
}

// handleCreateNetwork shows dialog to create a network
func (model Model) handleCreateNetwork() {
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
			return notifications.ShowError(fmt.Errorf("failed to create network: %w", err))
		}
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Created network: %s", networkID[:12])),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceNetwork,
					Operation: base.OperationCreated,
					IDs:       []string{networkID},
				}
			},
		)
	}
}
