package components

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/layout"
)

// progressTickMsg is sent periodically to update indeterminate progress
type progressTickMsg time.Time

// ProgressDialog shows a dialog with a progress bar and status message.
type ProgressDialog struct {
	base.WindowSize
	title         string
	status        string
	progress      progress.Model
	style         lipgloss.Style
	width         int
	height        int
	dialogSize    DialogSize
	showSpinner   bool
	maxPercent    float64 // Maximum percent to show for indeterminate progress
	tickIncrement float64 // Amount to increment per tick
	autoAdvance   bool
}

// NewProgressDialogWithBar creates a new progress dialog with a progress bar.
func NewProgressDialogWithBar(title string) ProgressDialog {
	width, height := state.GetWindowSize()

	prog := progress.New(
		progress.WithDefaultBlend(),
		progress.WithoutPercentage(),
	)

	dialog := ProgressDialog{
		title:         title,
		status:        "Initializing...",
		progress:      prog,
		width:         width,
		height:        height,
		dialogSize:    DialogSizeMedium,
		showSpinner:   false,
		maxPercent:    0.95, // Cap at 95% for indeterminate progress
		tickIncrement: 0.05, // Increment by 5% per tick
		autoAdvance:   false,
	}

	dialog.updateStyle()
	return dialog
}

// updateStyle applies the current dialog size to the style.
func (dialog *ProgressDialog) updateStyle() {
	dialog.style = lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder(), true, true).
		BorderForeground(colors.Primary())

	// Calculate dimensions based on size
	layoutManager := layout.NewLayoutManager(dialog.width, dialog.height)
	var dimensions layout.Dimensions
	switch dialog.dialogSize {
	case DialogSizeSmall:
		dimensions = layoutManager.CalculateSmall(dialog.style)
	case DialogSizeMedium:
		dimensions = layoutManager.CalculateMedium(dialog.style)
	case DialogSizeLarge:
		dimensions = layoutManager.CalculateLarge(dialog.style)
	default:
		dimensions = layoutManager.CalculateMedium(dialog.style)
	}

	dialog.style = dialog.style.Width(dimensions.Width).Height(dimensions.Height)

	// Update progress bar width
	dialog.progress.SetWidth(dimensions.ContentWidth - 4)
}

// UpdateWindowDimensions updates the dialog size.
func (dialog *ProgressDialog) UpdateWindowDimensions(msg tea.WindowSizeMsg) {
	dialog.width = msg.Width
	dialog.height = msg.Height
	dialog.updateStyle()
}

// SetStatus updates the status message.
func (dialog *ProgressDialog) SetStatus(status string) {
	dialog.status = status
}

// EnableAutoAdvance enables periodic indeterminate progress updates.
func (dialog *ProgressDialog) EnableAutoAdvance(maxPercent, tickIncrement float64) {
	dialog.autoAdvance = true
	if maxPercent > 0 {
		dialog.maxPercent = maxPercent
	}
	if tickIncrement > 0 {
		dialog.tickIncrement = tickIncrement
	}
}

// SetPercent updates the progress bar percentage (0.0 to 1.0).
func (dialog *ProgressDialog) SetPercent(percent float64) tea.Cmd {
	return dialog.progress.SetPercent(percent)
}

func (dialog ProgressDialog) Init() tea.Cmd {
	cmds := []tea.Cmd{dialog.progress.Init()}
	if dialog.autoAdvance {
		cmds = append(cmds, dialog.progress.SetPercent(0.05), tickProgressCmd())
	}
	return tea.Batch(cmds...)
}

// tickProgressCmd returns a command that sends a tick after a delay
func tickProgressCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return progressTickMsg(t)
	})
}

func (dialog ProgressDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dialog.UpdateWindowDimensions(msg)

	case progress.FrameMsg:
		progressModel, cmd := dialog.progress.Update(msg)
		dialog.progress = progressModel
		cmds = append(cmds, cmd)

	case progressTickMsg:
		if !dialog.autoAdvance {
			return dialog, nil
		}
		// Increment progress if below max
		if dialog.progress.Percent() < dialog.maxPercent {
			cmd := dialog.progress.IncrPercent(dialog.tickIncrement)
			cmds = append(cmds, cmd, tickProgressCmd())
		} else {
			// Still tick to keep checking
			cmds = append(cmds, tickProgressCmd())
		}
	}

	return dialog, tea.Batch(cmds...)
}

func (dialog ProgressDialog) View() tea.View {
	contentWidth := dialog.style.GetWidth()
	if contentWidth > 0 {
		frameSize := dialog.style.GetHorizontalFrameSize()
		contentWidth = contentWidth - frameSize
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Primary()).
		Width(contentWidth).
		Align(lipgloss.Center).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(dialog.title))
	b.WriteString("\n\n")

	// Status message
	statusStyle := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Width(contentWidth).
		Align(lipgloss.Left)
	b.WriteString(statusStyle.Render(dialog.status))
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(dialog.progress.View())
	b.WriteString("\n")

	return tea.NewView(dialog.style.Render(b.String()))
}

func (dialog ProgressDialog) String() string {
	contentWidth := dialog.style.GetWidth()
	if contentWidth > 0 {
		frameSize := dialog.style.GetHorizontalFrameSize()
		contentWidth = contentWidth - frameSize
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colors.Primary()).
		Width(contentWidth).
		Align(lipgloss.Center).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(dialog.title))
	b.WriteString("\n\n")

	// Status message
	statusStyle := lipgloss.NewStyle().
		Foreground(colors.Text()).
		Width(contentWidth).
		Align(lipgloss.Left)
	b.WriteString(statusStyle.Render(dialog.status))
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(dialog.progress.View())
	b.WriteString("\n")

	return dialog.style.Render(b.String())
}
