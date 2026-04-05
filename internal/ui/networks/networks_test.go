package networks

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
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
	item := NetworkItem{Network: client.Network{Name: name}, IsActive: true}
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
		{Network: client.Network{Name: "host", ID: "n-host"}, IsActive: false},
		{Network: client.Network{Name: "custom-unused", ID: "n-custom"}, IsActive: false},
	})

	if !model.hasPrunableNetworks() {
		t.Fatal("expected prunable networks when a non-system inactive network exists")
	}
}

func TestHasPrunableNetworksNone(t *testing.T) {
	model := newPruneTestModel([]NetworkItem{
		{Network: client.Network{Name: "bridge", ID: "n-bridge"}, IsActive: false},
		{Network: client.Network{Name: "custom-active", ID: "n-active"}, IsActive: true},
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
