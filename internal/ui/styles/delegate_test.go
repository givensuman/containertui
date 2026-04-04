package styles

import (
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/colors"
)

func TestChangeDelegateStylesDimmedTitleUsesMutedColor(t *testing.T) {
	delegate := list.NewDefaultDelegate()
	updated := ChangeDelegateStyles(delegate)

	if updated.Styles.DimmedTitle.GetForeground() != colors.Muted() {
		t.Fatalf("expected dimmed title foreground to be muted, got %#v", updated.Styles.DimmedTitle.GetForeground())
	}
}

func TestChangeDelegateStylesFilterMatchUnderlineDisabled(t *testing.T) {
	delegate := list.NewDefaultDelegate()
	updated := ChangeDelegateStyles(delegate)

	if updated.Styles.FilterMatch.GetUnderline() {
		t.Fatalf("expected filter match underline disabled, got %v", updated.Styles.FilterMatch.GetUnderline())
	}
}
