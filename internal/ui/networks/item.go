package networks

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
)

type NetworkItem struct {
	Network    client.Network
	isSelected bool
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
	switch context.GetConfig().NoNerdFonts {
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
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

func (networkItem NetworkItem) Title() string {
	titleOrnament := networkItem.getTitleOrnament()

	title := fmt.Sprintf("%s %s", titleOrnament, networkItem.Network.Name)
	title = lipgloss.NewStyle().
		Foreground(colors.Muted()).
		Render(title)

	statusIcon := networkItem.getIsSelectedIcon()
	var isSelectedColor color.Color
	switch networkItem.isSelected {
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

func (networkItem NetworkItem) Description() string {
	shortID := networkItem.Network.ID
	if len(networkItem.Network.ID) > 12 {
		shortID = networkItem.Network.ID[:12]
	}
	return "   " + shortID
}

func (networkItem NetworkItem) FilterValue() string {
	return networkItem.Title()
}
