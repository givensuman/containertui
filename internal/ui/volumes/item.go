package volumes

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/context"
)

type VolumeItem struct {
	Volume     client.Volume
	isSelected bool
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
	switch context.GetConfig().NoNerdFonts {
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
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

func (volumeItem VolumeItem) Title() string {
	// Return unstyled text to avoid ANSI escape code issues with filtering
	titleOrnament := volumeItem.getTitleOrnament()
	statusIcon := volumeItem.getIsSelectedIcon()
	return fmt.Sprintf("%s %s %s", statusIcon, titleOrnament, volumeItem.Volume.Name)
}

func (volumeItem VolumeItem) Description() string {
	return "   " + volumeItem.Volume.Driver
}

func (volumeItem VolumeItem) FilterValue() string {
	// Return the same value as Title() since we removed styling
	return volumeItem.Title()
}
