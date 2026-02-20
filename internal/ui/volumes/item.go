package volumes

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

type VolumeItem struct {
	Volume     client.Volume
	isSelected bool
	IsMounted  bool // Whether the volume is currently mounted by any containers
}

var (
	_ list.Item        = (*VolumeItem)(nil)
	_ list.DefaultItem = (*VolumeItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	return delegate
}

func (volumeItem VolumeItem) getIsSelectedIcon() string {
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch volumeItem.isSelected {
		case true:
			return iconSet.CheckedBox
		case false:
			return iconSet.UncheckedBox
		}
	case false: // Use nerd fonts.
		switch volumeItem.isSelected {
		case true:
			return iconSet.CheckedBox
		case false:
			return iconSet.UncheckedBox
		}
	}

	return iconSet.UncheckedBox
}

// getStatusIcon returns the appropriate icon for a volume based on its mount status
func (volumeItem VolumeItem) getStatusIcon() string {
	iconSet := icons.Get()

	var icon string
	var iconColor color.Color

	if volumeItem.IsMounted {
		icon = iconSet.Mounted
		iconColor = colors.Success()
	} else {
		icon = iconSet.Unmounted
		iconColor = colors.Text()
	}

	return icons.Styled(icon, iconColor)
}

func (volumeItem VolumeItem) Title() string {
	statusIcon := volumeItem.getIsSelectedIcon()
	statusStateIcon := volumeItem.getStatusIcon()

	// Apply themed coloring based on mount status
	var nameColor = colors.Text()
	if volumeItem.IsMounted {
		nameColor = colors.Success()
	}
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	styledName := nameStyle.Render(volumeItem.Volume.Name)

	return fmt.Sprintf("%s %s %s", statusIcon, statusStateIcon, styledName)
}

func (volumeItem VolumeItem) Description() string {
	return "   " + volumeItem.Volume.Driver
}

func (volumeItem VolumeItem) FilterValue() string {
	return volumeItem.Volume.Name
}
