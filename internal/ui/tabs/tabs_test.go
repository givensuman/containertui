package tabs

import (
	"strings"
	"testing"
)

func TestView_DoesNotRenderBoxBorders(t *testing.T) {
	m := New(Containers)
	m.WindowWidth = 120

	view := m.View()

	borderRunes := []string{"╭", "╮", "╰", "╯", "│", "─"}
	for _, r := range borderRunes {
		if strings.Contains(view, r) {
			t.Fatalf("expected tab view to render without box borders, found %q in %q", r, view)
		}
	}
}

func TestBrowseTabIsFifthTab(t *testing.T) {
	m := New(Containers)

	if len(m.Tabs) != 5 {
		t.Fatalf("expected 5 tabs, got %d", len(m.Tabs))
	}
	if m.Tabs[4] != Browse {
		t.Fatalf("expected Browse at index 4, got %v", m.Tabs[4])
	}
}

func TestServicesIsNotAValidTab(t *testing.T) {
	if IsValidTab("services") {
		t.Fatal("expected services to be an invalid tab name")
	}
}
