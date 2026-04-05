package images

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components"
)

func newPruneTestModel(items []ImageItem) Model {
	listItems := make([]list.Item, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, item)
	}

	listModel := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	return Model{ResourceView: components.ResourceView[string, ImageItem]{SplitView: splitView}}
}

func TestParsePullLayerProgressValid(t *testing.T) {
	raw := `{"id":"layer-1","progressDetail":{"current":50,"total":100}}`

	id, current, total, ok := parsePullLayerProgress(raw)
	if !ok {
		t.Fatal("expected valid progress payload")
	}
	if id != "layer-1" || current != 50 || total != 100 {
		t.Fatalf("unexpected parsed values: id=%q current=%d total=%d", id, current, total)
	}
}

func TestEstimatePullProgressAggregatesLayers(t *testing.T) {
	model := Model{pullLayers: map[string]pullLayerProgress{}}

	first, ok := model.estimatePullProgress(`{"id":"a","progressDetail":{"current":50,"total":100}}`)
	if !ok {
		t.Fatal("expected first progress estimate")
	}
	if first != 0.5 {
		t.Fatalf("expected 0.5, got %f", first)
	}

	second, ok := model.estimatePullProgress(`{"id":"b","progressDetail":{"current":50,"total":100}}`)
	if !ok {
		t.Fatal("expected second progress estimate")
	}
	if second != 0.5 {
		t.Fatalf("expected 0.5 aggregate, got %f", second)
	}
}

func TestEstimatePullProgressIsMonotonicAndCapped(t *testing.T) {
	model := Model{pullLayers: map[string]pullLayerProgress{}}

	_, ok := model.estimatePullProgress(`{"id":"a","progressDetail":{"current":99,"total":100}}`)
	if !ok {
		t.Fatal("expected progress estimate")
	}

	if model.pullPercent != 0.98 {
		t.Fatalf("expected cap at 0.98, got %f", model.pullPercent)
	}

	percent, ok := model.estimatePullProgress(`{"id":"a","progressDetail":{"current":10,"total":100}}`)
	if !ok {
		t.Fatal("expected progress estimate")
	}
	if percent != 0.98 {
		t.Fatalf("expected monotonic percent 0.98, got %f", percent)
	}
}

func TestCreateContainerStages(t *testing.T) {
	withoutStart := createContainerStages(false)
	if len(withoutStart) != 3 {
		t.Fatalf("expected 3 stages without auto-start, got %d", len(withoutStart))
	}

	withStart := createContainerStages(true)
	if len(withStart) != 4 {
		t.Fatalf("expected 4 stages with auto-start, got %d", len(withStart))
	}

	if withStart[len(withStart)-1] != "Starting container..." {
		t.Fatalf("expected last stage to be starting message, got %q", withStart[len(withStart)-1])
	}
}

func TestEstimateBuildPercentFromSteps(t *testing.T) {
	percent, ok := estimateBuildPercent(`{"stream":"Step 3/6 : RUN make build"}`, 0.0)
	if !ok {
		t.Fatal("expected to parse build step percent")
	}

	if percent <= 0 || percent >= 0.95 {
		t.Fatalf("expected step-based percent between 0 and 0.95, got %f", percent)
	}
}

func TestEstimateBuildPercentMonotonic(t *testing.T) {
	first, ok := estimateBuildPercent(`{"stream":"Step 4/4 : CMD run"}`, 0.0)
	if !ok {
		t.Fatal("expected first build percent")
	}

	second, ok := estimateBuildPercent(`{"stream":"Step 1/4 : FROM golang"}`, first)
	if !ok {
		t.Fatal("expected second build percent")
	}

	if second != first {
		t.Fatalf("expected monotonic percent, got first=%f second=%f", first, second)
	}
}

func TestBuildTempShellContainerConfigSetsLifecycleFields(t *testing.T) {
	config := buildTempShellContainerConfig("sha256:abcdef0123456789", time.Unix(1700000000, 0))

	if config.AutoStart {
		t.Fatal("expected AutoStart false so run-and-exec can explicitly start container")
	}

	if config.AutoRemove {
		t.Fatal("expected AutoRemove false so explicit cleanup can run")
	}

	if config.Network != "bridge" {
		t.Fatalf("expected bridge network, got %q", config.Network)
	}

	if config.ImageID != "sha256:abcdef0123456789" {
		t.Fatalf("unexpected image id %q", config.ImageID)
	}

	if !strings.HasPrefix(config.Name, "tmp-shell-") {
		t.Fatalf("name %q missing tmp-shell prefix", config.Name)
	}
}

func TestGenerateTempContainerNameIncludesPrefixAndImageID(t *testing.T) {
	name := generateTempContainerName("sha256:abcdef0123456789", time.Unix(1700000000, 123456000))

	if !strings.HasPrefix(name, "tmp-shell-") {
		t.Fatalf("name %q missing tmp-shell prefix", name)
	}

	if !strings.Contains(name, "abcdef012345") {
		t.Fatalf("name %q missing image ID segment", name)
	}
}

func TestGenerateTempContainerNameChangesWithTime(t *testing.T) {
	first := generateTempContainerName("sha256:abcdef0123456789", time.Unix(1700000000, 0))
	second := generateTempContainerName("sha256:abcdef0123456789", time.Unix(1700000001, 0))

	if first == second {
		t.Fatalf("expected unique names, got %q and %q", first, second)
	}
}

func TestImageTitleDoesNotWrapRepoNameWithANSI(t *testing.T) {
	state.SetConfig(config.DefaultConfig())

	repo := "library/nginx:latest"
	item := ImageItem{Image: client.Image{RepoTags: []string{repo}}, InUse: true}
	title := item.Title()

	if regexp.MustCompile("\\x1b\\[[0-9;]*m" + regexp.QuoteMeta(repo) + "\\x1b\\[[0-9;]*m").MatchString(title) {
		t.Fatalf("expected repo name to be plain text, got %q", title)
	}
	if strings.Contains(title, "\x1b[") {
		t.Fatalf("expected fully plain title without ANSI, got %q", title)
	}
}

func TestHasPrunableImages(t *testing.T) {
	model := newPruneTestModel([]ImageItem{
		{Image: client.Image{ID: "img-used"}, InUse: true},
		{Image: client.Image{ID: "img-unused"}, InUse: false},
	})

	if !model.hasPrunableImages() {
		t.Fatal("expected prunable images when at least one image is not in use")
	}
}

func TestHasPrunableImagesNone(t *testing.T) {
	model := newPruneTestModel([]ImageItem{
		{Image: client.Image{ID: "img-used-1"}, InUse: true},
		{Image: client.Image{ID: "img-used-2"}, InUse: true},
	})

	if model.hasPrunableImages() {
		t.Fatal("expected no prunable images when all images are in use")
	}
}

func TestDetailsKeybindingsSwitchHelpIncludesShiftTab(t *testing.T) {
	b := newDetailsKeybindings()
	if b.Switch.Help().Key != "tab/shift+tab" {
		t.Fatalf("switch help key = %q, want %q", b.Switch.Help().Key, "tab/shift+tab")
	}
}
