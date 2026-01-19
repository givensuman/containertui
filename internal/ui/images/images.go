// Package images defines the images component.
package images

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
)

// MsgPullProgress contains progress information from image pull.
type MsgPullProgress struct {
	Message string
}

// MsgPullComplete indicates the image pull has finished.
type MsgPullComplete struct {
	ImageName string
	Err       error
}

// MsgRefreshImages triggers a refresh of the images list.
type MsgRefreshImages struct{}

// MsgCreateContainerComplete indicates container creation has finished.
type MsgCreateContainerComplete struct {
	ContainerID string
	Err         error
}

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
	pullImage            key.Binding
	createContainer      key.Binding
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
		pullImage: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "pull image"),
		),
		createContainer: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create container"),
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

// validateImageName validates that an image name is not empty.
func validateImageName(input string) error {
	if input == "" {
		return fmt.Errorf("image name cannot be empty")
	}
	return nil
}

// validatePorts validates port mapping format (e.g., "8080:80,443:443").
func validatePorts(input string) error {
	if input == "" {
		return nil // Optional field
	}
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format, expected hostPort:containerPort")
		}
	}
	return nil
}

// validateVolumes validates volume mapping format (e.g., "/host:/container,vol:/data").
func validateVolumes(input string) error {
	if input == "" {
		return nil // Optional field
	}
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format, expected hostPath:containerPath")
		}
	}
	return nil
}

// validateEnv validates environment variable format (e.g., "KEY=value,FOO=bar").
func validateEnv(input string) error {
	if input == "" {
		return nil // Optional field
	}
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		if !strings.Contains(pair, "=") {
			return fmt.Errorf("invalid format, expected KEY=value")
		}
	}
	return nil
}

// validateBool validates yes/no input.
func validateBool(input string) error {
	if input == "" {
		return nil // Optional, defaults to no
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	if lower != "yes" && lower != "no" {
		return fmt.Errorf("expected 'yes' or 'no'")
	}
	return nil
}

// parsePorts parses port string into map.
func parsePorts(input string) map[string]string {
	result := make(map[string]string)
	if input == "" {
		return result
	}
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

// parseVolumes parses volume string into slice.
func parseVolumes(input string) []string {
	if input == "" {
		return []string{}
	}
	pairs := strings.Split(input, ",")
	result := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		trimmed := strings.TrimSpace(pair)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseEnv parses environment variable string into slice.
func parseEnv(input string) []string {
	if input == "" {
		return []string{}
	}
	pairs := strings.Split(input, ",")
	result := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		trimmed := strings.TrimSpace(pair)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseBool parses yes/no into boolean.
func parseBool(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	return lower == "yes"
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
	foreground         interface{} // Can be SmartDialog, FormDialog, etc.
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
	listModel.Styles.Filter.Focused.Prompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.Filter.Cursor.Color = colors.Primary()
	// listModel.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	// listModel.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	imageKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			imageKeybindings.toggleSelection,
			imageKeybindings.toggleSelectionOfAll,
			imageKeybindings.remove,
			imageKeybindings.pullImage,
			imageKeybindings.createContainer,
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

	return model
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global messages first
	switch msg := msg.(type) {
	case MsgPullComplete:
		if msg.Err != nil {
			// Show error dialog
			errorDialog := shared.NewSmartDialog(
				fmt.Sprintf("Failed to pull image:\n\n%v", msg.Err),
				[]shared.DialogButton{{Label: "OK", IsSafe: true}},
			)
			model.foreground = errorDialog
			model.sessionState = viewOverlay
		} else {
			// Success - close dialog and refresh list
			model.sessionState = viewMain
			model.foreground = nil

			// Trigger images refresh
			return model, func() tea.Msg {
				return MsgRefreshImages{}
			}
		}
		return model, nil
	case MsgRefreshImages:
		// Refresh the images list
		imageList, err := context.GetClient().GetImages()
		if err == nil {
			items := make([]list.Item, 0, len(imageList))
			for _, image := range imageList {
				items = append(items, ImageItem{Image: image})
			}
			model.splitView.List.SetItems(items)
		}
		return model, nil
	case MsgCreateContainerComplete:
		if msg.Err != nil {
			// Show error dialog
			errorDialog := shared.NewSmartDialog(
				fmt.Sprintf("Failed to create container:\n\n%v", msg.Err),
				[]shared.DialogButton{{Label: "OK", IsSafe: true}},
			)
			model.foreground = errorDialog
			model.sessionState = viewOverlay
		} else {
			// Success - show success message
			successDialog := shared.NewSmartDialog(
				fmt.Sprintf("Container created successfully!\n\nContainer ID: %s", msg.ContainerID[:12]),
				[]shared.DialogButton{{Label: "OK", IsSafe: true}},
			)
			model.foreground = successDialog
			model.sessionState = viewOverlay
		}
		return model, nil
	}

	switch model.sessionState {
	case viewOverlay:
		if model.foreground != nil {
			switch fg := model.foreground.(type) {
			case shared.SmartDialog:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			case shared.FormDialog:
				updated, cmd := fg.Update(msg)
				model.foreground = updated
				cmds = append(cmds, cmd)
			}
		}

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
			} else if confirmMsg.Action.Type == "PullImageAction" {
				// Extract image name from form values
				payload := confirmMsg.Action.Payload.(map[string]interface{})
				formValues := payload["values"].(map[string]string)
				imageName := formValues["Image"]

				// Show progress dialog
				progressDialog := shared.NewSmartDialog(
					fmt.Sprintf("Pulling image: %s\n\nThis may take a few moments...", imageName),
					[]shared.DialogButton{}, // No buttons while pulling
				)
				model.foreground = progressDialog

				// Start pull in goroutine
				return model, func() tea.Msg {
					err := context.GetClient().PullImage(imageName, nil)
					return MsgPullComplete{ImageName: imageName, Err: err}
				}
			} else if confirmMsg.Action.Type == "CreateContainerAction" {
				// Extract form values and image ID
				payload := confirmMsg.Action.Payload.(map[string]interface{})
				imageID := payload["imageID"].(string)
				formValues := payload["values"].(map[string]string)

				// Parse form values
				ports := parsePorts(formValues["Ports"])
				volumes := parseVolumes(formValues["Volumes"])
				env := parseEnv(formValues["Environment"])
				autoStart := parseBool(formValues["Auto-start"])

				// Create container config
				config := client.CreateContainerConfig{
					Name:      formValues["Name"],
					ImageID:   imageID,
					Ports:     ports,
					Volumes:   volumes,
					Env:       env,
					AutoStart: autoStart,
					Network:   "bridge",
				}

				// Show progress dialog
				progressDialog := shared.NewSmartDialog(
					"Creating container...",
					[]shared.DialogButton{}, // No buttons while creating
				)
				model.foreground = progressDialog

				// Create container
				return model, func() tea.Msg {
					containerID, err := context.GetClient().CreateContainer(config)
					if err != nil {
						return MsgCreateContainerComplete{Err: err}
					}
					return MsgCreateContainerComplete{ContainerID: containerID, Err: nil}
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
			case tea.KeyPressMsg:
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
				case key.Matches(msg, model.keybindings.pullImage):
					// Show form dialog to get image name
					formDialog := shared.NewFormDialog(
						"Pull Image",
						[]shared.FormField{
							{
								Label:       "Image",
								Placeholder: "nginx:latest",
								Required:    true,
								Validator:   validateImageName,
							},
						},
						"PullImageAction",
						nil,
					)
					model.foreground = formDialog
					model.sessionState = viewOverlay
				case key.Matches(msg, model.keybindings.createContainer):
					selectedItem := model.splitView.List.SelectedItem()
					if selectedItem != nil {
						if imageItem, ok := selectedItem.(ImageItem); ok {
							// Show form dialog to create container
							formDialog := shared.NewFormDialog(
								"Create Container from Image",
								[]shared.FormField{
									{
										Label:       "Name",
										Placeholder: "my-container (optional)",
										Required:    false,
									},
									{
										Label:       "Ports",
										Placeholder: "8080:80,443:443",
										Required:    false,
										Validator:   validatePorts,
									},
									{
										Label:       "Volumes",
										Placeholder: "/host:/container",
										Required:    false,
										Validator:   validateVolumes,
									},
									{
										Label:       "Environment",
										Placeholder: "KEY=value,FOO=bar",
										Required:    false,
										Validator:   validateEnv,
									},
									{
										Label:       "Auto-start",
										Placeholder: "yes/no",
										Required:    false,
										Validator:   validateBool,
									},
								},
								"CreateContainerAction",
								map[string]interface{}{"imageID": imageItem.Image.ID},
							)
							model.foreground = formDialog
							model.sessionState = viewOverlay
						}
					}
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

func (model Model) View() tea.View {
	if model.sessionState == viewOverlay && model.foreground != nil {
		var fgView string
		switch fg := model.foreground.(type) {
		case shared.SmartDialog:
			fgView = fg.View()
		case shared.FormDialog:
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
