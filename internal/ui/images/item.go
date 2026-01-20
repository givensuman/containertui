package images

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

type ImageItem struct {
	Image      client.Image
	isSelected bool
}

var (
	_ list.Item        = (*ImageItem)(nil)
	_ list.DefaultItem = (*ImageItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate = styles.ChangeDelegateStyles(delegate)

	return delegate
}

func (imageItem ImageItem) getIsSelectedIcon() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch imageItem.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts.
		switch imageItem.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (imageItem ImageItem) getTitleOrnament() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

func (imageItem ImageItem) FilterValue() string {
	return imageItem.Title()
}

func (imageItem ImageItem) Title() string {
	var repoTag string
	if len(imageItem.Image.RepoTags) > 0 {
		repoTag = imageItem.Image.RepoTags[0]
	} else {
		repoTag = "<none>"
	}

	titleOrnament := imageItem.getTitleOrnament()

	title := fmt.Sprintf("%s %s", titleOrnament, repoTag)
	title = lipgloss.NewStyle().
		Foreground(colors.Muted()).
		Render(title)

	statusIcon := imageItem.getIsSelectedIcon()
	var isSelectedColor color.Color
	switch imageItem.isSelected {
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

func (imageItem ImageItem) Description() string {
	shortID := imageItem.Image.ID
	if len(shortID) > 12 {
		shortID = shortID[7:19] // Remove "sha256:" prefix and take first 12 chars.
	}
	return "   " + shortID
}
