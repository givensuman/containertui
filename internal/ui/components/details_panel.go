// Package components provides reusable UI components.
package components

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

// DetailsPanel handles common details panel functionality including:
// - Format toggling (JSON/YAML)
// - Clipboard copying
// - Scroll position management
// - Content inspection display
type DetailsPanel struct {
	currentFormat   string
	scrollPositions map[string]int
	currentID       string // Current resource ID being displayed
}

// NewDetailsPanel creates a new details panel.
func NewDetailsPanel() DetailsPanel {
	return DetailsPanel{
		currentFormat:   "",
		scrollPositions: make(map[string]int),
		currentID:       "",
	}
}

// GetCurrentFormat returns the current format (json or yaml).
func (dp *DetailsPanel) GetCurrentFormat() string {
	if dp.currentFormat == "" {
		cfg := state.GetConfig()
		format := cfg.InspectionFormat
		if format == "" {
			return "yaml"
		}
		return format
	}
	return dp.currentFormat
}

// SetCurrentFormat sets the format explicitly.
func (dp *DetailsPanel) SetCurrentFormat(format string) {
	dp.currentFormat = format
}

// GetCurrentID returns the current resource ID.
func (dp *DetailsPanel) GetCurrentID() string {
	return dp.currentID
}

// SetCurrentID sets the current resource ID and saves scroll position.
func (dp *DetailsPanel) SetCurrentID(id string, vp *viewport.Model) {
	// Save scroll position for old ID
	if dp.currentID != "" && dp.currentID != id && vp != nil {
		dp.scrollPositions[dp.currentID] = vp.YOffset()
	}
	dp.currentID = id
}

// SaveScrollPosition saves the viewport scroll position for the current resource.
func (dp *DetailsPanel) SaveScrollPosition(vp *viewport.Model) {
	if dp.currentID != "" && vp != nil {
		dp.scrollPositions[dp.currentID] = vp.YOffset()
	}
}

// RestoreScrollPosition restores the viewport scroll position for the current resource.
// Returns a command that sets the viewport offset.
func (dp *DetailsPanel) RestoreScrollPosition(vp *viewport.Model) {
	if dp.currentID != "" && vp != nil {
		if offset, exists := dp.scrollPositions[dp.currentID]; exists {
			vp.SetYOffset(offset)
		}
	}
}

// HandleToggleFormat toggles between JSON and YAML format.
// Returns the new format and a notification command.
func (dp *DetailsPanel) HandleToggleFormat() (string, tea.Cmd) {
	currentFormat := dp.GetCurrentFormat()

	if currentFormat == "json" {
		dp.currentFormat = "yaml"
	} else {
		dp.currentFormat = "json"
	}

	return dp.currentFormat, notifications.ShowSuccess("Switched to " + dp.currentFormat)
}

// HandleCopyToClipboard copies the inspection data to clipboard.
// Returns a notification command (success or error).
func (dp *DetailsPanel) HandleCopyToClipboard(data any) tea.Cmd {
	if data == nil {
		return nil
	}

	format := infopanel.GetOutputFormat()
	currentFormat := dp.GetCurrentFormat()
	if currentFormat == "json" {
		format = infopanel.FormatJSON
	} else {
		format = infopanel.FormatYAML
	}

	bytes, err := infopanel.MarshalToFormat(data, format)
	if err != nil {
		return notifications.ShowError(err)
	}

	if err := clipboard.WriteAll(string(bytes)); err != nil {
		return notifications.ShowError(err)
	}

	return notifications.ShowSuccess("Copied to clipboard")
}

// GetFormatForDisplay returns the infopanel format constant for the current format.
func (dp *DetailsPanel) GetFormatForDisplay() infopanel.OutputFormat {
	currentFormat := dp.GetCurrentFormat()
	if currentFormat == "json" {
		return infopanel.FormatJSON
	}
	return infopanel.FormatYAML
}

// ClearScrollPositions clears all saved scroll positions.
func (dp *DetailsPanel) ClearScrollPositions() {
	dp.scrollPositions = make(map[string]int)
}

// ResetCurrentID resets the current ID (useful when resource is deleted).
func (dp *DetailsPanel) ResetCurrentID() {
	dp.currentID = ""
}
