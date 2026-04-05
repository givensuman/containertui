package services

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

type ServiceItem struct {
	Service client.Service
}

// getServiceIcon returns the service icon (no nerd fonts option check needed here since services always show icon)
func (i ServiceItem) getServiceIcon() string {
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true:
		return ""
	case false:
		return iconSet.Service
	}

	return ""
}

func (i ServiceItem) statusColor() color.Color {
	hasRunning := false
	allStopped := true

	for _, container := range i.Service.Containers {
		if container.State == "running" {
			hasRunning = true
			allStopped = false
		} else if container.State != "exited" {
			allStopped = false
		}
	}

	if hasRunning {
		return colors.Success()
	}
	if allStopped {
		return colors.Text()
	}

	return colors.Warning()
}

// getStatusIcon returns status icon based on whether service has running containers
func (i ServiceItem) getStatusIcon() string {
	iconSet := icons.Get()

	// Check if any container in this service is running
	hasRunning := false
	allStopped := true

	for _, container := range i.Service.Containers {
		if container.State == "running" {
			hasRunning = true
			allStopped = false
		} else if container.State != "exited" {
			allStopped = false
		}
	}

	var icon string
	if hasRunning {
		icon = iconSet.Running
	} else if allStopped {
		icon = iconSet.Stopped
	} else {
		icon = iconSet.Paused // Mixed state
	}

	return icons.Styled(icon, i.statusColor())
}

func (i ServiceItem) Title() string {
	statusIcon := i.getStatusIcon()
	nameStyle := lipgloss.NewStyle().Foreground(i.statusColor())
	styledName := nameStyle.Render(i.Service.Name)
	return fmt.Sprintf("   %s %s", statusIcon, styledName)
}

func (i ServiceItem) Description() string {
	return fmt.Sprintf("   Replicas: %d | Containers: %d", i.Service.Replicas, len(i.Service.Containers))
}

func (i ServiceItem) FilterValue() string {
	return i.Service.Name
}

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	return delegate
}

var (
	_ list.Item        = (*ServiceItem)(nil)
	_ list.DefaultItem = (*ServiceItem)(nil)
)
