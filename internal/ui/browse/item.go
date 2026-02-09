package browse

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/state"
)

// BrowseItem wraps RegistryImage for list display.
type BrowseItem struct {
	Image      registry.RegistryImage
	isSelected bool
	isWorking  bool
	spinner    spinner.Model
}

var (
	_ list.Item        = (*BrowseItem)(nil)
	_ list.DefaultItem = (*BrowseItem)(nil)
)

func newDefaultDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	delegate.UpdateFunc = func(msg tea.Msg, model *list.Model) tea.Cmd {
		if _, ok := msg.(spinner.TickMsg); ok {
			var cmds []tea.Cmd
			items := model.Items()
			for index, item := range items {
				if browseItem, ok := item.(BrowseItem); ok && browseItem.isWorking {
					var cmd tea.Cmd
					browseItem.spinner, cmd = browseItem.spinner.Update(msg)
					model.SetItem(index, browseItem)
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

func (item BrowseItem) getIsSelectedIcon() string {
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		switch item.isSelected {
		case true:
			return "[x]"
		case false:
			return "[ ]"
		}
	case false: // Use nerd fonts.
		switch item.isSelected {
		case true:
			return " "
		case false:
			return " "
		}
	}

	return "[ ]"
}

func (item BrowseItem) getTitleOrnament() string {
	// Use container/box icon for registry images
	switch state.GetConfig().NoNerdFonts {
	case true: // Don't use nerd fonts.
		return ""
	case false: // Use nerd fonts.
		if item.Image.IsOfficial {
			return " " // Star icon for official images
		}
		return " " // Box icon for regular images
	}

	return ""
}

func (item BrowseItem) formatCount(count int64) string {
	if count >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(count)/1_000_000_000)
	} else if count >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(count)/1_000_000)
	} else if count >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(count)/1_000)
	}
	return fmt.Sprintf("%d", count)
}

func (item BrowseItem) Title() string {
	// Return unstyled text to avoid ANSI escape code issues with filtering
	var statusIcon string
	if item.isWorking {
		statusIcon = item.spinner.View()
	} else {
		statusIcon = item.getIsSelectedIcon()
	}
	titleOrnament := item.getTitleOrnament()

	// Simple format without padding to avoid filtering artifacts
	stars := item.formatCount(int64(item.Image.StarCount))
	pulls := item.formatCount(item.Image.PullCount)

	return fmt.Sprintf("%s %s %s ★ %s ↓ %s",
		statusIcon, titleOrnament, item.Image.RepoName, stars, pulls)
}

func (item BrowseItem) Description() string {
	if item.Image.ShortDescription != "" {
		return "   " + item.Image.ShortDescription
	}
	return "   No description available"
}

func (item BrowseItem) FilterValue() string {
	// Return the same value as Title() since we removed styling
	return item.Title()
}
