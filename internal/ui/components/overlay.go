// Package shared defines shared UI logic
package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SimpleViewModel is a simple tea.Model that just renders a string.
// It is useful for wrapping a rendered string to pass as a model (e.g. for backgrounds).
type SimpleViewModel struct {
	Content string
}

func (m SimpleViewModel) Init() tea.Cmd {
	return nil
}

func (m SimpleViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m SimpleViewModel) View() tea.View {
	return tea.NewView(m.Content)
}

// RenderOverlay renders a foreground view centered on top of a background view.
// This replaces the bubbletea-overlay library using native Lipgloss v2.
// Takes string backgrounds and foregrounds, returns tea.View for Bubble Tea v2.
func RenderOverlay(background, foreground string, width, height int) tea.View {
	// Create background layer that fills the entire area
	bgLayer := lipgloss.NewLayer(background).Width(width).Height(height)

	// Create foreground layer centered
	fgLayer := lipgloss.NewLayer(foreground)

	// Center the foreground layer
	fgWidth := fgLayer.GetWidth()
	fgHeight := fgLayer.GetHeight()
	centerX := (width - fgWidth) / 2
	centerY := (height - fgHeight) / 2

	fgLayer = fgLayer.X(centerX).Y(centerY).Z(1) // Z=1 to render on top

	// Compose layers into a canvas
	canvas := lipgloss.NewCanvas(bgLayer, fgLayer)

	return tea.NewView(canvas.Render())
}

// RenderOverlayString is the same as RenderOverlay but returns a string.
func RenderOverlayString(background, foreground string, width, height int) string {
	// Create background layer that fills the entire area
	bgLayer := lipgloss.NewLayer(background).Width(width).Height(height)

	// Create foreground layer centered
	fgLayer := lipgloss.NewLayer(foreground)

	// Center the foreground layer
	fgWidth := fgLayer.GetWidth()
	fgHeight := fgLayer.GetHeight()
	centerX := (width - fgWidth) / 2
	centerY := (height - fgHeight) / 2

	fgLayer = fgLayer.X(centerX).Y(centerY).Z(1) // Z=1 to render on top

	// Compose layers into a canvas
	canvas := lipgloss.NewCanvas(bgLayer, fgLayer)

	return canvas.Render()
}
