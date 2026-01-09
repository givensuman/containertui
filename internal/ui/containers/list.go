package containers

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
)

type keybindings struct {
	pauseContainer       key.Binding
	unpauseContainer     key.Binding
	startContainer       key.Binding
	stopContainer        key.Binding
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
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
		toggleSelection: key.NewBinding(
			key.WithKeys(tea.KeySpace.String()),
			key.WithHelp("space", "toggle selection"),
		),
		toggleSelectionOfAll: key.NewBinding(
			key.WithKeys(tea.KeyCtrlA.String()),
			key.WithHelp("ctrl+a", "toggle selection of all"),
		),
	}
}

// selectedContainers is map of a container's ID to
// its index in the list
type selectedContainers struct {
	selections map[string]int
}

func newSelectedContainers() *selectedContainers {
	return &selectedContainers{
		selections: make(map[string]int),
	}
}

func (sc *selectedContainers) selectContainerInList(id string, index int) {
	sc.selections[id] = index
}

func (sc selectedContainers) unselectContainerInList(id string) {
	delete(sc.selections, id)
}

type ContainerList struct {
	style              lipgloss.Style
	list               list.Model
	selectedContainers *selectedContainers
	keybindings        *keybindings
}

var _ tea.Model = (*ContainerList)(nil)

func newContainerList() ContainerList {
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
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingTop(1)

	list := list.New(containerItems, newDefaultDelegate(), width, height)

	list.SetShowTitle(false)
	list.SetShowStatusBar(false)
	list.SetFilteringEnabled(false)

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

	return ContainerList{style, list, selectedContainers, keybindings}
}

func (m ContainerList) Init() tea.Cmd {
	return nil
}

func (m ContainerList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.style = m.style.Width(msg.Width).Height(msg.Height)
		widthOffset, heightOffset := m.style.GetFrameSize()

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

func (m ContainerList) View() string {
	return m.style.Render(m.list.View())
}
