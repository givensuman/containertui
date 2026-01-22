package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

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

func RenderOverlay(background, foreground string, width, height int) tea.View {
	bgLayer := lipgloss.NewLayer(background).Width(width).Height(height)

	fgLayer := lipgloss.NewLayer(foreground)

	fgWidth := fgLayer.GetWidth()
	fgHeight := fgLayer.GetHeight()
	centerX := (width - fgWidth) / 2
	centerY := (height - fgHeight) / 2

	fgLayer = fgLayer.X(centerX).Y(centerY).Z(1)

	canvas := lipgloss.NewCanvas(bgLayer, fgLayer)

	return tea.NewView(canvas.Render())
}

func RenderOverlayString(background, foreground string, width, height int) string {
	bgLayer := lipgloss.NewLayer(background).Width(width).Height(height)

	fgLayer := lipgloss.NewLayer(foreground)

	fgWidth := fgLayer.GetWidth()
	fgHeight := fgLayer.GetHeight()
	centerX := (width - fgWidth) / 2
	centerY := (height - fgHeight) / 2

	fgLayer = fgLayer.X(centerX).Y(centerY).Z(1)

	canvas := lipgloss.NewCanvas(bgLayer, fgLayer)

	return canvas.Render()
}
