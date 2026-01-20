package services

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/styles"
)

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

type MsgRefreshServices time.Time

type detailsKeybindings struct {
	Up     key.Binding
	Down   key.Binding
	Switch key.Binding
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
	}
}

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

type Model struct {
	base.Component
	splitView          components.SplitView
	keybindings        *keybindings
	sessionState       sessionState
	currentServiceName string
	detailsKeybindings detailsKeybindings
}

var (
	_ tea.Model           = (*Model)(nil)
	_ base.ComponentModel = (*Model)(nil)
)

func New() Model {
	services, err := context.GetClient().GetServices()
	if err != nil {
		services = []client.Service{}
	}
	serviceItems := make([]list.Item, 0, len(services))
	for _, service := range services {
		serviceItems = append(serviceItems, ServiceItem{Service: service})
	}

	width, height := context.GetWindowSize()

	delegate := list.NewDefaultDelegate()
	delegate = styles.ChangeDelegateStyles(delegate)
	listModel := list.New(serviceItems, delegate, width, height)

	listModel.SetShowHelp(false)
	listModel.SetShowTitle(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.Styles.Filter.Focused.Prompt = lipgloss.NewStyle().Foreground(colors.Primary())
	listModel.Styles.Filter.Cursor.Color = colors.Primary()
	// listModel.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(colors.Primary())
	// listModel.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	serviceKeybindings := newKeybindings()
	listModel.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			serviceKeybindings.switchTab,
		}
	}

	splitView := components.NewSplitView(listModel, components.NewViewportPane())

	model := Model{
		splitView:          splitView,
		keybindings:        serviceKeybindings,
		sessionState:       viewMain,
		detailsKeybindings: newDetailsKeybindings(),
	}

	return model
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return MsgRefreshServices(t)
	})
}

func (model *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	model.WindowWidth = msg.Width
	model.WindowHeight = msg.Height
	model.splitView.SetSize(msg.Width, msg.Height)
}

func (model Model) Init() tea.Cmd {
	return tickCmd()
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.UpdateWindowDimensions(msg)

	case MsgRefreshServices:
		cmds = append(cmds, tickCmd())
		// Refresh services data
		services, err := context.GetClient().GetServices()
		if err == nil {
			items := make([]list.Item, 0, len(services))
			for _, s := range services {
				items = append(items, ServiceItem{Service: s})
			}
			cmd := model.splitView.List.SetItems(items)
			cmds = append(cmds, cmd)
		}
	}

	// Forward messages to SplitView
	updatedSplitView, splitCmd := model.splitView.Update(msg)
	model.splitView = updatedSplitView
	cmds = append(cmds, splitCmd)

	// Handle keybindings when list is focused
	if model.splitView.Focus == components.FocusList {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if model.splitView.List.FilterState() != list.Filtering {
				if keyMsg.String() == "q" {
					return model, tea.Quit
				}
				if key.Matches(keyMsg, model.keybindings.switchTab) {
					return model, nil
				}
			}
		}
	}

	// Check if selection changed to update details
	selectedItem := model.splitView.List.SelectedItem()
	if selectedItem != nil {
		if serviceItem, ok := selectedItem.(ServiceItem); ok {
			if serviceItem.Service.Name != model.currentServiceName {
				model.currentServiceName = serviceItem.Service.Name
				model.updateDetails(serviceItem.Service)
			}
		}
	} else {
		// No selection
		if model.currentServiceName != "" {
			model.currentServiceName = ""
			if pane, ok := model.splitView.Detail.(*components.ViewportPane); ok {
				pane.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No service selected."))
			}
		}
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) updateDetails(service client.Service) {
	content := ""
	if service.ComposeFile != "" {
		data, err := os.ReadFile(service.ComposeFile)
		if err != nil {
			content = fmt.Sprintf("Error reading compose file: %v", err)
		} else {
			content = string(data)
		}
	} else {
		content = "No docker-compose file found for this service."
	}

	// Add some header info
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).MarginBottom(1)
	builder := strings.Builder{}

	builder.WriteString(headerStyle.Render(fmt.Sprintf("%s (%d replicas)", service.Name, service.Replicas)) + "\n")

	infoStyle := lipgloss.NewStyle().Foreground(colors.Text()).MarginBottom(1)
	if service.ComposeFile != "" {
		builder.WriteString(infoStyle.Render("Compose File: "+service.ComposeFile) + "\n\n")
	} else {
		builder.WriteString(infoStyle.Render("No compose file detected") + "\n\n")
	}

	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).Underline(true).MarginTop(1).MarginBottom(0)
	builder.WriteString(sectionHeader.Render("Compose Configuration") + "\n")
	builder.WriteString(content)

	if pane, ok := model.splitView.Detail.(*components.ViewportPane); ok {
		pane.SetContent(builder.String())
	}
}

func (model Model) View() tea.View {
	return tea.NewView(model.splitView.View())
}

func (model Model) ShortHelp() []key.Binding {
	switch model.splitView.Focus {
	case components.FocusList:
		return model.splitView.List.ShortHelp()
	case components.FocusDetail:
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
		}
	}
	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	switch model.splitView.Focus {
	case components.FocusList:
		return model.splitView.List.FullHelp()
	case components.FocusDetail:
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
		}
	}
	return nil
}
