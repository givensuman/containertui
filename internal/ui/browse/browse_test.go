package browse

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/state"
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

func TestBrowseTitleDoesNotWrapRepoNameWithANSI(t *testing.T) {
	state.SetConfig(config.DefaultConfig())
	repo := "library/nginx"
	model := newTestBrowseModel([]BrowseItem{{Image: registry.RegistryImage{RepoName: repo}}})
	title := model.GetItems()[0].Title()

	if regexp.MustCompile("\\x1b\\[[0-9;]*m" + regexp.QuoteMeta(repo) + "\\x1b\\[[0-9;]*m").MatchString(title) {
		t.Fatalf("expected repo name to be plain text, got %q", title)
	}
	if strings.Contains(title, "\x1b[") {
		t.Fatalf("expected fully plain title without ANSI, got %q", title)
	}
}

func TestAdditionalHelpBindingsSwitchTabFirst(t *testing.T) {
	bindings := newKeybindings()
	help := additionalHelpBindings(bindings)

	if len(help) == 0 {
		t.Fatal("expected non-empty additional help bindings")
	}
	if help[0].Help().Desc != "switch tab" {
		t.Fatalf("expected first additional help to be switch tab, got %q", help[0].Help().Desc)
	}
}
