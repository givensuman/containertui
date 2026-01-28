package components

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/ui/base"
)

type SessionState int

const (
	ViewMain SessionState = iota
	ViewOverlay
)

type ResourceView[ID comparable, Item list.Item] struct {
	base.Component
	SplitView       SplitView
	Selections      *SelectionManager[ID]
	SessionState    SessionState
	DetailsKeyBinds DetailsKeybindings
	Foreground      any

	Title          string
	AdditionalHelp []key.Binding

	LoadItems     func() ([]Item, error)
	GetItemID     func(Item) ID
	GetItemTitle  func(Item) string
	IsItemWorking func(Item) bool
	OnResize      func(w, h int)
}

func NewResourceView[ID comparable, Item list.Item](
	title string,
	loadItems func() ([]Item, error),
	getItemID func(Item) ID,
	getItemTitle func(Item) string,
	onResize func(w, h int),
) *ResourceView[ID, Item] {
	delegate := list.NewDefaultDelegate()
	listModel := list.New([]list.Item{}, delegate, 0, 0)
	listModel.Title = title
	listModel.SetShowTitle(false)
	listModel.SetShowHelp(false)

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

	rv.Refresh()

	return rv
}

func (rv *ResourceView[ID, Item]) Init() tea.Cmd {
	return nil
}

func (rv *ResourceView[ID, Item]) SetDelegate(delegate list.DefaultDelegate) {
	rv.SplitView.SetDelegates(delegate)
}

func (rv *ResourceView[ID, Item]) Refresh() tea.Cmd {
	if rv.LoadItems == nil {
		return nil
	}

	items, err := rv.LoadItems()
	if err != nil {
		return nil
	}

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	return rv.SplitView.List.SetItems(listItems)
}

func (rv *ResourceView[ID, Item]) SetOverlay(model any) {
	rv.Foreground = model
	rv.SessionState = ViewOverlay
}

func (rv *ResourceView[ID, Item]) CloseOverlay() {
	rv.Foreground = nil
	rv.SessionState = ViewMain
}

func (rv *ResourceView[ID, Item]) IsOverlayVisible() bool {
	return rv.SessionState == ViewOverlay
}

func (rv *ResourceView[ID, Item]) IsListFocused() bool {
	return rv.SplitView.Focus == FocusList
}

func (rv *ResourceView[ID, Item]) IsDetailFocused() bool {
	return rv.SplitView.Focus == FocusDetail
}

func (rv *ResourceView[ID, Item]) IsExtraFocused() bool {
	return rv.SplitView.Focus == FocusExtra
}

func (rv *ResourceView[ID, Item]) IsFiltering() bool {
	return rv.SplitView.List.FilterState() == list.Filtering
}

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

func (rv *ResourceView[ID, Item]) GetSelectedIndex() int {
	return rv.SplitView.List.Index()
}

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

func (rv *ResourceView[ID, Item]) SetItem(index int, item Item) {
	rv.SplitView.List.SetItem(index, item)
}

func (rv *ResourceView[ID, Item]) GetSelectedIDs() []ID {
	ids := make([]ID, 0)

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

func (rv *ResourceView[ID, Item]) ToggleSelection(id ID) {
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

func (rv *ResourceView[ID, Item]) GetContentWidth() int {
	if vp, ok := rv.SplitView.Detail.(*ViewportPane); ok {
		return vp.Viewport.Width()
	}
	return 0
}

func (rv *ResourceView[ID, Item]) SetContent(content string) {
	if vp, ok := rv.SplitView.Detail.(*ViewportPane); ok {
		vp.SetContent(content)
	}
}

func (rv *ResourceView[ID, Item]) SetExtraContent(content string) {
	if rv.SplitView.Extra != nil {
		if vp, ok := rv.SplitView.Extra.(*ViewportPane); ok {
			vp.SetContent(content)
		}
	}
}

func (rv *ResourceView[ID, Item]) SetDetailTitle(title string) {
	rv.SplitView.SetDetailTitle(title)
}

func (rv *ResourceView[ID, Item]) SetExtraTitle(title string) {
	rv.SplitView.SetExtraTitle(title)
}

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

func (rv *ResourceView[ID, Item]) Update(msg tea.Msg) (ResourceView[ID, Item], tea.Cmd) {
	var cmds []tea.Cmd

	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		rv.UpdateWindowDimensions(sizeMsg)
	}

	if rv.SessionState == ViewOverlay && rv.Foreground != nil {
		if model, ok := rv.Foreground.(tea.Model); ok {
			updatedModel, cmd := model.Update(msg)
			rv.Foreground = updatedModel
			cmds = append(cmds, cmd)
		}

		if _, ok := msg.(base.CloseDialogMessage); ok {
			rv.CloseOverlay()
		}

		return *rv, tea.Batch(cmds...)
	}

	if _, ok := msg.(base.CloseDialogMessage); ok {
		rv.CloseOverlay()
		return *rv, nil
	}

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

		if viewer, ok := rv.Foreground.(base.StringViewModel); ok {
			fgView = viewer.View()
		} else if viewer, ok := rv.Foreground.(interface{ View() string }); ok {
			fgView = viewer.View()
		} else if viewer, ok := rv.Foreground.(interface{ ViewString() string }); ok {
			fgView = viewer.ViewString()
		} else if model, ok := rv.Foreground.(tea.Model); ok {
			view := model.View()
			fgView = fmt.Sprint(view)
		} else {
			fgView = "Error: Overlay model has no View method"
		}

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
