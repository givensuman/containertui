// Package containers provides a component for viewing container logs in a scrollable overlay.
package containers

import (
	"bufio"
	stdcontext "context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
)

type logsKeybindings struct {
	ToggleFollow key.Binding
	Clear        key.Binding
	Search       key.Binding
	Copy         key.Binding
}

func newLogsKeybindings() logsKeybindings {
	return logsKeybindings{
		ToggleFollow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow/pause"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear logs"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search logs"),
		),
		Copy: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy logs"),
		),
	}
}

// ContainerLogs displays and scrolls logs for a specific container.
type ContainerLogs struct {
	viewport      viewport.Model // Log viewport.
	containerItem *ContainerItem // Reference to container for fetching logs.
	lines         []string       // Current log lines.
	isLoaded      bool           // Marks if log stream goroutine running.
	err           error          // Holds error from log fetching.
	width         int
	height        int
	isAtBottom    bool // If true, auto-scroll when new lines appear.
	follow        bool
	searchMode    bool
	searchQuery   string
	filteredLines []string
	keybindings   logsKeybindings
	cancelChannel chan struct{}         // To stop log streaming goroutine.
	cancelFunc    stdcontext.CancelFunc // To cancel the context used for OpenLogs.
	logCtx        stdcontext.Context    // Context for log streaming.
}

func NewContainerLogs(item ContainerItem, width, height int) *ContainerLogs {
	ctx, cancel := stdcontext.WithCancel(stdcontext.Background())
	model := &ContainerLogs{
		containerItem: &item,
		lines:         []string{},
		filteredLines: nil,
		isLoaded:      false,
		isAtBottom:    true,
		follow:        true,
		searchMode:    false,
		searchQuery:   "",
		keybindings:   newLogsKeybindings(),
		cancelChannel: make(chan struct{}),
		cancelFunc:    cancel,
		logCtx:        ctx,
		width:         width,
		height:        height,
	}
	model.setDimensions(width, height)
	return model
}

func (model *ContainerLogs) Init() tea.Cmd {
	return model.streamLogsCmd()
}

// streamLogsCmd streams logs live and sends new lines as they arrive, until cancelled.
func (model *ContainerLogs) streamLogsCmd() tea.Cmd {
	containerID := model.containerItem.ID
	cancelChannel := model.cancelChannel
	ctx := model.logCtx
	return func() tea.Msg {
		logs, err := state.GetBackend().OpenLogs(ctx, containerID)
		if err != nil {
			return logsLoadedMsg{lines: nil, err: err}
		}
		defer logs.Close()
		scanner := bufio.NewScanner(logs.Stream)
		for {
			select {
			case <-cancelChannel:
				return nil // Overlay closed, stop streaming.
			case <-ctx.Done():
				return nil // Context cancelled, stop streaming.
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
func (model *ContainerLogs) Update(msg tea.Msg) (*ContainerLogs, tea.Cmd) {
	switch msg := msg.(type) {
	case logsLoadedMsg:
		model.isLoaded = true
		model.err = msg.err
		if msg.err == nil {
			model.lines = msg.lines
			model.refreshViewport()
			model.isAtBottom = true
		} else {
			model.viewport.SetContent("Error loading logs: " + msg.err.Error())
		}
		return model, nil

	case newLogLineMsg:
		model.lines = append(model.lines, msg.line)
		model.refreshViewport()
		// If at the bottom or the log buffer size <= height, scroll to end.
		if model.follow && (model.isAtBottom || len(model.lines) <= model.viewport.Height()) {
			model.viewport.GotoBottom()
		}
		return model, model.streamLogsCmd()

	case tea.KeyPressMsg:
		if model.searchMode {
			switch msg.String() {
			case "esc":
				model.searchMode = false
				model.searchQuery = ""
				model.filteredLines = nil
				model.refreshViewport()
				return model, nil
			case "enter":
				model.searchMode = false
				model.applySearchFilter()
				return model, nil
			case "backspace":
				if len(model.searchQuery) > 0 {
					model.searchQuery = model.searchQuery[:len(model.searchQuery)-1]
				}
				return model, nil
			default:
				if msg.Text != "" {
					model.searchQuery += msg.Text
				}
				return model, nil
			}
		}

		switch msg.String() {
		case "q", "esc":
			// Cancel streaming goroutine.
			if model.cancelChannel != nil {
				close(model.cancelChannel)
			}
			// Cancel the context to stop the Docker log stream.
			if model.cancelFunc != nil {
				model.cancelFunc()
			}
			return model, func() tea.Msg { return base.CloseDialogMessage{} }
		case "f":
			model.follow = !model.follow
			if model.follow {
				model.viewport.GotoBottom()
			}
			return model, nil
		case "c":
			model.lines = []string{}
			model.filteredLines = nil
			model.refreshViewport()
			return model, nil
		case "/":
			model.searchMode = true
			model.searchQuery = ""
			return model, nil
		case "y":
			if err := clipboard.WriteAll(strings.Join(model.activeLines(), "\n")); err != nil {
				model.err = fmt.Errorf("failed to copy logs: %w", err)
			}
			return model, nil
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

func (model *ContainerLogs) activeLines() []string {
	if len(model.filteredLines) > 0 || (model.searchQuery != "" && model.filteredLines != nil) {
		return model.filteredLines
	}
	return model.lines
}

func (model *ContainerLogs) applySearchFilter() {
	query := strings.ToLower(strings.TrimSpace(model.searchQuery))
	if query == "" {
		model.filteredLines = nil
		model.refreshViewport()
		return
	}

	filtered := make([]string, 0, len(model.lines))
	for _, line := range model.lines {
		if strings.Contains(strings.ToLower(line), query) {
			filtered = append(filtered, line)
		}
	}
	model.filteredLines = filtered
	model.refreshViewport()
}

func (model *ContainerLogs) refreshViewport() {
	lines := model.activeLines()
	model.viewport.SetContent(strings.Join(lines, "\n"))
}

// setDimensions resizes the viewport and overlay on terminal window change.
func (model *ContainerLogs) setDimensions(width, height int) {
	model.width = width
	model.height = height
	model.viewport = viewport.New(
		viewport.WithWidth(width-4),   // Leave padding for overlay border.
		viewport.WithHeight(height-6), // Leave padding for title/controls.
	)
	if model.isLoaded && len(model.lines) > 0 {
		model.refreshViewport()
	}
}

// View renders the log overlay with controls and instructions.
func (model *ContainerLogs) View() string {
	return model.viewport.View()
}
