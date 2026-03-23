// Package browse defines the browse component for Docker Hub images.
package browse

import (
	stdcontext "context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

type keybindings struct {
	search               key.Binding
	pull                 key.Binding
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	switchTab            key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search registry"),
		),
		pull: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pull image"),
		),
		toggleSelection: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle selection"),
		),
		toggleSelectionOfAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle selection of all"),
		),
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5", "6"),
			key.WithHelp("1-6", "switch tab"),
		),
	}
}

type detailsKeybindings struct {
	Up         key.Binding
	Down       key.Binding
	Switch     key.Binding
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
		CopyOutput: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy to clipboard"),
		),
	}
}

// Model represents the browse component state.
type Model struct {
	components.ResourceView[string, BrowseItem]

	keybindings        *keybindings
	detailsKeybindings detailsKeybindings

	// Current state
	currentItemID      string
	inspection         registry.RegistryImageDetail
	isSearchMode       bool
	currentSearchQuery string

	// Scroll position memory
	scrollPositions map[string]int

	// Pull state
	isPulling      bool
	progressChan   <-chan string
	currentPulling string // Name of image currently being pulled
}

// New creates a new Browse model.
func New() Model {
	browseKeybindings := newKeybindings()

	// Initialize ResourceView
	fetchPopularImages := func() ([]BrowseItem, error) {
		client := state.GetClient().GetRegistryClient()
		images, err := client.GetPopularImages(stdcontext.Background(), 50)
		if err != nil {
			return nil, err
		}

		items := make([]BrowseItem, 0, len(images))
		for _, img := range images {
			items = append(items, BrowseItem{
				Image:      img,
				isSelected: false,
				isWorking:  false,
				spinner:    newSpinner(),
			})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, BrowseItem](
		"Browse",
		fetchPopularImages,
		func(item BrowseItem) string { return item.Image.RepoName },
		func(item BrowseItem) string { return item.Title() },
		func(w, h int) {},
	)

	// Set custom delegate for list styling
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	// Set detail panel title
	resourceView.SplitView.SetDetailTitle("README")

	// Add custom keybindings to help
	resourceView.AdditionalHelp = []key.Binding{
		browseKeybindings.search,
		browseKeybindings.pull,
		browseKeybindings.toggleSelection,
		browseKeybindings.toggleSelectionOfAll,
		browseKeybindings.switchTab,
	}

	return Model{
		ResourceView:       *resourceView,
		keybindings:        browseKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		scrollPositions:    make(map[string]int),
		isPulling:          false,
		currentSearchQuery: "",
	}
}

// Init initializes the browse model.
func (model Model) Init() tea.Cmd {
	return model.ResourceView.Init()
}

// Update handles messages and updates the model.
func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Forward to ResourceView first
	updatedView, cmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// Handle custom messages
	switch msg := msg.(type) {
	case MsgPopularImages:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		// Images are already loaded by ResourceView

	case MsgSearchResults:
		if msg.Err != nil {
			return model, notifications.ShowError(msg.Err)
		}
		// Update list with search results
		items := make([]BrowseItem, 0, len(msg.Images))
		for _, img := range msg.Images {
			items = append(items, BrowseItem{
				Image:      img,
				isSelected: false,
				isWorking:  false,
				spinner:    newSpinner(),
			})
		}
		// Convert to list.Item and set items
		listItems := make([]list.Item, len(items))
		for i, item := range items {
			listItems[i] = item
		}
		cmd := model.SplitView.List.SetItems(listItems)
		cmds = append(cmds, cmd)

		// Store the search query
		model.currentSearchQuery = msg.Query
		model.isSearchMode = true

		cmds = append(cmds, notifications.ShowInfo(fmt.Sprintf("Found %d results for '%s'", len(msg.Images), msg.Query)))
		return model, tea.Batch(cmds...)

	case MsgImageInspection:
		if msg.RepoName == model.currentItemID && msg.Err == nil {
			model.inspection = msg.Detail
			model.refreshInspectionContent()
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		} else if msg.Err != nil {
			model.SetContent(fmt.Sprintf("Error loading details: %v", msg.Err))
		}

	case MsgRestoreScroll:
		model.restoreScrollPosition()

	case MsgPullProgress:
		// Progress updates are handled internally, no UI update needed
		// Continue listening to the progress channel
		if model.progressChan != nil && msg.DoneChan != nil {
			return model, listenToProgressChannelWithDone(msg.ImageName, model.progressChan, msg.DoneChan)
		}

	case MsgPullComplete:
		model.isPulling = false
		model.progressChan = nil

		// Stop the spinner for the pulled image
		if model.currentPulling != "" {
			model.setWorkingState([]string{model.currentPulling}, false)
			model.currentPulling = ""
		}

		if msg.Err != nil {
			return model, notifications.ShowError(fmt.Errorf("pull failed: %w", msg.Err))
		}

		// Send message to refresh Images tab
		return model, tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Pulled %s successfully", msg.ImageName)),
			func() tea.Msg {
				return base.MsgImagePulled{ImageName: msg.ImageName}
			},
		)
	}

	// Handle dialog confirmations
	if model.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.SmartConfirmationMessage); ok {
			switch confirmMsg.Action.Type {
			case "PullImage":
				imageName := confirmMsg.Action.Payload.(string)
				model.CloseOverlay()

				// Start spinner for this image
				spinnerCmd := model.setWorkingState([]string{imageName}, true)

				model.isPulling = true
				model.currentPulling = imageName

				// Create channels for progress and completion
				progressChan := make(chan string, 100)
				doneChan := make(chan error, 1)
				model.progressChan = progressChan

				// Start pull in background
				go func() {
					ctx := stdcontext.Background()
					err := state.GetClient().PullImageFromRegistry(ctx, imageName, progressChan)
					// PullImage closes progressChan when done
					// Send the error result through doneChan
					doneChan <- err
					close(doneChan)
				}()

				// Start listening to progress
				return model, tea.Batch(spinnerCmd, listenToProgressChannelWithDone(imageName, progressChan, doneChan))

			case "SearchRegistry":
				// Extract query from form values
				var query string
				if payload, ok := confirmMsg.Action.Payload.(map[string]any); ok {
					if values, ok := payload["values"].(map[string]string); ok {
						query = strings.TrimSpace(values["Search Query"])
					}
				}
				model.CloseOverlay()

				// If query is empty, return to popular images
				if query == "" {
					model.currentSearchQuery = ""
					model.isSearchMode = false
					return model, tea.Batch(
						model.Refresh(),
						notifications.ShowInfo("Returned to popular images"),
					)
				}

				return model, model.performRemoteSearch(query)
			}
		}
		return model, tea.Batch(cmds...)
	}

	// Handle keybindings when list focused
	if model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.IsFiltering() {
				break
			}

			switch {
			case key.Matches(msg, model.keybindings.search):
				model.handleSearch()

			case key.Matches(msg, model.keybindings.pull):
				model.handlePull()

			case key.Matches(msg, model.keybindings.toggleSelection):
				model.handleToggleSelection()

			case key.Matches(msg, model.keybindings.toggleSelectionOfAll):
				model.handleToggleSelectionOfAll()
			}
		}
	}

	// Handle keybindings when detail pane focused
	if model.IsDetailFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, model.detailsKeybindings.CopyOutput):
				cmd := model.handleCopyToClipboard()
				cmds = append(cmds, cmd)
			}
		}
	}

	// Update detail content if selection changed
	detailCmd := model.updateDetailContent()
	cmds = append(cmds, detailCmd)

	return model, tea.Batch(cmds...)
}

// View renders the browse component.
func (model Model) View() string {
	return model.ResourceView.View()
}

// updateDetailContent fetches and displays detailed info for the selected item.
func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent("No image selected")
		return nil
	}

	itemID := selectedItem.Image.RepoName
	if itemID != model.currentItemID {
		// Save scroll position of previous item
		model.saveScrollPosition()

		model.currentItemID = itemID

		// Fetch detailed data asynchronously
		return func() tea.Msg {
			client := state.GetClient().GetRegistryClient()

			// Determine namespace and name from RepoName
			// Docker Hub format: "namespace/name" or just "name" for official images
			var namespace, name string
			if selectedItem.Image.IsOfficial {
				// Official images are in the "library" namespace
				namespace = "library"
				name = itemID
			} else {
				// Parse namespace/name from RepoName
				// RepoName format: "user/image" or "org/image"
				parts := strings.SplitN(itemID, "/", 2)
				if len(parts) == 2 {
					namespace = parts[0]
					name = parts[1]
				} else {
					// Fallback: assume library namespace
					namespace = "library"
					name = itemID
				}
			}

			detail, err := client.GetRepository(stdcontext.Background(), namespace, name)
			return MsgImageInspection{
				RepoName: itemID,
				Detail:   detail,
				Err:      err,
			}
		}
	}

	return nil
}

// refreshInspectionContent updates the detail panel with current inspection data.
func (model *Model) refreshInspectionContent() {
	content := builders.BuildBrowsePanel(model.inspection, model.WindowWidth)
	model.SetContent(content)
}

// saveScrollPosition saves the current scroll position.
func (model *Model) saveScrollPosition() {
	if model.currentItemID != "" {
		if vp := model.getViewport(); vp != nil {
			model.scrollPositions[model.currentItemID] = vp.YOffset()
		}
	}
}

// restoreScrollPosition restores the scroll position for the current item.
func (model *Model) restoreScrollPosition() {
	if model.currentItemID != "" {
		if vp := model.getViewport(); vp != nil {
			if offset, exists := model.scrollPositions[model.currentItemID]; exists {
				vp.SetYOffset(offset)
			} else {
				vp.SetYOffset(0)
			}
		}
	}
}

// getViewport returns the detail viewport.
func (model *Model) getViewport() *viewport.Model {
	if vp, ok := model.SplitView.Detail.(*components.ViewportPane); ok {
		return &vp.Viewport
	}
	return nil
}

// handleSearch shows a search dialog.
func (model *Model) handleSearch() {
	searchDialog := components.NewFormDialog(
		"Search Docker Hub",
		[]components.FormField{
			{
				Label:       "Search Query",
				Placeholder: "e.g., nginx, postgres, redis (leave empty to return to popular images)",
				Value:       model.currentSearchQuery,
				Required:    false,
			},
		},
		base.SmartDialogAction{
			Type: "SearchRegistry",
		},
		nil,
	)
	model.SetOverlay(searchDialog)
}

// handlePull shows a pull confirmation dialog.
func (model *Model) handlePull() {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	imageName := selectedItem.Image.RepoName

	confirmDialog := components.NewDialog(
		fmt.Sprintf("Pull image '%s' from Docker Hub?", imageName),
		[]components.DialogButton{
			{Label: "Cancel"},
			{
				Label: "Pull",
				Action: base.SmartDialogAction{
					Type:    "PullImage",
					Payload: imageName,
				},
			},
		},
	)
	model.SetOverlay(confirmDialog)
}

// handleToggleSelection toggles selection of the current item.
func (model *Model) handleToggleSelection() {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return
	}

	// Use the built-in toggle method
	model.HandleToggleSelection()

	// Update the item's visual state and set it back
	index := model.GetSelectedIndex()
	if item := model.GetSelectedItem(); item != nil {
		item.isSelected = model.Selections.IsSelected(item.Image.RepoName)
		model.SetItem(index, *item)
	}
}

// handleToggleSelectionOfAll toggles selection of all items.
func (model *Model) handleToggleSelectionOfAll() {
	// Use the built-in toggle all method
	model.HandleToggleAll()

	// Update visual state of all items
	items := model.GetItems()
	for i := range items {
		items[i].isSelected = model.Selections.IsSelected(items[i].Image.RepoName)
		model.SetItem(i, items[i])
	}
}

// handleCopyToClipboard copies the detail panel content to clipboard.
func (model *Model) handleCopyToClipboard() tea.Cmd {
	if model.currentItemID == "" {
		return nil
	}

	// Build the content
	content := builders.BuildBrowsePanel(model.inspection, model.WindowWidth)

	if err := clipboard.WriteAll(content); err != nil {
		return notifications.ShowError(fmt.Errorf("failed to copy: %w", err))
	}
	return notifications.ShowSuccess("Copied to clipboard")
}

// performRemoteSearch performs a remote search on Docker Hub.
func (model *Model) performRemoteSearch(query string) tea.Cmd {
	return func() tea.Msg {
		client := state.GetClient().GetRegistryClient()
		response, err := client.Search(stdcontext.Background(), query, 25)
		return MsgSearchResults{
			Query:  query,
			Images: response.Results,
			Err:    err,
		}
	}
}

// setWorkingState sets the working/spinner state for specific images.
func (model *Model) setWorkingState(imageNames []string, working bool) tea.Cmd {
	var cmds []tea.Cmd

	currentItems := model.GetItems()
	for i, item := range currentItems {
		// Check if this item's image name matches
		if contains(imageNames, item.Image.RepoName) {
			item.isWorking = working
			if working {
				item.spinner = newSpinner()
				// Capture the spinner in a local variable to avoid closure capture bug
				spinner := item.spinner
				cmds = append(cmds, func() tea.Msg {
					return spinner.Tick()
				})
			}
			model.SetItem(i, item)
		}
	}

	return tea.Batch(cmds...)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func listenToProgressChannelWithDone(imageName string, progressChan <-chan string, doneChan <-chan error) tea.Cmd {
	return func() tea.Msg {
		select {
		case status, ok := <-progressChan:
			if !ok {
				// Progress channel closed, check for completion error
				err := <-doneChan
				return MsgPullComplete{
					ImageName: imageName,
					Err:       err,
				}
			}

			// Regular progress update
			return MsgPullProgress{
				ImageName: imageName,
				Status:    status,
				DoneChan:  doneChan,
			}

		case err := <-doneChan:
			// Completion received before progress channel closed
			return MsgPullComplete{
				ImageName: imageName,
				Err:       err,
			}
		}
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (model Model) ShortHelp() []key.Binding {
	// If detail pane is focused, show detail keybindings
	if model.IsDetailFocused() {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
			model.detailsKeybindings.CopyOutput,
		}
	}
	return model.ResourceView.ShortHelp()
}

// FullHelp returns keybindings for the expanded help view.
func (model Model) FullHelp() [][]key.Binding {
	// If detail pane is focused, show detail keybindings
	if model.IsDetailFocused() {
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
			{
				model.detailsKeybindings.CopyOutput,
			},
		}
	}
	return model.ResourceView.FullHelp()
}
