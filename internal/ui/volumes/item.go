package volumes

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/styles"
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
	delegate = styles.ChangeDelegateStyles(delegate)

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
	titleOrnament := volumeItem.getTitleOrnament()

	title := fmt.Sprintf("%s %s", titleOrnament, volumeItem.Volume.Name)
	title = lipgloss.NewStyle().
		Foreground(colors.Muted()).
		Render(title)

	statusIcon := volumeItem.getIsSelectedIcon()
	var isSelectedColor color.Color
	switch volumeItem.isSelected {
	case true:
		isSelectedColor = colors.Selected()
	case false:
		isSelectedColor = colors.Text()
	}
	statusIcon = lipgloss.NewStyle().
		Foreground(isSelectedColor).
		Render(statusIcon)

	return fmt.Sprintf("%s %s", statusIcon, title)
}

func (volumeItem VolumeItem) Description() string {
	return "   " + volumeItem.Volume.Driver
}

func (volumeItem VolumeItem) FilterValue() string {
	return volumeItem.Title()
}
