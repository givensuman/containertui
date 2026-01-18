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

// MsgStatsTick is sent to trigger a stats refresh
type MsgStatsTick time.Time

// MsgContainerStats contains the stats for a container
type MsgContainerStats struct {
	ID    string
	Stats client.ContainerStats
	Err   error
}

// MsgContainerInspection contains the inspection data for a container
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

// Model is the main containers screen for the TUI.
// Manages two views: either normal container list or overlay (logs, delete dialog).
type Model struct {
	shared.Component
	// sessionState governs whether we're in main or overlay view
	sessionState sessionState
	// focusedView governs which panel is active (0: List, 1: Details)
	focusedView int

	// foreground is the currently active overlay model (logs, confirmation, etc)
	foreground tea.Model
	// background is the container list model
	background tea.Model
	// overlayModel manages overlay transitions and rendering
	overlayModel *overlay.Model

	// Stats related fields
	currentContainerID string
	cpuHistory         []float64
	lastStats          client.ContainerStats

	// Details view
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

	vp := viewport.New(0, 0)

	return Model{
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
		viewport:           vp,
		detailsKeybindings: newDetailsKeybindings(),
	}
}

// tickCmd generates a command to tick the stats loop
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return MsgStatsTick(t)
	})
}

// UpdateWindowDimensions updates all sub-models with latest terminal size.
// Called during resize events or window size changes.
func (m *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	m.WindowWidth = msg.Width
	m.WindowHeight = msg.Height

	// Calculate detail width/height for viewport
	lm := shared.NewLayoutManager(m.WindowWidth, m.WindowHeight)
	_, detail := lm.CalculateMasterDetail(lipgloss.NewStyle())

	// Update viewport size
	// We need to account for borders (2) and padding (2)
	width := detail.Width - 4
	height := detail.Height - 2
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	m.viewport.Width = width
	m.viewport.Height = height

	switch m.sessionState {
	case viewMain:
		if cl, ok := m.background.(ContainerList); ok {
			cl.UpdateWindowDimensions(msg)
			m.background = cl
		}
	case viewOverlay:
		// Forward dimension updates to correct overlay model
		switch fg := m.foreground.(type) {
		case *ContainerLogs:
			fg.setDimensions(msg.Width, msg.Height)
			m.foreground = fg
		case DeleteConfirmation:
			fg.UpdateWindowDimensions(msg)
			m.foreground = fg
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// Update handles all Bubbletea messages relevant to this containers screen.
// Manages both main view logic and overlay/dialog/confirmation sub-models.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle focus switching
	if m.sessionState == viewMain {
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "tab" {
				if m.focusedView == focusList {
					m.focusedView = focusDetails
				} else {
					m.focusedView = focusList
				}
				return m, nil
			}
		}
	}

	// Update foreground/background model based on session mode.
	switch m.sessionState {
	case viewMain:
		// Only update the focused model to prevent key stealing,
		// but always update background if it's not a key msg or if it's the focused one.
		// Actually, we usually want background to handle list updates even if not focused?
		// No, if we want to scroll details with j/k, list must NOT receive j/k.

		isKeyMsg := false
		if _, ok := msg.(tea.KeyMsg); ok {
			isKeyMsg = true
		}

		if !isKeyMsg || m.focusedView == focusList {
			bg, cmd := m.background.Update(msg)
			m.background = bg.(ContainerList)
			cmds = append(cmds, cmd)
		}

		if !isKeyMsg || m.focusedView == focusDetails {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

		// Check for selection change (if list updated)
		if cl, ok := m.background.(ContainerList); ok {
			item := cl.list.SelectedItem()
			if item != nil {
				if c, ok := item.(ContainerItem); ok {
					if c.ID != m.currentContainerID {
						m.currentContainerID = c.ID
						m.cpuHistory = make([]float64, 0)
						// Trigger inspection
						cmds = append(cmds, func() tea.Msg {
							info, err := context.GetClient().InspectContainer(c.ID)
							return MsgContainerInspection{ID: c.ID, Container: info, Err: err}
						})
					}
				}
			}
		}

	case viewOverlay:
		fg, cmd := m.foreground.Update(msg)
		m.foreground = fg
		cmds = append(cmds, cmd)
	}

	// Handle special message types (resize, open/close dialogs, etc)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.UpdateWindowDimensions(msg)

	case MessageCloseOverlay:
		m.sessionState = viewMain

	case MessageOpenDeleteConfirmationDialog:
		m.foreground = newDeleteConfirmation(msg.requestedContainersToDelete...)
		m.sessionState = viewOverlay
		cmds = append(cmds, m.foreground.Init())

	case MsgStatsTick:
		cmds = append(cmds, tickCmd())

		// Only fetch stats if we are in main view and have a selected container
		if m.sessionState == viewMain {
			if cl, ok := m.background.(ContainerList); ok {
				item := cl.list.SelectedItem()
				if item != nil {
					if c, ok := item.(ContainerItem); ok && c.State == "running" {
						// Fetch stats
						cmds = append(cmds, func() tea.Msg {
							stats, err := context.GetClient().GetContainerStats(c.ID)
							return MsgContainerStats{ID: c.ID, Stats: stats, Err: err}
						})
					}
				}
			}
		}

	case MsgContainerInspection:
		if msg.ID == m.currentContainerID && msg.Err == nil {
			m.inspection = msg.Container
			content := formatInspection(m.inspection, m.lastStats, m.cpuHistory, m.viewport.Width)
			m.viewport.SetContent(content)
		}

	case MsgContainerStats:
		if msg.ID == m.currentContainerID && msg.Err == nil {
			m.lastStats = msg.Stats
			m.cpuHistory = append(m.cpuHistory, msg.Stats.CPUPercent)
			if len(m.cpuHistory) > 30 {
				m.cpuHistory = m.cpuHistory[1:]
			}
			// Update content to reflect stats
			content := formatInspection(m.inspection, m.lastStats, m.cpuHistory, m.viewport.Width)
			m.viewport.SetContent(content)
		}
	}

	// Always update the overlay model and sync submodels
	m.overlayModel.Foreground = m.foreground
	m.overlayModel.Background = m.background

	updatedOverlayModel, cmd := m.overlayModel.Update(msg)
	overlayModel, ok := updatedOverlayModel.(*overlay.Model)
	if ok {
		m.overlayModel = overlayModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI for containers.
// If an overlay is open, delegates to overlay model’s view,
// Otherwise renders the main container background view.
func (m Model) View() string {
	if m.sessionState == viewOverlay && m.foreground != nil {
		return m.overlayModel.View()
	}

	// Get dimensions for master (list) and detail (inspect)
	lm := shared.NewLayoutManager(m.WindowWidth, m.WindowHeight)
	_, detail := lm.CalculateMasterDetail(lipgloss.NewStyle())

	// Render the list (background)
	listView := m.background.View()

	borderColor := colors.Muted()
	if m.focusedView == focusDetails {
		borderColor = colors.Primary()
	}

	detailStyle := lipgloss.NewStyle().
		Width(detail.Width - 2). // Subtract 2 for border width compensation
		Height(detail.Height).   // Ensure height matches master list
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1)

	var detailContent string
	if m.currentContainerID != "" {
		detailContent = m.viewport.View()
	} else {
		detailContent = lipgloss.NewStyle().Foreground(colors.Muted()).Render("No container selected")
	}

	detailView := detailStyle.Render(detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}

// Helper to expose keys for help menu
func (m Model) ShortHelp() []key.Binding {
	// If in overlay mode, delegate to overlay if possible, or just return empty
	if m.sessionState == viewOverlay {
		if h, ok := m.foreground.(help.KeyMap); ok {
			return h.ShortHelp()
		}
		// Special case for delete confirmation if it implements something custom,
		// but currently DeleteConfirmation uses standard bubbletea, we might need to cast.
		// For now, let's focus on main view.
		return nil
	}

	if m.focusedView == focusList {
		if cl, ok := m.background.(ContainerList); ok {
			return cl.list.ShortHelp()
		}
	} else if m.focusedView == focusDetails {
		// Details view keybindings (scroll)
		return []key.Binding{
			m.detailsKeybindings.Up,
			m.detailsKeybindings.Down,
		}
	}

	return nil
}

// FullHelp returns keybindings for the expanded help view.
// It uses pointer receiver *Model because in the future we might want to cache or mutate state,
// but for interface satisfaction it needs to match. Since ShortHelp uses value receiver (because Model is often passed by value in bubbletea updates),
// we should check how it is called. ui/ui.go converts to interface helpProvider.
// If Model is large, passing by value is expensive.
// However, the current signature in UI is func (m Model) View().
func (m Model) FullHelp() [][]key.Binding {
	if m.sessionState == viewOverlay {
		return nil // Or implement overlay specific help
	}

	if m.focusedView == focusList {
		if cl, ok := m.background.(ContainerList); ok {
			return cl.list.FullHelp()
		}
	} else if m.focusedView == focusDetails {
		return [][]key.Binding{
			{
				m.detailsKeybindings.Up,
				m.detailsKeybindings.Down,
				m.detailsKeybindings.Switch,
			},
		}
	}
	return nil
}

func formatInspection(c types.ContainerJSON, stats client.ContainerStats, cpuHistory []float64, width int) string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).MarginBottom(1)
	b.WriteString(headerStyle.Render(fmt.Sprintf("%s (%s)", c.Name, c.ID[:12])) + "\n")

	// Basic Info
	infoStyle := lipgloss.NewStyle().Foreground(colors.Text()).MarginBottom(1)
	stateColor := colors.Success()
	if !c.State.Running {
		stateColor = colors.Muted()
	}
	stateStr := lipgloss.NewStyle().Foreground(stateColor).Render(c.State.Status)

	b.WriteString(infoStyle.Render(fmt.Sprintf("Image: %s", c.Config.Image)) + "\n")
	b.WriteString(fmt.Sprintf("State: %s\n\n", stateStr))

	// Stats Graph (if running)
	if c.State.Running && len(cpuHistory) > 0 {
		graphHeight := 8
		graphWidth := width - 10
		if graphWidth > 10 {
			graph := asciigraph.Plot(cpuHistory,
				asciigraph.Height(graphHeight),
				asciigraph.Width(graphWidth),
				asciigraph.Caption("CPU Usage (%)"),
				asciigraph.SeriesColors(
					asciigraph.Blue,
				),
			)
			b.WriteString(graph + "\n\n")

			// Additional Stats
			statsStyle := lipgloss.NewStyle().Foreground(colors.Secondary())
			memUsageMB := stats.MemUsage / 1024 / 1024
			memLimitMB := stats.MemLimit / 1024 / 1024
			b.WriteString(statsStyle.Render(fmt.Sprintf("CPU: %.2f%% | Mem: %.0fMB / %.0fMB",
				stats.CPUPercent, memUsageMB, memLimitMB)) + "\n\n")
		}
	}

	// Helper for sections
	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(colors.Secondary()).Underline(true).MarginTop(1).MarginBottom(0)

	// Config
	b.WriteString(sectionHeader.Render("Configuration") + "\n")
	b.WriteString(fmt.Sprintf("Cmd: %v\n", c.Config.Cmd))
	b.WriteString(fmt.Sprintf("Entrypoint: %v\n", c.Config.Entrypoint))
	b.WriteString(fmt.Sprintf("WorkingDir: %s\n", c.Config.WorkingDir))

	// Environment
	if len(c.Config.Env) > 0 {
		b.WriteString("\n" + sectionHeader.Render("Environment Variables") + "\n")
		for _, e := range c.Config.Env {
			b.WriteString(e + "\n")
		}
	}

	// Mounts
	if len(c.Mounts) > 0 {
		b.WriteString("\n" + sectionHeader.Render("Mounts") + "\n")
		for _, m := range c.Mounts {
			b.WriteString(fmt.Sprintf("%s -> %s (%s)\n", m.Source, m.Destination, m.Type))
		}
	}

	// Ports
	if c.NetworkSettings != nil && len(c.NetworkSettings.Ports) > 0 {
		b.WriteString("\n" + sectionHeader.Render("Ports") + "\n")
		for port, bindings := range c.NetworkSettings.Ports {
			var binds []string
			for _, b := range bindings {
				binds = append(binds, fmt.Sprintf("%s:%s", b.HostIP, b.HostPort))
			}
			b.WriteString(fmt.Sprintf("%s -> %v\n", port, binds))
		}
	}

	return b.String()
}
