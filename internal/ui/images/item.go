package images

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

type ImageItem struct {
	Image      client.Image
	isSelected bool
	InUse      bool // Whether the image is being used by any containers
}

var (
	_ list.Item        = (*ImageItem)(nil)
	_ list.DefaultItem = (*ImageItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	return delegate
}

func (imageItem ImageItem) getTitleOrnament() string {
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch imageItem.InUse {
		case true:
			return "(in use) "
		case false:
			return ""
		}
	case false: // Use nerd fonts.
		// Color the icon based on usage
		var icon string
		var iconColor color.Color

		switch imageItem.InUse {
		case true:
			icon = iconSet.InUse
			iconColor = colors.Success()
		case false:
			icon = iconSet.Unused
			iconColor = colors.Text()
		}

		return icons.Styled(icon, iconColor) + " "
	}

	return ""
}

func (imageItem ImageItem) Title() string {
	var repoTag string
	if len(imageItem.Image.RepoTags) > 0 {
		repoTag = imageItem.Image.RepoTags[0]
	} else {
		repoTag = "<none>"
	}

	titleOrnament := imageItem.getTitleOrnament()

	// Apply themed coloring based on usage status
	nameColor := colors.Text()
	if imageItem.InUse {
		nameColor = colors.Success()
	}
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	styledRepoTag := nameStyle.Render(repoTag)

	return fmt.Sprintf("%s%s", titleOrnament, styledRepoTag)
}

func (imageItem ImageItem) Description() string {
	shortID := imageItem.Image.ID
	if len(shortID) > 12 {
		shortID = shortID[7:19] // Remove "sha256:" prefix and take first 12 chars.
	}
	return "   " + shortID
}

func (imageItem ImageItem) FilterValue() string {
	if len(imageItem.Image.RepoTags) > 0 {
		return imageItem.Image.RepoTags[0]
	}
	return imageItem.Image.ID
}
