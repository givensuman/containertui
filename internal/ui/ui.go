// Package ui implements the terminal user interface
package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"

	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/containers"
	"github.com/givensuman/containertui/internal/ui/images"
	"github.com/givensuman/containertui/internal/ui/networks"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/tabs"
	"github.com/givensuman/containertui/internal/ui/volumes"
)

// Model represents the top-level Bubbletea UI model.
// Contains terminal dimensions and the containers model (main TUI view).
type Model struct {
	width              int                 // current terminal width
	height             int                 // current terminal height
	tabsModel          tabs.Model          // tabs component
	containersModel    containers.Model    // main containers view/model
	imagesModel        images.Model        // images view/model
	volumesModel       volumes.Model       // volumes view/model
	networksModel      networks.Model      // networks view/model
	notificationsModel notifications.Model // notifications view/model
	overlayModel       *overlay.Model      // global overlay for notifications
	help               help.Model          // global help model
}

func NewModel() Model {
	width, height := context.GetWindowSize()

	tabsM := tabs.New()
	conts := containers.New()
	imgs := images.New()
	vols := volumes.New()
	nets := networks.New()
	notifs := notifications.New()

	ov := overlay.New(notifs, conts, overlay.Right, overlay.Top, 0, 0)
	helpM := help.New()

	return Model{
		width:              width,
		height:             height,
		tabsModel:          tabsM,
		containersModel:    conts,
		imagesModel:        imgs,
		volumesModel:       vols,
		networksModel:      nets,
		notificationsModel: notifs,
		overlayModel:       ov,
		help:               helpM,
	}
}

// Init performs any initial commands for the Bubbletea app
// (no async initialization needed here)
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all Bubbletea update messages dispatched by the tea runtime.
// Manages window resizing and quit keys, delegates other updates to containersModel.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Message to pass to the overlay (and content)
	// By default it's the original message, but for resize events we adjust it.
	overlayMsg := msg

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update local size and global context
		m.width = msg.Width
		m.height = msg.Height
		context.SetWindowSize(msg.Width, msg.Height)

		// Update tabs with full width
		newTabs, _ := m.tabsModel.Update(msg)
		m.tabsModel = newTabs.(tabs.Model)

		// Calculate content height (subtract tabs height which is usually 1-2 lines)
		// Tabs normally take up 3 lines (border + content + padding/margin potentially)
		// We'll reserve 3 lines to be safe and ensure the content doesn't overflow
		// Reserve 1 line for short help, so the bottom border isn't eaten by it.
		// Content = Height - Tabs(1) - Help(1) - Borders(2) = Height - 4
		contentHeight := msg.Height - 4
		if contentHeight < 0 {
			contentHeight = 0
		}

		contentMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: contentHeight,
		}
		overlayMsg = contentMsg

		// Update sub-models with adjusted height
		newConts, _ := m.containersModel.Update(contentMsg)
		m.containersModel = newConts.(containers.Model)

		newImgs, _ := m.imagesModel.Update(contentMsg)
		m.imagesModel = newImgs.(images.Model)

		newVols, _ := m.volumesModel.Update(contentMsg)
		m.volumesModel = newVols.(volumes.Model)

		newNets, _ := m.networksModel.Update(contentMsg)
		m.networksModel = newNets.(networks.Model)

		m.help.Width = msg.Width

	case tea.KeyMsg:
		// Handle quit signals (Ctrl-C, Ctrl-D)
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit
		}

		// Delegate tab switching to tabs model
		// This now handles 1-4
		newTabs, cmd := m.tabsModel.Update(msg)
		m.tabsModel = newTabs.(tabs.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Update help model for expansion toggling
		if msg.String() == "?" {
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	// Update Notifications
	notifsModel, cmd := m.notificationsModel.Update(msg)
	m.notificationsModel = notifsModel.(notifications.Model)
	cmds = append(cmds, cmd)

	// Update active view based on selected tab
	var activeView tea.Model
	switch m.tabsModel.ActiveTab {
	case tabs.Containers:
		activeView = m.containersModel
		// Forward non-window/non-quit messages to containers view model
		// (WindowSizeMsg is handled specifically above to adjust for tab height)
		// We forward everything to the containers model, including tab presses,
		// because the containers model now handles tab to switch focus between panes.
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			containersModel, cmd := m.containersModel.Update(msg)
			m.containersModel = containersModel.(containers.Model)
			cmds = append(cmds, cmd)
			activeView = m.containersModel
		}
	case tabs.Images:
		activeView = m.imagesModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			imagesModel, cmd := m.imagesModel.Update(msg)
			m.imagesModel = imagesModel.(images.Model)
			cmds = append(cmds, cmd)
			activeView = m.imagesModel
		}
	case tabs.Volumes:
		activeView = m.volumesModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			volumesModel, cmd := m.volumesModel.Update(msg)
			m.volumesModel = volumesModel.(volumes.Model)
			cmds = append(cmds, cmd)
			activeView = m.volumesModel
		}
	case tabs.Networks:
		activeView = m.networksModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			networksModel, cmd := m.networksModel.Update(msg)
			m.networksModel = networksModel.(networks.Model)
			cmds = append(cmds, cmd)
			activeView = m.networksModel
		}
	}

	// Sync overlay
	m.overlayModel.Foreground = m.notificationsModel
	m.overlayModel.Background = activeView

	// Update overlay (mostly for resizing if needed, though we handle window size manually too)
	// We ignore the returned model here as we hold the reference in m.overlayModel pointer,
	// but the library might return a new one or update internal state.
	updatedOverlay, cmd := m.overlayModel.Update(overlayMsg)
	if ov, ok := updatedOverlay.(*overlay.Model); ok {
		m.overlayModel = ov
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Helper interface for components that provide help
type helpProvider interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

// View renders the terminal as a string (delegated to containersModel).
func (m Model) View() string {
	// Render tabs at the top
	tabsView := m.tabsModel.View()

	// Render content below
	contentView := m.overlayModel.View()

	// Render help at the bottom
	var helpView string
	var currentHelp helpProvider

	switch m.tabsModel.ActiveTab {
	case tabs.Containers:
		currentHelp = m.containersModel
	case tabs.Images:
		currentHelp = m.imagesModel
	case tabs.Volumes:
		currentHelp = m.volumesModel
	case tabs.Networks:
		currentHelp = m.networksModel
	}

	if currentHelp != nil {
		helpView = m.help.View(currentHelp)
	}

	// If help is expanded, we need to handle the layout differently to overlay it.
	// However, simple string joining pushes content up.
	// To overlay, we can use lipgloss.Place or similar, but bubbletea doesn't support z-index natively for strings.
	// The standard way to "overlay" at the bottom without shifting top content is to:
	// 1. Render the main content (tabs + view) with full height.
	// 2. Render the help view.
	// 3. Subtract lines from the bottom of the main content equal to help height? No, that just clips it.

	// If we want it to *overlay* (cover) the bottom content:
	// We need to take the full view, split it by lines, replace the last N lines with the help view lines.

	fullView := lipgloss.JoinVertical(lipgloss.Top, tabsView, contentView)

	if helpView != "" {
		// Apply border to help view to distinguish it, ONLY if expanded
		// We need to apply it to the help content
		helpStyle := lipgloss.NewStyle().Width(m.width)

		if m.help.ShowAll {
			helpStyle = helpStyle.
				Border(lipgloss.ASCIIBorder(), true, false, false, false).
				BorderForeground(colors.Muted())
		}

		helpView = helpStyle.Render(helpView)

		helpLines := strings.Split(helpView, "\n")
		helpHeight := len(helpLines)
		fullLines := strings.Split(fullView, "\n")

		// Ensure we don't go out of bounds
		if len(fullLines) >= helpHeight {
			// Replace last N lines
			// We can't just replace because width might be different (transparent background?)
			// Help component usually has transparent background.
			// If we want a solid background for the help overlay so it hides content behind it:
			// We need to apply a style to helpView.

			// Let's assume helpView is opaque or we make it opaque.
			// Actually, just returning JoinVertical pushes content up because the total height exceeds terminal height.
			// We need to ensure the `contentView` height + `tabsView` height + `helpView` height <= m.height.
			// But we WANT it to overlay, meaning `contentView` uses full height, and help sits ON TOP.

			// We can use shared.Layer or just manual line replacement.
			// Let's stick to lines replacement for simplicity as it effectively "overlays" in a terminal.

			// Truncate fullView to make room for help if we were stacking, but we are overlaying.
			// So we take fullView (which is full screen), and overwrite the bottom lines.

			topLines := fullLines[:len(fullLines)-helpHeight]
			// We might need to pad help lines to full width to ensure they cover underlying text

			// Join top lines and help lines
			return strings.Join(append(topLines, helpLines...), "\n")
		}
	}

	return fullView
}

// Start the Bubbletea UI main loop.
// Returns error if Bubbletea program terminates abnormally.
func Start() error {
	model := NewModel()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
