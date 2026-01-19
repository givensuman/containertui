package services

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type keybindings struct {
	switchTab key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5", "tab", "shift+tab"),
			key.WithHelp("1-5/tab", "switch tab"),
		),
	}
}

type ServiceList struct {
	shared.Component
	style       lipgloss.Style
	list        list.Model
	keybindings *keybindings
}

func newServiceList() ServiceList {
	services, err := context.GetClient().GetServices()
	if err != nil {
		services = []client.Service{}
	}
	serviceItems := make([]list.Item, 0, len(services))
	for _, service := range services {
		serviceItems = append(serviceItems, ServiceItem{Service: service})
	}

	width, height := context.GetWindowSize()
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		PaddingTop(1)

	delegate := list.NewDefaultDelegate()
	delegate = shared.ChangeDelegateStyles(delegate)
	listModel := list.New(serviceItems, delegate, width, height)

	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.FilterCursor = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	serviceKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			serviceKeybindings.switchTab,
		}
	}

	return ServiceList{
		style:       style,
		list:        listModel,
		keybindings: serviceKeybindings,
	}
}

func (s *ServiceList) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	s.WindowWidth = msg.Width
	s.WindowHeight = msg.Height

	layoutManager := shared.NewLayoutManager(msg.Width, msg.Height)
	masterLayout, _ := layoutManager.CalculateMasterDetail(s.style)

	s.style = s.style.Width(masterLayout.Width).Height(masterLayout.Height)
	s.list.SetWidth(masterLayout.ContentWidth)
	s.list.SetHeight(masterLayout.ContentHeight)
}

func (s ServiceList) Init() tea.Cmd {
	return nil
}

func (s ServiceList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case shared.MsgFocusChanged:
		delegate := list.NewDefaultDelegate()
		if msg.IsDetailsFocused {
			delegate = shared.UnfocusDelegateStyles(delegate)
		} else {
			delegate = shared.ChangeDelegateStyles(delegate)
		}
		s.list.SetDelegate(delegate)

	case tea.KeyMsg:
		if s.list.FilterState() == list.Filtering {
			break
		}
		if msg.String() == "q" {
			return s, tea.Quit
		}
	}

	updatedList, listCmd := s.list.Update(msg)
	s.list = updatedList
	cmds = append(cmds, listCmd)

	return s, tea.Batch(cmds...)
}

func (s ServiceList) View() string {
	return s.style.Render(s.list.View())
}
