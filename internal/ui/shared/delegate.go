package shared

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/givensuman/containertui/internal/colors"
)

func ChangeDelegateStyles(delegate list.DefaultDelegate) list.DefaultDelegate {
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		BorderLeftForeground(colors.Primary()).
		Foreground(colors.Primary())

	delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.
		Foreground(colors.Error())

	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		BorderLeftForeground(colors.Primary()).
		Foreground(colors.Primary())

	delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.
		Foreground(colors.Muted()).
		Bold(false)

	delegate.Styles.FilterMatch = delegate.Styles.FilterMatch.
		Foreground(colors.Primary()).
		Bold(true)

	return delegate
}
