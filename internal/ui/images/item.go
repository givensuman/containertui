package images

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
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

func (imageItem ImageItem) getIsSelectedIcon() string {
	switch state.GetConfig().NoNerdFonts {
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
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

// getStatusIcon returns the appropriate icon for an image based on its usage status
func (imageItem ImageItem) getStatusIcon() string {
	icons := infopanel.GetIcons()

	if imageItem.InUse {
		return icons.InUse
	}
	return icons.Unused
}

func (imageItem ImageItem) Title() string {
	var repoTag string
	if len(imageItem.Image.RepoTags) > 0 {
		repoTag = imageItem.Image.RepoTags[0]
	} else {
		repoTag = "<none>"
	}

	titleOrnament := imageItem.getTitleOrnament()
	statusIcon := imageItem.getIsSelectedIcon()
	statusStateIcon := imageItem.getStatusIcon()

	// Apply themed coloring based on usage status
	var nameColor = colors.Text()
	if imageItem.InUse {
		nameColor = colors.Success()
	}
	nameStyle := lipgloss.NewStyle().Foreground(nameColor)
	styledRepoTag := nameStyle.Render(repoTag)

	return fmt.Sprintf("%s %s%s %s", statusIcon, titleOrnament, statusStateIcon, styledRepoTag)
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
