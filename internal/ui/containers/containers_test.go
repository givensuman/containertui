package containers

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
)

func newContainersTestModel() Model {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())

	return Model{
		ResourceView:       components.ResourceView[string, ContainerItem]{SplitView: splitView},
		detailsKeybindings: newDetailsKeybindings(),
		detailsPanel:       components.NewDetailsPanel(),
	}
}

func TestContainersViewHasNoStatsExtraPane(t *testing.T) {
	model := newContainersTestModel()
	configureContainersSplitView(&model.ResourceView)

	if model.SplitView.Extra != nil {
		t.Fatal("expected containers view to have no extra stats pane")
	}
}

func TestUpdate_ContainerCrossTabMessagesTriggerRefresh(t *testing.T) {
	model := newContainersTestModel()

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
	model := newContainersTestModel()
	model.detailsPanel.SetCurrentID("id-1", nil)

	msg := MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: backend.Container{ID: "id-1", Name: "c1", Image: "img", State: "running"}},
			{Container: backend.Container{ID: "id-2", Name: "c2", Image: "img", State: "exited"}},
		},
	}

	updated, cmd := model.Update(msg)
	if cmd != nil {
		if followUp := cmd(); followUp != nil {
			updated, _ = updated.Update(followUp)
		}
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
	model := newContainersTestModel()
	model.detailsPanel.SetCurrentID("c1", nil)
	updated, cmd := model.Update(MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: backend.Container{ID: "c1", State: "running"}},
			{Container: backend.Container{ID: "c2", State: "exited"}},
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
	model := newContainersTestModel()
	model.detailsPanel.SetCurrentID("c1", nil)
	updated, cmd := model.Update(MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: backend.Container{ID: "c1", State: "running"}},
			{Container: backend.Container{ID: "c2", State: "paused"}},
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

func TestPruneConfirmationUsesSafetyHelper(t *testing.T) {
	model := newContainersTestModel()
	updated, cmd := model.Update(MsgContainersRefreshed{
		Items: []ContainerItem{
			{Container: backend.Container{ID: "c1", Name: "api", State: "exited"}},
			{Container: backend.Container{ID: "c2", Name: "worker", State: "created"}},
		},
	})
	_ = cmd

	updated.showPruneContainersConfirmation()
	if !updated.IsOverlayVisible() {
		t.Fatal("expected prune action to show confirmation overlay")
	}

	dialog, ok := updated.Foreground.(components.Dialog)
	if !ok {
		t.Fatalf("expected dialog overlay, got %T", updated.Foreground)
	}

	text := fmt.Sprint(dialog.View())
	if !strings.Contains(text, "Prune 2 containers") {
		t.Fatalf("expected prune count in dialog, got %q", text)
	}
	if !strings.Contains(text, "api") {
		t.Fatalf("expected sample name in dialog, got %q", text)
	}
	if !strings.Contains(strings.ToLower(text), "destructive") {
		t.Fatalf("expected destructive warning in dialog, got %q", text)
	}
}
