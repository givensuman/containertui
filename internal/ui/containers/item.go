package containers

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/moby/moby/api/types/container"
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
		return ""
	case false: // Use nerd fonts
		return " "
	}

	return ""
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

func newDefaultDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		BorderLeftForeground(colors.Primary())
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(colors.White()).
		BorderLeftForeground(colors.Primary())

	d.Styles.DimmedDesc = d.Styles.DimmedDesc.
		Foreground(colors.Gray()).
		Bold(false)

	d.Styles.FilterMatch = d.Styles.FilterMatch.
		Foreground(colors.Primary()).
		Bold(true)

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		if _, ok := msg.(spinner.TickMsg); ok {
			items := m.Items()
			for i, item := range items {
				if c, ok := item.(ContainerItem); ok && c.isWorking {
					var cmd tea.Cmd
					c.spinner, cmd = c.spinner.Update(msg)
					m.SetItem(i, c)
					return cmd
				}
			}
		}
		return nil
	}

	return d
}

func newSpinner() spinner.Model {
	spinnerModel := spinner.New()
	spinnerModel.Spinner = spinner.Dot
	spinnerModel.Style = lipgloss.NewStyle().Foreground(colors.Primary())

	return spinnerModel
}

func (ci ContainerItem) FilterValue() string {
	return ci.Name
}

func (ci ContainerItem) Title() string {
	var statusIcon string
	if ci.isWorking {
		statusIcon = ci.spinner.View()
	} else {
		statusIcon = ci.getIsSelectedIcon()
	}
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

	if !ci.isWorking {
		var isSelectedColor lipgloss.Color
		switch ci.isSelected {
		case true:
			isSelectedColor = colors.Blue()
		case false:
			isSelectedColor = colors.White()
		}
		statusIcon = lipgloss.NewStyle().
			Foreground(isSelectedColor).
			Render(statusIcon)
	}

	return fmt.Sprintf("%s %s", statusIcon, title)
}

func (ci ContainerItem) Description() string {
	return fmt.Sprintf("   %s", ansi.Truncate(ci.Image, 24, "..."))
}
