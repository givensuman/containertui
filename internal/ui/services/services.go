package services

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

const (
	focusList = iota
	focusDetails
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

type Model struct {
	shared.Component
	sessionState sessionState
	focusedView  int

	background   tea.Model
	overlayModel *overlay.Model

	currentServiceName string
	viewport           viewport.Model
	detailsKeybindings detailsKeybindings
}

func New() Model {
	serviceList := newServiceList()
	detailViewport := viewport.New(0, 0)

	model := Model{
		sessionState: viewMain,
		focusedView:  focusList,
		background:   serviceList,
		overlayModel: overlay.New(
			nil, // No foreground initially
			serviceList,
			overlay.Center,
			overlay.Center,
			0,
			0,
		),
		viewport:           detailViewport,
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

	layoutManager := shared.NewLayoutManager(model.WindowWidth, model.WindowHeight)
	_, detailLayout := layoutManager.CalculateMasterDetail(lipgloss.NewStyle())

	viewportWidth := detailLayout.Width - 4
	viewportHeight := detailLayout.Height - 2
	if viewportWidth < 0 {
		viewportWidth = 0
	}
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	model.viewport.Width = viewportWidth
	model.viewport.Height = viewportHeight

	if serviceList, ok := model.background.(ServiceList); ok {
		serviceList.UpdateWindowDimensions(msg)
		model.background = serviceList
	}
}

func (model Model) Init() tea.Cmd {
	return tickCmd()
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.UpdateWindowDimensions(msg)

	case tea.KeyMsg:
		if msg.String() == "tab" {
			if model.focusedView == focusList {
				model.focusedView = focusDetails
				cmds = append(cmds, func() tea.Msg { return shared.MsgFocusChanged{IsDetailsFocused: true} })
			} else {
				model.focusedView = focusList
				cmds = append(cmds, func() tea.Msg { return shared.MsgFocusChanged{IsDetailsFocused: false} })
			}
			return model, tea.Batch(cmds...)
		}
	}

	if model.focusedView == focusList {
		updatedBackground, cmd := model.background.Update(msg)
		model.background = updatedBackground
		cmds = append(cmds, cmd)
	} else if model.focusedView == focusDetails {
		updatedViewport, cmd := model.viewport.Update(msg)
		model.viewport = updatedViewport
		cmds = append(cmds, cmd)
	}

	// Check if selection changed to update details
	if serviceList, ok := model.background.(ServiceList); ok {
		selectedItem := serviceList.list.SelectedItem()
		if selectedItem != nil {
			if serviceItem, ok := selectedItem.(ServiceItem); ok {
				if serviceItem.Service.Name != model.currentServiceName {
					model.currentServiceName = serviceItem.Service.Name
					model.updateDetails(serviceItem.Service)
				}
			}
		}
	}

	// Handle refresh
	if _, ok := msg.(MsgRefreshServices); ok {
		cmds = append(cmds, tickCmd())
		// Trigger a refresh of the list data here if needed,
		// but typically we'd send a msg to the list component or reload data.
		// For simplicity, we can reload in the list component or passing a command.
		// Let's implement reloading in the background list component if we had more time.
		// For now we just refresh the loop.

		// To truly refresh, we should re-fetch services:
		services, err := context.GetClient().GetServices()
		if err == nil {
			// Update the list items
			if serviceList, ok := model.background.(ServiceList); ok {
				items := make([]list.Item, 0, len(services))
				for _, s := range services {
					items = append(items, ServiceItem{Service: s})
				}
				cmd := serviceList.list.SetItems(items)
				cmds = append(cmds, cmd)
				model.background = serviceList
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

	model.viewport.SetContent(builder.String())
}

func (model Model) renderMainView() string {
	layoutManager := shared.NewLayoutManager(model.WindowWidth, model.WindowHeight)
	_, detailLayout := layoutManager.CalculateMasterDetail(lipgloss.NewStyle())

	listView := model.background.View()

	borderColor := colors.Muted()
	if model.focusedView == focusDetails {
		borderColor = colors.Primary()
	}

	detailStyle := lipgloss.NewStyle().
		Width(detailLayout.Width - 2).
		Height(detailLayout.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1)

	var detailContent string
	if model.currentServiceName != "" {
		detailContent = model.viewport.View()
	} else {
		detailContent = lipgloss.NewStyle().Foreground(colors.Muted()).Render("No service selected.")
	}

	detailView := detailStyle.Render(detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

func (model Model) View() string {
	return model.renderMainView()
}

func (model Model) ShortHelp() []key.Binding {
	if serviceList, ok := model.background.(ServiceList); ok {
		return serviceList.list.ShortHelp()
	}
	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	if model.focusedView == focusList {
		if serviceList, ok := model.background.(ServiceList); ok {
			return serviceList.list.FullHelp()
		}
	} else {
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
