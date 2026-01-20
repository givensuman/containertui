package components

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/ui/base"
)

// SessionState defines the active view mode (main or overlay).
type SessionState int

const (
	ViewMain SessionState = iota
	ViewOverlay
)

// ResourceView encapsulates common patterns for resource management views.
// It handles split view, selection management, data loading, and basic state.
type ResourceView[ID comparable, Item list.Item] struct {
	base.Component
	SplitView       SplitView
	Selections      *SelectionManager[ID]
	SessionState    SessionState
	DetailsKeyBinds DetailsKeybindings
	Foreground      interface{} // Dialogs, logs, etc.

	// Configuration
	Title          string
	AdditionalHelp []key.Binding

	// Callbacks
	LoadItems     func() ([]Item, error)
	GetItemID     func(Item) ID
	GetItemTitle  func(Item) string
	IsItemWorking func(Item) bool
	OnResize      func(w, h int)
}

// NewResourceView creates a new initialized ResourceView.
func NewResourceView[ID comparable, Item list.Item](
	title string,
	loadItems func() ([]Item, error),
	getItemID func(Item) ID,
	getItemTitle func(Item) string,
	onResize func(w, h int),
) *ResourceView[ID, Item] {
	// Initialize default list
	delegate := list.NewDefaultDelegate()
	listModel := list.New([]list.Item{}, delegate, 0, 0)
	listModel.Title = title
	listModel.SetShowTitle(false) // We usually handle title outside or via splitview headers if needed
	listModel.SetShowHelp(false)  // Disable built-in help, we use global help

	// Initialize split view
	splitView := NewSplitView(listModel, NewViewportPane())

	rv := &ResourceView[ID, Item]{
		SplitView:       splitView,
		Selections:      NewSelectionManager[ID](),
		SessionState:    ViewMain,
		DetailsKeyBinds: NewDetailsKeybindings(),
		Title:           title,
		LoadItems:       loadItems,
		GetItemID:       getItemID,
		GetItemTitle:    getItemTitle,
		OnResize:        onResize,
	}

	// Initial load
	rv.Refresh()

	return rv
}

// Init initializes the component.
func (rv *ResourceView[ID, Item]) Init() tea.Cmd {
	return nil
}

// SetDelegate sets the list delegate and stores focused/unfocused versions.
func (rv *ResourceView[ID, Item]) SetDelegate(delegate list.DefaultDelegate) {
	rv.SplitView.SetDelegates(delegate)
}

// Refresh reloads the items using the LoadItems callback.
func (rv *ResourceView[ID, Item]) Refresh() tea.Cmd {
	if rv.LoadItems == nil {
		return nil
	}

	items, err := rv.LoadItems()
	if err != nil {
		// In a real app we might want to show an error, but for now just log or ignore
		return nil
	}

	// Cast items to list.Item interface
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	return rv.SplitView.List.SetItems(listItems)
}

// SetOverlay shows a modal/overlay.
func (rv *ResourceView[ID, Item]) SetOverlay(model interface{}) {
	rv.Foreground = model
	rv.SessionState = ViewOverlay
}

// CloseOverlay hides the overlay.
func (rv *ResourceView[ID, Item]) CloseOverlay() {
	rv.Foreground = nil
	rv.SessionState = ViewMain
}

// IsOverlayVisible returns true if an overlay is currently shown.
func (rv *ResourceView[ID, Item]) IsOverlayVisible() bool {
	return rv.SessionState == ViewOverlay
}

// IsListFocused returns true if the list pane is focused.
func (rv *ResourceView[ID, Item]) IsListFocused() bool {
	return rv.SplitView.Focus == FocusList
}

// IsFiltering returns true if the list is currently filtering.
func (rv *ResourceView[ID, Item]) IsFiltering() bool {
	return rv.SplitView.List.FilterState() == list.Filtering
}

// GetSelectedItem returns the currently selected item, or nil.
func (rv *ResourceView[ID, Item]) GetSelectedItem() *Item {
	item := rv.SplitView.List.SelectedItem()
	if item == nil {
		return nil
	}
	if casted, ok := item.(Item); ok {
		return &casted
	}
	return nil
}

// GetSelectedIndex returns the index of the currently selected item.
func (rv *ResourceView[ID, Item]) GetSelectedIndex() int {
	return rv.SplitView.List.Index()
}

// GetItems returns all items cast to the specific type.
func (rv *ResourceView[ID, Item]) GetItems() []Item {
	rawItems := rv.SplitView.List.Items()
	items := make([]Item, 0, len(rawItems))
	for _, raw := range rawItems {
		if casted, ok := raw.(Item); ok {
			items = append(items, casted)
		}
	}
	return items
}

// SetItem updates an item at a specific index.
func (rv *ResourceView[ID, Item]) SetItem(index int, item Item) {
	rv.SplitView.List.SetItem(index, item)
}

// GetSelectedIDs returns the IDs of all selected items (multi-selection).
func (rv *ResourceView[ID, Item]) GetSelectedIDs() []ID {
	// We delegate to the SelectionManager
	// But SelectionManager stores IDs. We might want to ensure they are valid?
	// For now just return what's in the manager.
	// Actually, SelectionManager might need to know the current list to validate?
	// Let's just return the keys from the map.
	ids := make([]ID, 0)
	// Accessing the internal map of SelectionManager would be ideal if exposed,
	// otherwise we iterate.
	// Assuming SelectionManager has IsSelected.
	// A better way for the consumer is to iterate items and check selection.
	// But if we use the SelectionManager as the source of truth:
	// We need to iterate the list items to maintain order.

	items := rv.SplitView.List.Items()
	for _, raw := range items {
		if item, ok := raw.(Item); ok {
			id := rv.GetItemID(item)
			if rv.Selections.IsSelected(id) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// ToggleSelection toggles selection for the given ID.
func (rv *ResourceView[ID, Item]) ToggleSelection(id ID) {
	// We need the index for the SelectionManager
	// Find index
	items := rv.SplitView.List.Items()
	index := -1
	for i, raw := range items {
		if item, ok := raw.(Item); ok {
			if rv.GetItemID(item) == id {
				index = i
				break
			}
		}
	}

	if index != -1 {
		rv.Selections.Toggle(id, index)
	}
}

// GetContentWidth returns the width available for content in the detail pane.
func (rv *ResourceView[ID, Item]) GetContentWidth() int {
	// SplitView logic: (Width - 2 for border) - 2 for padding = Width - 4
	// But we need the current size of the Detail pane.
	// SplitView.Detail is a Pane.
	// We can check the viewport width if it's a ViewportPane.
	if vp, ok := rv.SplitView.Detail.(*ViewportPane); ok {
		return vp.Viewport.Width()
	}
	return 0
}

// SetContent sets the content string for the detail pane (assumes ViewportPane).
func (rv *ResourceView[ID, Item]) SetContent(content string) {
	if vp, ok := rv.SplitView.Detail.(*ViewportPane); ok {
		vp.SetContent(content)
	}
}

// HandleToggleSelection toggles the selection of the currently focused item.
func (rv *ResourceView[ID, Item]) HandleToggleSelection() {
	index := rv.SplitView.List.Index()
	selectedItem := rv.SplitView.List.SelectedItem()
	if selectedItem == nil {
		return
	}

	item, ok := selectedItem.(Item)
	if !ok {
		return
	}

	if rv.IsItemWorking != nil && rv.IsItemWorking(item) {
		return
	}

	id := rv.GetItemID(item)
	rv.Selections.Toggle(id, index)
}

// HandleToggleAll toggles selection for all items.
func (rv *ResourceView[ID, Item]) HandleToggleAll() {
	items := rv.SplitView.List.Items()
	allSelected := true

	for _, rawItem := range items {
		item, ok := rawItem.(Item)
		if !ok {
			continue
		}
		if rv.IsItemWorking != nil && rv.IsItemWorking(item) {
			continue
		}
		id := rv.GetItemID(item)
		if !rv.Selections.IsSelected(id) {
			allSelected = false
			break
		}
	}

	if allSelected {
		rv.Selections.Clear()
	} else {
		rv.Selections.Clear()
		for index, rawItem := range items {
			item, ok := rawItem.(Item)
			if !ok {
				continue
			}
			if rv.IsItemWorking != nil && rv.IsItemWorking(item) {
				continue
			}
			id := rv.GetItemID(item)
			rv.Selections.Select(id, index)
		}
	}
}

// DetailsKeybindings are standard keys for the detail pane.
type DetailsKeybindings struct {
	Up     key.Binding
	Down   key.Binding
	Switch key.Binding
}

func NewDetailsKeybindings() DetailsKeybindings {
	return DetailsKeybindings{
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

// Update handles standard resource view updates (resizing, basic navigation).
func (rv *ResourceView[ID, Item]) Update(msg tea.Msg) (ResourceView[ID, Item], tea.Cmd) {
	var cmds []tea.Cmd

	// Handle Window Resize
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		rv.UpdateWindowDimensions(sizeMsg)
	}

	// Handle Overlay
	if rv.SessionState == ViewOverlay && rv.Foreground != nil {
		// If it's a model, update it
		if model, ok := rv.Foreground.(tea.Model); ok {
			updatedModel, cmd := model.Update(msg)
			rv.Foreground = updatedModel
			cmds = append(cmds, cmd)
		}

		// Handle generic close messages if bubbling up
		if _, ok := msg.(base.CloseDialogMessage); ok {
			rv.CloseOverlay()
		}

		// We generally don't process other keys if overlay is active, unless we want to allow background updates
		return *rv, tea.Batch(cmds...)
	}

	// Handle close dialog message (from global event loop or self)
	if _, ok := msg.(base.CloseDialogMessage); ok {
		rv.CloseOverlay()
		return *rv, nil
	}

	// Update SplitView
	var cmd tea.Cmd
	rv.SplitView, cmd = rv.SplitView.Update(msg)
	cmds = append(cmds, cmd)

	return *rv, tea.Batch(cmds...)
}

func (rv *ResourceView[ID, Item]) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	rv.WindowWidth = msg.Width
	rv.WindowHeight = msg.Height
	rv.SplitView.SetSize(msg.Width, msg.Height)

	if rv.SessionState == ViewOverlay && rv.Foreground != nil {
		if component, ok := rv.Foreground.(base.ComponentModel); ok {
			component.UpdateWindowDimensions(msg)
		}
	}

	if rv.OnResize != nil {
		rv.OnResize(msg.Width, msg.Height)
	}
}

func (rv *ResourceView[ID, Item]) View() string {
	if rv.SessionState == ViewOverlay && rv.Foreground != nil {
		var fgView string

		// Try to get View() string
		if viewer, ok := rv.Foreground.(base.StringViewModel); ok {
			fgView = viewer.View()
		} else if viewer, ok := rv.Foreground.(interface{ View() string }); ok {
			fgView = viewer.View()
		} else if _, ok := rv.Foreground.(tea.Model); ok {
			// tea.View in v2 is a struct, not string.
			// RenderOverlay expects string for background and foreground.
			// So we can't directly use tea.View in RenderOverlay's signature if it expects string.

			// If we are here, we are using a tea.Model that is NOT a StringViewModel.
			// This means we might be trying to render a full component in the overlay.
			// RenderOverlay is designed for string-based overlays.

			// For now, let's just return a placeholder error if we hit this,
			// because our app mainly uses string view models for dialogs.
			fgView = "Error: Overlay model does not support string view"
		}

		// Since RenderOverlay returns tea.View, we assume RenderOverlay handles conversion internally
		// However, the signature of RenderOverlay is `func(..., ) tea.View`.
		// If we want to return a string from ResourceView.View, we need RenderOverlay to return string
		// OR we change ResourceView.View to return tea.View.

		// Wait, previously `ResourceView.View` returned `tea.View`.
		// But in `containers.go` we tried calling `.String()` on it which failed.
		// That suggests `tea.View` is a struct that DOES NOT have String().

		// If `ResourceView` is embedded in `Model`, and `Model.View()` returns `tea.View`,
		// then `ResourceView.View()` should probably return `tea.View` too.

		// The error in `containers.go` was `model.ResourceView.View().String() undefined`.
		// This means `model.ResourceView.View()` returned `tea.View` which has no `.String()`.
		// But `RenderOverlay` takes string arguments for background.

		// So `ResourceView.View()` MUST return a string if we want to use it AS A BACKGROUND in `RenderOverlay` inside `Model.View()`.

		// BUT `ResourceView` logic *itself* handles overlay rendering too (lines 374-402).
		// So `ResourceView.View()` is ALREADY returning a composed overlay if needed.

		// Let's make `ResourceView.View()` return string.
		// Then `RenderOverlay` should be adjusted or used carefully.

		// `RenderOverlay` returns `tea.View`. We can't easily convert `tea.View` to string unless we access its internal buffer, which might not be exposed.
		// Actually, `RenderOverlay` uses `lipgloss.NewCanvas(...).Render()` which returns string, then wraps it in `tea.NewView(...)`.

		// Let's change `RenderOverlay` to return string in `overlay.go`?
		// Or assume `RenderOverlay` returns `tea.View` and we extract the string?
		// tea.View (v2) is `type View struct { body string }`? No, it's likely opaque.

		// Let's look at `RenderOverlay` implementation again.
		// It returns `tea.NewView(canvas.Render())`. `canvas.Render()` is string.

		// So if `ResourceView` wants to return string, it should call `RenderOverlayAsString` or similar.
		// Or we modify `RenderOverlay` to return string.

		// Let's modify `ResourceView.View` to return `string`.
		// And we need `RenderOverlay` to return string.

		// But `RenderOverlay` is in `components/overlay.go`.

		// TEMPORARY FIX:
		// Let's assume we can just implement the overlay logic here returning string.

		// Re-implement RenderOverlay logic here locally to return string
		bg := rv.SplitView.View()
		return RenderOverlayString(bg, fgView, rv.WindowWidth, rv.WindowHeight)
	}
	return rv.SplitView.View()
}

func (rv *ResourceView[ID, Item]) ShortHelp() []key.Binding {
	if rv.SplitView.Focus == FocusList {
		return rv.SplitView.List.ShortHelp()
	}
	return []key.Binding{
		rv.DetailsKeyBinds.Up,
		rv.DetailsKeyBinds.Down,
	}
}

func (rv *ResourceView[ID, Item]) FullHelp() [][]key.Binding {
	if rv.SplitView.Focus == FocusList {
		help := rv.SplitView.List.FullHelp()
		if len(rv.AdditionalHelp) > 0 {
			help = append(help, rv.AdditionalHelp)
		}
		return help
	}
	return [][]key.Binding{
		{
			rv.DetailsKeyBinds.Up,
			rv.DetailsKeyBinds.Down,
			rv.DetailsKeyBinds.Switch,
		},
	}
}
