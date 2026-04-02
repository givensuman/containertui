package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/tabs"
)

func TestModelKeyHandlingIgnoresKeyReleaseForQuit(t *testing.T) {
	model := Model{
		tabsModel:          tabs.Model{ActiveTab: tabs.Tab(-1)},
		previousTab:        tabs.Tab(-1),
		notificationsModel: notifications.New(),
	}

	newModel, cmd := model.Update(tea.KeyReleaseMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd != nil {
		t.Fatal("expected nil command for ctrl+c key release")
	}

	typedModel, ok := newModel.(Model)
	if !ok {
		t.Fatalf("returned model has unexpected type %T", newModel)
	}

	if typedModel.tabsModel.ActiveTab != tabs.Tab(-1) {
		t.Fatalf("active tab after key release = %v, want %v", typedModel.tabsModel.ActiveTab, tabs.Tab(-1))
	}

	newModel, cmd = typedModel.Update(tea.KeyReleaseMsg{Code: 'q', Text: "q"})
	if cmd != nil {
		t.Fatal("expected nil command for q key release")
	}

	typedModel, ok = newModel.(Model)
	if !ok {
		t.Fatalf("returned model has unexpected type %T", newModel)
	}

	if typedModel.tabsModel.ActiveTab != tabs.Tab(-1) {
		t.Fatalf("active tab after q key release = %v, want %v", typedModel.tabsModel.ActiveTab, tabs.Tab(-1))
	}
}
