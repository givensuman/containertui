package networks

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
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
	titleOrnament := networkItem.getTitleOrnament()
	statusIcon := networkItem.getIsSelectedIcon()
	name := networkItem.Network.Name

	// Add lock icon for system networks
	if isSystemNetwork(name) {
		name = "🔒 " + name
	}

	// Apply themed coloring based on activity status
	var nameColor = colors.Text()
	if networkItem.IsActive {
		nameColor = colors.Success()
	}
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	styledName := nameStyle.Render(name)

	return fmt.Sprintf("%s %s %s", statusIcon, titleOrnament, styledName)
}

func (networkItem NetworkItem) Description() string {
	shortID := networkItem.Network.ID
	if len(networkItem.Network.ID) > 12 {
		shortID = networkItem.Network.ID[:12]
	}
	return "   " + shortID
}

func (networkItem NetworkItem) FilterValue() string {
	return networkItem.Network.Name
}
