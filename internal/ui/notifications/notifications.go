// Package notifications provides a component for displaying toast-like notifications.
package notifications

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Level int

const (
	Info Level = iota
	Error
	Success
	Progress
)

type Notification struct {
	ID         int64
	Message    string
	Level      Level
	Timestamp  time.Time
	Duration   time.Duration
	Persistent bool // If true, won't auto-dismiss
}

type Model struct {
	notifications []Notification
	nextID        int64
	width         int
	height        int
}

type AddNotificationMsg struct {
	Message    string
	Level      Level
	Duration   time.Duration
	Persistent bool
	ID         int64 // Optional: if set, replaces notification with this ID
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
		id := msg.ID
		if id == 0 {
			id = m.nextID
			m.nextID++
		}

		// Check if we should replace existing notification with same ID
		replaced := false
		for i, n := range m.notifications {
			if n.ID == id {
				m.notifications[i] = Notification{
					ID:         id,
					Message:    msg.Message,
					Level:      msg.Level,
					Timestamp:  time.Now(),
					Duration:   msg.Duration,
					Persistent: msg.Persistent,
				}
				replaced = true
				break
			}
		}

		if !replaced {
			n := Notification{
				ID:         id,
				Message:    msg.Message,
				Level:      msg.Level,
				Timestamp:  time.Now(),
				Duration:   msg.Duration,
				Persistent: msg.Persistent,
			}
			m.notifications = append(m.notifications, n)
		}

		// Only set auto-dismiss if not persistent
		if !msg.Persistent && msg.Duration > 0 {
			cmds = append(cmds, tick(id, msg.Duration))
		}

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
			Message:    msg,
			Level:      Info,
			Duration:   5 * time.Second,
			Persistent: false,
		}
	}
}

func ShowError(err error) tea.Cmd {
	return func() tea.Msg {
		return AddNotificationMsg{
			Message:    err.Error(),
			Level:      Error,
			Duration:   10 * time.Second,
			Persistent: false,
		}
	}
}

func ShowSuccess(msg string) tea.Cmd {
	return func() tea.Msg {
		return AddNotificationMsg{
			Message:    msg,
			Level:      Success,
			Duration:   5 * time.Second,
			Persistent: false,
		}
	}
}

// ShowProgress shows a persistent progress notification that must be manually dismissed
// Returns the notification ID so it can be updated or removed later
func ShowProgress(msg string, id int64) tea.Cmd {
	return func() tea.Msg {
		return AddNotificationMsg{
			Message:    msg,
			Level:      Progress,
			Duration:   0,
			Persistent: true,
			ID:         id,
		}
	}
}

// DismissNotification removes a notification by ID
func DismissNotification(id int64) tea.Cmd {
	return func() tea.Msg {
		return RemoveNotificationMsg{ID: id}
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

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(lipgloss.Color("#E0A856")). // Orange/amber
			Padding(0, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#E0A856"))
)

func (m Model) View() tea.View {
	return tea.NewView(m.ViewString())
}

func (m Model) ViewString() string {
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
		case Progress:
			style = progressStyle
		}

		content = lipgloss.JoinVertical(lipgloss.Left, content, style.Render(n.Message))
	}

	return content
}
