// Package images defines the images component.
package images

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

const (
	focusList = iota
	focusDetails
)

// Model represents the images component state.
type Model struct {
	shared.Component
	style          lipgloss.Style
	list           list.Model
	viewport       viewport.Model
	selectedImages *selectedImages
	keybindings    *keybindings

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
	imageList, err := context.GetClient().GetImages()
	if err != nil {
		imageList = []client.Image{}
	}
	items := make([]list.Item, 0, len(imageList))
	for _, image := range imageList {
		items = append(items, ImageItem{Image: image})
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

	imageKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			imageKeybindings.toggleSelection,
			imageKeybindings.toggleSelectionOfAll,
			imageKeybindings.remove,
			imageKeybindings.switchTab,
		}
	}

	detailViewport := viewport.New(0, 0)

	model := Model{
		style:              style,
		list:               listModel,
		viewport:           detailViewport,
		selectedImages:     newSelectedImages(),
		keybindings:        imageKeybindings,
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
			if confirmMsg.Action.Type == "DeleteImage" {
				imageID := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveImage(imageID)
				if err != nil {
					// Error handling could be added here.
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
			updatedList, listCmd := model.list.Update(msg)
			model.list = updatedList
			cmds = append(cmds, listCmd)
		}

		selectedItem := model.list.SelectedItem()
		if selectedItem != nil {
			if imageItem, ok := selectedItem.(ImageItem); ok {
				detailsContent := fmt.Sprintf(
					"ID: %s\nSize: %d\nTags: %v",
					imageItem.Image.ID, imageItem.Image.Size, imageItem.Image.RepoTags,
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
	selectedItem, ok := model.list.SelectedItem().(ImageItem)
	if ok {
		isSelected := selectedItem.isSelected

		if isSelected {
			model.selectedImages.unselectImageInList(selectedItem.Image.ID)
		} else {
			model.selectedImages.selectImageInList(selectedItem.Image.ID, currentIndex)
		}

		selectedItem.isSelected = !isSelected
		model.list.SetItem(currentIndex, selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	allImagesSelected := true
	items := model.list.Items()

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

		for index, item := range model.list.Items() {
			if imageItem, ok := item.(ImageItem); ok {
				imageItem.isSelected = false
				model.list.SetItem(index, imageItem)
			}
		}
	} else {
		// Select all items.
		model.selectedImages = newSelectedImages()

		for index, item := range model.list.Items() {
			if imageItem, ok := item.(ImageItem); ok {
				imageItem.isSelected = true
				model.list.SetItem(index, imageItem)
				model.selectedImages.selectImageInList(imageItem.Image.ID, index)
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
		detailContent = lipgloss.NewStyle().Foreground(colors.Muted()).Render("No image selected.")
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
	if model.focusedView == focusList {
		return model.list.ShortHelp()
	} else if model.focusedView == focusDetails {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
		}
	}
	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	if model.focusedView == focusList {
		return model.list.FullHelp()
	} else if model.focusedView == focusDetails {
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
