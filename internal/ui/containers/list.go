package containers

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
)

var style lipgloss.Style = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	list               list.Model
	selectedContainers *selectedContainers
	keybindings        *keybindings
}

var _ tea.Model = (*Model)(nil)

func NewContainersList() Model {
	containers := context.GetClient().GetContainers()
	var containerItems []list.Item
	for _, container := range containers {
		containerItems = append(
			containerItems,
			ContainerItem{
				Container:  container,
				isSelected: false,
			},
		)
	}

	width, height := context.GetWindowSize()
	list := list.New(containerItems, newDefaultDelegate(), width, height)

	list.SetShowTitle(false)
	list.SetShowStatusBar(false)
	list.SetFilteringEnabled(false) // TODO: Workout styling issues with filtering

	selectedContainers := newSelectedContainers()

	keybindings := newKeybindings()
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keybindings.pauseContainer,
			keybindings.unpauseContainer,
			keybindings.startContainer,
			keybindings.stopContainer,
			keybindings.toggleSelection,
			keybindings.toggleSelectionOfAll,
		}
	}

	return Model{list, selectedContainers, keybindings}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		widthOffset, heightOffset := style.GetFrameSize()
		m.list.SetWidth(msg.Width - widthOffset)
		m.list.SetHeight(msg.Height - heightOffset)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(msg, m.keybindings.pauseContainer):
			m.handlePauseContainers()
		case key.Matches(msg, m.keybindings.unpauseContainer):
			m.handleUnpauseContainers()
		case key.Matches(msg, m.keybindings.startContainer):
			m.handleStartContainers()
		case key.Matches(msg, m.keybindings.stopContainer):
			m.handleStopContainers()
		case key.Matches(msg, m.keybindings.toggleSelection):
			m.handleToggleSelection()
		case key.Matches(msg, m.keybindings.toggleSelectionOfAll):
			m.handleToggleSelectionOfAll()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return style.Render(m.list.View())
}
