package images

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/icons"
)

type ImageItem struct {
	Image      backend.Image
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
		switch imageItem.InUse {
		case true:
			return iconSet.InUse + " "
		case false:
			return iconSet.Unused + " "
		}
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

	return fmt.Sprintf("%s%s", titleOrnament, repoTag)
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
