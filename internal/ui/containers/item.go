package containers

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/context"
	"github.com/moby/moby/api/types/container"
)

// ContainerItem represents a container in the ContainerList.
type ContainerItem struct {
	client.Container
	isSelected bool
}

var _ list.Item = ContainerItem{}

// NewContainerItem creates a new ContainerItem.
func NewContainerItem(container client.Container) ContainerItem {
	return ContainerItem{
		Container:  container,
		isSelected: false,
	}
}

// toggleSelected toggles the selection state of the container item.
func (c ContainerItem) toggleSelected() ContainerItem {
	c.isSelected = !c.isSelected
	return c
}

func (c ContainerItem) getIsSelectedIcon() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use Nerd Fonts
		switch c.isSelected {
		case true:
			return "[x] "
		case false:
			return "[ ] "
		}
	case false: // Use Nerd Fonts
		switch c.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ] "
}

func (c ContainerItem) getStateIcon() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use Nerd Fonts
		switch c.State {
		case string(container.StateRunning):
			return "> "
		case string(container.StatePaused):
			return "= "
		case string(container.StateExited):
			return "# "
		}
	case false: // Use Nerd Fonts
		switch c.State {
		case string(container.StateRunning):
			return "▶︎ "
		case string(container.StatePaused):
			return "⏸︎ "
		case string(container.StateExited):
			return "⏹ "
		}
	}

	return "="
}

// FilterValue implements list.Item.
func (c ContainerItem) FilterValue() string {
	return c.Name
}

// Title implements list.DefaultDelegate.
func (c ContainerItem) Title() string {
	isSelectedIcon := c.getIsSelectedIcon()
	stateIcon := c.getStateIcon()
	shortID := c.ID[len(c.ID)-12:]

	containerOrnament := " "
	if context.GetConfig().NoNerdFonts {
		containerOrnament = ""
	}

	title := fmt.Sprintf("%s %s %s %s (%s)",
		isSelectedIcon,
		containerOrnament,
		stateIcon,
		c.Name,
		shortID,
	)
	return title
}

// Description implements list.DefaultDelegate.
func (c ContainerItem) Description() string {
	return fmt.Sprintf("   %s - %s", c.Image, c.State)
}
