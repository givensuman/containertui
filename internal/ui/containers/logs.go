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
	viewport  viewport.Model // log viewport
	container *ContainerItem // reference to container for fetching logs
	lines     []string       // current log lines
	loaded    bool           // Marks if log stream goroutine running
	error     error          // holds error from log fetching
	width     int
	height    int
	atBottom  bool          // If true, auto-scroll when new lines appear
	cancelCh  chan struct{} // To stop log streaming goroutine
}

func (cl *ContainerLogs) Init() tea.Cmd {
	return cl.streamLogsCmd()
}

// streamLogsCmd streams logs live and sends new lines as they arrive, until cancelled.
func (cl *ContainerLogs) streamLogsCmd() tea.Cmd {
	containerID := cl.container.ID
	cancelCh := cl.cancelCh
	return func() tea.Msg {
		reader, err := contxt.GetClient().OpenLogs(containerID)
		if err != nil {
			return logsLoadedMsg{lines: nil, err: err}
		}
		scanner := bufio.NewScanner(reader)
		for {
			select {
			case <-cancelCh:
				return nil // Overlay closed, stop streaming
			default:
				if !scanner.Scan() {
					return nil // end of stream
				}
				line := scanner.Text()
				return newLogLineMsg{line: line}
			}
		}
	}
}

// logsLoadedMsg used internally to dispatch logs from async fetch to UI
// or signal error
type logsLoadedMsg struct {
	lines []string
	err   error
}

type newLogLineMsg struct {
	line string
}

// Update implements the Bubbletea update loop for the logs overlay.
func (cl *ContainerLogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logsLoadedMsg:
		cl.loaded = true
		cl.error = msg.err
		if msg.err == nil {
			cl.lines = msg.lines
			cl.viewport.SetContent(strings.Join(msg.lines, "\n"))
			cl.atBottom = true
		} else {
			cl.viewport.SetContent("Error loading logs: " + msg.err.Error())
		}
		return cl, nil

	case newLogLineMsg:
		cl.lines = append(cl.lines, msg.line)
		cl.viewport.SetContent(strings.Join(cl.lines, "\n"))
		// If at the bottom or the log buffer size <= height, scroll to end
		if cl.atBottom || len(cl.lines) <= cl.viewport.Height {
			cl.viewport.GotoBottom()
		}
		return cl, cl.streamLogsCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Cancel streaming goroutine
			if cl.cancelCh != nil {
				close(cl.cancelCh)
			}
			return cl, CloseOverlay()
		case "up", "down", "pgup", "pgdown", "mouse wheel up", "mouse wheel down":
			// Let viewport handle
			vp, _ := cl.viewport.Update(msg)
			cl.viewport = vp
			newScroll := cl.viewport.ScrollPercent()
			if newScroll < 0.99 {
				cl.atBottom = false
			} else {
				cl.atBottom = true
			}
			return cl, nil
		}
		return cl, nil
	case tea.WindowSizeMsg:
		cl.setDimensions(msg.Width, msg.Height)
		return cl, nil
	}
	// Pass through viewport and mouse messages
	vp, cmd := cl.viewport.Update(msg)
	cl.viewport = vp
	return cl, cmd
}

// setDimensions resizes the viewport and overlay on terminal window change.
func (cl *ContainerLogs) setDimensions(width, height int) {
	cl.width = width
	cl.height = height
	cl.viewport.Width = width - 4   // leave padding for overlay border
	cl.viewport.Height = height - 6 // leave padding for title/controls
	if cl.loaded && len(cl.lines) > 0 {
		cl.viewport.SetContent(strings.Join(cl.lines, "\n"))
	}
}

// View renders the log overlay with controls and instructions.
func (cl *ContainerLogs) View() string {
	return cl.viewport.View()
}
