package containers

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/ui/base"
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

func TestUpdate_ContainerCrossTabMessagesTriggerRefresh(t *testing.T) {
	model := newStatsTestModel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{name: "container created message", msg: base.MsgContainerCreated{ContainerID: "abc"}},
		{name: "resource changed container created", msg: base.MsgResourceChanged{Resource: base.ResourceContainer, Operation: base.OperationCreated}},
		{name: "resource changed container pruned", msg: base.MsgResourceChanged{Resource: base.ResourceContainer, Operation: base.OperationPruned}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := model.Update(tt.msg)
			if cmd == nil {
				t.Fatalf("expected non-nil refresh command for %T", tt.msg)
			}
		})
	}
}

func TestUpdate_AppliesRefreshedContainerItems(t *testing.T) {
	model := newStatsTestModel()

	msg := MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: client.Container{ID: "id-1", Name: "c1", Image: "img", State: "running"}},
			{Container: client.Container{ID: "id-2", Name: "c2", Image: "img", State: "exited"}},
		},
	}

	updated, cmd := model.Update(msg)
	if cmd == nil {
		t.Fatal("expected command to apply refreshed items")
	}

	if followUp := cmd(); followUp != nil {
		updated, _ = updated.Update(followUp)
	}

	items := updated.GetItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 items after refresh, got %d", len(items))
	}

	if items[0].Name != "c1" || items[1].Name != "c2" {
		t.Fatalf("unexpected refreshed items: %#v", items)
	}
}

func TestHasPrunableContainers(t *testing.T) {
	model := newStatsTestModel()
	updated, cmd := model.Update(MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: client.Container{ID: "c1", State: "running"}},
			{Container: client.Container{ID: "c2", State: "exited"}},
		},
	})
	if cmd != nil {
		if followUp := cmd(); followUp != nil {
			updated, _ = updated.Update(followUp)
		}
	}

	if !updated.hasPrunableContainers() {
		t.Fatal("expected prunable containers when at least one container is stopped")
	}
}

func TestHasPrunableContainersNone(t *testing.T) {
	model := newStatsTestModel()
	updated, cmd := model.Update(MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: client.Container{ID: "c1", State: "running"}},
			{Container: client.Container{ID: "c2", State: "paused"}},
		},
	})
	if cmd != nil {
		if followUp := cmd(); followUp != nil {
			updated, _ = updated.Update(followUp)
		}
	}

	if updated.hasPrunableContainers() {
		t.Fatal("expected no prunable containers when no container is in prune-eligible stopped state")
	}
}

func TestDetailsKeybindingsSwitchHelpIncludesShiftTab(t *testing.T) {
	b := newDetailsKeybindings()
	if b.Switch.Help().Key != "tab/shift+tab" {
		t.Fatalf("switch help key = %q, want %q", b.Switch.Help().Key, "tab/shift+tab")
	}
}
