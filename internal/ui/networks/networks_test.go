package networks

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/ui/components"
)

func newPruneTestModel(items []NetworkItem) Model {
	listItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, item)
	}

	listModel := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	return Model{ResourceView: components.ResourceView[string, NetworkItem]{SplitView: splitView}}
}

func TestHandleCreateNetworkCompleteSuccess(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateNetworkComplete(MsgCreateNetworkComplete{NetworkID: "123456789012345"})
	if cmd == nil {
		t.Fatal("expected command for successful create network completion")
	}
}

func TestHandleCreateNetworkCompleteError(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateNetworkComplete(MsgCreateNetworkComplete{Err: errTestCreateNetwork})
	if cmd == nil {
		t.Fatal("expected command for failed create network completion")
	}
}

func TestWithCreateNetworkDialogShowsOverlay(t *testing.T) {
	model := Model{}

	model = model.withCreateNetworkDialog()

	if !model.IsOverlayVisible() {
		t.Fatal("expected create network dialog to be visible")
	}
}

var errTestCreateNetwork = testError("create network failed")

type testError string

func (e testError) Error() string {
	return string(e)
}

func TestNetworkTitleDoesNotWrapNameWithANSI(t *testing.T) {
	name := "very-long-network-name"
	item := NetworkItem{Network: backend.Network{Name: name}, IsActive: true}
	title := item.Title()

	if regexp.MustCompile("\\x1b\\[[0-9;]*m" + regexp.QuoteMeta(name) + "\\x1b\\[[0-9;]*m").MatchString(title) {
		t.Fatalf("expected network name to be plain text, got %q", title)
	}
	if strings.Contains(title, "\x1b[") {
		t.Fatalf("expected fully plain title without ANSI, got %q", title)
	}
}

func TestHasPrunableNetworks(t *testing.T) {
	model := newPruneTestModel([]NetworkItem{
		{Network: backend.Network{Name: "host", ID: "n-host"}, IsActive: false},
		{Network: backend.Network{Name: "custom-unused", ID: "n-custom"}, IsActive: false},
	})

	if !model.hasPrunableNetworks() {
		t.Fatal("expected prunable networks when a non-system inactive network exists")
	}
}

func TestHasPrunableNetworksNone(t *testing.T) {
	model := newPruneTestModel([]NetworkItem{
		{Network: backend.Network{Name: "bridge", ID: "n-bridge"}, IsActive: false},
		{Network: backend.Network{Name: "custom-active", ID: "n-active"}, IsActive: true},
	})

	if model.hasPrunableNetworks() {
		t.Fatal("expected no prunable networks when only system/in-use networks exist")
	}
}

func TestDetailsKeybindingsSwitchHelpIncludesShiftTab(t *testing.T) {
	b := newDetailsKeybindings()
	if b.Switch.Help().Key != "tab/shift+tab" {
		t.Fatalf("switch help key = %q, want %q", b.Switch.Help().Key, "tab/shift+tab")
	}
}

func TestPruneNetworksConfirmationUsesSafetyHelper(t *testing.T) {
	model := newPruneTestModel([]NetworkItem{
		{Network: backend.Network{Name: "custom-a", ID: "n1"}, IsActive: false},
		{Network: backend.Network{Name: "custom-b", ID: "n2"}, IsActive: false},
		{Network: backend.Network{Name: "bridge", ID: "n3"}, IsActive: false},
	})

	model.showPruneNetworksConfirmation()
	if !model.IsOverlayVisible() {
		t.Fatal("expected prune networks confirmation overlay")
	}

	dialog, ok := model.Foreground.(components.Dialog)
	if !ok {
		t.Fatalf("expected dialog overlay, got %T", model.Foreground)
	}

	text := fmt.Sprint(dialog.View())
	if !strings.Contains(text, "Prune 2 networks") {
		t.Fatalf("expected prune count in dialog, got %q", text)
	}
	if !strings.Contains(text, "custom-a") {
		t.Fatalf("expected sample network name in dialog, got %q", text)
	}
}

func TestBuildNetworkConnectivityContentIncludesDependencyTrace(t *testing.T) {
	content := buildNetworkConnectivityContent(
		backend.NetworkDetail{Network: backend.Network{Name: "app-net", Driver: "bridge", Scope: "local", ID: "n123"}},
		[]string{"api", "worker"},
	)

	if !strings.Contains(content, "Driver: bridge") {
		t.Fatalf("expected driver metadata in connectivity content, got %q", content)
	}
	if !strings.Contains(strings.ToLower(content), "connected") {
		t.Fatalf("expected dependency section in connectivity content, got %q", content)
	}
}

func TestBuildNetworkConnectivityContentIncludesEndpointDetails(t *testing.T) {
	inspection := backend.NetworkDetail{
		Network: backend.Network{
			Name:   "app-net",
			Driver: "bridge",
			Scope:  "local",
			ID:     "n123",
		},
		Containers: map[string]backend.EndpointResource{
			"abc": {
				Name:        "api",
				IPv4Address: "172.20.0.2/16",
			},
		},
	}

	content := buildNetworkConnectivityContent(inspection, nil)
	if !strings.Contains(content, "api") {
		t.Fatalf("expected endpoint container name in connectivity content, got %q", content)
	}
	if !strings.Contains(content, "172.20.0.2/16") {
		t.Fatalf("expected endpoint IP in connectivity content, got %q", content)
	}
}

func TestHandleAttachContainerShowsDialogWhenNetworkSelected(t *testing.T) {
	model := newPruneTestModel([]NetworkItem{{Network: backend.Network{Name: "app-net", ID: "n1"}, IsActive: false}})

	model.handleAttachContainer()
	if !model.IsOverlayVisible() {
		t.Fatal("expected attach container dialog overlay")
	}
}
