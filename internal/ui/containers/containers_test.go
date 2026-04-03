package containers

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/ui/components"
)

func newStatsTestModel() Model {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	splitView.SetExtraPane(components.NewViewportPane(), 0.3)

	return Model{ResourceView: components.ResourceView[string, ContainerItem]{SplitView: splitView}}
}

func TestUpdateStatsPaneShowsLoadingForUnknownState(t *testing.T) {
	model := newStatsTestModel()

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

func TestUpdateStatsPaneRendersGraphForRunningContainer(t *testing.T) {
	model := newStatsTestModel()
	model.activeStatsContainerID = "abc123"
	model.statsHistoryByContainer = map[string]*statsHistory{"abc123": newStatsHistory(8)}

	history := model.statsHistoryByContainer["abc123"]
	base := time.Unix(100, 0)
	history.push(client.ContainerStats{CPUPercent: 20, MemUsage: 64, MemLimit: 128, NetRx: 100, NetTx: 200}, base)
	history.push(client.ContainerStats{CPUPercent: 40, MemUsage: 80, MemLimit: 128, NetRx: 180, NetTx: 260}, base.Add(time.Second))

	stats := &client.ContainerStats{CPUPercent: 40, MemUsage: 80, MemLimit: 128, NetRx: 180, NetTx: 260}
	updated := model.updateStatsPane("running", stats, nil)

	vp, ok := updated.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane to be a viewport pane")
	}
	vp.SetSize(80, 8)

	view := vp.View()
	if !strings.Contains(view, "CPU trend") {
		t.Fatalf("expected graph output to include CPU trend, got %q", view)
	}

	if !strings.Contains(view, "MEM trend") {
		t.Fatalf("expected graph output to include MEM trend, got %q", view)
	}

	if !strings.Contains(view, "Network RX:") {
		t.Fatalf("expected output to include network RX line, got %q", view)
	}

	if !strings.Contains(view, "/s") {
		t.Fatalf("expected output to include per-second rate units, got %q", view)
	}
}

func TestUpdateStatsPaneShowsLoadingForRunningWithoutStatsYet(t *testing.T) {
	model := newStatsTestModel()

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

func TestShouldShortCircuitStatsFetch(t *testing.T) {
	tests := []struct {
		name           string
		containerState string
		want           bool
	}{
		{name: "running container does not short-circuit", containerState: "running", want: false},
		{name: "unknown container state does not short-circuit", containerState: "", want: false},
		{name: "paused container short-circuits", containerState: "paused", want: true},
		{name: "exited container short-circuits", containerState: "exited", want: true},
		{name: "created container short-circuits", containerState: "created", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldShortCircuitStatsFetch(tt.containerState)
			if got != tt.want {
				t.Fatalf("shouldShortCircuitStatsFetch(%q) = %v, want %v", tt.containerState, got, tt.want)
			}
		})
	}
}
