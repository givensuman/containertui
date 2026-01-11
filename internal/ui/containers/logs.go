package containers

import (
	"bytes"
	"io"
	"math"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/shared"
)

type (
	LogChunkMsg []byte
	LogErrorMsg error
)

type ContainerLogs struct {
	shared.Component
	style     lipgloss.Style
	container *ContainerItem
	logs      client.Logs
	content   bytes.Buffer
	viewport  viewport.Model
}

var (
	_ tea.Model             = (*ContainerLogs)(nil)
	_ shared.ComponentModel = (*ContainerLogs)(nil)
)

func waitForLogs(reader io.Reader) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil {
			return LogErrorMsg(err)
		}

		return LogChunkMsg(buf[:n])
	}
}

func newContainerLogs(container *ContainerItem) ContainerLogs {
	logs := context.GetClient().OpenLogs(container.ID)
	width, height := context.GetWindowSize()

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Primary()).
		Padding(0, 1)

	return ContainerLogs{
		style:     style,
		container: container,
		logs:      logs,
		viewport: viewport.New( // TODO: Can we improve this?
			int(math.Floor(float64(width)*0.7)),
			int(math.Floor(float64(height)*0.7)),
		),
	}
}

func (cl *ContainerLogs) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	cl.WindowWidth = msg.Width
	cl.WindowHeight = msg.Height
}

func (cl ContainerLogs) Init() tea.Cmd {
	return waitForLogs(cl.logs)
}

func (cl ContainerLogs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case LogChunkMsg:
		cl.content.Write(msg)

		cl.viewport.SetContent(cl.content.String())
		cl.viewport.GotoBottom()

		cmds = append(cmds, waitForLogs(cl.logs))

	case LogErrorMsg:
		if msg == io.EOF {
			return cl, nil
		}

		return cl, nil

	case tea.WindowSizeMsg:
		cl.UpdateWindowDimensions(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyEscape.String(), tea.KeyEsc.String():
			cl.logs.Close()
			return cl, CloseOverlay()
		}
	}

	cl.viewport, cmd = cl.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return cl, tea.Batch(cmds...)
}

func (cl ContainerLogs) View() string {
	return cl.style.Render(cl.viewport.View())
}
