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

func TestNewExcludesServicesTabFromVisibleTabs(t *testing.T) {
	m := New(Containers)

	for _, tab := range m.Tabs {
		if tab == Services {
			t.Fatal("expected services tab to be excluded from visible tabs")
		}
	}
}

func TestServicesTabIsNotValidStartupTab(t *testing.T) {
	if IsValidTab("services") {
		t.Fatal("expected services to be invalid while temporarily disabled")
	}
}
