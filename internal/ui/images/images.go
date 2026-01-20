// Package images defines the images component.
package images

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

// MsgImageInspection contains the inspection data for an image.
type MsgImageInspection struct {
	ID    string
	Image types.ImageInspect
	Err   error
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

// Model represents the images component state.
type Model struct {
	components.ResourceView[string, ImageItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	currentImageID     string
	inspection         types.ImageInspect
	scrollPositions    map[string]int
	currentFormat      string
}

func New() Model {
	imageKeybindings := newKeybindings()

	fetchImages := func() ([]ImageItem, error) {
		imageList, err := context.GetClient().GetImages()
		if err != nil {
			return nil, err
		}
		items := make([]ImageItem, 0, len(imageList))
		for _, image := range imageList {
			items = append(items, ImageItem{Image: image})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, ImageItem](
		"Images",
		fetchImages,
		func(item ImageItem) string { return item.Image.ID },
		func(item ImageItem) string { return item.Title() },
		func(w, h int) {
			// Window resize handled by base component
		},
	)

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        imageKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		scrollPositions:    make(map[string]int),
		currentFormat:      "",
	}

	// Add custom keybindings to help
	model.ResourceView.AdditionalHelp = []key.Binding{
		imageKeybindings.toggleSelection,
		imageKeybindings.toggleSelectionOfAll,
		imageKeybindings.remove,
		imageKeybindings.pullImage,
		imageKeybindings.createContainer,
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
	case MsgImageInspection:
		if msg.ID == model.currentImageID && msg.Err == nil {
			model.inspection = msg.Image
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.restoreScrollPosition()

	case MsgPullComplete:
		if msg.Err != nil {
			// Show error dialog
			errorDialog := components.NewSmartDialog(
				fmt.Sprintf("Failed to pull image:\n\n%v", msg.Err),
				[]components.DialogButton{{Label: "OK", IsSafe: true}},
			)
			model.ResourceView.SetOverlay(errorDialog)
		} else {
			// Success - close dialog and refresh list
			model.ResourceView.CloseOverlay()
			// Trigger images refresh
			return model, func() tea.Msg {
				return MsgRefreshImages{}
			}
		}
		return model, nil
	case MsgRefreshImages:
		// Refresh the images list via ResourceView
		return model, model.ResourceView.Refresh()

	case MsgCreateContainerComplete:
		if msg.Err != nil {
			// Show error dialog
			errorDialog := components.NewSmartDialog(
				fmt.Sprintf("Failed to create container:\n\n%v", msg.Err),
				[]components.DialogButton{{Label: "OK", IsSafe: true}},
			)
			model.ResourceView.SetOverlay(errorDialog)
		} else {
			// Success - show success message
			successDialog := components.NewSmartDialog(
				fmt.Sprintf("Container created successfully!\n\nContainer ID: %s", msg.ContainerID[:12]),
				[]components.DialogButton{{Label: "OK", IsSafe: true}},
			)
			model.ResourceView.SetOverlay(successDialog)
		}
		return model, nil
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.ResourceView.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.ConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteImage" {
				imageID := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveImage(imageID)
				if err == nil {
					// We need to refresh list.
					// ResourceView has Refresh() command but for immediate feedback we might want local update.
					// Refresh() is safer.
					return model, model.ResourceView.Refresh()
				} else {
					// Show error
					errorDialog := components.NewSmartDialog(
						fmt.Sprintf("Failed to remove image:\n\n%v", err),
						[]components.DialogButton{{Label: "OK", IsSafe: true}},
					)
					model.ResourceView.SetOverlay(errorDialog)
					return model, nil
				}
			} else if confirmMsg.Action.Type == "PullImageAction" {
				// Extract image name from form values
				payload := confirmMsg.Action.Payload.(map[string]interface{})
				formValues := payload["values"].(map[string]string)
				imageName := formValues["Image"]

				// Show progress dialog
				progressDialog := components.NewSmartDialog(
					fmt.Sprintf("Pulling image: %s\n\nThis may take a few moments...", imageName),
					[]components.DialogButton{}, // No buttons while pulling
				)
				model.ResourceView.SetOverlay(progressDialog)

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
				progressDialog := components.NewSmartDialog(
					"Creating container...",
					[]components.DialogButton{}, // No buttons while creating
				)
				model.ResourceView.SetOverlay(progressDialog)

				// Create container
				return model, func() tea.Msg {
					containerID, err := context.GetClient().CreateContainer(config)
					if err != nil {
						return MsgCreateContainerComplete{Err: err}
					}
					return MsgCreateContainerComplete{ContainerID: containerID, Err: nil}
				}
			}

			model.ResourceView.CloseOverlay()
			return model, nil
		}

		// Let ResourceView handle forwarding to overlay
		return model, tea.Batch(cmds...)
	}

	// 4. Main View Logic
	if model.ResourceView.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.ResourceView.IsFiltering() {
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

			case key.Matches(msg, model.keybindings.pullImage):
				// Show form dialog to get image name
				formDialog := components.NewFormDialog(
					"Pull Image",
					[]components.FormField{
						{
							Label:       "Image",
							Placeholder: "nginx:latest",
							Required:    true,
							Validator:   validateImageName,
						},
					},
					base.SmartDialogAction{Type: "PullImageAction"},
					nil,
				)
				model.ResourceView.SetOverlay(formDialog)

			case key.Matches(msg, model.keybindings.createContainer):
				selectedItem := model.ResourceView.GetSelectedItem()
				if selectedItem != nil {
					// Show form dialog to create container
					formDialog := components.NewFormDialog(
						"Create Container from Image",
						[]components.FormField{
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
						base.SmartDialogAction{
							Type:    "CreateContainerAction",
							Payload: map[string]interface{}{"imageID": selectedItem.Image.ID},
						},
						nil,
					)
					model.ResourceView.SetOverlay(formDialog)
				}

			case key.Matches(msg, model.keybindings.remove):
				model.handleRemove()
				return model, nil
			}
		}
	} else {
		// Detail pane is focused
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
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

	// 5. Update Detail Content
	detailCmd := model.updateDetailContent()
	if detailCmd != nil {
		cmds = append(cmds, detailCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	return model.ResourceView.View()
}

func (model *Model) handleToggleSelection() {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem != nil {
		model.ResourceView.ToggleSelection(selectedItem.Image.ID)

		// Update visual state
		index := model.ResourceView.GetSelectedIndex()
		selectedItem.isSelected = !selectedItem.isSelected
		model.ResourceView.SetItem(index, *selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	// Similar logic to container selection toggling
	// If any item is not selected, select all. Otherwise deselect all.

	items := model.ResourceView.GetItems()
	selectedIDs := model.ResourceView.GetSelectedIDs()

	shouldSelectAll := false
	for _, item := range items {
		found := false
		for _, id := range selectedIDs {
			if id == item.Image.ID {
				found = true
				break
			}
		}
		if !found {
			shouldSelectAll = true
			break
		}
	}

	if shouldSelectAll {
		// Select all
		for i, item := range items {
			found := false
			for _, id := range selectedIDs {
				if id == item.Image.ID {
					found = true
					break
				}
			}
			if !found {
				model.ResourceView.ToggleSelection(item.Image.ID)
			}
			item.isSelected = true
			model.ResourceView.SetItem(i, item)
		}
	} else {
		// Deselect all
		for i, item := range items {
			model.ResourceView.ToggleSelection(item.Image.ID)
			item.isSelected = false
			model.ResourceView.SetItem(i, item)
		}
	}
}

func (model *Model) handleRemove() {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	containersUsingImage, _ := context.GetClient().GetContainersUsingImage(selectedItem.Image.ID)
	if len(containersUsingImage) > 0 {
		warningDialog := components.NewSmartDialog(
			fmt.Sprintf("Image %s is used by %d containers (%v).\nCannot delete.", selectedItem.Image.ID[:12], len(containersUsingImage), containersUsingImage),
			[]components.DialogButton{
				{Label: "OK", IsSafe: true},
			},
		)
		model.ResourceView.SetOverlay(warningDialog)
	} else {
		confirmationDialog := components.NewSmartDialog(
			fmt.Sprintf("Are you sure you want to delete image %s?", selectedItem.Image.ID[:12]),
			[]components.DialogButton{
				{Label: "Cancel", IsSafe: true},
				{Label: "Delete", IsSafe: false, Action: base.SmartDialogAction{Type: "DeleteImage", Payload: selectedItem.Image.ID}},
			},
		)
		model.ResourceView.SetOverlay(confirmationDialog)
	}
}

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem == nil {
		model.ResourceView.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No image selected."))
		return nil
	}

	imageID := selectedItem.Image.ID
	// If we've switched to a different image, OR we don't have inspection data yet, fetch it
	if imageID != model.currentImageID || model.inspection.ID == "" {
		// Save scroll position of previous image
		if model.currentImageID != "" && model.currentImageID != imageID {
			model.saveScrollPosition()
		}

		model.currentImageID = imageID
		// Fetch inspection data asynchronously
		return func() tea.Msg {
			imageInfo, err := context.GetClient().InspectImage(imageID)
			return MsgImageInspection{ID: imageID, Image: imageInfo, Err: err}
		}
	}

	return nil
}

// saveScrollPosition saves the current viewport scroll position for the current image
func (model *Model) saveScrollPosition() {
	if model.currentImageID != "" {
		if vp := model.getViewport(); vp != nil {
			model.scrollPositions[model.currentImageID] = vp.YOffset()
		}
	}
}

// restoreScrollPosition restores the viewport scroll position for the current image
func (model *Model) restoreScrollPosition() {
	if model.currentImageID != "" {
		if vp := model.getViewport(); vp != nil {
			if offset, exists := model.scrollPositions[model.currentImageID]; exists {
				vp.SetYOffset(offset)
			} else {
				vp.SetYOffset(0) // New image, start at top
			}
		}
	}
}

// getViewport returns the viewport from the detail pane if available
func (model *Model) getViewport() *viewport.Model {
	if vp, ok := model.ResourceView.SplitView.Detail.(*components.ViewportPane); ok {
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
	content := builders.BuildImagePanel(model.inspection, model.ResourceView.GetContentWidth(), format)
	model.ResourceView.SetContent(content)
}

// handleCopyToClipboard copies the current inspection output to clipboard
func (model *Model) handleCopyToClipboard() tea.Cmd {
	if model.inspection.ID == "" {
		return nil
	}

	// Determine which format to use
	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	// Marshal the data without syntax highlighting
	data, err := infopanel.MarshalToFormat(model.inspection, format)
	if err != nil {
		return notifications.ShowError(err)
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(string(data)); err != nil {
		return notifications.ShowError(err)
	}

	return notifications.ShowSuccess("Copied to clipboard")
}

// handleToggleFormat toggles between JSON and YAML format
func (model *Model) handleToggleFormat() tea.Cmd {
	// Determine current effective format
	currentFormat := model.currentFormat
	if currentFormat == "" {
		cfg := context.GetConfig()
		currentFormat = cfg.InspectionFormat
		if currentFormat == "" {
			currentFormat = "yaml"
		}
	}

	// Toggle to the opposite format
	if currentFormat == "json" {
		model.currentFormat = "yaml"
	} else {
		model.currentFormat = "json"
	}

	// Refresh content with new format
	model.refreshInspectionContent()

	return notifications.ShowSuccess("Switched to " + model.currentFormat)
}

func (model Model) ShortHelp() []key.Binding {
	// If detail pane is focused, show detail keybindings
	if !model.ResourceView.IsListFocused() {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
			model.detailsKeybindings.ToggleJSON,
			model.detailsKeybindings.CopyOutput,
		}
	}
	return model.ResourceView.ShortHelp()
}

func (model Model) FullHelp() [][]key.Binding {
	// If detail pane is focused, show detail keybindings
	if !model.ResourceView.IsListFocused() {
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
	}
	return model.ResourceView.FullHelp()
}
