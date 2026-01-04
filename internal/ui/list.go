package ui

import (
	"fmt"

	"github.com/givensuman/containertui/internal/color"
	"github.com/givensuman/containertui/internal/context"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2F99EE"))
	titleStyle        = lipgloss.NewStyle().Padding(0, 1).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#2F99EE"))
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type Item struct {
	ID       string
	Name     string
	Image    string
	State    string
	Selected bool
}

// FilterValue implements list.Item.
func (i Item) FilterValue() string {
	return i.Name
}
var _ list.Item = Item{}

func (i *Item) Title() string {
	selectedIcon := "[ ]"
	if i.Selected {
		selectedIcon = selectedItemStyle.Render("[x]")
	}

	stateIcon := "⏸︎"
	if i.State == "running" {
		stateIcon = "▶︎"
	} else if i.State == "exited" {
		stateIcon = "⏹"
	}

	// coloring
	sTitle := fmt.Sprintf("   %s (%s) %s", i.Name, i.ID[:12], stateIcon)
	if i.State == "running" {
		sTitle = color.FgGreen(sTitle)
	} else if i.State == "exited" {
		sTitle = color.FgGray(sTitle)
	} else {
		sTitle = color.FgYellow(sTitle)
	}
	return fmt.Sprintf("%s %s", selectedIcon, sTitle)
}

func (i *Item) Description() string {
	return fmt.Sprintf("   %s - %s", i.Image, i.State)
}

// key binding
type listKeyMap struct {
	pauseContainer   key.Binding
	unpauseContainer key.Binding
	startContainer   key.Binding
	stopContainer    key.Binding
	toggleSelect     key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		pauseContainer: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause container"),
		),
		unpauseContainer: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "unpause container"),
		),
		startContainer: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start container"),
		),
		stopContainer: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stop container"),
		),
		toggleSelect: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle select"),
		),
	}
}

func getContainerItems() []list.Item {
	containers := context.GetClient().GetContainers()
	var items []list.Item
	for _, container := range containers {
		items = append(
			items,
			Item{
				ID:       container.ID,
				Name:     container.Name,
				Image:    container.Image,
				State:    container.State,
				Selected: false,
			},
		)
	}
	return items
}

func NewList() list.Model {
	items := getContainerItems()
	d := list.NewDefaultDelegate()

	dockerColor := lipgloss.Color("#2F99EE")
	white := lipgloss.Color("#FFFDF5")
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(white).
		BorderLeftForeground(dockerColor)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy() // reuse the title style here

	l := list.New(items, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return l
}

type ListModel struct {
	list list.Model
	keys *listKeyMap
}

func (lm ListModel) Init() tea.Cmd {
	return nil
}

func (lm ListModel) getSelectedItems() ([]Item, []int) {
	var selected []Item
	var indices []int
	items := lm.list.Items()
	for i, item := range items {
		if it, ok := item.(Item); ok && it.Selected {
			selected = append(selected, it)
			indices = append(indices, i)
		}
	}
	return selected, indices
}

func (lm ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if lm.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, lm.keys.toggleSelect):
			// toggle selection
			selectedItem, ok := lm.list.SelectedItem().(Item)
			if ok {
				selectedItem.Selected = !selectedItem.Selected
				lm.list.SetItem(lm.list.Index(), selectedItem)
			}

			return lm, nil

		case key.Matches(msg, lm.keys.pauseContainer):
			selectedItems, selectedIndexes := lm.getSelectedItems()

			if len(selectedItems) > 0 {
				// pause selected containers
				var ids []string
				for _, it := range selectedItems {
					ids = append(ids, it.ID)
				}
				context.GetClient().PauseContainers(ids)
				items := lm.list.Items()
				for _, index := range selectedIndexes {
					curItem := items[index].(Item)
					curItem.State = "paused"
					lm.list.SetItem(index, curItem)
				}
			} else {
				// pause current container
				selectedItem, ok := lm.list.SelectedItem().(Item)
				if ok {
					context.GetClient().PauseContainer(selectedItem.ID)
					selectedItem.State = "paused"
					lm.list.SetItem(lm.list.Index(), selectedItem)
				}
			}

			return lm, nil

		case key.Matches(msg, lm.keys.unpauseContainer):
			selectedItems, selectedIndexes := lm.getSelectedItems()

			if len(selectedItems) > 0 {
				// unpause selected containers
				var ids []string
				for _, it := range selectedItems {
					ids = append(ids, it.ID)
				}
				context.GetClient().UnpauseContainers(ids)
				items := lm.list.Items()
				for _, index := range selectedIndexes {
					curItem := items[index].(Item)
					curItem.State = "running"
					lm.list.SetItem(index, curItem)
				}
			} else {
				// unpause current container
				selectedItem, ok := lm.list.SelectedItem().(Item)
				if ok {
					context.GetClient().UnpauseContainer(selectedItem.ID)
					selectedItem.State = "running"
					lm.list.SetItem(lm.list.Index(), selectedItem)
				}

			}

			return lm, nil

		case key.Matches(msg, lm.keys.startContainer):
			selectedItems, selectedIndexes := lm.getSelectedItems()

			if len(selectedItems) > 0 {
				// start selected containers
				var ids []string
				for _, it := range selectedItems {
					ids = append(ids, it.ID)
				}
				context.GetClient().StartContainers(ids)
				items := lm.list.Items()
				for _, index := range selectedIndexes {
					curItem := items[index].(Item)
					curItem.State = "running"
					lm.list.SetItem(index, curItem)
				}
			} else {
				// start current container
				selectedItem, ok := lm.list.SelectedItem().(Item)
				if ok {
					context.GetClient().StartContainer(selectedItem.ID)
					selectedItem.State = "running"
					lm.list.SetItem(lm.list.Index(), selectedItem)
				}
			}

			return lm, nil

		case key.Matches(msg, lm.keys.stopContainer):
			selectedItems, selectedIndexes := lm.getSelectedItems()

			if len(selectedItems) > 0 {
				// stop selected containers
				var ids []string
				for _, it := range selectedItems {
					ids = append(ids, it.ID)
				}
				context.GetClient().StopContainers(ids)
				items := lm.list.Items()
				for _, index := range selectedIndexes {
					curItem := items[index].(Item)
					curItem.State = "exited"
					lm.list.SetItem(index, curItem)
				}
			} else {
				// stop current container
				selectedItem, ok := lm.list.SelectedItem().(Item)
				if ok {
					context.GetClient().StopContainer(selectedItem.ID)
					selectedItem.State = "exited"
					lm.list.SetItem(lm.list.Index(), selectedItem)
				}
			}

			return lm, nil
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		lm.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	lm.list, cmd = lm.list.Update(msg)
	return lm, cmd
}

func (lm ListModel) View() string {
	return lm.list.View()
}
