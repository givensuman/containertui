package containers

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/ui/components"
)

func TestUpdateStatsPaneShowsLoadingForUnknownState(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	splitView.SetExtraPane(components.NewViewportPane(), 0.3)

	model := Model{ResourceView: components.ResourceView[string, ContainerItem]{SplitView: splitView}}

	updated := model.updateStatsPane("", nil, nil)

	vp, ok := updated.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane to be a viewport pane")
	}
	vp.SetSize(40, 5)

	if !strings.Contains(vp.View(), "Loading stats...") {
		t.Fatalf("expected loading message, got %q", vp.View())
	}
}

func TestUpdateStatsPaneShowsLoadingForRunningWithoutStatsYet(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	splitView.SetExtraPane(components.NewViewportPane(), 0.3)

	model := Model{ResourceView: components.ResourceView[string, ContainerItem]{SplitView: splitView}}

	updated := model.updateStatsPane("running", nil, nil)

	vp, ok := updated.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane to be a viewport pane")
	}
	vp.SetSize(40, 5)

	if !strings.Contains(vp.View(), "Loading stats...") {
		t.Fatalf("expected loading message, got %q", vp.View())
	}
}
