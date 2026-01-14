// Package containers defines the containers component
package containers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/givensuman/containertui/internal/ui/shared"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type sessionState int

const (
	viewMain sessionState = iota
	viewOverlay
)

// Model is the main containers screen for the TUI.
// Manages two views: either normal container list or overlay (logs, delete dialog).
type Model struct {
	shared.Component
	// sessionState governs whether we're in main or overlay view
	sessionState sessionState
	// foreground is the currently active overlay model (logs, confirmation, etc)
	foreground tea.Model
	// background is the container list model
	background tea.Model
	// overlayModel manages overlay transitions and rendering
	overlayModel *overlay.Model
}

var (
	_ tea.Model             = (*Model)(nil)
	_ shared.ComponentModel = (*Model)(nil)
)

func New() Model {
	deleteConfirmation := newDeleteConfirmation(nil)
	containerList := newContainerList()

	return Model{
		sessionState: viewMain,
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
	}
}

// UpdateWindowDimensions updates all sub-models with latest terminal size.
// Called during resize events or window size changes.
func (m *Model) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	m.WindowWidth = msg.Width
	m.WindowHeight = msg.Height

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
	return nil
}

// Update handles all Bubbletea messages relevant to this containers screen.
// Manages both main view logic and overlay/dialog/confirmation sub-models.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Update foreground/background model based on session mode.
	switch m.sessionState {
	case viewMain:
		bg, cmd := m.background.Update(msg)
		m.background = bg.(ContainerList)
		cmds = append(cmds, cmd)
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

	case MessageOpenContainerLogs:
		m.foreground = newContainerLogs(msg.container)
		m.sessionState = viewOverlay
		cmds = append(cmds, m.foreground.Init())
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

	return m.background.View()
}
