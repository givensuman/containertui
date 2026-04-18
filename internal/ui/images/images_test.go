package images

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
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

func TestPruneImagesConfirmationUsesSafetyHelper(t *testing.T) {
	model := newPruneTestModel([]ImageItem{
		{Image: client.Image{ID: "img-used", RepoTags: []string{"busybox:latest"}}, InUse: true},
		{Image: client.Image{ID: "img-unused-1", RepoTags: []string{"nginx:latest"}}, InUse: false},
		{Image: client.Image{ID: "img-unused-2", RepoTags: []string{"redis:7"}}, InUse: false},
	})

	model.showPruneImagesConfirmation()
	if !model.IsOverlayVisible() {
		t.Fatal("expected prune images confirmation overlay")
	}

	dialog, ok := model.Foreground.(components.Dialog)
	if !ok {
		t.Fatalf("expected dialog overlay, got %T", model.Foreground)
	}

	text := fmt.Sprint(dialog.View())
	if !strings.Contains(text, "Prune 2 images") {
		t.Fatalf("expected prune count in dialog, got %q", text)
	}
	if !strings.Contains(text, "nginx:latest") {
		t.Fatalf("expected sample image name in dialog, got %q", text)
	}
}

func TestPruneImageCandidatesFallsBackToImageID(t *testing.T) {
	model := newPruneTestModel([]ImageItem{
		{Image: client.Image{ID: "sha256:abcdef1234567890", RepoTags: nil}, InUse: false},
	})

	candidates := model.pruneImageCandidates()
	if len(candidates) != 1 {
		t.Fatalf("candidate count = %d, want 1", len(candidates))
	}
	if candidates[0] != "abcdef123456" {
		t.Fatalf("fallback candidate = %q, want %q", candidates[0], "abcdef123456")
	}
}

func TestEscCancelsActiveImageOperation(t *testing.T) {
	model := Model{}
	cancelled := false
	model.activeOperation = "pull"
	model.operationCancel = func() { cancelled = true }
	model.SetOverlay(components.NewProgressDialogWithBar("Pulling image"))

	updated, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if !cancelled {
		t.Fatal("expected active image operation to be cancelled on esc")
	}
	if updated.activeOperation != "" {
		t.Fatalf("expected active operation to be cleared, got %q", updated.activeOperation)
	}
	if updated.IsOverlayVisible() {
		t.Fatal("expected overlay to be closed after cancellation")
	}
	if cmd == nil {
		t.Fatal("expected cancellation to return a user feedback command")
	}
}

func TestBuildImageUsageAndHistoryContentIncludesMetadataAndCleanupHint(t *testing.T) {
	inspection := types.ImageInspect{
		ID:              "sha256:abcdef1234567890",
		Size:            1_024,
		VirtualSize:     2_048,
		Created:         "2025-01-01T00:00:00Z",
		RootFS:          types.RootFS{Layers: []string{"sha256:l1", "sha256:l2"}},
		ContainerConfig: &container.Config{},
	}
	history := []image.HistoryResponseItem{
		{CreatedBy: "RUN apk add curl"},
		{CreatedBy: "CMD [\"sh\"]"},
	}

	content := buildImageUsageAndHistoryContent(inspection, []string{"api", "worker"}, history)
	if !strings.Contains(content, "Layers: 2") {
		t.Fatalf("expected layer count in metadata content, got %q", content)
	}
	if !strings.Contains(content, "History") {
		t.Fatalf("expected history section in metadata content, got %q", content)
	}
	if !strings.Contains(strings.ToLower(content), "cleanup recommendation") {
		t.Fatalf("expected cleanup recommendation in content, got %q", content)
	}
}

func TestBuildImageUsageAndHistoryContentWithNoDependenciesMarksSafeCleanup(t *testing.T) {
	inspection := types.ImageInspect{ID: "sha256:abcdef", RootFS: types.RootFS{Layers: []string{}}, Size: 0, VirtualSize: 0}
	content := buildImageUsageAndHistoryContent(inspection, nil, nil)

	if !strings.Contains(strings.ToLower(content), "safe cleanup candidate") {
		t.Fatalf("expected safe cleanup recommendation for unused image, got %q", content)
	}
}

func TestExtractTagImageActionPayloadSupportsFlatMetadata(t *testing.T) {
	payload := map[string]any{
		"imageID": "sha256:abc123",
		"values":  map[string]string{"New Tag": "demo/app:v1"},
	}

	imageID, newTag, err := extractTagImageActionPayload(payload)
	if err != nil {
		t.Fatalf("expected payload parsing success, got error: %v", err)
	}
	if imageID != "sha256:abc123" {
		t.Fatalf("unexpected imageID %q", imageID)
	}
	if newTag != "demo/app:v1" {
		t.Fatalf("unexpected new tag %q", newTag)
	}
}

func TestExtractTagImageActionPayloadSupportsLegacyNestedMetadata(t *testing.T) {
	payload := map[string]any{
		"metadata": map[string]any{"imageID": "sha256:def456"},
		"values":   map[string]string{"New Tag": "demo/app:v2"},
	}

	imageID, newTag, err := extractTagImageActionPayload(payload)
	if err != nil {
		t.Fatalf("expected payload parsing success, got error: %v", err)
	}
	if imageID != "sha256:def456" {
		t.Fatalf("unexpected imageID %q", imageID)
	}
	if newTag != "demo/app:v2" {
		t.Fatalf("unexpected new tag %q", newTag)
	}
}

func TestExtractTagImageActionPayloadRequiresImageID(t *testing.T) {
	payload := map[string]any{
		"values": map[string]string{"New Tag": "demo/app:v1"},
	}

	_, _, err := extractTagImageActionPayload(payload)
	if err == nil {
		t.Fatal("expected error when imageID is missing")
	}
	if !strings.Contains(err.Error(), "invalid image ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImagePulledMessageRefreshesImagesList(t *testing.T) {
	refreshed := false
	model := newPruneTestModel(nil)
	model.ResourceView.LoadItems = func() ([]ImageItem, error) {
		refreshed = true
		return nil, nil
	}

	updated, cmd := model.Update(base.MsgImagePulled{ImageName: "nginx:latest"})

	// Execute the returned cmd to trigger the async load and deliver the message back
	if cmd != nil {
		msg := cmd()
		updated, _ = updated.Update(msg)
	}

	if !refreshed {
		t.Fatal("expected images list refresh when MsgImagePulled is received")
	}

	if updated.ResourceView.LoadItems == nil {
		t.Fatal("expected model to keep resource view configuration")
	}
}

func TestImageResourceChangedMessageRefreshesImagesList(t *testing.T) {
	refreshed := false
	model := newPruneTestModel(nil)
	model.ResourceView.LoadItems = func() ([]ImageItem, error) {
		refreshed = true
		return nil, nil
	}

	updated, cmd := model.Update(base.MsgResourceChanged{Resource: base.ResourceImage, Operation: base.OperationUpdated})

	// Execute the returned cmd to trigger the async load and deliver the message back
	if cmd != nil {
		msg := cmd()
		updated, _ = updated.Update(msg)
	}

	if !refreshed {
		t.Fatal("expected images list refresh when image MsgResourceChanged is received")
	}

	if updated.ResourceView.LoadItems == nil {
		t.Fatal("expected model to keep resource view configuration")
	}
}
