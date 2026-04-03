package browse

import (
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/ui/components"
)

func newTestBrowseModel(items []BrowseItem) Model {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	listModel := list.New(listItems, list.NewDefaultDelegate(), 80, 20)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())

	rv := components.ResourceView[string, BrowseItem]{
		SplitView:    splitView,
		Selections:   components.NewSelectionManager[string](),
		GetItemID:    func(item BrowseItem) string { return item.Image.RepoName },
		GetItemTitle: func(item BrowseItem) string { return item.Title() },
	}

	return Model{
		ResourceView:       rv,
		keybindings:        newKeybindings(),
		detailsKeybindings: newDetailsKeybindings(),
		scrollPositions:    make(map[string]int),
		currentRegistry:    registryDockerHub,
	}
}

func TestBrowseSpaceTogglesSelection(t *testing.T) {
	model := newTestBrowseModel([]BrowseItem{{Image: registry.RegistryImage{RepoName: "library/nginx", Registry: registryQuay}}})

	updated, _ := model.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	items := updated.GetItems()
	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	if !items[0].isSelected {
		t.Fatal("expected item to be selected after pressing space")
	}
}

func TestPullImageTargetsPrefersMultiSelection(t *testing.T) {
	model := newTestBrowseModel([]BrowseItem{
		{Image: registry.RegistryImage{RepoName: "library/nginx", Registry: registryQuay}},
		{Image: registry.RegistryImage{RepoName: "library/redis", Registry: registryQuay}},
	})

	model.Selections.Select("library/nginx", 0)
	model.Selections.Select("library/redis", 1)

	targets := model.pullImageTargets()
	if len(targets) != 2 {
		t.Fatalf("expected two pull targets, got %d", len(targets))
	}
	if targets[0].ImageName != "library/nginx" || targets[1].ImageName != "library/redis" {
		t.Fatalf("unexpected pull targets: %#v", targets)
	}
}
