// Package tabs implements the tab navigation component for the TUI.
package tabs

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type Tab int

const (
	Containers Tab = iota
	Images
	Volumes
	Networks
	Services
)

func (t Tab) String() string {
	return [...]string{
		"Containers",
		"Images",
		"Volumes",
		"Networks",
		"Services",
	}[t]
}

type KeyMap struct {
	SwitchToContainers key.Binding
	SwitchToImages     key.Binding
	SwitchToVolumes    key.Binding
	SwitchToNetworks   key.Binding
	SwitchToServices   key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		SwitchToContainers: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "containers"),
		),
		SwitchToImages: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "images"),
		),
		SwitchToVolumes: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "volumes"),
		),
		SwitchToNetworks: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "networks"),
		),
		SwitchToServices: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "services"),
		),
	}
}

type Model struct {
	shared.Component
	ActiveTab Tab
	Tabs      []Tab
	KeyMap    KeyMap
}

func New() Model {
	return Model{
		ActiveTab: Containers,
		Tabs:      []Tab{Containers, Images, Volumes, Networks, Services},
		KeyMap:    NewKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.SwitchToContainers):
			m.ActiveTab = Containers
		case key.Matches(msg, m.KeyMap.SwitchToImages):
			m.ActiveTab = Images
		case key.Matches(msg, m.KeyMap.SwitchToVolumes):
			m.ActiveTab = Volumes
		case key.Matches(msg, m.KeyMap.SwitchToNetworks):
			m.ActiveTab = Networks
		case key.Matches(msg, m.KeyMap.SwitchToServices):
			m.ActiveTab = Services
		}
	case tea.WindowSizeMsg:
		m.WindowWidth = msg.Width
		m.WindowHeight = msg.Height
	}
	return m, nil
}

func (m Model) View() tea.View {
	var tabs []string
	for _, t := range m.Tabs {
		if m.ActiveTab == t {
			tabs = append(tabs, activeTabStyle.Render(t.String()))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(t.String()))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	// Fill the rest of the line with the gap style
	// We need to account for borders in width calculation
	gapWidth := maxInt(0, m.WindowWidth-lipgloss.Width(row)-2) // -2 for safety margin
	gap := strings.Repeat(" ", gapWidth)

	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	// Active tab: Primary color background, distinct text
	activeTabStyle = lipgloss.NewStyle().
			Foreground(colors.Primary()).
			Padding(0, 1).
			Bold(true)

	// Inactive tab: Muted color, still needs rounded border structure to align with active tab but we hide top/sides?
	// Actually, standard TUI tab design often puts a bottom border on the GAP, and NO bottom border on the active tab.
	// But let's try to make them all visible first.
	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colors.Muted()).
				Padding(0, 1).
				Bold(false)
)
