// Package containers provides a component for viewing container logs in a scrollable overlay.
package containers

import (
	"bufio"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	contxt "github.com/givensuman/containertui/internal/context"
)

// ContainerLogs displays and scrolls logs for a specific container.
type ContainerLogs struct {
	viewport      viewport.Model // Log viewport.
	containerItem *ContainerItem // Reference to container for fetching logs.
	lines         []string       // Current log lines.
	isLoaded      bool           // Marks if log stream goroutine running.
	err           error          // Holds error from log fetching.
	width         int
	height        int
	isAtBottom    bool          // If true, auto-scroll when new lines appear.
	cancelChannel chan struct{} // To stop log streaming goroutine.
}

func (model *ContainerLogs) Init() tea.Cmd {
	return model.streamLogsCmd()
}

// streamLogsCmd streams logs live and sends new lines as they arrive, until cancelled.
func (model *ContainerLogs) streamLogsCmd() tea.Cmd {
	containerID := model.containerItem.ID
	cancelChannel := model.cancelChannel
	return func() tea.Msg {
		reader, err := contxt.GetClient().OpenLogs(containerID)
		if err != nil {
			return logsLoadedMsg{lines: nil, err: err}
		}
		scanner := bufio.NewScanner(reader)
		for {
			select {
			case <-cancelChannel:
				return nil // Overlay closed, stop streaming.
			default:
				if !scanner.Scan() {
					return nil // End of stream.
				}
				line := scanner.Text()
				return newLogLineMsg{line: line}
			}
		}
	}
}

// logsLoadedMsg used internally to dispatch logs from async fetch to UI
// or signal error.
type logsLoadedMsg struct {
	lines []string
	err   error
}

type newLogLineMsg struct {
	line string
}

// Update implements the Bubbletea update loop for the logs overlay.
func (model *ContainerLogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logsLoadedMsg:
		model.isLoaded = true
		model.err = msg.err
		if msg.err == nil {
			model.lines = msg.lines
			model.viewport.SetContent(strings.Join(msg.lines, "\n"))
			model.isAtBottom = true
		} else {
			model.viewport.SetContent("Error loading logs: " + msg.err.Error())
		}
		return model, nil

	case newLogLineMsg:
		model.lines = append(model.lines, msg.line)
		model.viewport.SetContent(strings.Join(model.lines, "\n"))
		// If at the bottom or the log buffer size <= height, scroll to end.
		if model.isAtBottom || len(model.lines) <= model.viewport.Height {
			model.viewport.GotoBottom()
		}
		return model, model.streamLogsCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Cancel streaming goroutine.
			if model.cancelChannel != nil {
				close(model.cancelChannel)
			}
			return model, CloseOverlay()
		case "up", "down", "pgup", "pgdown", "mouse wheel up", "mouse wheel down":
			// Let viewport handle.
			viewportModel, _ := model.viewport.Update(msg)
			model.viewport = viewportModel
			newScroll := model.viewport.ScrollPercent()
			if newScroll < 0.99 {
				model.isAtBottom = false
			} else {
				model.isAtBottom = true
			}
			return model, nil
		}
		return model, nil
	case tea.WindowSizeMsg:
		model.setDimensions(msg.Width, msg.Height)
		return model, nil
	}
	// Pass through viewport and mouse messages.
	viewportModel, cmd := model.viewport.Update(msg)
	model.viewport = viewportModel
	return model, cmd
}

// setDimensions resizes the viewport and overlay on terminal window change.
func (model *ContainerLogs) setDimensions(width, height int) {
	model.width = width
	model.height = height
	model.viewport.Width = width - 4   // Leave padding for overlay border.
	model.viewport.Height = height - 6 // Leave padding for title/controls.
	if model.isLoaded && len(model.lines) > 0 {
		model.viewport.SetContent(strings.Join(model.lines, "\n"))
	}
}

// View renders the log overlay with controls and instructions.
func (model *ContainerLogs) View() string {
	return model.viewport.View()
}
