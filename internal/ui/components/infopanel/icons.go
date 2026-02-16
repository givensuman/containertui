// Package infopanel provides components for building informational panels.
package infopanel

import (
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

	// Container status indicators
	Running    string
	Stopped    string
	Paused     string
	Restarting string
	Removing   string
	Created    string
	Dead       string
	Warning    string
	Error      string
	Info       string

	// Resource state indicators
	InUse     string
	Unused    string
	Active    string
	Empty     string
	Mounted   string
	Unmounted string

	// Misc
	Port   string
	Mount  string
	CPU    string
	Memory string
	NetIO  string
	Disk   string
	Time   string
	Tag    string
	Link   string
	Check  string
	Cross  string
}

var (
	// Nerd Font icon set
	nerdFontIcons = IconSet{
		// Resource types (nerd fonts)
		Container: "󰡨",
		Image:     "",
		Network:   "",
		Volume:    "󰋊",
		Service:   "",

		// Container status indicators
		Running:    "",
		Stopped:    "",
		Paused:     "",
		Restarting: "󰑐",
		Removing:   "󰩺",
		Created:    "",
		Dead:       "",
		Warning:    "",
		Error:      "",
		Info:       "",

		// Resource state indicators
		InUse:     "󰌹",
		Unused:    "󰌺",
		Active:    "󰐾",
		Empty:     "󰐿",
		Mounted:   "󰉋",
		Unmounted: "󰋊",

		// Misc
		Port:   "",
		Mount:  "",
		CPU:    "",
		Memory: "󰍛",
		NetIO:  "󰛳",
		Disk:   "󰋊",
		Time:   "",
		Tag:    "",
		Link:   "",
		Check:  "",
		Cross:  "",
	}

	// Fallback text-based icon set
	textIcons = IconSet{
		// Resource types (ASCII/Unicode symbols)
		Container: "●",
		Image:     "◆",
		Network:   "⬡",
		Volume:    "◇",
		Service:   "◈",

		// Container status indicators
		Running:    "●",
		Stopped:    "○",
		Paused:     "◐",
		Restarting: "↻",
		Removing:   "✗",
		Created:    "○",
		Dead:       "✗",
		Warning:    "⚠",
		Error:      "✗",
		Info:       "ⓘ",

		// Resource state indicators
		InUse:     "◆",
		Unused:    "◇",
		Active:    "⬢",
		Empty:     "⬡",
		Mounted:   "▣",
		Unmounted: "▢",

		// Misc
		Port:   "→",
		Mount:  "⟷",
		CPU:    "CPU",
		Memory: "MEM",
		NetIO:  "NET",
		Disk:   "DSK",
		Time:   "⏱",
		Tag:    "⚐",
		Link:   "⎘",
		Check:  "✓",
		Cross:  "✗",
	}
)

// GetIcons returns the appropriate icon set based on the configuration.
// If --no-nerd-fonts is set, returns text-based icons, otherwise returns nerd font icons.
func GetIcons() IconSet {
	cfg := state.GetConfig()
	if cfg != nil && bool(cfg.NoNerdFonts) {
		return textIcons
	}
	return nerdFontIcons
}
