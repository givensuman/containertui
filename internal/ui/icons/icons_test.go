package icons

import (
	"testing"

	"github.com/givensuman/containertui/internal/colors"
)

func TestSelectionCheckboxUsesPrimaryOnlyForChecked(t *testing.T) {
	iconSet := Get()

	checked := SelectionCheckbox(true)
	if checked != Styled(iconSet.CheckedBox, colors.Primary()) {
		t.Fatalf("checked icon should use primary color, got %q", checked)
	}

	unchecked := SelectionCheckbox(false)
	if unchecked != iconSet.UncheckedBox {
		t.Fatalf("unchecked icon should keep default style, got %q", unchecked)
	}
}
