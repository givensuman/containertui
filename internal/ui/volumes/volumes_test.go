package volumes

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/ui/components"
)

func newPruneTestModel(items []VolumeItem) Model {
	listItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, item)
	}

	listModel := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	return Model{ResourceView: components.ResourceView[string, VolumeItem]{SplitView: splitView}}
}

func TestHandleCreateVolumeCompleteSuccess(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateVolumeComplete(MsgCreateVolumeComplete{VolumeName: "my-volume"})
	if cmd == nil {
		t.Fatal("expected command for successful create volume completion")
	}
}

func TestHandleCreateVolumeCompleteError(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateVolumeComplete(MsgCreateVolumeComplete{Err: errTestCreateVolume})
	if cmd == nil {
		t.Fatal("expected command for failed create volume completion")
	}
}

func TestWithCreateVolumeDialogShowsOverlay(t *testing.T) {
	model := Model{}

	model = model.withCreateVolumeDialog()

	if !model.IsOverlayVisible() {
		t.Fatal("expected create volume dialog to be visible")
	}
}

func TestHasPrunableVolumes(t *testing.T) {
	model := newPruneTestModel([]VolumeItem{
		{Volume: backend.Volume{Name: "vol-mounted"}, IsMounted: true},
		{Volume: backend.Volume{Name: "vol-unused"}, IsMounted: false},
	})

	if !model.hasPrunableVolumes() {
		t.Fatal("expected prunable volumes when at least one volume is not mounted")
	}
}

func TestHasPrunableVolumesNone(t *testing.T) {
	model := newPruneTestModel([]VolumeItem{
		{Volume: backend.Volume{Name: "vol-mounted-1"}, IsMounted: true},
		{Volume: backend.Volume{Name: "vol-mounted-2"}, IsMounted: true},
	})

	if model.hasPrunableVolumes() {
		t.Fatal("expected no prunable volumes when all volumes are mounted")
	}
}

func TestDetailsKeybindingsSwitchHelpIncludesShiftTab(t *testing.T) {
	b := components.NewDetailsKeybindings()
	if b.Switch.Help().Key != "tab/shift+tab" {
		t.Fatalf("switch help key = %q, want %q", b.Switch.Help().Key, "tab/shift+tab")
	}
}

func TestPruneVolumesConfirmationUsesSafetyHelper(t *testing.T) {
	model := newPruneTestModel([]VolumeItem{
		{Volume: backend.Volume{Name: "data"}, IsMounted: false},
		{Volume: backend.Volume{Name: "cache"}, IsMounted: false},
		{Volume: backend.Volume{Name: "live"}, IsMounted: true},
	})

	model.showPruneVolumesConfirmation()
	if !model.IsOverlayVisible() {
		t.Fatal("expected prune volumes confirmation overlay")
	}

	dialog, ok := model.Foreground.(components.Dialog)
	if !ok {
		t.Fatalf("expected dialog overlay, got %T", model.Foreground)
	}

	text := fmt.Sprint(dialog.View())
	if !strings.Contains(text, "Prune 2 volumes") {
		t.Fatalf("expected prune count in dialog, got %q", text)
	}
	if !strings.Contains(text, "data") {
		t.Fatalf("expected sample volume name in dialog, got %q", text)
	}
}

func TestBuildVolumeUsageContentIncludesDependencyTrace(t *testing.T) {
	inspection := backend.VolumeDetail{Volume: backend.Volume{Name: "cache", Driver: "local", Mountpoint: "/var/lib/docker/volumes/cache/_data"}}
	content := buildVolumeUsageContent(inspection, []string{"api", "worker"})

	if !strings.Contains(content, "Driver: local") {
		t.Fatalf("expected driver metadata in usage content, got %q", content)
	}
	if !strings.Contains(content, "Mountpoint") {
		t.Fatalf("expected mountpoint metadata in usage content, got %q", content)
	}
	if !strings.Contains(strings.ToLower(content), "depend") {
		t.Fatalf("expected dependency section in usage content, got %q", content)
	}
}

func TestWithAttachVolumeDialogShowsOverlay(t *testing.T) {
	model := newPruneTestModel([]VolumeItem{{Volume: backend.Volume{Name: "cache"}, IsMounted: false}})

	model = model.withAttachVolumeDialog()
	if !model.IsOverlayVisible() {
		t.Fatal("expected attach volume dialog to be visible")
	}
}

func TestWithDetachVolumeDialogShowsOverlay(t *testing.T) {
	model := newPruneTestModel([]VolumeItem{{Volume: backend.Volume{Name: "cache"}, IsMounted: true}})

	model = model.withDetachVolumeDialog()
	if !model.IsOverlayVisible() {
		t.Fatal("expected detach volume dialog to be visible")
	}
}

var errTestCreateVolume = testError("create volume failed")

type testError string

func (e testError) Error() string {
	return string(e)
}
