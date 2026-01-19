package shared

import (
	"charm.land/bubbles/v2/list"
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

// UnfocusDelegateStyles modifies the delegate styles to look "unfocused" (no primary color highlights).
func UnfocusDelegateStyles(delegate list.DefaultDelegate) list.DefaultDelegate {
	// Revert to normal/muted colors but keep layout
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		BorderLeftForeground(colors.Muted()). // Or clear/transparent? Muted shows position but not focus.
		Foreground(colors.Text())

	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		BorderLeftForeground(colors.Muted()).
		Foreground(colors.Muted())

	return delegate
}
