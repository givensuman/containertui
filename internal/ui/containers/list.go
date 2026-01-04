package containers

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/context"
)

// ListModel wraps the bubbles list model.
type ListModel struct {
	list list.Model
}

// NewListModel creates a new list model for containers.
func NewListModel() ListModel {
	// Get containers from client
	client := context.GetClient()
	containers := client.GetContainers()

	items := make([]list.Item, len(containers))
	for i, c := range containers {
		items[i] = NewContainerItem(c)
	}

	width, height := context.GetWindowSize()
	l := list.New(items, list.NewDefaultDelegate(), width, height)

	l.SetShowTitle(false)

	return ListModel{
		list: l,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return m.list.View()
}
