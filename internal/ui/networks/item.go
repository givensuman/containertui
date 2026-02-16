package networks

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
)

type NetworkItem struct {
	Network    client.Network
	isSelected bool
	IsActive   bool // Whether the network has any containers connected
}

var (
	_ list.Item        = (*NetworkItem)(nil)
	_ list.DefaultItem = (*NetworkItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	return delegate
}

func (networkItem NetworkItem) getIsSelectedIcon() string {
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch networkItem.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts.
		switch networkItem.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (networkItem NetworkItem) getTitleOrnament() string {
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

// getStatusIcon returns the appropriate icon for a network based on its activity status
func (networkItem NetworkItem) getStatusIcon() string {
	icons := infopanel.GetIcons()

	if networkItem.IsActive {
		return icons.Active
	}
	return icons.Empty
}

func (networkItem NetworkItem) Title() string {
	// Return unstyled text to avoid ANSI escape code issues with filtering
	titleOrnament := networkItem.getTitleOrnament()
	statusIcon := networkItem.getIsSelectedIcon()
	name := networkItem.Network.Name

	// Add lock icon for system networks
	if isSystemNetwork(name) {
		name = "🔒 " + name
	}

	return fmt.Sprintf("%s %s %s", statusIcon, titleOrnament, name)
}

func (networkItem NetworkItem) Description() string {
	shortID := networkItem.Network.ID
	if len(networkItem.Network.ID) > 12 {
		shortID = networkItem.Network.ID[:12]
	}
	return "   " + shortID
}

func (networkItem NetworkItem) FilterValue() string {
	// Return the same value as Title() since we removed styling
	return networkItem.Title()
}
