// Package networks defines the networks component.
package networks

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type keybindings struct {
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	remove               key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		toggleSelection: key.NewBinding(
			key.WithKeys(tea.KeySpace.String()),
			key.WithHelp("space", "toggle selection"),
		),
		toggleSelectionOfAll: key.NewBinding(
			key.WithKeys(tea.KeyCtrlA.String()),
			key.WithHelp("ctrl+a", "toggle selection of all"),
		),
		remove: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove"),
		),
	}
}

// selectedNetworks is map of a network's ID to
// its index in the list
type selectedNetworks struct {
	selections map[string]int
}

func newSelectedNetworks() *selectedNetworks {
	return &selectedNetworks{
		selections: make(map[string]int),
	}
}

func (sn *selectedNetworks) selectNetworkInList(id string, index int) {
	sn.selections[id] = index
}

func (sn selectedNetworks) unselectNetworkInList(id string) {
	delete(sn.selections, id)
}

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

type Model struct {
	shared.Component
	style            lipgloss.Style
	list             list.Model
	selectedNetworks *selectedNetworks
	keybindings      *keybindings

	// Overlay support
	sessionState sessionState
	foreground   tea.Model
	overlayModel *overlay.Model
}

var (
	_ tea.Model             = (*Model)(nil)
	_ shared.ComponentModel = (*Model)(nil)
)

func New() Model {
	networks, err := context.GetClient().GetNetworks()
	if err != nil {
		networks = []client.Network{}
	}
	items := make([]list.Item, 0, len(networks))
	for _, net := range networks {
		items = append(items, NetworkItem{Network: net})
	}

	width, height := context.GetWindowSize()
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingTop(1)

	delegate := newDefaultDelegate()
	l := list.New(items, delegate, width, height)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(colors.Primary())
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(colors.Primary())
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	l.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	keybindings := newKeybindings()
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keybindings.toggleSelection,
			keybindings.toggleSelectionOfAll,
			keybindings.remove,
		}
	}

	m := Model{
		style:            style,
		list:             l,
		selectedNetworks: newSelectedNetworks(),
		keybindings:      keybindings,
		sessionState:     viewMain,
	}

	// Initialize overlay with nil content initially
	m.overlayModel = overlay.New(nil, m.list, overlay.Center, overlay.Center, 0, 0)
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.sessionState {
	case viewOverlay:
		// If overlay is active, route messages there
		fg, cmd := m.foreground.Update(msg)
		m.foreground = fg
		cmds = append(cmds, cmd)

		// Check if overlay wants to close or confirm
		if _, ok := msg.(shared.CloseDialogMessage); ok {
			m.sessionState = viewMain
			m.foreground = nil
		} else if confirm, ok := msg.(shared.ConfirmationMessage); ok {
			// Handle confirmation
			if confirm.Action.Type == "DeleteNetwork" {
				id := confirm.Action.Payload.(string)
				// Perform deletion
				err := context.GetClient().RemoveNetwork(id)
				if err != nil {
					// TODO: Show error
				} else {
					// Remove from list or refresh
				}
			}
			m.sessionState = viewMain
			m.foreground = nil
		}
	case viewMain:
		// Main loop handling
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.UpdateWindowDimensions(msg)
		case tea.KeyMsg:
			if m.list.FilterState() == list.Filtering {
				break
			}

			switch {
			case key.Matches(msg, m.keybindings.toggleSelection):
				m.handleToggleSelection()
			case key.Matches(msg, m.keybindings.toggleSelectionOfAll):
				m.handleToggleSelectionOfAll()
			case key.Matches(msg, m.keybindings.remove):
				// Trigger remove dialog
				item := m.list.SelectedItem()
				if item != nil {
					if n, ok := item.(NetworkItem); ok {
						// Check usage
						usedBy, _ := context.GetClient().GetContainersUsingNetwork(n.Network.ID)
						if len(usedBy) > 0 {
							// Show warning dialog with navigation option
							dialog := shared.NewSmartDialog(
								fmt.Sprintf("Network %s is used by %d containers (%v).\nCannot delete.", n.Network.Name, len(usedBy), usedBy),
								[]shared.DialogButton{
									{Label: "OK", IsSafe: true},
								},
							)
							m.foreground = dialog
							m.sessionState = viewOverlay
						} else {
							// Show confirmation
							dialog := shared.NewSmartDialog(
								fmt.Sprintf("Are you sure you want to delete network %s?", n.Network.Name),
								[]shared.DialogButton{
									{Label: "Cancel", IsSafe: true},
									{Label: "Delete", IsSafe: false, Action: shared.SmartDialogAction{Type: "DeleteNetwork", Payload: n.Network.ID}},
								},
							)
							m.foreground = dialog
							m.sessionState = viewOverlay
						}
					}
				}
			}
		}

		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Always update overlay model to sync state
	m.overlayModel.Foreground = m.foreground
	m.overlayModel.Background = m.list

	return m, tea.Batch(cmds...)
}

func (m *Model) handleToggleSelection() {
	index := m.list.Index()
	selectedItem, ok := m.list.SelectedItem().(NetworkItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			m.selectedNetworks.unselectNetworkInList(selectedItem.Network.ID)
		} else {
			m.selectedNetworks.selectNetworkInList(selectedItem.Network.ID, index)
		}

		selectedItem.isSelected = !isSelected
		m.list.SetItem(index, selectedItem)
	}
}

func (m *Model) handleToggleSelectionOfAll() {
	allAlreadySelected := true
	items := m.list.Items()

	for _, item := range items {
		if c, ok := item.(NetworkItem); ok {
			if _, selected := m.selectedNetworks.selections[c.Network.ID]; !selected {
				allAlreadySelected = false
				break
			}
		}
	}

	if allAlreadySelected {
		// Unselect all items
		m.selectedNetworks = newSelectedNetworks()

		for index, item := range m.list.Items() {
			item, ok := item.(NetworkItem)
			if ok {
				item.isSelected = false
				m.list.SetItem(index, item)
			}
		}
	} else {
		// Select all items
		m.selectedNetworks = newSelectedNetworks()

		for index, item := range m.list.Items() {
			item, ok := item.(NetworkItem)
			if ok {
				item.isSelected = true
				m.list.SetItem(index, item)
				m.selectedNetworks.selectNetworkInList(item.Network.ID, index)
			}
		}
	}
}

func (m Model) View() string {
	if m.sessionState == viewOverlay && m.foreground != nil {
		return m.overlayModel.View()
	}

	// Get dimensions for master (list) and detail (inspect)
	lm := shared.NewLayoutManager(m.WindowWidth, m.WindowHeight)
	_, detail := lm.CalculateMasterDetail(lipgloss.NewStyle())

	// Render the list (background)
	listView := m.style.Render(m.list.View())

	// Render the detail view (side pane)
	detailStyle := lipgloss.NewStyle().
		Width(detail.Width - 2). // Subtract 2 for border width compensation
		Height(detail.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Muted()).
		Padding(1)

	var detailContent string
	item := m.list.SelectedItem()
	if item != nil {
		if n, ok := item.(NetworkItem); ok {
			detailContent = fmt.Sprintf(
				"ID: %s\nName: %s\nDriver: %s\nScope: %s",
				n.Network.ID, n.Network.Name, n.Network.Driver, n.Network.Scope,
			)
		}
	}

	detailView := detailStyle.Render(detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

func (m *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	m.WindowWidth = msg.Width
	m.WindowHeight = msg.Height

	lm := shared.NewLayoutManager(msg.Width, msg.Height)
	master, _ := lm.CalculateMasterDetail(m.style)

	m.style = m.style.Width(master.Width).Height(master.Height) // Update style dimensions

	switch m.sessionState {
	case viewMain:
		if m.list.Width() != master.ContentWidth || m.list.Height() != master.ContentHeight {
			m.list.SetWidth(master.ContentWidth)
			m.list.SetHeight(master.ContentHeight)
		}
	case viewOverlay:
		if d, ok := m.foreground.(shared.SmartDialog); ok {
			d.UpdateWindowDimensions(msg)
			m.foreground = d
		}
	}
}

func (m Model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	return m.list.FullHelp()
}
