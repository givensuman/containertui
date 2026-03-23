// Package icons provides centralized icon management for the UI.
package icons

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/state"
)

// IconSet contains icons for various resources and states.
type IconSet struct {
	// Resource types
	Container string
	Image     string
	Network   string
	Volume    string
	Service   string

	// Container/Service status indicators
	Running    string
	Stopped    string
	Paused     string
	Restarting string
	Removing   string
	Created    string
	Dead       string

	// Visual state indicators (for Containers tab title ornament)
	PlayIcon   string // Running state visual
	PauseIcon  string // Paused state visual
	StopIcon   string // Stopped state visual
	DockerIcon string // Default/generic container

	// Resource state indicators
	InUse     string
	Unused    string
	Active    string
	Empty     string
	Mounted   string
	Unmounted string

	// Selection/UI
	CheckedBox   string
	UncheckedBox string

	// Misc
	Lock    string // For system resources
	Star    string // For official/featured items
	Box     string // Generic container/package
	Port    string
	Mount   string
	CPU     string
	Memory  string
	NetIO   string
	Disk    string
	Time    string
	Tag     string
	Link    string
	Check   string
	Cross   string
	Warning string
	Error   string
	Info    string
}

var (
	// Nerd Font icon set
	nerdFontIcons = IconSet{
		// Resource types
		Container: " ",
		Image:     " ",
		Network:   " ",
		Volume:    " ",
		Service:   " ",

		// Status indicators - use play/pause/stop icons
		Running:    " ",
		Stopped:    " ",
		Paused:     " ",
		Restarting: "󰑐 ",
		Removing:   "󰩺 ",
		Created:    "󰐾 ",
		Dead:       " ",

		// Visual state icons (Containers tab - same as status)
		PlayIcon:   " ",
		PauseIcon:  " ",
		StopIcon:   " ",
		DockerIcon: " ",

		// Resource states
		InUse:     "󰌹 ",
		Unused:    "󰌺 ",
		Active:    "󰐾 ",
		Empty:     "󰐿 ",
		Mounted:   "󰉋 ",
		Unmounted: "󰋊 ",
		// Selection
		CheckedBox:   " ",
		UncheckedBox: " ",

		// Misc
		Lock:    "",
		Star:    "★",
		Box:     "□",
		Port:    "→",
		Mount:   "⟷",
		CPU:     "CPU",
		Memory:  "󰍛",
		NetIO:   "󰛳",
		Disk:    "󰋊",
		Time:    "⏱",
		Tag:     "⚐",
		Link:    "⎘",
		Check:   "✓",
		Cross:   "✗",
		Warning: "⚠",
		Error:   "✗",
		Info:    "ⓘ",
	}

	// Text-based fallback icon set
	textIcons = IconSet{
		// Resource types
		Container: "●",
		Image:     "◆",
		Network:   "⬡",
		Volume:    "◇",
		Service:   "◈",

		// Status indicators
		Running:    "●",
		Stopped:    "○",
		Paused:     "◐",
		Restarting: "↻",
		Removing:   "✗",
		Created:    "○",
		Dead:       "✗",

		// Visual state icons (same as status for text mode)
		PlayIcon:   "●",
		PauseIcon:  "◐",
		StopIcon:   "○",
		DockerIcon: "●",

		// Resource states
		InUse:     "◆",
		Unused:    "◇",
		Active:    "⬢",
		Empty:     "⬡",
		Mounted:   "▣",
		Unmounted: "▢",

		// Selection
		CheckedBox:   "[x]",
		UncheckedBox: "[ ]",

		// Misc
		Lock:    "(lock)",
		Star:    "★",
		Box:     "□",
		Port:    "→",
		Mount:   "⟷",
		CPU:     "CPU",
		Memory:  "MEM",
		NetIO:   "NET",
		Disk:    "DSK",
		Time:    "⏱",
		Tag:     "⚐",
		Link:    "⎘",
		Check:   "✓",
		Cross:   "✗",
		Warning: "⚠",
		Error:   "✗",
		Info:    "ⓘ",
	}
)

// Get returns the appropriate icon set based on configuration.
func Get() IconSet {
	cfg := state.GetConfig()
	if cfg != nil && bool(cfg.NoNerdFonts) {
		return textIcons
	}
	return nerdFontIcons
}

// Styled returns an icon with color applied via lipgloss.
func Styled(icon string, color color.Color) string {
	if icon == "" {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(icon)
}
