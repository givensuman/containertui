// Package images defines the images component.
package images

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/utils"
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
	ImageName    string
	Message      string
	ProgressChan <-chan string
	DoneChan     <-chan error
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

// MsgRunAndExecComplete indicates run-and-exec flow has finished.
type MsgRunAndExecComplete struct {
	ContainerID string
	Err         error
}

// MsgTagImageComplete indicates image tagging has finished.
type MsgTagImageComplete struct {
	ImageID string
	NewTag  string
	Err     error
}

// MsgBuildImageComplete indicates image build has finished.
type MsgBuildImageComplete struct {
	Tag string
	Err error
}

// MsgPruneComplete is sent when the prune operation completes
type MsgPruneComplete struct {
	SpaceReclaimed uint64
	Err            error
}

type keybindings struct {
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	remove               key.Binding
	pruneImages          key.Binding
	tagImage             key.Binding
	runAndExec           key.Binding
	buildImage           key.Binding
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
		pruneImages: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prune unused"),
		),
		tagImage: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tag image"),
		),
		runAndExec: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "run & exec"),
		),
		buildImage: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "build image"),
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
			key.WithKeys("1", "2", "3", "4", "5", "6"),
			key.WithHelp("1-6", "switch tab"),
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
	inspection         types.ImageInspect
	detailsPanel       components.DetailsPanel
}

func New() Model {
	imageKeybindings := newKeybindings()

	fetchImages := func() ([]ImageItem, error) {
		imageList, err := state.GetClient().GetImages(stdcontext.Background())
		if err != nil {
			return nil, err
		}

		// Get all containers to determine which images are in use
		containers, err := state.GetClient().GetContainers(stdcontext.Background())
		if err != nil {
			return nil, err
		}

		// Build a map of image IDs that are in use
		inUseImages := make(map[string]bool)
		for _, container := range containers {
			// Mark by image name (e.g., "nginx:latest") and ID
			inUseImages[container.Image] = true
		}

		items := make([]ImageItem, 0, len(imageList))
		for _, image := range imageList {
			// Check if this image is in use by checking its ID or any of its tags
			inUse := inUseImages[image.ID]
			if !inUse {
				// Also check if any repo tags match
				for _, tag := range image.RepoTags {
					if inUseImages[tag] {
						inUse = true
						break
					}
				}
			}
			items = append(items, ImageItem{
				Image: image,
				InUse: inUse,
			})
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

	// Add extra pane below detail pane
	extraPane := components.NewViewportPane()
	extraPane.SetContent("")                            // Will be populated when an image is selected
	resourceView.SplitView.SetExtraPane(extraPane, 0.3) // 30% of height

	// Set titles for the panes
	resourceView.SplitView.SetDetailTitle("Inspect")
	resourceView.SplitView.SetExtraTitle("Used By")

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        imageKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		inspection:         types.ImageInspect{},
		detailsPanel:       components.NewDetailsPanel(),
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		imageKeybindings.switchTab,
		imageKeybindings.remove,
		imageKeybindings.pruneImages,
		imageKeybindings.tagImage,
		imageKeybindings.runAndExec,
		imageKeybindings.buildImage,
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
		if msg.ID == model.detailsPanel.GetCurrentID() && msg.Err == nil {
			model.inspection = msg.Image
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.detailsPanel.RestoreScrollPosition(model.getViewport())
		return model, nil

	case MsgPullComplete:
		if msg.Err != nil {
			// Show error dialog
			errorDialog := components.NewDialog(
				fmt.Sprintf("Failed to pull image:\n\n%v", msg.Err),
				[]components.DialogButton{{Label: "OK"}},
			)
			model.SetOverlay(errorDialog)
		} else {
			// Success - close dialog and refresh list
			model.CloseOverlay()
			// Trigger images refresh
			return model, func() tea.Msg {
				return MsgRefreshImages{}
			}
		}
		return model, nil

	case MsgPullProgress:
		if progressDialog, ok := model.Foreground.(components.ProgressDialog); ok {
			progressDialog.SetStatus(msg.Message)
			model.Foreground = progressDialog
		}

		if msg.ProgressChan != nil && msg.DoneChan != nil {
			return model, listenToPullProgress(msg.ImageName, msg.ProgressChan, msg.DoneChan)
		}
		return model, nil
	case MsgRefreshImages:
		// Refresh the images list via ResourceView
		return model, model.Refresh()

	case MsgCreateContainerComplete:
		// Close the progress dialog
		model.CloseOverlay()

		if msg.Err != nil {
			// Show error notification
			return model, notifications.ShowError(msg.Err)
		}

		// Success - show success notification
		successMsg := fmt.Sprintf("Container created: %s", msg.ContainerID[:12])
		// Emit container created message to trigger refresh
		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
			func() tea.Msg {
				return base.MsgContainerCreated{ContainerID: msg.ContainerID}
			},
		)

	case MsgRunAndExecComplete:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}

		successMsg := fmt.Sprintf("Created ephemeral container: %s", msg.ContainerID[:12])
		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceContainer,
					Operation: base.OperationCreated,
					IDs:       []string{msg.ContainerID},
				}
			},
		)

	case MsgTagImageComplete:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}

		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Tagged image: %s", msg.NewTag)),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceImage,
					Operation: base.OperationUpdated,
					IDs:       []string{msg.ImageID},
				}
			},
		)

	case MsgBuildImageComplete:
		if model.IsOverlayVisible() {
			model.CloseOverlay()
		}

		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}

		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Built image: %s", msg.Tag)),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceImage,
					Operation: base.OperationCreated,
					IDs:       []string{msg.Tag},
				}
			},
		)

	case MsgPruneComplete:
		model.CloseOverlay()
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		successMsg := fmt.Sprintf("Pruned unused images, freed %s", utils.HumanizeBytes(msg.SpaceReclaimed))
		if msg.SpaceReclaimed == 0 {
			successMsg = "No unused images to prune"
		}
		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
			model.Refresh(),
			func() tea.Msg {
				return base.MsgResourceChanged{
					Resource:  base.ResourceImage,
					Operation: base.OperationPruned,
					IDs:       nil,
					Metadata: map[string]any{
						"spaceReclaimed": msg.SpaceReclaimed,
					},
				}
			},
		)
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			switch confirmMsg.Action.Type {
			case "DeleteImage":
				imageID := confirmMsg.Action.Payload.(string)
				err := state.GetClient().RemoveImage(stdcontext.Background(), imageID, false)
				if err == nil {
					model.CloseOverlay()
					return model, tea.Batch(
						notifications.ShowSuccess(fmt.Sprintf("Image removed: %s", imageID[:12])),
						model.Refresh(),
					)
				} else {
					// Show error notification
					model.CloseOverlay()
					return model, notifications.ShowError(err)
				}
			case "ForceDeleteImage":
				imageID := confirmMsg.Action.Payload.(string)
				err := state.GetClient().RemoveImage(stdcontext.Background(), imageID, true)
				model.CloseOverlay()
				if err != nil {
					return model, notifications.ShowError(fmt.Errorf("failed to force delete image: %w", err))
				}
				return model, tea.Batch(
					notifications.ShowSuccess(fmt.Sprintf("Force deleted image: %s", imageID[:12])),
					model.Refresh(),
				)
			case "PullImageAction":
				// Extract image name from form values
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				formValues, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}
				imageName := formValues["Image"]
				if imageName == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("image name is required"))
				}

				// Show progress dialog
				progressDialog := components.NewProgressDialogWithBar(fmt.Sprintf("Pulling image: %s", imageName))
				progressDialog.SetStatus("Preparing pull...")
				model.SetOverlay(progressDialog)

				progressChan := make(chan string, 100)
				doneChan := make(chan error, 1)

				go func() {
					err := state.GetClient().PullImage(stdcontext.Background(), imageName, progressChan)
					doneChan <- err
					close(doneChan)
				}()

				return model, listenToPullProgress(imageName, progressChan, doneChan)
			case "CreateContainerAction":
				// Extract form values and image ID
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				imageID, ok := payload["imageID"].(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid image ID"))
				}
				formValues, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}

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

				// Close the form overlay and show progress dialog
				model.CloseOverlay()

				// Show progress dialog (like image pull)
				progressDialog := components.NewProgressDialog("Creating container...\n\nThis may take a few moments...")
				model.SetOverlay(progressDialog)

				// Create container
				return model, func() tea.Msg {
					containerID, err := state.GetClient().CreateContainer(stdcontext.Background(), config)
					if err != nil {
						return MsgCreateContainerComplete{Err: err}
					}
					return MsgCreateContainerComplete{ContainerID: containerID, Err: nil}
				}
			case "TagImageAction":
				// Extract form values and image ID from metadata
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				metadata, ok := payload["metadata"].(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid metadata"))
				}
				imageID, ok := metadata["imageID"].(string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid image ID"))
				}
				formValues, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}

				newTag := formValues["New Tag"]
				if newTag == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("tag is required"))
				}

				model.CloseOverlay()
				return model, model.performTagImage(imageID, newTag)

			case "BuildImageAction":
				// Extract form values
				payload, ok := confirmMsg.Action.Payload.(map[string]any)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid payload type"))
				}
				formValues, ok := payload["values"].(map[string]string)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid form values"))
				}

				dockerfile := formValues["Dockerfile Path"]
				tag := formValues["Tag"]
				contextPath := formValues["Build Context"]
				buildArgs := formValues["Build Args"]

				if dockerfile == "" || tag == "" || contextPath == "" {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("dockerfile, tag, and context are required"))
				}

				// Parse build args if provided
				var buildArgsMap map[string]string
				if buildArgs != "" {
					buildArgsMap = make(map[string]string)
					pairs := strings.Split(buildArgs, ",")
					for _, pair := range pairs {
						kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
						if len(kv) == 2 {
							buildArgsMap[kv[0]] = kv[1]
						}
					}
				}

				model.CloseOverlay()
				progressDialog := components.NewProgressDialog("Building image...\n\nThis may take a few moments...")
				model.SetOverlay(progressDialog)
				return model, model.performBuildImage(dockerfile, tag, contextPath, buildArgsMap)
			}

			model.CloseOverlay()
			return model, nil
		}

		// Let ResourceView handle forwarding to overlay
		return model, tea.Batch(cmds...)
	}

	// 4. Main View Logic
	if model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.IsFiltering() {
				break
			}

			switch {
			case key.Matches(msg, model.keybindings.switchTab):
				return model, tea.Batch(cmds...) // Handled by parent

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
				model.SetOverlay(formDialog)

			case key.Matches(msg, model.keybindings.createContainer):
				selectedItem := model.GetSelectedItem()
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
							Payload: map[string]any{"imageID": selectedItem.Image.ID},
						},
						nil,
					)
					model.SetOverlay(formDialog)
				}

			case key.Matches(msg, model.keybindings.remove):
				model.handleRemove()
				return model, nil

			case key.Matches(msg, model.keybindings.pruneImages):
				if cmd := model.handlePruneImages(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)

			case key.Matches(msg, model.keybindings.tagImage):
				model.handleTagImage()
				return model, nil

			case key.Matches(msg, model.keybindings.runAndExec):
				if cmd := model.handleRunAndExec(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)

			case key.Matches(msg, model.keybindings.buildImage):
				model.handleBuildImage()
				return model, nil
			}
		}
	} else {
		// Detail or extra pane is focused
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			// Only handle these actions when detail pane is focused (not extra)
			if model.IsDetailFocused() {
				if key.Matches(msg, model.detailsKeybindings.ToggleJSON) {
					newFormat, cmd := model.detailsPanel.HandleToggleFormat()
					_ = newFormat // format is tracked internally
					model.refreshInspectionContent()
					cmds = append(cmds, cmd)
				}
				if key.Matches(msg, model.detailsKeybindings.CopyOutput) {
					cmd := model.detailsPanel.HandleCopyToClipboard(model.inspection)
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

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model *Model) handleToggleSelection() {
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil {
		model.ToggleSelection(selectedItem.Image.ID)

		// Update visual state
		index := model.GetSelectedIndex()
		selectedItem.isSelected = !selectedItem.isSelected
		model.SetItem(index, *selectedItem)
	}
}

func (model *Model) handleToggleSelectionOfAll() {
	// Similar logic to container selection toggling
	// If any item is not selected, select all. Otherwise deselect all.

	items := model.GetItems()
	selectedIDs := model.GetSelectedIDs()

	shouldSelectAll := false
	for _, item := range items {
		if !slices.Contains(selectedIDs, item.Image.ID) {
			shouldSelectAll = true
			break
		}
	}

	if shouldSelectAll {
		// Select all
		for i, item := range items {
			if !slices.Contains(selectedIDs, item.Image.ID) {
				model.ToggleSelection(item.Image.ID)
			}
			item.isSelected = true
			model.SetItem(i, item)
		}
	} else {
		// Deselect all
		for i, item := range items {
			model.ToggleSelection(item.Image.ID)
			item.isSelected = false
			model.SetItem(i, item)
		}
	}
}

func (model *Model) handleRemove() {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	containersUsingImage, err := state.GetClient().GetContainersUsingImage(stdcontext.Background(), selectedItem.Image.ID)
	if err != nil {
		// If we can't check usage, show error and don't proceed with deletion
		errorDialog := components.NewDialog(
			fmt.Sprintf("Failed to check image usage: %v\nCannot safely delete image.", err),
			[]components.DialogButton{
				{Label: "OK"},
			},
		)
		model.SetOverlay(errorDialog)
		return
	}
	if len(containersUsingImage) > 0 {
		confirmationDialog := components.NewDialog(
			fmt.Sprintf("Image %s is used by %d containers (%v).\n\nForce delete anyway?", selectedItem.Image.ID[:12], len(containersUsingImage), containersUsingImage),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Force Delete", Action: base.SmartDialogAction{Type: "ForceDeleteImage", Payload: selectedItem.Image.ID}},
			},
		)
		model.SetOverlay(confirmationDialog)
	} else {
		confirmationDialog := components.NewDialog(
			fmt.Sprintf("Are you sure you want to delete image %s?", selectedItem.Image.ID[:12]),
			[]components.DialogButton{
				{Label: "Cancel"},
				{Label: "Delete", Action: base.SmartDialogAction{Type: "DeleteImage", Payload: selectedItem.Image.ID}},
			},
		)
		model.SetOverlay(confirmationDialog)
	}
}

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No image selected"))
		model.SetExtraContent("") // Clear extra pane when no image selected
		return nil
	}

	imageID := selectedItem.Image.ID
	currentID := model.detailsPanel.GetCurrentID()

	// If we've switched to a different image, OR we don't have inspection data yet, fetch it
	if imageID != currentID || model.inspection.ID == "" {
		// SetCurrentID will save scroll position for previous ID
		model.detailsPanel.SetCurrentID(imageID, model.getViewport())

		// Fetch inspection data asynchronously
		return func() tea.Msg {
			imageInfo, err := state.GetClient().InspectImage(stdcontext.Background(), imageID)
			return MsgImageInspection{ID: imageID, Image: imageInfo, Err: err}
		}
	}

	return nil
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
	// Use DetailsPanel to get the current format
	format := model.detailsPanel.GetFormatForDisplay()

	// Build content with current format
	content := builders.BuildImagePanel(model.inspection, model.GetContentWidth(), format)
	model.SetContent(content)

	// Update "Used By" panel
	model.updateUsedByPanel()
}

// updateUsedByPanel updates the extra pane with containers using this image
func (model *Model) updateUsedByPanel() {
	if model.inspection.ID == "" {
		model.SetExtraContent("")
		return
	}

	// Fetch containers using this image
	usedBy, err := state.GetClient().GetContainersUsingImage(stdcontext.Background(), model.inspection.ID)
	if err != nil {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	if len(usedBy) == 0 {
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No containers using this image"))
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

func (model Model) ShortHelp() []key.Binding {
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

func (model Model) FullHelp() [][]key.Binding {
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

// handlePruneImages prunes unused images
func (model *Model) handlePruneImages() tea.Cmd {
	// Show progress dialog
	progressDialog := components.NewProgressDialog(
		"Pruning unused images...\n\nThis may take a few moments...",
	)
	model.SetOverlay(progressDialog)

	// Start async prune operation
	return func() tea.Msg {
		ctx := stdcontext.Background()
		spaceReclaimed, err := state.GetClient().PruneImages(ctx)
		return MsgPruneComplete{
			SpaceReclaimed: spaceReclaimed,
			Err:            err,
		}
	}
}

// handleTagImage shows dialog to tag an image
func (model *Model) handleTagImage() {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	fields := []components.FormField{
		{
			Label:       "New Tag",
			Placeholder: "myrepo/myimage:v1.0",
			Required:    true,
		},
	}

	metadata := map[string]any{
		"imageID": selectedItem.Image.ID,
	}

	dialog := components.NewFormDialog(
		"Tag Image",
		fields,
		base.SmartDialogAction{Type: "TagImageAction"},
		metadata,
	)

	model.SetOverlay(dialog)
}

// handleRunAndExec creates an ephemeral container and execs into it
func (model *Model) handleRunAndExec() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	return func() tea.Msg {
		ctx := stdcontext.Background()

		// Create ephemeral container
		config := client.CreateContainerConfig{
			Name:      fmt.Sprintf("temp-%d", selectedItem.Image.Created),
			ImageID:   selectedItem.Image.ID,
			Ports:     nil,
			Volumes:   nil,
			Env:       nil,
			AutoStart: true,
			Network:   "bridge",
		}

		containerID, err := state.GetClient().CreateContainer(ctx, config)
		if err != nil {
			return MsgRunAndExecComplete{Err: fmt.Errorf("failed to create container: %w", err)}
		}

		// Exec into it
		shell := []string{"/bin/sh"}
		if _, err := state.GetClient().ExecShell(ctx, containerID, shell); err != nil {
			// Try bash if sh fails
			shell = []string{"/bin/bash"}
			if _, err := state.GetClient().ExecShell(ctx, containerID, shell); err != nil {
				return MsgRunAndExecComplete{Err: fmt.Errorf("failed to exec shell: %w", err)}
			}
		}

		return MsgRunAndExecComplete{ContainerID: containerID}
	}
}

// handleBuildImage shows dialog to build an image
func (model *Model) handleBuildImage() {
	fields := []components.FormField{
		{
			Label:       "Dockerfile Path",
			Placeholder: "./Dockerfile",
			Required:    true,
		},
		{
			Label:       "Tag",
			Placeholder: "myimage:latest",
			Required:    true,
		},
		{
			Label:       "Build Context",
			Placeholder: ".",
			Required:    true,
		},
		{
			Label:       "Build Args",
			Placeholder: "KEY=value,FOO=bar",
			Required:    false,
		},
	}

	dialog := components.NewFormDialog(
		"Build Image",
		fields,
		base.SmartDialogAction{Type: "BuildImageAction"},
		nil,
	)

	model.SetOverlay(dialog)
}

// performTagImage tags an image with a new name
func (model *Model) performTagImage(imageID, newTag string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		err := state.GetClient().TagImage(ctx, imageID, newTag)
		if err != nil {
			return MsgTagImageComplete{Err: fmt.Errorf("failed to tag image: %w", err)}
		}
		return MsgTagImageComplete{ImageID: imageID, NewTag: newTag}
	}
}

// performBuildImage builds an image from a Dockerfile
func (model *Model) performBuildImage(dockerfile, tag, contextPath string, buildArgs map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()

		// Convert buildArgs to map[string]*string
		buildArgsPtr := make(map[string]*string)
		for k, v := range buildArgs {
			val := v
			buildArgsPtr[k] = &val
		}

		buildOutput, err := state.GetClient().BuildImage(ctx, dockerfile, tag, contextPath, buildArgsPtr)
		if err != nil {
			return MsgBuildImageComplete{Err: fmt.Errorf("failed to build image: %w", err)}
		}
		defer buildOutput.Close()

		if _, err := io.Copy(io.Discard, buildOutput); err != nil {
			return MsgBuildImageComplete{Err: fmt.Errorf("failed to read build output: %w", err)}
		}

		return MsgBuildImageComplete{Tag: tag}
	}
}

func listenToPullProgress(imageName string, progressChan <-chan string, doneChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case status, ok := <-progressChan:
			if !ok {
				err := <-doneChan
				return MsgPullComplete{ImageName: imageName, Err: err}
			}

			return MsgPullProgress{
				ImageName:    imageName,
				Message:      parsePullStatusMessage(status),
				ProgressChan: progressChan,
				DoneChan:     doneChan,
			}

		case err := <-doneChan:
			return MsgPullComplete{ImageName: imageName, Err: err}
		}
	}
}

func parsePullStatusMessage(raw string) string {
	type pullProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	}

	type pullStatus struct {
		ID             string             `json:"id"`
		Status         string             `json:"status"`
		Error          string             `json:"error"`
		Progress       string             `json:"progress"`
		ProgressDetail pullProgressDetail `json:"progressDetail"`
	}

	var status pullStatus
	if err := json.Unmarshal([]byte(raw), &status); err != nil {
		return "Pulling image..."
	}

	if status.Error != "" {
		return status.Error
	}

	message := strings.TrimSpace(status.Status)
	if message == "" {
		message = "Pulling image..."
	}

	if status.ID != "" {
		message = fmt.Sprintf("%s (%s)", message, status.ID)
	}

	if status.Progress != "" {
		return fmt.Sprintf("%s %s", message, strings.TrimSpace(status.Progress))
	}

	if status.ProgressDetail.Total > 0 {
		percent := int((status.ProgressDetail.Current * 100) / status.ProgressDetail.Total)
		return fmt.Sprintf("%s (%d%%)", message, percent)
	}

	return message
}
