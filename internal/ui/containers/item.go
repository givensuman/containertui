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
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
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
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch containerItem.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts.
		switch containerItem.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (containerItem ContainerItem) getTitleOrnament() string {
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
}

// getStatusIcon returns the appropriate icon for a container based on its state
func (containerItem ContainerItem) getStatusIcon() string {
	icons := infopanel.GetIcons()

	switch containerItem.State {
	case "running":
		return icons.Running
	case "paused":
		return icons.Paused
	case "exited":
		return icons.Stopped
	case "restarting":
		return icons.Restarting
	case "removing":
		return icons.Removing
	case "created":
		return icons.Created
	case "dead":
		return icons.Dead
	default:
		return icons.Stopped
	}
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
	titleOrnament := containerItem.getTitleOrnament()
	statusStateIcon := containerItem.getStatusIcon()

	// Apply status-based coloring
	statusColor := getStatusColor(containerItem.State)
	nameStyle := lipgloss.NewStyle().Foreground(statusColor)
	styledName := nameStyle.Render(containerItem.Name)

	return fmt.Sprintf("%s %s%s %s", statusIcon, titleOrnament, statusStateIcon, styledName)
}

func (containerItem ContainerItem) Description() string {
	shortID := containerItem.ID
	if len(containerItem.ID) > 12 {
		shortID = containerItem.ID[:12]
	}

	// Apply status-based coloring to description as well
	statusColor := getStatusColor(containerItem.State)
	descStyle := lipgloss.NewStyle().Foreground(statusColor)
	styledID := descStyle.Render(shortID)

	return "   " + styledID
}

func (containerItem ContainerItem) FilterValue() string {
	// Return unstyled text for filtering to work correctly
	return containerItem.Name
}
