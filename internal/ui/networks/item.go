package networks

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/icons"
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
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch networkItem.isSelected {
		case true:
			return iconSet.CheckedBox
		case false:
			return iconSet.UncheckedBox
		}
	case false: // Use nerd fonts.
		switch networkItem.isSelected {
		case true:
			return iconSet.CheckedBox
		case false:
			return iconSet.UncheckedBox
		}
	}

	return iconSet.UncheckedBox
}

func (networkItem NetworkItem) getNetworkIcon() string {
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true:
		return ""
	case false:
		// Color the network icon based on active state
		iconColor := colors.Text()
		if networkItem.IsActive {
			iconColor = colors.Success()
		}
		return icons.Styled(iconSet.Network, iconColor)
	}

	return ""
}

// getStatusIcon returns the appropriate icon for a network based on its activity status
func (networkItem NetworkItem) getStatusIcon() string {
	iconSet := icons.Get()

	var icon string
	var iconColor color.Color

	if networkItem.IsActive {
		icon = iconSet.Active
		iconColor = colors.Success()
	} else {
		icon = iconSet.Empty
		iconColor = colors.Text()
	}

	return icons.Styled(icon, iconColor)
}

func (networkItem NetworkItem) Title() string {
	selectionIcon := networkItem.getIsSelectedIcon() // Checkbox
	statusIcon := networkItem.getStatusIcon()        // Active/Empty (colored)

	// Apply themed coloring to name based on activity status
	nameColor := colors.Text()
	if networkItem.IsActive {
		nameColor = colors.Success()
	}
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	styledName := nameStyle.Render(networkItem.Network.Name)

	return fmt.Sprintf("%s %s %s", selectionIcon, statusIcon, styledName)
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
