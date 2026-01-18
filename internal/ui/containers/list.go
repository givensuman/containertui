// Package containers defines the containers component.
package containers

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
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
	execShell            key.Binding
	toggleSelection      key.Binding
	toggleSelectionOfAll key.Binding
	switchTab            key.Binding
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
			key.WithKeys("L"),
			key.WithHelp("L", "show container logs"),
		),
		execShell: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "exec shell"),
		),
		toggleSelection: key.NewBinding(
			key.WithKeys(tea.KeySpace.String()),
			key.WithHelp("space", "toggle selection"),
		),
		toggleSelectionOfAll: key.NewBinding(
			key.WithKeys(tea.KeyCtrlA.String()),
			key.WithHelp("ctrl+a", "toggle selection of all"),
		),
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "tab", "shift+tab"),
			key.WithHelp("1-4/tab", "switch tab"),
		),
	}
}

// selectedContainers maps a container's ID to its index in the list.
type selectedContainers struct {
	selections map[string]int
}

func newSelectedContainers() *selectedContainers {
	return &selectedContainers{
		selections: make(map[string]int),
	}
}

func (selectedContainers *selectedContainers) selectContainerInList(id string, index int) {
	selectedContainers.selections[id] = index
}

func (selectedContainers selectedContainers) unselectContainerInList(id string) {
	delete(selectedContainers.selections, id)
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
	containers, err := context.GetClient().GetContainers()
	if err != nil {
		containers = []client.Container{}
	}
	containerItems := make([]list.Item, 0, len(containers))
	for _, container := range containers {
		containerItems = append(
			containerItems,
			ContainerItem{
				Container:  container,
				isSelected: false,
				isWorking:  false,
				spinner:    newSpinner(),
			},
		)
	}

	width, height := context.GetWindowSize()
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingTop(1)

	delegate := newDefaultDelegate()
	listModel := list.New(containerItems, delegate, width, height)

	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.FilterCursor = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	containerKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			containerKeybindings.pauseContainer,
			containerKeybindings.unpauseContainer,
			containerKeybindings.startContainer,
			containerKeybindings.stopContainer,
			containerKeybindings.removeContainer,
			containerKeybindings.showLogs,
			containerKeybindings.execShell,
			containerKeybindings.toggleSelection,
			containerKeybindings.toggleSelectionOfAll,
			containerKeybindings.switchTab,
		}
	}

	return ContainerList{
		style:              style,
		list:               listModel,
		selectedContainers: newSelectedContainers(),
		keybindings:        containerKeybindings,
	}
}

func (cl *ContainerList) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	cl.WindowWidth = msg.Width
	cl.WindowHeight = msg.Height

	layoutManager := shared.NewLayoutManager(msg.Width, msg.Height)
	masterLayout, _ := layoutManager.CalculateMasterDetail(cl.style)

	cl.style = cl.style.Width(masterLayout.Width).Height(masterLayout.Height)
	cl.list.SetWidth(masterLayout.ContentWidth)
	cl.list.SetHeight(masterLayout.ContentHeight)
}

func (cl ContainerList) Init() tea.Cmd {
	return nil
}

func (cl ContainerList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case MessageConfirmDelete:
		cmds = append(cmds, cl.handleConfirmationOfRemoveContainers())

	case MessageContainerOperationResult:
		if cmd := cl.handleContainerOperationResult(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		if cl.list.FilterState() == list.Filtering {
			break
		}

		if msg.String() == "q" {
			return cl, tea.Quit
		}

		switch {
		case key.Matches(msg, cl.keybindings.switchTab):
			return cl, nil
		case key.Matches(msg, cl.keybindings.pauseContainer):
			cmds = append(cmds, cl.handlePauseContainers())
		case key.Matches(msg, cl.keybindings.unpauseContainer):
			cmds = append(cmds, cl.handleUnpauseContainers())
		case key.Matches(msg, cl.keybindings.startContainer):
			cmds = append(cmds, cl.handleStartContainers())
		case key.Matches(msg, cl.keybindings.stopContainer):
			cmds = append(cmds, cl.handleStopContainers())
		case key.Matches(msg, cl.keybindings.removeContainer):
			cmds = append(cmds, cl.handleRemoveContainers())
		case key.Matches(msg, cl.keybindings.showLogs):
			if cmd := cl.handleShowLogs(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case key.Matches(msg, cl.keybindings.execShell):
			if cmd := cl.handleExecShell(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case key.Matches(msg, cl.keybindings.toggleSelection):
			cl.handleToggleSelection()
		case key.Matches(msg, cl.keybindings.toggleSelectionOfAll):
			cl.handleToggleSelectionOfAll()
		}
	}

	updatedList, listCmd := cl.list.Update(msg)
	cl.list = updatedList
	cmds = append(cmds, listCmd)

	if _, ok := msg.(spinner.TickMsg); !ok {
		for _, item := range cl.list.Items() {
			if containerItem, ok := item.(ContainerItem); ok && containerItem.isWorking {
				cmds = append(cmds, containerItem.spinner.Tick)
			}
		}
	}

	return cl, tea.Batch(cmds...)
}

func (cl ContainerList) View() string {
	return cl.style.Render(cl.list.View())
}
