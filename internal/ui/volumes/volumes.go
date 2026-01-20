// Package volumes defines the volumes component.
package volumes

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/docker/docker/api/types/volume"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

// MsgVolumeInspection contains the inspection data for a volume.
type MsgVolumeInspection struct {
	Name   string
	Volume volume.Volume
	Err    error
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

// Model represents the volumes component state.
type Model struct {
	components.ResourceView[string, VolumeItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	currentVolumeName  string
	inspection         volume.Volume
	scrollPositions    map[string]int
	currentFormat      string
}

func New() *Model {
	volumeKeybindings := newKeybindings()

	fetchVolumes := func() ([]VolumeItem, error) {
		volumeList, err := context.GetClient().GetVolumes()
		if err != nil {
			return nil, err
		}
		items := make([]VolumeItem, 0, len(volumeList))
		for _, volume := range volumeList {
			items = append(items, VolumeItem{Volume: volume})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, VolumeItem](
		"Volumes",
		fetchVolumes,
		func(item VolumeItem) string { return item.Volume.Name },
		func(item VolumeItem) string { return item.Title() },
		func(w, h int) {
			// Window resize handled by base component
		},
	)

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        volumeKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		scrollPositions:    make(map[string]int),
		currentFormat:      "",
	}

	// Add custom keybindings to help
	model.ResourceView.AdditionalHelp = []key.Binding{
		volumeKeybindings.toggleSelection,
		volumeKeybindings.toggleSelectionOfAll,
		volumeKeybindings.remove,
	}

	return &model
}

func (model *Model) Init() tea.Cmd {
	return nil
}

func (model *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	// 1. Try standard ResourceView updates first (resizing, dialog closing, basic navigation)
	updatedView, cmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)

	// 2. Handle Messages
	switch msg := msg.(type) {
	case MsgVolumeInspection:
		if msg.Name == model.currentVolumeName && msg.Err == nil {
			model.inspection = msg.Volume
			model.refreshInspectionContent()
			// Send a message to restore scroll position on next update
			cmds = append(cmds, func() tea.Msg { return MsgRestoreScroll{} })
		}

	case MsgRestoreScroll:
		// Restore scroll position after viewport has processed content
		model.restoreScrollPosition()
	}

	// 3. Handle Overlay/Dialog logic specifically for ConfirmationMessage
	if model.ResourceView.IsOverlayVisible() {
		if confirmMsg, ok := msg.(base.ConfirmationMessage); ok {
			if confirmMsg.Action.Type == "DeleteVolume" {
				volumeName := confirmMsg.Action.Payload.(string)
				err := context.GetClient().RemoveVolume(volumeName)
				if err == nil {
					// Refresh list
					return model, model.ResourceView.Refresh()
				}
			}
			model.ResourceView.CloseOverlay()
			return model, nil
		}

		// Let ResourceView handle forwarding to overlay
		return model, tea.Batch(cmds...)
	}

	// 3. Main View Logic
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
				selectedItem := model.ResourceView.GetSelectedItem()
				if selectedItem != nil {
					model.ResourceView.ToggleSelection(selectedItem.Volume.Name)

					// Update the visual state of the item
					index := model.ResourceView.GetSelectedIndex()
					selectedItem.isSelected = !selectedItem.isSelected
					model.ResourceView.SetItem(index, *selectedItem)
				}
				return model, nil

			case key.Matches(msg, model.keybindings.toggleSelectionOfAll):
				// Check if we need to select all or deselect all
				items := model.ResourceView.GetItems()
				selectedIDs := model.ResourceView.GetSelectedIDs()

				shouldSelectAll := false
				for _, item := range items {
					found := false
					for _, id := range selectedIDs {
						if id == item.Volume.Name {
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
							if id == item.Volume.Name {
								found = true
								break
							}
						}
						if !found {
							model.ResourceView.ToggleSelection(item.Volume.Name)
						}
						// Visual update
						item.isSelected = true
						model.ResourceView.SetItem(i, item)
					}
				} else {
					// Deselect all
					for i, item := range items {
						for _, id := range selectedIDs {
							if id == item.Volume.Name {
								model.ResourceView.ToggleSelection(item.Volume.Name)
								break
							}
						}
						// Visual update
						item.isSelected = false
						model.ResourceView.SetItem(i, item)
					}
				}
				return model, nil

			case key.Matches(msg, model.keybindings.remove):
				// TODO: Implement remove functionality
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

	// Update Detail Content
	detailCmd := model.updateDetailContent()
	if detailCmd != nil {
		cmds = append(cmds, detailCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.ResourceView.GetSelectedItem()
	if selectedItem == nil {
		model.ResourceView.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No volume selected."))
		return nil
	}

	volumeName := selectedItem.Volume.Name
	// If we've switched to a different volume, OR we don't have inspection data yet, fetch it
	if volumeName != model.currentVolumeName || model.inspection.Name == "" {
		// Save scroll position of previous volume
		if model.currentVolumeName != "" && model.currentVolumeName != volumeName {
			model.saveScrollPosition()
		}

		model.currentVolumeName = volumeName
		// Fetch inspection data asynchronously
		return func() tea.Msg {
			volumeInfo, err := context.GetClient().InspectVolume(volumeName)
			return MsgVolumeInspection{Name: volumeName, Volume: volumeInfo, Err: err}
		}
	}

	return nil
}

// saveScrollPosition saves the current viewport scroll position for the current volume
func (model *Model) saveScrollPosition() {
	if model.currentVolumeName != "" {
		if vp := model.getViewport(); vp != nil {
			model.scrollPositions[model.currentVolumeName] = vp.YOffset()
		}
	}
}

// restoreScrollPosition restores the viewport scroll position for the current volume
func (model *Model) restoreScrollPosition() {
	if model.currentVolumeName != "" {
		if vp := model.getViewport(); vp != nil {
			if offset, exists := model.scrollPositions[model.currentVolumeName]; exists {
				vp.SetYOffset(offset)
			} else {
				vp.SetYOffset(0) // New volume, start at top
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
	content := builders.BuildVolumePanel(model.inspection, model.ResourceView.GetContentWidth(), format)
	model.ResourceView.SetContent(content)
}

// handleCopyToClipboard copies the current inspection output to clipboard
func (model *Model) handleCopyToClipboard() tea.Cmd {
	if model.inspection.Name == "" {
		return nil
	}

	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	data, err := infopanel.MarshalToFormat(model.inspection, format)
	if err != nil {
		return notifications.ShowError(err)
	}

	if err := clipboard.WriteAll(string(data)); err != nil {
		return notifications.ShowError(err)
	}

	return notifications.ShowSuccess("Copied to clipboard")
}

// handleToggleFormat toggles between JSON and YAML format
func (model *Model) handleToggleFormat() tea.Cmd {
	currentFormat := model.currentFormat
	if currentFormat == "" {
		cfg := context.GetConfig()
		currentFormat = cfg.InspectionFormat
		if currentFormat == "" {
			currentFormat = "yaml"
		}
	}

	if currentFormat == "json" {
		model.currentFormat = "yaml"
	} else {
		model.currentFormat = "json"
	}

	model.refreshInspectionContent()
	return notifications.ShowSuccess("Switched to " + model.currentFormat)
}

func (model *Model) View() string {
	return model.ResourceView.View()
}

func (model *Model) ShortHelp() []key.Binding {
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

func (model *Model) FullHelp() [][]key.Binding {
	if !model.ResourceView.IsListFocused() {
		return [][]key.Binding{
			{model.detailsKeybindings.Up, model.detailsKeybindings.Down, model.detailsKeybindings.Switch},
			{model.detailsKeybindings.ToggleJSON, model.detailsKeybindings.CopyOutput},
		}
	}
	return model.ResourceView.FullHelp()
}
