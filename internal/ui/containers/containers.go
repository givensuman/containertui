// Package containers defines the containers component.
package containers

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
	"github.com/guptarohit/asciigraph"
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

// MsgStatsTick is sent to trigger a stats refresh.
type MsgStatsTick time.Time

// MsgContainerStats contains the stats for a container.
type MsgContainerStats struct {
	ID    string
	Stats client.ContainerStats
	Err   error
}

// MsgContainerInspection contains the inspection data for a container.
type MsgContainerInspection struct {
	ID        string
	Container types.ContainerJSON
	Err       error
}

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

// Model represents the containers component state.
type Model struct {
	shared.Component
	sessionState sessionState
	focusedView  int

	foreground   tea.Model
	background   tea.Model
	overlayModel *overlay.Model

	currentContainerID string
	cpuHistory         []float64
	lastStats          client.ContainerStats

	viewport           viewport.Model
	inspection         types.ContainerJSON
	detailsKeybindings detailsKeybindings
}

var (
	_ tea.Model             = (*Model)(nil)
	_ shared.ComponentModel = (*Model)(nil)
)

func New() Model {
	deleteConfirmation := newDeleteConfirmation(nil)
	containerList := newContainerList()

	detailViewport := viewport.New(0, 0)

	model := Model{
		sessionState: viewMain,
		focusedView:  focusList,
		foreground:   deleteConfirmation,
		background:   containerList,
		overlayModel: overlay.New(
			deleteConfirmation,
			containerList,
			overlay.Center,
			overlay.Center,
			0,
			0,
		),
		cpuHistory:         make([]float64, 0),
		viewport:           detailViewport,
		detailsKeybindings: newDetailsKeybindings(),
	}

	return model
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return MsgStatsTick(t)
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

	switch model.sessionState {
	case viewMain:
		if containerList, ok := model.background.(ContainerList); ok {
			containerList.UpdateWindowDimensions(msg)
			model.background = containerList
		}
	case viewOverlay:
		switch foregroundModel := model.foreground.(type) {
		case *ContainerLogs:
			foregroundModel.setDimensions(msg.Width, msg.Height)
			model.foreground = foregroundModel
		case DeleteConfirmation:
			foregroundModel.UpdateWindowDimensions(msg)
			model.foreground = foregroundModel
		}
	}
}

func (model Model) Init() tea.Cmd {
	return tickCmd()
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if model.sessionState == viewMain {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "tab" {
				if model.focusedView == focusList {
					model.focusedView = focusDetails
				} else {
					model.focusedView = focusList
				}
				return model, nil
			}
		}
	}

	switch model.sessionState {
	case viewMain:
		isKeyMessage := false
		if _, ok := msg.(tea.KeyMsg); ok {
			isKeyMessage = true
		}

		if !isKeyMessage || model.focusedView == focusList {
			backgroundModel, backgroundCmd := model.background.Update(msg)
			model.background = backgroundModel.(ContainerList)
			cmds = append(cmds, backgroundCmd)
		}

		if !isKeyMessage || model.focusedView == focusDetails {
			updatedViewport, viewportCmd := model.viewport.Update(msg)
			model.viewport = updatedViewport
			cmds = append(cmds, viewportCmd)
		}

		if containerList, ok := model.background.(ContainerList); ok {
			selectedItem := containerList.list.SelectedItem()
			if selectedItem != nil {
				if containerItem, ok := selectedItem.(ContainerItem); ok {
					if containerItem.ID != model.currentContainerID {
						model.currentContainerID = containerItem.ID
						model.cpuHistory = make([]float64, 0)
						cmds = append(cmds, func() tea.Msg {
							containerInfo, err := context.GetClient().InspectContainer(containerItem.ID)
							return MsgContainerInspection{ID: containerItem.ID, Container: containerInfo, Err: err}
						})
					}
				}
			}
		}

	case viewOverlay:
		foregroundModel, foregroundCmd := model.foreground.Update(msg)
		model.foreground = foregroundModel
		cmds = append(cmds, foregroundCmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.UpdateWindowDimensions(msg)

	case MessageCloseOverlay:
		model.sessionState = viewMain

	case MessageOpenDeleteConfirmationDialog:
		model.foreground = newDeleteConfirmation(msg.requestedContainersToDelete...)
		model.sessionState = viewOverlay
		cmds = append(cmds, model.foreground.Init())

	case MsgStatsTick:
		cmds = append(cmds, tickCmd())

		if model.sessionState == viewMain {
			if containerList, ok := model.background.(ContainerList); ok {
				selectedItem := containerList.list.SelectedItem()
				if selectedItem != nil {
					if containerItem, ok := selectedItem.(ContainerItem); ok && containerItem.State == "running" {
						cmds = append(cmds, func() tea.Msg {
							containerStats, err := context.GetClient().GetContainerStats(containerItem.ID)
							return MsgContainerStats{ID: containerItem.ID, Stats: containerStats, Err: err}
						})
					}
				}
			}
		}

	case MsgContainerInspection:
		if msg.ID == model.currentContainerID && msg.Err == nil {
			model.inspection = msg.Container
			inspectionContent := formatInspection(model.inspection, model.lastStats, model.cpuHistory, model.viewport.Width)
			model.viewport.SetContent(inspectionContent)
		}

	case MsgContainerStats:
		if msg.ID == model.currentContainerID && msg.Err == nil {
			model.lastStats = msg.Stats
			model.cpuHistory = append(model.cpuHistory, msg.Stats.CPUPercent)
			if len(model.cpuHistory) > 30 {
				model.cpuHistory = model.cpuHistory[1:]
			}
			inspectionContent := formatInspection(model.inspection, model.lastStats, model.cpuHistory, model.viewport.Width)
			model.viewport.SetContent(inspectionContent)
		}
	}

	model.overlayModel.Foreground = model.foreground
	model.overlayModel.Background = model.background

	updatedOverlayModel, overlayCmd := model.overlayModel.Update(msg)
	if overlayModel, ok := updatedOverlayModel.(*overlay.Model); ok {
		model.overlayModel = overlayModel
		cmds = append(cmds, overlayCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model Model) View() string {
	if model.sessionState == viewOverlay && model.foreground != nil {
		return model.overlayModel.View()
	}

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
	if model.currentContainerID != "" {
		detailContent = model.viewport.View()
	} else {
		detailContent = lipgloss.NewStyle().Foreground(colors.Muted()).Render("No container selected.")
	}

	detailView := detailStyle.Render(detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

func (model Model) ShortHelp() []key.Binding {
	if model.sessionState == viewOverlay {
		if helpKeyMap, ok := model.foreground.(help.KeyMap); ok {
			return helpKeyMap.ShortHelp()
		}
		return nil
	}

	if model.focusedView == focusList {
		if containerList, ok := model.background.(ContainerList); ok {
			return containerList.list.ShortHelp()
		}
	} else if model.focusedView == focusDetails {
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
		}
	}

	return nil
}

func (model Model) FullHelp() [][]key.Binding {
	if model.sessionState == viewOverlay {
		return nil
	}

	if model.focusedView == focusList {
		if containerList, ok := model.background.(ContainerList); ok {
			return containerList.list.FullHelp()
		}
	} else if model.focusedView == focusDetails {
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

func formatInspection(container types.ContainerJSON, containerStats client.ContainerStats, cpuHistory []float64, viewportWidth int) string {
	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).MarginBottom(1)
	builder.WriteString(headerStyle.Render(fmt.Sprintf("%s (%s)", container.Name, container.ID[:12])) + "\n")

	infoStyle := lipgloss.NewStyle().Foreground(colors.Text()).MarginBottom(1)
	stateColor := colors.Success()
	if !container.State.Running {
		stateColor = colors.Muted()
	}
	stateString := lipgloss.NewStyle().Foreground(stateColor).Render(container.State.Status)

	builder.WriteString(infoStyle.Render(fmt.Sprintf("Image: %s", container.Config.Image)) + "\n")
	builder.WriteString(fmt.Sprintf("State: %s\n\n", stateString))

	if container.State.Running && len(cpuHistory) > 0 {
		graphHeight := 8
		graphWidth := viewportWidth - 10
		if graphWidth > 10 {
			usageGraph := asciigraph.Plot(cpuHistory,
				asciigraph.Height(graphHeight),
				asciigraph.Width(graphWidth),
				asciigraph.Caption("CPU Usage (%)"),
				asciigraph.SeriesColors(
					asciigraph.Blue,
				),
			)
			builder.WriteString(usageGraph + "\n\n")

			statsStyle := lipgloss.NewStyle().Foreground(colors.Primary())
			memUsageMB := containerStats.MemUsage / 1024 / 1024
			memLimitMB := containerStats.MemLimit / 1024 / 1024
			builder.WriteString(statsStyle.Render(fmt.Sprintf("CPU: %.2f%% | Mem: %.0fMB / %.0fMB",
				containerStats.CPUPercent, memUsageMB, memLimitMB)) + "\n\n")
		}
	}

	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).Underline(true).MarginTop(1).MarginBottom(0)

	builder.WriteString(sectionHeader.Render("Configuration") + "\n")
	builder.WriteString(fmt.Sprintf("Cmd: %v\n", container.Config.Cmd))
	builder.WriteString(fmt.Sprintf("Entrypoint: %v\n", container.Config.Entrypoint))
	builder.WriteString(fmt.Sprintf("WorkingDir: %s\n", container.Config.WorkingDir))

	if len(container.Config.Env) > 0 {
		builder.WriteString("\n" + sectionHeader.Render("Environment Variables") + "\n")
		for _, envVar := range container.Config.Env {
			builder.WriteString(envVar + "\n")
		}
	}

	if len(container.Mounts) > 0 {
		builder.WriteString("\n" + sectionHeader.Render("Mounts") + "\n")
		for _, mount := range container.Mounts {
			builder.WriteString(fmt.Sprintf("%s -> %s (%s)\n", mount.Source, mount.Destination, mount.Type))
		}
	}

	if container.NetworkSettings != nil && len(container.NetworkSettings.Ports) > 0 {
		builder.WriteString("\n" + sectionHeader.Render("Ports") + "\n")
		for port, bindings := range container.NetworkSettings.Ports {
			var portBindings []string
			for _, portBinding := range bindings {
				portBindings = append(portBindings, fmt.Sprintf("%s:%s", portBinding.HostIP, portBinding.HostPort))
			}
			builder.WriteString(fmt.Sprintf("%s -> %v\n", port, portBindings))
		}
	}

	return builder.String()
}
