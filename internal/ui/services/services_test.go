package services

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/ui/components"
)

func TestNewKeybindingsIncludeServiceActions(t *testing.T) {
	b := newKeybindings()

	if b.startService.Help().Desc != "start service" {
		t.Fatalf("start service help = %q, want %q", b.startService.Help().Desc, "start service")
	}

	if b.stopService.Help().Desc != "stop service" {
		t.Fatalf("stop service help = %q, want %q", b.stopService.Help().Desc, "stop service")
	}

	if b.restartService.Help().Desc != "restart service" {
		t.Fatalf("restart service help = %q, want %q", b.restartService.Help().Desc, "restart service")
	}
}

func TestServiceContainerIDs(t *testing.T) {
	svc := client.Service{
		Name: "api",
		Containers: []client.Container{
			{ID: "abc123", State: "running"},
			{ID: "def456", State: "exited"},
		},
	}

	ids := serviceContainerIDs(svc)
	if len(ids) != 2 {
		t.Fatalf("serviceContainerIDs length = %d, want %d", len(ids), 2)
	}

	if ids[0] != "abc123" || ids[1] != "def456" {
		t.Fatalf("serviceContainerIDs = %#v, want %#v", ids, []string{"abc123", "def456"})
	}
}

func TestNewConfiguresComposeExtraPane(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	rv := components.ResourceView[string, ServiceItem]{
		SplitView: components.NewSplitView(listModel, components.NewViewportPane()),
	}

	configureServiceSplitView(&rv)

	if rv.SplitView.Extra == nil {
		t.Fatal("expected extra pane to be configured")
	}
}

func TestUpdateDetailContentClearsBothPanesWhenNoSelection(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	splitView.SetExtraPane(components.NewViewportPane(), 0.4)

	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")

	model := Model{
		ResourceView: components.ResourceView[string, ServiceItem]{
			SplitView: splitView,
		},
		detailsPanel: detailsPanel,
	}

	model.SetContent("inspect old")
	model.SetExtraContent("compose old")

	_ = model.updateDetailContent()

	vp, ok := model.SplitView.Detail.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected detail pane viewport")
	}
	vp.SetSize(80, 8)
	if !strings.Contains(vp.View(), "No service selected") {
		t.Fatalf("expected no service selected content, got %q", vp.View())
	}

	extraVp, ok := model.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane viewport")
	}
	extraVp.SetSize(80, 8)
	if !strings.Contains(extraVp.View(), "No compose file available") {
		t.Fatalf("expected no compose file content, got %q", extraVp.View())
	}
}
