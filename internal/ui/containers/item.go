package containers

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/moby/moby/api/types/container"
)

type ContainerItem struct {
	client.Container
	isSelected bool
}

var (
	_ list.Item        = (*ContainerItem)(nil)
	_ list.DefaultItem = (*ContainerItem)(nil)
)

func (ci ContainerItem) getIsSelectedIcon() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts
		switch ci.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts
		switch ci.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (ci ContainerItem) getTitleOrnament() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts
		return "|"
	case false: // Use nerd fonts
		return " "
	}

	return "|"
}

func (ci ContainerItem) getContainerStateIcon() string {
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts
		switch ci.State {
		case container.StateRunning:
			return ">"
		case container.StatePaused:
			return "="
		case container.StateExited:
			return "#"
		}
	case false: // Use nerd fonts
		switch ci.State {
		case container.StateRunning:
			return " "
		case container.StatePaused:
			return " "
		case container.StateExited:
			return " "
		}
	}

	return ">"
}

func (ci ContainerItem) FilterValue() string {
	return ci.Name
}

func (ci ContainerItem) Title() string {
	isSelectedIcon := ci.getIsSelectedIcon()
	titleOrnament := ci.getTitleOrnament()
	containerStateIcon := ci.getContainerStateIcon()
	shortID := ci.ID[len(ci.ID)-12:]

	title := fmt.Sprintf("%s %s %s (%s)",
		titleOrnament,
		containerStateIcon,
		ci.Name,
		shortID,
	)

	var titleColor lipgloss.Color
	switch ci.State {
	case container.StateRunning:
		titleColor = colors.Green()
	case container.StatePaused:
		titleColor = colors.Yellow()
	case container.StateExited:
		titleColor = colors.Gray()
	}
	title = lipgloss.NewStyle().
		Foreground(titleColor).
		Render(title)

	var isSelectedColor lipgloss.Color
	switch ci.isSelected {
	case true:
		isSelectedColor = colors.Blue()
	case false:
		isSelectedColor = colors.White()
	}
	isSelectedIcon = lipgloss.NewStyle().
		Foreground(isSelectedColor).
		Render(isSelectedIcon)

	return fmt.Sprintf("%s %s", isSelectedIcon, title)
}

func (ci ContainerItem) Description() string {
	var color lipgloss.Color
	switch ci.State {
	case container.StateRunning:
		color = colors.Green()
	case container.StatePaused:
		color = colors.Yellow()
	case container.StateExited:
		color = colors.Gray()
	}

	description := fmt.Sprintf("   %s - %s", ci.Image, ci.State)
	return lipgloss.NewStyle().
		Foreground(color).
		Render(description)
}
