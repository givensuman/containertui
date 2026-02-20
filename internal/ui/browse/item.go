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
	"github.com/givensuman/containertui/internal/ui/icons"
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
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true:
		if item.isSelected {
			return iconSet.CheckedBox
		}
		return iconSet.UncheckedBox
	case false:
		if item.isSelected {
			return iconSet.CheckedBox
		}
		return iconSet.UncheckedBox
	}

	return iconSet.UncheckedBox
}

func (item BrowseItem) getTitleOrnament() string {
	iconSet := icons.Get()

	switch state.GetConfig().NoNerdFonts {
	case true:
		return ""
	case false:
		if item.Image.IsOfficial {
			return icons.Styled(iconSet.Star, colors.Warning()) + " " // Star for official
		}
		return icons.Styled(iconSet.Box, colors.Text()) + " " // Box for regular
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
	var statusIcon string
	if item.isWorking {
		statusIcon = item.spinner.View()
	} else {
		statusIcon = item.getIsSelectedIcon()
	}
	titleOrnament := item.getTitleOrnament()

	// Format counts
	stars := item.formatCount(int64(item.Image.StarCount))
	pulls := item.formatCount(item.Image.PullCount)

	// Apply themed coloring - all images use default text color
	nameStyle := lipgloss.NewStyle().Foreground(colors.Text())
	styledName := nameStyle.Render(item.Image.RepoName)

	return fmt.Sprintf("%s %s %s ★ %s ↓ %s",
		statusIcon, titleOrnament, styledName, stars, pulls)
}

func (item BrowseItem) Description() string {
	if item.Image.ShortDescription != "" {
		return "   " + item.Image.ShortDescription
	}
	return "   No description available"
}

func (item BrowseItem) FilterValue() string {
	return item.Image.RepoName
}
