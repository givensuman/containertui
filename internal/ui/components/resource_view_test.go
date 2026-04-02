package components

import (
	"fmt"
	"strings"
	"testing"
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
