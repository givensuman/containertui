// Package volumes defines the volumes component.
package volumes

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

// selectedVolumes is map of a volume's Name (volumes don't always have IDs like images/containers) to
// its index in the list. The client uses Name as the primary identifier for volumes usually.
// Wait, client.Volume struct doesn't have ID field, only Name.
type selectedVolumes struct {
	selections map[string]int
}

func newSelectedVolumes() *selectedVolumes {
	return &selectedVolumes{
		selections: make(map[string]int),
	}
}

func (sv *selectedVolumes) selectVolumeInList(name string, index int) {
	sv.selections[name] = index
}

func (sv selectedVolumes) unselectVolumeInList(name string) {
	delete(sv.selections, name)
}

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

type Model struct {
	shared.Component
	style           lipgloss.Style
	list            list.Model
	selectedVolumes *selectedVolumes
	keybindings     *keybindings

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
	volumes, err := context.GetClient().GetVolumes()
	if err != nil {
		volumes = []client.Volume{}
	}
	items := make([]list.Item, 0, len(volumes))
	for _, vol := range volumes {
		items = append(items, VolumeItem{Volume: vol})
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
		style:           style,
		list:            l,
		selectedVolumes: newSelectedVolumes(),
		keybindings:     keybindings,
		sessionState:    viewMain,
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
			if confirm.Action.Type == "DeleteVolume" {
				name := confirm.Action.Payload.(string)
				// Perform deletion
				err := context.GetClient().RemoveVolume(name)
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
					if v, ok := item.(VolumeItem); ok {
						// Check usage
						usedBy, _ := context.GetClient().GetContainersUsingVolume(v.Volume.Name)
						if len(usedBy) > 0 {
							// Show warning dialog with navigation option
							dialog := shared.NewSmartDialog(
								fmt.Sprintf("Volume %s is used by %d containers (%v).\nCannot delete.", v.Volume.Name, len(usedBy), usedBy),
								[]shared.DialogButton{
									{Label: "OK", IsSafe: true},
								},
							)
							m.foreground = dialog
							m.sessionState = viewOverlay
						} else {
							// Show confirmation
							dialog := shared.NewSmartDialog(
								fmt.Sprintf("Are you sure you want to delete volume %s?", v.Volume.Name),
								[]shared.DialogButton{
									{Label: "Cancel", IsSafe: true},
									{Label: "Delete", IsSafe: false, Action: shared.SmartDialogAction{Type: "DeleteVolume", Payload: v.Volume.Name}},
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
	selectedItem, ok := m.list.SelectedItem().(VolumeItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			m.selectedVolumes.unselectVolumeInList(selectedItem.Volume.Name)
		} else {
			m.selectedVolumes.selectVolumeInList(selectedItem.Volume.Name, index)
		}

		selectedItem.isSelected = !isSelected
		m.list.SetItem(index, selectedItem)
	}
}

func (m *Model) handleToggleSelectionOfAll() {
	allAlreadySelected := true
	items := m.list.Items()

	for _, item := range items {
		if c, ok := item.(VolumeItem); ok {
			if _, selected := m.selectedVolumes.selections[c.Volume.Name]; !selected {
				allAlreadySelected = false
				break
			}
		}
	}

	if allAlreadySelected {
		// Unselect all items
		m.selectedVolumes = newSelectedVolumes()

		for index, item := range m.list.Items() {
			item, ok := item.(VolumeItem)
			if ok {
				item.isSelected = false
				m.list.SetItem(index, item)
			}
		}
	} else {
		// Select all items
		m.selectedVolumes = newSelectedVolumes()

		for index, item := range m.list.Items() {
			item, ok := item.(VolumeItem)
			if ok {
				item.isSelected = true
				m.list.SetItem(index, item)
				m.selectedVolumes.selectVolumeInList(item.Volume.Name, index)
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
		if v, ok := item.(VolumeItem); ok {
			detailContent = fmt.Sprintf(
				"Name: %s\nDriver: %s\nMountpoint: %s",
				v.Volume.Name, v.Volume.Driver, v.Volume.Mountpoint,
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
