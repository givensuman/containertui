package browse

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/state"
)

// BrowseItem wraps RegistryImage for list display.
type BrowseItem struct {
	Image      registry.RegistryImage
	isSelected bool
}

var (
	_ list.Item        = (*BrowseItem)(nil)
	_ list.DefaultItem = (*BrowseItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	return delegate
}

func (item BrowseItem) getIsSelectedIcon() string {
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch item.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts.
		switch item.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (item BrowseItem) getTitleOrnament() string {
	// Use container/box icon for registry images
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		if item.Image.IsOfficial {
			return " " // Star icon for official images
		}
		return " " // Box icon for regular images
	}

	return ""
}

func (item BrowseItem) formatCount(count int64) string {
	if count >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(count)/1_000_000_000)
	} else if count >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	} else if count >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	}
	return fmt.Sprintf("%d", count)
}

func (item BrowseItem) Title() string {
	statusIcon := item.getIsSelectedIcon()
	titleOrnament := item.getTitleOrnament()

	// Format: [x]  nginx  ★ 21.2K  ↓ 12.8B
	stars := item.formatCount(int64(item.Image.StarCount))
	pulls := item.formatCount(item.Image.PullCount)

	return fmt.Sprintf("%s %s %-30s ★ %-8s ↓ %s",
		statusIcon, titleOrnament, item.Image.RepoName, stars, pulls)
}

func (item BrowseItem) Description() string {
	if item.Image.ShortDescription != "" {
		return "   " + item.Image.ShortDescription
	}
	return "   No description available"
}

func (item BrowseItem) FilterValue() string {
	// Include repo name and description for filtering
	return item.Image.RepoName + " " + item.Image.ShortDescription
}
