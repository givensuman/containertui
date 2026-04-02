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
