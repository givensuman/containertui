package services

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/icons"
)

type ServiceItem struct {
	Service    client.Service
	isSelected bool
	isWorking  bool
	spinner    spinner.Model
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

	leading := icons.SelectionCheckbox(i.isSelected)
	if i.isWorking {
		leading = i.spinner.View()
	}

	return fmt.Sprintf("%s %s %s", leading, statusIcon, styledName)
}

func (i ServiceItem) Description() string {
	return fmt.Sprintf("   Replicas: %d | Containers: %d", i.Service.Replicas, len(i.Service.Containers))
}

func (i ServiceItem) FilterValue() string {
	return i.Service.Name
}

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	delegate.UpdateFunc = func(msg tea.Msg, model *list.Model) tea.Cmd {
		if _, ok := msg.(spinner.TickMsg); ok {
			var cmds []tea.Cmd
			items := model.Items()
			for index, item := range items {
				serviceItem, ok := item.(ServiceItem)
				if !ok || !serviceItem.isWorking {
					continue
				}

				var cmd tea.Cmd
				serviceItem.spinner, cmd = serviceItem.spinner.Update(msg)
				model.SetItem(index, serviceItem)
				cmds = append(cmds, cmd)
			}

			return tea.Batch(cmds...)
		}

		return nil
	}

	return delegate
}

func newSpinner() spinner.Model {
	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = lipgloss.NewStyle().Foreground(colors.Primary())
	return spinnerModel
}

var (
	_ list.Item        = (*ServiceItem)(nil)
	_ list.DefaultItem = (*ServiceItem)(nil)
)
