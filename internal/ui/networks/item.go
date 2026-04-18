package networks

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/state"

	"github.com/givensuman/containertui/internal/ui/icons"
)

type NetworkItem struct {
	Network    backend.Network
	isSelected bool
	IsActive   bool // Whether the network has any containers connected
}

var (
	_ list.Item        = (*NetworkItem)(nil)
	_ list.DefaultItem = (*NetworkItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	return delegate
}

func (networkItem NetworkItem) getNetworkIcon() string {
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true:
		return ""
	case false:
		return iconSet.Network
	}

	return ""
}

// getStatusIcon returns the appropriate icon for a network based on its activity status
func (networkItem NetworkItem) getStatusIcon() string {
	iconSet := icons.Get()

	if networkItem.IsActive {
		return iconSet.Active
	}
	return iconSet.Empty
}

func (networkItem NetworkItem) Title() string {
	statusIcon := networkItem.getStatusIcon() // Active/Empty (colored)

	return fmt.Sprintf("%s %s", statusIcon, networkItem.Network.Name)
}

func (networkItem NetworkItem) Description() string {
	shortID := networkItem.Network.ID
	if len(networkItem.Network.ID) > 12 {
		shortID = networkItem.Network.ID[:12]
	}
	return "   " + shortID
}

func (networkItem NetworkItem) FilterValue() string {
	return networkItem.Network.Name
}
