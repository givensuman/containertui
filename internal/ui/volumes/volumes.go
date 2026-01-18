// Package volumes defines the volumes component.
package volumes

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
	overlay "github.com/rmhubbert/bubbletea-overlay"
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

const (
	focusList = iota
	focusDetails
)

// Model represents the volumes component state.
type Model struct {
	shared.Component
	style           lipgloss.Style
	list            list.Model
	viewport        viewport.Model
	selectedVolumes *selectedVolumes
	keybindings     *keybindings

	sessionState       sessionState
	focusedView        int
	detailsKeybindings detailsKeybindings
	foreground         tea.Model
	overlayModel       *overlay.Model
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
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingTop(1)

	delegate := newDefaultDelegate()
	listModel := list.New(items, delegate, width, height)
	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.FilterCursor = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	volumeKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			volumeKeybindings.toggleSelection,
			volumeKeybindings.toggleSelectionOfAll,
			volumeKeybindings.remove,
			volumeKeybindings.switchTab,
		}
	}

	detailViewport := viewport.New(0, 0)

	model := Model{
		style:              style,
		list:               listModel,
		viewport:           detailViewport,
		selectedVolumes:    newSelectedVolumes(),
		keybindings:        volumeKeybindings,
		sessionState:       viewMain,
		focusedView:        focusList,
		detailsKeybindings: newDetailsKeybindings(),
	}

	model.overlayModel = overlay.New(nil, model.list, overlay.Center, overlay.Center, 0, 0)
	return model
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch model.sessionState {
	case viewOverlay:
		foregroundModel, foregroundCmd := model.foreground.Update(msg)
		model.foreground = foregroundModel
		cmds = append(cmds, foregroundCmd)

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
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "tab" && model.list.FilterState() != list.Filtering {
				if model.focusedView == focusList {
					model.focusedView = focusDetails
				} else {
					model.focusedView = focusList
				}
				return model, nil
			}
		}

		isKeyMessage := false
		if _, ok := msg.(tea.KeyMsg); ok {
			isKeyMessage = true
		}

		if !isKeyMessage || model.focusedView == focusList {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				model.UpdateWindowDimensions(msg)
			case tea.KeyMsg:
				if model.list.FilterState() == list.Filtering {
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
					selectedItem := model.list.SelectedItem()
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
			updatedList, listCmd := model.list.Update(msg)
			model.list = updatedList
			cmds = append(cmds, listCmd)
		}

		selectedItem := model.list.SelectedItem()
		if selectedItem != nil {
			if volumeItem, ok := selectedItem.(VolumeItem); ok {
				detailsContent := fmt.Sprintf(
					"Name: %s\nDriver: %s\nMountpoint: %s",
					volumeItem.Volume.Name, volumeItem.Volume.Driver, volumeItem.Volume.Mountpoint,
				)
				model.viewport.SetContent(detailsContent)
			}
		}

		if !isKeyMessage || model.focusedView == focusDetails {
			updatedViewport, viewportCmd := model.viewport.Update(msg)
			model.viewport = updatedViewport
			cmds = append(cmds, viewportCmd)
		}
	}

	model.overlayModel.Foreground = model.foreground
	model.overlayModel.Background = model.list

	return model, tea.Batch(cmds...)
}

func (model *Model) handleToggleSelection() {
	currentIndex := model.list.Index()
	selectedItem, ok := model.list.SelectedItem().(VolumeItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			model.selectedVolumes.unselectVolumeInList(selectedItem.Volume.Name)
		} else {
			model.selectedVolumes.selectVolumeInList(selectedItem.Volume.Name, currentIndex)
		}

		selectedItem.isSelected = !isSelected
		model.list.SetItem(currentIndex, selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	allVolumesSelected := true
	items := model.list.Items()

	for _, item := range items {
		if volumeItem, ok := item.(VolumeItem); ok {
			if _, isSelected := model.selectedVolumes.selections[volumeItem.Volume.Name]; !isSelected {
				allVolumesSelected = false
				break
			}
		}
	}

	if allVolumesSelected {
		// Unselect all items.
		model.selectedVolumes = newSelectedVolumes()

		for index, item := range model.list.Items() {
			if volumeItem, ok := item.(VolumeItem); ok {
				volumeItem.isSelected = false
				model.list.SetItem(index, volumeItem)
			}
		}
	} else {
		// Select all items.
		model.selectedVolumes = newSelectedVolumes()

		for index, item := range model.list.Items() {
			if volumeItem, ok := item.(VolumeItem); ok {
				volumeItem.isSelected = true
				model.list.SetItem(index, volumeItem)
				model.selectedVolumes.selectVolumeInList(volumeItem.Volume.Name, index)
			}
		}
	}
}

func (model Model) View() string {
	if model.sessionState == viewOverlay && model.foreground != nil {
		return model.overlayModel.View()
	}

	layoutManager := shared.NewLayoutManager(model.WindowWidth, model.WindowHeight)
	_, detailLayout := layoutManager.CalculateMasterDetail(lipgloss.NewStyle())

	listView := model.style.Render(model.list.View())

	borderColor := colors.Muted()
	if model.focusedView == focusDetails {
		borderColor = colors.Primary()
	}

	detailStyle := lipgloss.NewStyle().
		Width(detailLayout.Width - 2).
		Height(detailLayout.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1)

	var detailContent string
	if model.list.SelectedItem() != nil {
		detailContent = model.viewport.View()
	} else {
		detailContent = lipgloss.NewStyle().Foreground(colors.Muted()).Render("No volume selected.")
	}

	detailView := detailStyle.Render(detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

func (model *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.WindowWidth = msg.Width
	model.WindowHeight = msg.Height

	layoutManager := shared.NewLayoutManager(msg.Width, msg.Height)
	masterLayout, detailLayout := layoutManager.CalculateMasterDetail(model.style)

	model.style = model.style.Width(masterLayout.Width).Height(masterLayout.Height)

	viewportWidth := detailLayout.Width - 4
	viewportHeight := detailLayout.Height - 2
	if viewportWidth < 0 {
		viewportWidth = 0
	}
	if viewportHeight < 0 {
		viewportHeight = 0
	}
	model.viewport.Width = viewportWidth
	model.viewport.Height = viewportHeight

	switch model.sessionState {
	case viewMain:
		if model.list.Width() != masterLayout.ContentWidth || model.list.Height() != masterLayout.ContentHeight {
			model.list.SetWidth(masterLayout.ContentWidth)
			model.list.SetHeight(masterLayout.ContentHeight)
		}
	case viewOverlay:
		if smartDialog, ok := model.foreground.(shared.SmartDialog); ok {
			smartDialog.UpdateWindowDimensions(msg)
			model.foreground = smartDialog
		}
	}
}

func (model Model) ShortHelp() []key.Binding {
	switch model.focusedView {
	case focusList:
		return model.list.ShortHelp()
	case focusDetails:
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
		}
	}
	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	switch model.focusedView {
	case focusList:
		return model.list.FullHelp()
	case focusDetails:
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
