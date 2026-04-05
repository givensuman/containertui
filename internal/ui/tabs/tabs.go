// Package tabs implements the tab navigation component for the TUI.
package tabs

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/ui/base"
)

type Tab int

const (
	Containers Tab = iota
	Images
	Volumes
	Networks
	Services
	Browse
)

func (t Tab) String() string {
	return [...]string{
		"Containers",
		"Images",
		"Volumes",
		"Networks",
		"Services",
		"Browse",
	}[t]
}

// TabFromString converts a string to a Tab, returns -1 if invalid
func TabFromString(s string) Tab {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "containers":
		return Containers
	case "images":
		return Images
	case "volumes":
		return Volumes
	case "networks":
		return Networks
	case "browse":
		return Browse
	default:
		return -1
	}
}

// IsValidTab checks if a tab string is valid
func IsValidTab(s string) bool {
	return TabFromString(s) != -1
}

// AllTabNames returns all valid tab names
func AllTabNames() []string {
	return []string{"containers", "images", "volumes", "networks", "browse"}
}

type KeyMap struct {
	SwitchToContainers key.Binding
	SwitchToImages     key.Binding
	SwitchToVolumes    key.Binding
	SwitchToNetworks   key.Binding
	SwitchToServices   key.Binding
	SwitchToBrowse     key.Binding
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
		SwitchToBrowse: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "browse"),
		),
	}
}

type Model struct {
	base.Component
	ActiveTab Tab
	Tabs      []Tab
	KeyMap    KeyMap
}

func New(startupTab Tab) Model {
	if startupTab == Services {
		startupTab = Containers
	}

	return Model{
		ActiveTab: startupTab,
		Tabs:      []Tab{Containers, Images, Volumes, Networks, Browse},
		KeyMap:    NewKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
		case key.Matches(msg, m.KeyMap.SwitchToBrowse):
			m.ActiveTab = Browse
		}
	case tea.WindowSizeMsg:
		m.WindowWidth = msg.Width
		m.WindowHeight = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
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

	return lipgloss.JoinHorizontal(lipgloss.Top, row, gap)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	// Active tab: filled pill/box to make selected section explicit.
	activeTabStyle = lipgloss.NewStyle().
			Foreground(colors.PrimaryText()).
			Background(colors.Primary()).
			Padding(0, 1).
			MarginRight(1).
			Bold(true)

	// Inactive tab: outlined pill/box for stronger affordance than plain text.
	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colors.Text()).
				Background(colors.Border()).
				Padding(0, 1).
				MarginRight(1).
				Bold(false)
)
