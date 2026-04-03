package containers

import (
	"fmt"
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/ui/icons"
)

type ContainerItem struct {
	client.Container
	isSelected bool
	isWorking  bool
	spinner    spinner.Model
}

var (
	_ list.Item        = (*ContainerItem)(nil)
	_ list.DefaultItem = (*ContainerItem)(nil)
)

func (containerItem ContainerItem) getIsSelectedIcon() string {
	return icons.SelectionCheckbox(containerItem.isSelected)
}

// getStatusIcon returns the appropriate icon for a container based on its state
func (containerItem ContainerItem) getStatusIcon() string {
	iconSet := icons.Get()
	statusColor := getStatusColor(containerItem.State)

	var icon string
	switch containerItem.State {
	case "running":
		icon = iconSet.Running
	case "paused":
		icon = iconSet.Paused
	case "exited":
		icon = iconSet.Stopped
	case "restarting":
		icon = iconSet.Restarting
	case "removing":
		icon = iconSet.Removing
	case "created":
		icon = iconSet.Created
	case "dead":
		icon = iconSet.Dead
	default:
		icon = iconSet.Stopped
	}

	return icons.Styled(icon, statusColor)
}

// getStatusColor returns the appropriate color for a container based on its state
func getStatusColor(state string) color.Color {
	switch state {
	case "running":
		return colors.Success()
	case "paused":
		return colors.Warning()
	case "restarting":
		return colors.Success()
	case "removing":
		return colors.Warning()
	case "dead":
		return colors.Error()
	case "exited":
		return colors.Text()
	default:
		return colors.Text()
	}
}

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	delegate.UpdateFunc = func(msg tea.Msg, model *list.Model) tea.Cmd {
		if _, ok := msg.(spinner.TickMsg); ok {
			var cmds []tea.Cmd
			items := model.Items()
			for index, item := range items {
				if container, ok := item.(ContainerItem); ok && container.isWorking {
					var cmd tea.Cmd
					container.spinner, cmd = container.spinner.Update(msg)
					model.SetItem(index, container)
					cmds = append(cmds, cmd)
				}
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

func (containerItem ContainerItem) Title() string {
	var statusIcon string
	if containerItem.isWorking {
		statusIcon = containerItem.spinner.View()
	} else {
		statusIcon = containerItem.getIsSelectedIcon()
	}
	statusStateIcon := containerItem.getStatusIcon()

	// Apply status-based coloring
	statusColor := getStatusColor(containerItem.State)
	nameStyle := lipgloss.NewStyle().Foreground(statusColor)
	styledName := nameStyle.Render(containerItem.Name)

	return fmt.Sprintf("%s %s %s", statusIcon, statusStateIcon, styledName)
}

func (containerItem ContainerItem) Description() string {
	shortID := containerItem.ID
	if len(containerItem.ID) > 12 {
		shortID = containerItem.ID[:12]
	}

	return "   " + shortID
}

func (containerItem ContainerItem) FilterValue() string {
	return containerItem.ID
}
