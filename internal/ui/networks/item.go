package networks

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
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
	// Return unstyled text to avoid ANSI escape code issues with filtering
	titleOrnament := networkItem.getTitleOrnament()
	statusIcon := networkItem.getIsSelectedIcon()
	return fmt.Sprintf("%s %s %s", statusIcon, titleOrnament, networkItem.Network.Name)
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
