package containers

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type keybindings struct {
	pauseContainer       key.Binding
	unpauseContainer     key.Binding
	startContainer       key.Binding
	stopContainer        key.Binding
	removeContainer      key.Binding
	showLogs             key.Binding
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
		removeContainer: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "remove container"),
		),
		showLogs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "show container logs"),
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
	shared.Component
	style              lipgloss.Style
	list               list.Model
	selectedContainers *selectedContainers
	keybindings        *keybindings
}

var (
	_ tea.Model             = (*ContainerList)(nil)
	_ shared.ComponentModel = (*ContainerList)(nil)
)

func newContainerList() ContainerList {
	containers := context.GetClient().GetContainers()
	var containerItems []list.Item
	for _, container := range containers {
		containerItems = append(
			containerItems,
			ContainerItem{
				Container:  container,
				isSelected: false,
				isWorking:  false,
				spinner:    spinner.New(),
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

	keybindings := newKeybindings()
	list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keybindings.pauseContainer,
			keybindings.unpauseContainer,
			keybindings.startContainer,
			keybindings.stopContainer,
			keybindings.removeContainer,
			keybindings.showLogs,
			keybindings.toggleSelection,
			keybindings.toggleSelectionOfAll,
		}
	}

	return ContainerList{
		style:              style,
		list:               list,
		selectedContainers: newSelectedContainers(),
		keybindings:        keybindings,
	}
}

func (cl *ContainerList) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	cl.WindowWidth = msg.Width
	cl.WindowHeight = msg.Height
}

func (cl ContainerList) Init() tea.Cmd {
	return nil
}

func (cl ContainerList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cl.style = cl.style.Width(msg.Width).Height(msg.Height)
		widthOffset, heightOffset := cl.style.GetFrameSize()

		cl.list.SetWidth(msg.Width - widthOffset)
		cl.list.SetHeight(msg.Height - heightOffset)

	case MessageConfirmDelete:
		cmd = cl.handleConfirmationOfRemoveContainers()
		cmds = append(cmds, cmd)

	case MessageContainerOperationResult:
		cl.handleContainerOperationResult(msg)

	case tea.KeyMsg:
		if cl.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, cl.keybindings.pauseContainer):
			cmd = cl.handlePauseContainers()
			cmds = append(cmds, cmd)
		case key.Matches(msg, cl.keybindings.unpauseContainer):
			cmd = cl.handleUnpauseContainers()
			cmds = append(cmds, cmd)
		case key.Matches(msg, cl.keybindings.startContainer):
			cmd = cl.handleStartContainers()
			cmds = append(cmds, cmd)
		case key.Matches(msg, cl.keybindings.stopContainer):
			cmd = cl.handleStopContainers()
			cmds = append(cmds, cmd)
		case key.Matches(msg, cl.keybindings.removeContainer):
			cmd = cl.handleRemoveContainers()
			cmds = append(cmds, cmd)
		case key.Matches(msg, cl.keybindings.showLogs):
			cmd = cl.handleShowLogs()
			cmds = append(cmds, cmd)
		case key.Matches(msg, cl.keybindings.toggleSelection):
			cl.handleToggleSelection()
		case key.Matches(msg, cl.keybindings.toggleSelectionOfAll):
			cl.handleToggleSelectionOfAll()
		}
	}

	listModel, cmd := cl.list.Update(msg)
	cl.list = listModel
	cmds = append(cmds, cmd)

	items := cl.list.Items()
	for _, item := range items {
		if c, ok := item.(ContainerItem); ok && c.isWorking {
			cmds = append(cmds, c.spinner.Tick)
		}
	}

	return cl, tea.Batch(cmds...)
}

func (cl ContainerList) View() string {
	return cl.style.Render(cl.list.View())
}
