// Package networks defines the networks component.
package networks

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
)

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
			key.WithKeys("1", "2", "3", "4", "tab", "shift+tab"),
			key.WithHelp("1-4/tab", "switch tab"),
		),
	}
}

// Model represents the networks component state.
type Model struct {
	components.ResourceView[string, NetworkItem]
	keybindings *keybindings
}

var (
	_ tea.Model           = (*Model)(nil)
	_ base.ComponentModel = (*Model)(nil)
)

func New() *Model {
	networkKeybindings := newKeybindings()

	fetchNetworks := func() ([]NetworkItem, error) {
		networkList, err := context.GetClient().GetNetworks()
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

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView: *resourceView,
		keybindings:  networkKeybindings,
	}

	// Add custom keybindings to help
	model.ResourceView.AdditionalHelp = []key.Binding{
		networkKeybindings.toggleSelection,
		networkKeybindings.toggleSelectionOfAll,
		networkKeybindings.remove,
	}

	return &model
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 1. Try standard ResourceView updates first (resizing, dialog closing, basic navigation)
	updatedView, cmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// 2. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.ResourceView.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.ConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteNetwork" {
				networkID := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveNetwork(networkID)
				if err == nil {
					// Refresh list
					return model, model.ResourceView.Refresh()
				}
			}
			model.ResourceView.CloseOverlay()
			return model, nil
		}

		// Let ResourceView handle forwarding to overlay
		return model, tea.Batch(cmds...)
	}

	// 3. Main View Logic
	if model.ResourceView.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.ResourceView.IsFiltering() {
				break
			}

			switch {
			case key.Matches(msg, model.keybindings.switchTab):
				return model, nil // Handled by parent

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
	}

	// 4. Update Detail Content
	model.updateDetailContent()

	return model, tea.Batch(cmds...)
}

func (model *Model) View() tea.View {
	return tea.NewView(model.ResourceView.View())
}

func (model *Model) handleToggleSelection() {
	model.ResourceView.HandleToggleSelection()

	index := model.ResourceView.GetSelectedIndex()
	if selectedItem := model.ResourceView.GetSelectedItem(); selectedItem != nil {
		selectedItem.isSelected = model.ResourceView.Selections.IsSelected(selectedItem.Network.ID)
		model.ResourceView.SetItem(index, *selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	model.ResourceView.HandleToggleAll()

	items := model.ResourceView.GetItems()
	for i, item := range items {
		item.isSelected = model.ResourceView.Selections.IsSelected(item.Network.ID)
		model.ResourceView.SetItem(i, item)
	}
}

func (model *Model) handleRemove() {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	containersUsingNetwork, _ := context.GetClient().GetContainersUsingNetwork(selectedItem.Network.ID)
	if len(containersUsingNetwork) > 0 {
		warningDialog := components.NewSmartDialog(
			fmt.Sprintf("Network %s is used by %d containers (%v).\nCannot delete.", selectedItem.Network.Name, len(containersUsingNetwork), containersUsingNetwork),
			[]components.DialogButton{
				{Label: "OK", IsSafe: true},
			},
		)
		model.ResourceView.SetOverlay(warningDialog)
	} else {
		confirmationDialog := components.NewSmartDialog(
			fmt.Sprintf("Are you sure you want to delete network %s?", selectedItem.Network.Name),
			[]components.DialogButton{
				{Label: "Cancel", IsSafe: true},
				{Label: "Delete", IsSafe: false, Action: base.SmartDialogAction{Type: "DeleteNetwork", Payload: selectedItem.Network.ID}},
			},
		)
		model.ResourceView.SetOverlay(confirmationDialog)
	}
}

func (model *Model) updateDetailContent() {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem != nil {
		detailsContent := fmt.Sprintf(
			"ID: %s\nName: %s\nDriver: %s\nScope: %s",
			selectedItem.Network.ID, selectedItem.Network.Name, selectedItem.Network.Driver, selectedItem.Network.Scope,
		)
		model.ResourceView.SetContent(detailsContent)
	} else {
		model.ResourceView.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No network selected."))
	}
}

func (model *Model) removeNetworkFromList(id string) {
	// Replaced by Refresh
}

func (model *Model) ShortHelp() []key.Binding {
	return model.ResourceView.ShortHelp()
}

func (model *Model) FullHelp() [][]key.Binding {
	return model.ResourceView.FullHelp()
}
