// Package browse defines the browse component for registry images.
package browse

import (
	stdcontext "context"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

const (
	registryDockerHub = "dockerhub"
	registryQuay      = "quay"
)

var supportedRegistries = []string{registryDockerHub, registryQuay}

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
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "switch tab"),
		),
	}
}

func additionalHelpBindings(bindings *keybindings) []key.Binding {
	return []key.Binding{
		bindings.switchTab,
		bindings.search,
		bindings.pull,
	}
}

// Model represents the browse component state.
type Model struct {
	components.ResourceView[string, BrowseItem]

	keybindings        *keybindings
	detailsKeybindings components.DetailsKeybindings

	// Current state
	currentItemID      string
	inspection         registry.RegistryImageDetail
	isSearchMode       bool
	currentSearchQuery string
	currentRegistry    string

	// Scroll position memory
	scrollPositions map[string]int

	// Pull state
	isPulling      bool
	progressChan   <-chan string
	currentPulling string // Name of image currently being pulled
	pendingPulls   []pullTarget
	batchPullTotal int
	batchPulled    int

	// Pull progress tracking
	pullLayers     map[string]pullLayerProgress
	pullPercent    float64
	progressDialog *components.ProgressDialog
}

type pullLayerProgress struct {
	current int64
	total   int64
}

type pullTarget struct {
	ImageName string
	Registry  string
}

// New creates a new Browse model.
func New() Model {
	browseKeybindings := newKeybindings()

	// Initialize ResourceView
	fetchPopularImages := func() ([]BrowseItem, error) {
		client := state.GetRegistryClient()
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
	resourceView.AdditionalHelp = additionalHelpBindings(browseKeybindings)

	return Model{
		ResourceView:       *resourceView,
		keybindings:        browseKeybindings,
		detailsKeybindings: components.NewDetailsKeybindings(),
		scrollPositions:    make(map[string]int),
		isPulling:          false,
		currentSearchQuery: "",
		currentRegistry:    registryDockerHub,
		pendingPulls:       nil,
		batchPullTotal:     0,
		batchPulled:        0,
		pullLayers:         make(map[string]pullLayerProgress),
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
		model.currentRegistry = normalizeRegistry(msg.Registry)
		model.isSearchMode = true

		cmds = append(cmds, notifications.ShowInfo(fmt.Sprintf("Found %d results for '%s' in %s", len(msg.Images), msg.Query, displayRegistryName(model.currentRegistry))))
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
		var progressCmd tea.Cmd
		if progressDialog, ok := model.Foreground.(components.ProgressDialog); ok {
			progressDialog.SetStatus(parsePullStatusMessage(msg.Status))
			if percent, hasPercent := model.estimatePullProgress(msg.Status); hasPercent {
				progressCmd = progressDialog.SetPercent(percent)
			}
			model.Foreground = progressDialog
		}
		// Continue listening to the progress channel
		if msg.ProgressChan != nil && msg.DoneChan != nil {
			return model, tea.Batch(progressCmd, listenToProgressChannelWithDone(msg.ImageName, msg.ProgressChan, msg.DoneChan))
		}
		return model, progressCmd

	case MsgPullComplete:
		model.isPulling = false
		model.progressChan = nil
		model.progressDialog = nil
		model.pullLayers = make(map[string]pullLayerProgress)
		model.pullPercent = 0

		// Close the progress overlay
		model.CloseOverlay()

		// Stop the spinner for the pulled image
		if model.currentPulling != "" {
			model.setWorkingState([]string{model.currentPulling}, false)
			model.currentPulling = ""
		}

		if msg.Err != nil {
			model.pendingPulls = nil
			model.batchPullTotal = 0
			model.batchPulled = 0
			return model, notifications.ShowError(fmt.Errorf("pull failed: %w", msg.Err))
		}

		model.batchPulled++
		if len(model.pendingPulls) > 0 {
			next := model.pendingPulls[0]
			model.pendingPulls = model.pendingPulls[1:]
			return model, tea.Batch(
				notifications.ShowInfo(fmt.Sprintf("Pulled %s (%d/%d)", msg.ImageName, model.batchPulled, model.batchPullTotal)),
				model.startPull(next),
			)
		}

		successMsg := fmt.Sprintf("Pulled %s successfully", msg.ImageName)
		if model.batchPullTotal > 1 {
			successMsg = fmt.Sprintf("Pulled %d images successfully", model.batchPullTotal)
		}
		model.batchPullTotal = 0
		model.batchPulled = 0

		// Send message to refresh Images tab
		return model, tea.Batch(
			notifications.ShowSuccess(successMsg),
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
				payload, ok := confirmMsg.Action.Payload.([]pullTarget)
				if !ok {
					model.CloseOverlay()
					return model, notifications.ShowError(fmt.Errorf("invalid pull payload"))
				}
				model.CloseOverlay()
				if len(payload) == 0 {
					return model, nil
				}

				model.batchPullTotal = len(payload)
				model.batchPulled = 0
				model.pendingPulls = append([]pullTarget(nil), payload[1:]...)
				return model, model.startPull(payload[0])

			case "SearchRegistry":
				// Extract query from form values
				var query string
				selectedRegistry := model.currentRegistry
				if payload, ok := confirmMsg.Action.Payload.(map[string]any); ok {
					if values, ok := payload["values"].(map[string]string); ok {
						selectedRegistry = normalizeRegistry(values["Search Registry"])
						query = strings.TrimSpace(values["Search Query"])
					}
				}
				model.CloseOverlay()

				// If query is empty, return to popular images
				if query == "" {
					if selectedRegistry != registryDockerHub {
						return model, notifications.ShowInfo("Enter a search query for non-Docker Hub registries")
					}

					model.currentSearchQuery = ""
					model.currentRegistry = registryDockerHub
					model.isSearchMode = false
					return model, tea.Batch(model.Refresh(), notifications.ShowInfo("Returned to popular Docker Hub images"))
				}

				return model, model.performRemoteSearch(query, selectedRegistry)
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
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No image selected"))
		return nil
	}

	itemID := selectedItem.Image.RepoName
	if itemID != model.currentItemID {
		// Save scroll position of previous item
		model.saveScrollPosition()

		model.currentItemID = itemID

		// Fetch detailed data asynchronously
		if selectedItem.Image.Registry == registryQuay {
			model.inspection = detailFromBrowseItem(selectedItem.Image)
			model.refreshInspectionContent()
			return nil
		}

		return func() tea.Msg {
			client := state.GetRegistryClient()

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
		"Search Registry",
		[]components.FormField{
			{
				Label:    "Search Registry",
				Value:    model.currentRegistry,
				Options:  supportedRegistries,
				Required: false,
			},
			{
				Label:       "Search Query",
				Placeholder: "e.g., nginx, postgres, redis",
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
	targets := model.pullImageTargets()
	if len(targets) == 0 {
		return
	}

	message := ""
	if len(targets) == 1 {
		message = fmt.Sprintf("Pull image '%s' from %s?", targets[0].ImageName, displayRegistryName(targets[0].Registry))
	} else {
		message = fmt.Sprintf("Pull %d selected images?", len(targets))
	}

	confirmDialog := components.NewDialog(
		message,
		[]components.DialogButton{
			{Label: "Cancel"},
			{
				Label: "Pull",
				Action: base.SmartDialogAction{
					Type:    "PullImage",
					Payload: targets,
				},
			},
		},
	)
	model.SetOverlay(confirmDialog)
}

func (model *Model) pullImageTargets() []pullTarget {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		targets := make([]pullTarget, 0, len(selectedIDs))
		for _, item := range model.GetItems() {
			if !contains(selectedIDs, item.Image.RepoName) {
				continue
			}
			registryName := normalizeRegistry(item.Image.Registry)
			if item.Image.Registry == "" {
				registryName = model.currentRegistry
			}
			targets = append(targets, pullTarget{ImageName: item.Image.RepoName, Registry: registryName})
		}
		return targets
	}

	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	registryName := normalizeRegistry(selectedItem.Image.Registry)
	if selectedItem.Image.Registry == "" {
		registryName = model.currentRegistry
	}

	return []pullTarget{{ImageName: selectedItem.Image.RepoName, Registry: registryName}}
}

func (model *Model) startPull(target pullTarget) tea.Cmd {
	imageName := target.ImageName
	registryName := normalizeRegistry(target.Registry)

	spinnerCmd := model.setWorkingState([]string{imageName}, true)

	model.isPulling = true
	model.currentPulling = imageName
	model.pullLayers = make(map[string]pullLayerProgress)
	model.pullPercent = 0

	progressDialog := components.NewProgressDialogWithBar("Pulling " + imageName)
	model.progressDialog = &progressDialog
	model.SetOverlay(progressDialog)

	progressChan := make(chan string, 100)
	doneChan := make(chan error, 1)
	model.progressChan = progressChan

	go func() {
		ctx := stdcontext.Background()
		err := state.GetBackend().PullImage(ctx, imageName, progressChan)
		doneChan <- err
		close(doneChan)
	}()

	return tea.Batch(
		spinnerCmd,
		notifications.ShowInfo(fmt.Sprintf("Pulling from %s", displayRegistryName(registryName))),
		listenToProgressChannelWithDone(imageName, progressChan, doneChan),
	)
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

// performRemoteSearch performs a remote search on a selected registry.
func (model *Model) performRemoteSearch(query, registryName string) tea.Cmd {
	return func() tea.Msg {
		var (
			response registry.SearchResponse
			err      error
		)

		normalizedRegistry := normalizeRegistry(registryName)
		switch normalizedRegistry {
		case registryQuay:
			client := state.GetQuayRegistryClient()
			response, err = client.Search(stdcontext.Background(), query, 25)
		default:
			client := state.GetRegistryClient()
			response, err = client.Search(stdcontext.Background(), query, 25)
		}

		return MsgSearchResults{
			Query:    query,
			Registry: normalizedRegistry,
			Images:   response.Results,
			Err:      err,
		}
	}
}

func normalizeRegistry(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, candidate := range supportedRegistries {
		if normalized == candidate {
			return candidate
		}
	}

	if strings.Contains(normalized, "quay") {
		return registryQuay
	}

	return registryDockerHub
}

func displayRegistryName(name string) string {
	if normalizeRegistry(name) == registryQuay {
		return "Quay"
	}

	return "Docker Hub"
}

func detailFromBrowseItem(img registry.RegistryImage) registry.RegistryImageDetail {
	fullName := strings.TrimPrefix(img.RepoName, "quay.io/")
	namespace := ""
	name := fullName
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) == 2 {
		namespace = parts[0]
		name = parts[1]
	}

	return registry.RegistryImageDetail{
		Name:            name,
		Namespace:       namespace,
		Description:     img.ShortDescription,
		FullDescription: img.ShortDescription,
		StarCount:       img.StarCount,
		PullCount:       img.PullCount,
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

func (model *Model) estimatePullProgress(raw string) (float64, bool) {
	layerID, current, total, ok := parsePullLayerProgress(raw)
	if !ok {
		return 0, false
	}

	model.pullLayers[layerID] = pullLayerProgress{current: current, total: total}

	var sumCurrent int64
	var sumTotal int64
	for _, layer := range model.pullLayers {
		sumCurrent += layer.current
		sumTotal += layer.total
	}

	if sumTotal <= 0 {
		return 0, false
	}

	percent := float64(sumCurrent) / float64(sumTotal)
	if percent > 0.98 {
		percent = 0.98
	}
	if percent < model.pullPercent {
		percent = model.pullPercent
	}

	model.pullPercent = percent
	return percent, true
}

func parsePullLayerProgress(raw string) (string, int64, int64, bool) {
	type pullProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	}
	type pullStatus struct {
		ID             string             `json:"id"`
		ProgressDetail pullProgressDetail `json:"progressDetail"`
	}

	var status pullStatus
	if err := json.Unmarshal([]byte(raw), &status); err != nil {
		return "", 0, 0, false
	}
	if status.ID == "" || status.ProgressDetail.Total <= 0 {
		return "", 0, 0, false
	}

	current := status.ProgressDetail.Current
	if current < 0 {
		current = 0
	}
	if current > status.ProgressDetail.Total {
		current = status.ProgressDetail.Total
	}
	return status.ID, current, status.ProgressDetail.Total, true
}

// parsePullStatusMessage extracts a human-readable status string from raw pull JSON.
func parsePullStatusMessage(raw string) string {
	type pullMsg struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	}
	var msg pullMsg
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return raw
	}
	if msg.ID != "" {
		return fmt.Sprintf("[%s] %s", msg.ID[:min(len(msg.ID), 12)], msg.Status)
	}
	return msg.Status
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
				ImageName:    imageName,
				Status:       status,
				ProgressChan: progressChan,
				DoneChan:     doneChan,
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
