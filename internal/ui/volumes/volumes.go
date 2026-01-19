// Package volumes defines the volumes component.
package volumes

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type detailsKeybindings struct {
	Up     key.Binding
	Down   key.Binding
	Switch key.Binding
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
			key.WithKeys("1", "2", "3", "4", "tab", "shift+tab"),
			key.WithHelp("1-4/tab", "switch tab"),
		),
	}
}

// selectedVolumes maps a volume's Name to its index in the list.
type selectedVolumes struct {
	selections map[string]int
}

func newSelectedVolumes() *selectedVolumes {
	return &selectedVolumes{
		selections: make(map[string]int),
	}
}

func (selectedVolumes *selectedVolumes) selectVolumeInList(name string, index int) {
	selectedVolumes.selections[name] = index
}

func (selectedVolumes selectedVolumes) unselectVolumeInList(name string) {
	delete(selectedVolumes.selections, name)
}

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

// Model represents the volumes component state.
type Model struct {
	shared.Component
	splitView       shared.SplitView
	selectedVolumes *selectedVolumes
	keybindings     *keybindings

	sessionState       sessionState
	detailsKeybindings detailsKeybindings
	foreground         interface{} // Can be SmartDialog, etc.
}

var (
	_ tea.Model             = (*Model)(nil)
	_ shared.ComponentModel = (*Model)(nil)
)

func New() Model {
	volumeList, err := context.GetClient().GetVolumes()
	if err != nil {
		volumeList = []client.Volume{}
	}
	items := make([]list.Item, 0, len(volumeList))
	for _, volume := range volumeList {
		items = append(items, VolumeItem{Volume: volume})
	}

	width, height := context.GetWindowSize()

	delegate := newDefaultDelegate()
	listModel := list.New(items, delegate, width, height)
	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.Filter.Focused.Prompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.Filter.Cursor.Color = colors.Primary()
	// listModel.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	// listModel.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	volumeKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			volumeKeybindings.toggleSelection,
			volumeKeybindings.toggleSelectionOfAll,
			volumeKeybindings.remove,
			volumeKeybindings.switchTab,
		}
	}

	splitView := shared.NewSplitView(listModel, shared.NewViewportPane())

	model := Model{
		splitView:          splitView,
		selectedVolumes:    newSelectedVolumes(),
		keybindings:        volumeKeybindings,
		sessionState:       viewMain,
		detailsKeybindings: newDetailsKeybindings(),
	}

	return model
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch model.sessionState {
	case viewOverlay:
		// Update SplitView for background resize but don't process keys
		if _, ok := msg.(tea.WindowSizeMsg); ok {
			updatedSplitView, splitCmd := model.splitView.Update(msg)
			model.splitView = updatedSplitView
			cmds = append(cmds, splitCmd)
		}

		if model.foreground != nil {
			switch fg := model.foreground.(type) {
			case shared.SmartDialog:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			}
		}

		if _, ok := msg.(shared.CloseDialogMessage); ok {
			model.sessionState = viewMain
			model.foreground = nil
		} else if confirmMsg, ok := msg.(shared.ConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteVolume" {
				volumeName := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveVolume(volumeName)
				if err != nil {
					break
				}
			}
			model.sessionState = viewMain
			model.foreground = nil
		}
	case viewMain:
		updatedSplitView, splitCmd := model.splitView.Update(msg)
		model.splitView = updatedSplitView
		cmds = append(cmds, splitCmd)

		if model.splitView.Focus == shared.FocusList {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				model.UpdateWindowDimensions(msg)
			case tea.KeyPressMsg:
				if model.splitView.List.FilterState() == list.Filtering {
					break
				}

				switch {
				case key.Matches(msg, model.keybindings.switchTab):
					return model, nil
				case key.Matches(msg, model.keybindings.toggleSelection):
					model.handleToggleSelection()
				case key.Matches(msg, model.keybindings.toggleSelectionOfAll):
					model.handleToggleSelectionOfAll()
				case key.Matches(msg, model.keybindings.remove):
					selectedItem := model.splitView.List.SelectedItem()
					if selectedItem != nil {
						if volumeItem, ok := selectedItem.(VolumeItem); ok {
							containersUsingVolume, _ := context.GetClient().GetContainersUsingVolume(volumeItem.Volume.Name)
							if len(containersUsingVolume) > 0 {
								warningDialog := shared.NewSmartDialog(
									fmt.Sprintf("Volume %s is used by %d containers (%v).\nCannot delete.", volumeItem.Volume.Name, len(containersUsingVolume), containersUsingVolume),
									[]shared.DialogButton{
										{Label: "OK", IsSafe: true},
									},
								)
								model.foreground = warningDialog
								model.sessionState = viewOverlay
							} else {
								confirmationDialog := shared.NewSmartDialog(
									fmt.Sprintf("Are you sure you want to delete volume %s?", volumeItem.Volume.Name),
									[]shared.DialogButton{
										{Label: "Cancel", IsSafe: true},
										{Label: "Delete", IsSafe: false, Action: shared.SmartDialogAction{Type: "DeleteVolume", Payload: volumeItem.Volume.Name}},
									},
								)
								model.foreground = confirmationDialog
								model.sessionState = viewOverlay
							}
						}
					}
				}
			}
		}

		// Update Detail Content
		selectedItem := model.splitView.List.SelectedItem()
		if selectedItem != nil {
			if volumeItem, ok := selectedItem.(VolumeItem); ok {
				detailsContent := fmt.Sprintf(
					"Name: %s\nDriver: %s\nMountpoint: %s",
					volumeItem.Volume.Name, volumeItem.Volume.Driver, volumeItem.Volume.Mountpoint,
				)
				if pane, ok := model.splitView.Detail.(*shared.ViewportPane); ok {
					pane.SetContent(detailsContent)
				}
			}
		} else {
			if pane, ok := model.splitView.Detail.(*shared.ViewportPane); ok {
				pane.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No volume selected."))
			}
		}
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) handleToggleSelection() {
	currentIndex := model.splitView.List.Index()
	selectedItem, ok := model.splitView.List.SelectedItem().(VolumeItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			model.selectedVolumes.unselectVolumeInList(selectedItem.Volume.Name)
		} else {
			model.selectedVolumes.selectVolumeInList(selectedItem.Volume.Name, currentIndex)
		}

		selectedItem.isSelected = !isSelected
		model.splitView.List.SetItem(currentIndex, selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	allVolumesSelected := true
	items := model.splitView.List.Items()

	for _, item := range items {
		if volumeItem, ok := item.(VolumeItem); ok {
			if _, isSelected := model.selectedVolumes.selections[volumeItem.Volume.Name]; !isSelected {
				allVolumesSelected = false
				break
			}
		}
	}

	if allVolumesSelected {
		model.selectedVolumes = newSelectedVolumes()
		for index, item := range model.splitView.List.Items() {
			if volumeItem, ok := item.(VolumeItem); ok {
				volumeItem.isSelected = false
				model.splitView.List.SetItem(index, volumeItem)
			}
		}
	} else {
		model.selectedVolumes = newSelectedVolumes()
		for index, item := range model.splitView.List.Items() {
			if volumeItem, ok := item.(VolumeItem); ok {
				volumeItem.isSelected = true
				model.splitView.List.SetItem(index, volumeItem)
				model.selectedVolumes.selectVolumeInList(volumeItem.Volume.Name, index)
			}
		}
	}
}

func (model Model) View() tea.View {
	if model.sessionState == viewOverlay && model.foreground != nil {
		var fgView string
		switch fg := model.foreground.(type) {
		case shared.SmartDialog:
			fgView = fg.View()
		}

		return shared.RenderOverlay(
			model.splitView.View(),
			fgView,
			model.WindowWidth,
			model.WindowHeight,
		)
	}

	return tea.NewView(model.splitView.View())
}

func (model *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.WindowWidth = msg.Width
	model.WindowHeight = msg.Height
	model.splitView.SetSize(msg.Width, msg.Height)

	switch model.sessionState {
	case viewOverlay:
		if smartDialog, ok := model.foreground.(shared.SmartDialog); ok {
			smartDialog.UpdateWindowDimensions(msg)
			model.foreground = smartDialog
		}
	}
}

func (model Model) ShortHelp() []key.Binding {
	switch model.splitView.Focus {
	case shared.FocusList:
		return model.splitView.List.ShortHelp()
	case shared.FocusDetail:
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
		}
	}
	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	switch model.splitView.Focus {
	case shared.FocusList:
		return model.splitView.List.FullHelp()
	case shared.FocusDetail:
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
		}
	}
	return nil
}
