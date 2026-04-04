package volumes

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/ui/icons"
)

const volumeTitleMaxLength = 32

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

// getStatusIcon returns the appropriate icon for a volume based on its mount status
func (volumeItem VolumeItem) getStatusIcon() string {
	iconSet := icons.Get()
	if volumeItem.IsMounted {
		return iconSet.Mounted
	}
	return iconSet.Unmounted
}

func (volumeItem VolumeItem) Title() string {
	statusStateIcon := volumeItem.getStatusIcon()
	return fmt.Sprintf("%s %s", statusStateIcon, truncateVolumeName(volumeItem.Volume.Name, volumeTitleMaxLength))
}

func truncateVolumeName(name string, maxLen int) string {
	runes := []rune(name)
	if maxLen <= 3 || len(runes) <= maxLen {
		return name
	}

	return string(runes[:maxLen-3]) + "..."
}

func (volumeItem VolumeItem) Description() string {
	return "   " + volumeItem.Volume.Driver
}

func (volumeItem VolumeItem) FilterValue() string {
	return volumeItem.Volume.Name
}
