package components

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

type testListItem struct {
	value string
}

func (i testListItem) Title() string       { return i.value }
func (i testListItem) Description() string { return "" }
func (i testListItem) FilterValue() string { return i.value }

// deliverRefresh executes a Refresh cmd synchronously and delivers the result to Update.
func deliverRefresh[ID comparable, Item list.Item](rv *ResourceView[ID, Item]) {
	cmd := rv.Refresh()
	if cmd == nil {
		return
	}
	msg := cmd()
	updated, _ := rv.Update(msg)
	*rv = updated
}

func TestResourceViewRefreshTracksLoadErrors(t *testing.T) {
	errLoad := fmt.Errorf("docker unavailable")

	rv := NewResourceView[string, testListItem](
		"Test",
		func() ([]testListItem, error) {
			return nil, errLoad
		},
		func(item testListItem) string { return item.value },
		func(item testListItem) string { return item.value },
		nil,
	)

	// Simulate async delivery of initial load (which fails)
	deliverRefresh(rv)

	if rv.loadErr == nil {
		t.Fatal("loadErr should be set after failing refresh")
	}

	if rv.loadErr.Error() != errLoad.Error() {
		t.Fatalf("loadErr = %q, want %q", rv.loadErr.Error(), errLoad.Error())
	}

	rv.LoadItems = func() ([]testListItem, error) {
		return []testListItem{{value: "item-1"}}, nil
	}

	deliverRefresh(rv)

	if rv.loadErr != nil {
		t.Fatalf("loadErr should be cleared on successful refresh, got %v", rv.loadErr)
	}

	items := rv.SplitView.List.Items()
	if len(items) != 1 {
		t.Fatalf("item count after successful refresh = %d, want 1", len(items))
	}
}

func TestResourceViewViewShowsLoadErrorWhenEmpty(t *testing.T) {
	errLoad := fmt.Errorf("daemon not reachable")

	rv := NewResourceView[string, testListItem](
		"Services",
		func() ([]testListItem, error) {
			return nil, errLoad
		},
		func(item testListItem) string { return item.value },
		func(item testListItem) string { return item.value },
		nil,
	)

	// Simulate async delivery of initial load (which fails)
	deliverRefresh(rv)

	rv.SplitView.SetSize(80, 20)
	view := rv.View()

	if !contains(view, "Failed to load Services") {
		t.Fatalf("view does not contain load error header: %q", view)
	}

	if !contains(view, errLoad.Error()) {
		t.Fatalf("view does not contain load error details: %q", view)
	}

	if !contains(view, "Try: verify Docker daemon is running and accessible") {
		t.Fatalf("view does not contain actionable recovery guidance: %q", view)
	}
}

func contains(s, needle string) bool {
	return strings.Contains(s, needle)
}

func TestResourceViewShortHelpIncludesFocusSwitchInListView(t *testing.T) {
	rv := NewResourceView[string, testListItem](
		"Test",
		func() ([]testListItem, error) {
			return []testListItem{{value: "item-1"}}, nil
		},
		func(item testListItem) string { return item.value },
		func(item testListItem) string { return item.value },
		nil,
	)

	help := rv.ShortHelp()
	if !hasHelpDesc(help, "switch focus") {
		t.Fatal("expected short help to include switch focus in list view")
	}
}

func TestResourceViewShortHelpStaysMinimalInListView(t *testing.T) {
	rv := NewResourceView[string, testListItem](
		"Test",
		func() ([]testListItem, error) {
			return []testListItem{{value: "item-1"}}, nil
		},
		func(item testListItem) string { return item.value },
		func(item testListItem) string { return item.value },
		nil,
	)

	rv.AdditionalHelp = []key.Binding{
		key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "do thing")),
	}

	help := rv.ShortHelp()
	if hasHelpDesc(help, "do thing") {
		t.Fatal("expected short help to remain minimal and exclude additional contextual hints")
	}
}

func TestResourceViewFullHelpIncludesFocusSwitchInListView(t *testing.T) {
	rv := NewResourceView[string, testListItem](
		"Test",
		func() ([]testListItem, error) {
			return []testListItem{{value: "item-1"}}, nil
		},
		func(item testListItem) string { return item.value },
		func(item testListItem) string { return item.value },
		nil,
	)

	groups := rv.FullHelp()
	if !hasHelpDescInGroups(groups, "switch focus") {
		t.Fatal("expected full help to include switch focus in list view")
	}
}

func TestResourceViewFullHelpPutsFocusSwitchInFirstColumn(t *testing.T) {
	rv := NewResourceView[string, testListItem](
		"Test",
		func() ([]testListItem, error) {
			return []testListItem{{value: "item-1"}}, nil
		},
		func(item testListItem) string { return item.value },
		func(item testListItem) string { return item.value },
		nil,
	)

	groups := rv.FullHelp()
	if len(groups) == 0 {
		t.Fatal("expected at least one help group")
	}

	if !hasHelpDesc(groups[0], "switch focus") {
		t.Fatal("expected switch focus help in first help column")
	}

	if len(groups) > 0 && len(groups[0]) == 1 {
		t.Fatal("did not expect switch focus to occupy a dedicated single-binding first column")
	}
}

func hasHelpDesc(bindings []key.Binding, desc string) bool {
	for _, binding := range bindings {
		if binding.Help().Desc == desc {
			return true
		}
	}

	return false
}

func hasHelpDescInGroups(groups [][]key.Binding, desc string) bool {
	for _, group := range groups {
		if hasHelpDesc(group, desc) {
			return true
		}
	}

	return false
}
