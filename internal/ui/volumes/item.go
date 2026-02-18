package volumes

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
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
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch volumeItem.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts.
		switch volumeItem.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (volumeItem VolumeItem) getTitleOrnament() string {
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

// getStatusIcon returns the appropriate icon for a volume based on its mount status
func (volumeItem VolumeItem) getStatusIcon() string {
	icons := infopanel.GetIcons()

	if volumeItem.IsMounted {
		return icons.Mounted
	}
	return icons.Unmounted
}

func (volumeItem VolumeItem) Title() string {
	titleOrnament := volumeItem.getTitleOrnament()
	statusIcon := volumeItem.getIsSelectedIcon()
	statusStateIcon := volumeItem.getStatusIcon()

	// Apply themed coloring based on mount status
	var nameColor = colors.Text()
	if volumeItem.IsMounted {
		nameColor = colors.Success()
	}
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	styledName := nameStyle.Render(volumeItem.Volume.Name)

	return fmt.Sprintf("%s %s%s %s", statusIcon, titleOrnament, statusStateIcon, styledName)
}

func (volumeItem VolumeItem) Description() string {
	return "   " + volumeItem.Volume.Driver
}

func (volumeItem VolumeItem) FilterValue() string {
	return volumeItem.Volume.Name
}
