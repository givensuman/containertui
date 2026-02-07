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
	"github.com/atotto/clipboard"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
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
	currentNetworkID   string
	inspection         types.NetworkResource
	scrollPositions    map[string]int
	currentFormat      string
}

func New() *Model {
	networkKeybindings := newKeybindings()

	fetchNetworks := func() ([]NetworkItem, error) {
		networkList, err := context.GetClient().GetNetworks(stdcontext.Background())
		if err != nil {
			return nil, err
		}
		items := make([]NetworkItem, 0, len(networkList))
		for _, network := range networkList {
			items = append(items, NetworkItem{Network: network})
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
		scrollPositions:    make(map[string]int),
		currentFormat:      "",
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		networkKeybindings.toggleSelection,
		networkKeybindings.toggleSelectionOfAll,
		networkKeybindings.remove,
	}

	return &model
}

func (model *Model) Init() tea.Cmd {
	return model.ResourceView.Init()
}

func (model *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	// 1. Try standard ResourceView updates first (resizing, dialog closing, basic navigation)
	updatedView, cmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// 2. Handle Messages
	switch msg := msg.(type) {
	case MsgNetworkInspection:
		if msg.ID == model.currentNetworkID && msg.Err == nil {
			model.inspection = msg.Network
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.restoreScrollPosition()
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteNetwork" {
				networkID := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveNetwork(stdcontext.Background(), networkID)
				if err == nil {
					// Close the overlay and refresh list
					model.CloseOverlay()
					return model, model.Refresh()
				} else {
					// Show error
					errorDialog := components.NewDialog(
						fmt.Sprintf("Failed to remove network:\n\n%v", err),
						[]components.DialogButton{{Label: "OK"}},
					)
					model.SetOverlay(errorDialog)
					return model, nil
				}
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
				model.handleRemove()
				return model, nil
			}
		}
	} else {
		// Detail or extra pane is focused
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			// Only handle these actions when detail pane is focused (not extra)
			if model.IsDetailFocused() {
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
	}

	// 4. Update Detail Content
	detailCmd := model.updateDetailContent()
	if detailCmd != nil {
		cmds = append(cmds, detailCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) View() string {
	return model.ResourceView.View()
}

func (model *Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model *Model) handleToggleSelection() {
	model.HandleToggleSelection()

	index := model.GetSelectedIndex()
	if selectedItem := model.GetSelectedItem(); selectedItem != nil {
		selectedItem.isSelected = model.Selections.IsSelected(selectedItem.Network.ID)
		model.SetItem(index, *selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	model.HandleToggleAll()

	items := model.GetItems()
	for i, item := range items {
		item.isSelected = model.Selections.IsSelected(item.Network.ID)
		model.SetItem(i, item)
	}
}

func (model *Model) handleRemove() {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	containersUsingNetwork, err := context.GetClient().GetContainersUsingNetwork(stdcontext.Background(), selectedItem.Network.ID)
	if err != nil {
		// If we can't check usage, show error and don't proceed with deletion
		errorDialog := components.NewDialog(
			fmt.Sprintf("Failed to check network usage: %v\nCannot safely delete network.", err),
			[]components.DialogButton{
				{Label: "OK"},
			},
		)
		model.SetOverlay(errorDialog)
		return
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
}

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No network selected."))
		model.SetExtraContent("") // Clear extra pane when no network selected
		return nil
	}

	networkID := selectedItem.Network.ID
	// If we've switched to a different network, OR we don't have inspection data yet, fetch it
	if networkID != model.currentNetworkID || model.inspection.ID == "" {
		// Save scroll position of previous network
		if model.currentNetworkID != "" && model.currentNetworkID != networkID {
			model.saveScrollPosition()
		}

		model.currentNetworkID = networkID
		// Fetch inspection data asynchronously
		return func() tea.Msg {
			networkInfo, err := context.GetClient().InspectNetwork(stdcontext.Background(), networkID)
			return MsgNetworkInspection{ID: networkID, Network: networkInfo, Err: err}
		}
	}

	return nil
}

// saveScrollPosition saves the current viewport scroll position for the current network
func (model *Model) saveScrollPosition() {
	if model.currentNetworkID != "" {
		if vp := model.getViewport(); vp != nil {
			model.scrollPositions[model.currentNetworkID] = vp.YOffset()
		}
	}
}

// restoreScrollPosition restores the viewport scroll position for the current network
func (model *Model) restoreScrollPosition() {
	if model.currentNetworkID != "" {
		if vp := model.getViewport(); vp != nil {
			if offset, exists := model.scrollPositions[model.currentNetworkID]; exists {
				vp.SetYOffset(offset)
			} else {
				vp.SetYOffset(0) // New network, start at top
			}
		}
	}
}

// getViewport returns the viewport from the detail pane if available
func (model *Model) getViewport() *viewport.Model {
	if vp, ok := model.SplitView.Detail.(*components.ViewportPane); ok {
		return &vp.Viewport
	}
	return nil
}

// refreshInspectionContent refreshes the detail content with current inspection data
func (model *Model) refreshInspectionContent() {
	// Determine format to use
	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

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
	usedBy, err := context.GetClient().GetContainersUsingNetwork(stdcontext.Background(), model.inspection.ID)
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

// handleCopyToClipboard copies the current inspection output to clipboard
func (model *Model) handleCopyToClipboard() tea.Cmd {
	if model.inspection.ID == "" {
		return nil
	}

	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	data, err := infopanel.MarshalToFormat(model.inspection, format)
	if err != nil {
		return notifications.ShowError(err)
	}

	if err := clipboard.WriteAll(string(data)); err != nil {
		return notifications.ShowError(err)
	}

	return notifications.ShowSuccess("Copied to clipboard")
}

// handleToggleFormat toggles between JSON and YAML format
func (model *Model) handleToggleFormat() tea.Cmd {
	currentFormat := model.currentFormat
	if currentFormat == "" {
		cfg := context.GetConfig()
		currentFormat = cfg.InspectionFormat
		if currentFormat == "" {
			currentFormat = "yaml"
		}
	}

	if currentFormat == "json" {
		model.currentFormat = "yaml"
	} else {
		model.currentFormat = "json"
	}

	model.refreshInspectionContent()
	return notifications.ShowSuccess("Switched to " + model.currentFormat)
}

func (model *Model) ShortHelp() []key.Binding {
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

func (model *Model) FullHelp() [][]key.Binding {
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
