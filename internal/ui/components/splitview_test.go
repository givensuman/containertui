package components

import (
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func TestSplitViewTabCyclesFocusOnlyOnKeyPress(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := NewSplitView(listModel, NewViewportPane())

	if splitView.Focus != FocusList {
		t.Fatalf("initial focus = %v, want %v", splitView.Focus, FocusList)
	}

	updated, _ := splitView.Update(tea.KeyReleaseMsg{Code: tea.KeyTab})
	if updated.Focus != FocusList {
		t.Fatalf("focus after key release = %v, want %v", updated.Focus, FocusList)
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if updated.Focus != FocusDetail {
		t.Fatalf("focus after key press = %v, want %v", updated.Focus, FocusDetail)
	}
}

func TestSplitViewShiftTabCyclesFocusBackwardsTwoPane(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := NewSplitView(listModel, NewViewportPane())

	if splitView.Focus != FocusList {
		t.Fatalf("initial focus = %v, want %v", splitView.Focus, FocusList)
	}

	updated, _ := splitView.Update(tea.KeyPressMsg{Text: "shift+tab"})
	if updated.Focus != FocusDetail {
		t.Fatalf("focus after shift+tab = %v, want %v", updated.Focus, FocusDetail)
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Text: "shift+tab"})
	if updated.Focus != FocusList {
		t.Fatalf("focus after second shift+tab = %v, want %v", updated.Focus, FocusList)
	}
}

func TestSplitViewShiftTabCyclesFocusBackwardsThreePane(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	splitView := NewSplitView(listModel, NewViewportPane())
	splitView.SetExtraPane(NewViewportPane(), 0.3)

	updated, _ := splitView.Update(tea.KeyPressMsg{Text: "shift+tab"})
	if updated.Focus != FocusExtra {
		t.Fatalf("focus after shift+tab from list = %v, want %v", updated.Focus, FocusExtra)
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Text: "shift+tab"})
	if updated.Focus != FocusDetail {
		t.Fatalf("focus after shift+tab from extra = %v, want %v", updated.Focus, FocusDetail)
	}

	updated, _ = updated.Update(tea.KeyPressMsg{Text: "shift+tab"})
	if updated.Focus != FocusList {
		t.Fatalf("focus after shift+tab from detail = %v, want %v", updated.Focus, FocusList)
	}
}
