package components

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
)

type testListItem struct {
	value string
}

func (i testListItem) Title() string       { return i.value }
func (i testListItem) Description() string { return "" }
func (i testListItem) FilterValue() string { return i.value }

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

	if rv.loadErr == nil {
		t.Fatal("loadErr should be set after failing refresh")
	}

	if rv.loadErr.Error() != errLoad.Error() {
		t.Fatalf("loadErr = %q, want %q", rv.loadErr.Error(), errLoad.Error())
	}

	rv.LoadItems = func() ([]testListItem, error) {
		return []testListItem{{value: "item-1"}}, nil
	}

	rv.Refresh()

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

	rv.SplitView.SetSize(80, 20)
	view := rv.View()

	if !contains(view, "Failed to load Services") {
		t.Fatalf("view does not contain load error header: %q", view)
	}

	if !contains(view, errLoad.Error()) {
		t.Fatalf("view does not contain load error details: %q", view)
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
