// Package notifications provides a component for displaying toast-like notifications.
package notifications

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Level int

const (
	Info Level = iota
	Error
	Success
)

type Notification struct {
	ID        int64
	Message   string
	Level     Level
	Timestamp time.Time
	Duration  time.Duration
}

type Model struct {
	notifications []Notification
	nextID        int64
	width         int
	height        int
}

type AddNotificationMsg struct {
	Message  string
	Level    Level
	Duration time.Duration
}

type RemoveNotificationMsg struct {
	ID int64
}

func New() Model {
	return Model{
		notifications: []Notification{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case AddNotificationMsg:
		id := m.nextID
		m.nextID++
		n := Notification{
			ID:        id,
			Message:   msg.Message,
			Level:     msg.Level,
			Timestamp: time.Now(),
			Duration:  msg.Duration,
		}
		m.notifications = append(m.notifications, n)
		cmds = append(cmds, tick(id, msg.Duration))

	case RemoveNotificationMsg:
		var newNotifs []Notification
		for _, n := range m.notifications {
			if n.ID != msg.ID {
				newNotifs = append(newNotifs, n)
			}
		}
		m.notifications = newNotifs

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, tea.Batch(cmds...)
}

func tick(id int64, d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return RemoveNotificationMsg{ID: id}
	})
}

// Helper commands

func ShowInfo(msg string) tea.Cmd {
	return func() tea.Msg {
		return AddNotificationMsg{
			Message:  msg,
			Level:    Info,
			Duration: 5 * time.Second,
		}
	}
}

func ShowError(err error) tea.Cmd {
	return func() tea.Msg {
		return AddNotificationMsg{
			Message:  err.Error(),
			Level:    Error,
			Duration: 10 * time.Second,
		}
	}
}

func ShowSuccess(msg string) tea.Cmd {
	return func() tea.Msg {
		return AddNotificationMsg{
			Message:  msg,
			Level:    Success,
			Duration: 5 * time.Second,
		}
	}
}

// Styling

var (
	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#5A56E0")). // Purple-ish
			Padding(0, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5A56E0"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#E05656")). // Red
			Padding(0, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#E05656"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#56E095")). // Green
			Padding(0, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#56E095"))
)

func (m Model) View() string {
	if len(m.notifications) == 0 {
		return ""
	}

	var content string
	// Stack notifications from bottom up or top down?
	// Usually top-right means they stack downwards.

	for _, n := range m.notifications {
		var style lipgloss.Style
		switch n.Level {
		case Info:
			style = infoStyle
		case Error:
			style = errorStyle
		case Success:
			style = successStyle
		}

		content = lipgloss.JoinVertical(lipgloss.Left, content, style.Render(n.Message))
	}

	return content
}
