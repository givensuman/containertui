// Package images defines the images component.
package images

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

// selectedImages maps an image's ID to its index in the list.
type selectedImages struct {
	selections map[string]int
}

func newSelectedImages() *selectedImages {
	return &selectedImages{
		selections: make(map[string]int),
	}
}

func (selectedImages *selectedImages) selectImageInList(id string, index int) {
	selectedImages.selections[id] = index
}

func (selectedImages selectedImages) unselectImageInList(id string) {
	delete(selectedImages.selections, id)
}

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

// Model represents the images component state.
type Model struct {
	shared.Component
	splitView      shared.SplitView
	selectedImages *selectedImages
	keybindings    *keybindings

	sessionState       sessionState
	detailsKeybindings detailsKeybindings
	foreground         tea.Model
	overlayModel       *overlay.Model
}

var (
	_ tea.Model             = (*Model)(nil)
	_ shared.ComponentModel = (*Model)(nil)
)

func New() Model {
	imageList, err := context.GetClient().GetImages()
	if err != nil {
		imageList = []client.Image{}
	}
	items := make([]list.Item, 0, len(imageList))
	for _, image := range imageList {
		items = append(items, ImageItem{Image: image})
	}

	width, height := context.GetWindowSize()

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

	imageKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			imageKeybindings.toggleSelection,
			imageKeybindings.toggleSelectionOfAll,
			imageKeybindings.remove,
			imageKeybindings.switchTab,
		}
	}

	splitView := shared.NewSplitView(listModel, shared.NewViewportPane())

	model := Model{
		splitView:          splitView,
		selectedImages:     newSelectedImages(),
		keybindings:        imageKeybindings,
		sessionState:       viewMain,
		detailsKeybindings: newDetailsKeybindings(),
	}

	model.overlayModel = overlay.New(nil, model.splitView.List, overlay.Center, overlay.Center, 0, 0)
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
			if confirmMsg.Action.Type == "DeleteImage" {
				imageID := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveImage(imageID)
				if err != nil {
					break
				}
			}
			model.sessionState = viewMain
			model.foreground = nil
		}
	case viewMain:
		// Forward message to SplitView first
		updatedSplitView, splitCmd := model.splitView.Update(msg)
		model.splitView = updatedSplitView
		cmds = append(cmds, splitCmd)

		if model.splitView.Focus == shared.FocusList {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				model.UpdateWindowDimensions(msg)
			case tea.KeyMsg:
				if model.splitView.List.FilterState() == list.Filtering {
					break
				}

				switch {
				case key.Matches(msg, model.keybindings.switchTab):
					// Handled by parent container (or ignored here if we want parent to switch tabs)
					// The generic tab switching logic for FOCUS is handled by SplitView.
					// The numeric keys for switching TABS are handled by the main UI loop usually,
					// but here we might need to bubble them up or just ignore them so they bubble.
					return model, nil
				case key.Matches(msg, model.keybindings.toggleSelection):
					model.handleToggleSelection()
				case key.Matches(msg, model.keybindings.toggleSelectionOfAll):
					model.handleToggleSelectionOfAll()
				case key.Matches(msg, model.keybindings.remove):
					selectedItem := model.splitView.List.SelectedItem()
					if selectedItem != nil {
						if imageItem, ok := selectedItem.(ImageItem); ok {
							containersUsingImage, _ := context.GetClient().GetContainersUsingImage(imageItem.Image.ID)
							if len(containersUsingImage) > 0 {
								warningDialog := shared.NewSmartDialog(
									fmt.Sprintf("Image %s is used by %d containers (%v).\nCannot delete.", imageItem.Image.ID[:12], len(containersUsingImage), containersUsingImage),
									[]shared.DialogButton{
										{Label: "OK", IsSafe: true},
									},
								)
								model.foreground = warningDialog
								model.sessionState = viewOverlay
							} else {
								confirmationDialog := shared.NewSmartDialog(
									fmt.Sprintf("Are you sure you want to delete image %s?", imageItem.Image.ID[:12]),
									[]shared.DialogButton{
										{Label: "Cancel", IsSafe: true},
										{Label: "Delete", IsSafe: false, Action: shared.SmartDialogAction{Type: "DeleteImage", Payload: imageItem.Image.ID}},
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
			if imageItem, ok := selectedItem.(ImageItem); ok {
				detailsContent := fmt.Sprintf(
					"ID: %s\nSize: %d\nTags: %v",
					imageItem.Image.ID, imageItem.Image.Size, imageItem.Image.RepoTags,
				)
				if pane, ok := model.splitView.Detail.(*shared.ViewportPane); ok {
					pane.SetContent(detailsContent)
				}
			}
		} else {
			if pane, ok := model.splitView.Detail.(*shared.ViewportPane); ok {
				pane.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No image selected."))
			}
		}
	}

	model.overlayModel.Foreground = model.foreground
	model.overlayModel.Background = model.splitView // SplitView implements View()

	return model, tea.Batch(cmds...)
}

func (model *Model) handleToggleSelection() {
	currentIndex := model.splitView.List.Index()
	selectedItem, ok := model.splitView.List.SelectedItem().(ImageItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			model.selectedImages.unselectImageInList(selectedItem.Image.ID)
		} else {
			model.selectedImages.selectImageInList(selectedItem.Image.ID, currentIndex)
		}

		selectedItem.isSelected = !isSelected
		model.splitView.List.SetItem(currentIndex, selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	allImagesSelected := true
	items := model.splitView.List.Items()

	for _, item := range items {
		if imageItem, ok := item.(ImageItem); ok {
			if _, isSelected := model.selectedImages.selections[imageItem.Image.ID]; !isSelected {
				allImagesSelected = false
				break
			}
		}
	}

	if allImagesSelected {
		// Unselect all items.
		model.selectedImages = newSelectedImages()

		for index, item := range model.splitView.List.Items() {
			if imageItem, ok := item.(ImageItem); ok {
				imageItem.isSelected = false
				model.splitView.List.SetItem(index, imageItem)
			}
		}
	} else {
		// Select all items.
		model.selectedImages = newSelectedImages()

		for index, item := range model.splitView.List.Items() {
			if imageItem, ok := item.(ImageItem); ok {
				imageItem.isSelected = true
				model.splitView.List.SetItem(index, imageItem)
				model.selectedImages.selectImageInList(imageItem.Image.ID, index)
			}
		}
	}
}

func (model Model) View() string {
	if model.sessionState == viewOverlay && model.foreground != nil {
		model.overlayModel.Background = shared.SimpleViewModel{Content: model.splitView.View()}
		return model.overlayModel.View()
	}

	return model.splitView.View()
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
