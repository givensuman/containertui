package services

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
)

type ServiceItem struct {
	Service client.Service
}

// getServiceIcon returns the service icon (no nerd fonts option check needed here since services always show icon)
func (i ServiceItem) getServiceIcon() string {
	switch state.GetConfig().NoNerdFonts {
	case true:
		return ""
	case false:
		return " "
	}
	return ""
}

// getStatusIcon returns status icon based on whether service has running containers
func (i ServiceItem) getStatusIcon() string {
	icons := infopanel.GetIcons()

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

	if hasRunning {
		return icons.Running
	} else if allStopped {
		return icons.Stopped
	}
	return icons.Paused // Mixed state or other states
}

func (i ServiceItem) Title() string {
	serviceIcon := i.getServiceIcon()
	statusIcon := i.getStatusIcon()
	return fmt.Sprintf("%s%s %s", serviceIcon, statusIcon, i.Service.Name)
}

func (i ServiceItem) Description() string {
	return fmt.Sprintf("Replicas: %d | Containers: %d", i.Service.Replicas, len(i.Service.Containers))
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
