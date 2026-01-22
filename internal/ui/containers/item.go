package containers

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
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
	switch context.GetConfig().NoNerdFonts {
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
	switch context.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		return " "
	}

	return ""
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
	// Return unstyled text to avoid ANSI escape code issues with filtering
	var statusIcon string
	if containerItem.isWorking {
		statusIcon = containerItem.spinner.View()
	} else {
		statusIcon = containerItem.getIsSelectedIcon()
	}
	titleOrnament := containerItem.getTitleOrnament()
	return fmt.Sprintf("%s %s %s", statusIcon, titleOrnament, containerItem.Name)
}

func (containerItem ContainerItem) Description() string {
	shortID := containerItem.ID
	if len(containerItem.ID) > 12 {
		shortID = containerItem.ID[:12]
	}
	return "   " + shortID
}

func (containerItem ContainerItem) FilterValue() string {
	// Return unstyled text for filtering to avoid ANSI code artifacts
	statusIcon := containerItem.getIsSelectedIcon()
	titleOrnament := containerItem.getTitleOrnament()
	return fmt.Sprintf("%s %s %s", statusIcon, titleOrnament, containerItem.Name)
}
