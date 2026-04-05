package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/containers"
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

func TestCrossTabRefreshTargets(t *testing.T) {
	tests := []struct {
		name           string
		msg            tea.Msg
		wantContainers bool
		wantImages     bool
		wantVolumes    bool
		wantNetworks   bool
		wantBrowse     bool
	}{
		{
			name:           "container created refreshes containers",
			msg:            base.MsgContainerCreated{ContainerID: "abc"},
			wantContainers: true,
		},
		{
			name:       "image pulled refreshes images",
			msg:        base.MsgImagePulled{ImageName: "nginx:latest"},
			wantImages: true,
		},
		{
			name:           "resource changed container refreshes containers",
			msg:            base.MsgResourceChanged{Resource: base.ResourceContainer, Operation: base.OperationCreated},
			wantContainers: true,
		},
		{
			name:       "resource changed image refreshes images",
			msg:        base.MsgResourceChanged{Resource: base.ResourceImage, Operation: base.OperationUpdated},
			wantImages: true,
		},
		{
			name:        "resource changed volume refreshes volumes",
			msg:         base.MsgResourceChanged{Resource: base.ResourceVolume, Operation: base.OperationPruned},
			wantVolumes: true,
		},
		{
			name:         "resource changed network refreshes networks",
			msg:          base.MsgResourceChanged{Resource: base.ResourceNetwork, Operation: base.OperationDeleted},
			wantNetworks: true,
		},
		{
			name:           "containers refreshed message refreshes containers",
			msg:            containers.MsgContainersRefreshed{},
			wantContainers: true,
		},
		{
			name:           "containers refresh periodic message refreshes containers",
			msg:            containers.MsgRefreshContainers{},
			wantContainers: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			refreshContainers, refreshImages, refreshVolumes, refreshNetworks, refreshBrowse := crossTabRefreshTargets(tc.msg)

			if refreshContainers != tc.wantContainers ||
				refreshImages != tc.wantImages ||
				refreshVolumes != tc.wantVolumes ||
				refreshNetworks != tc.wantNetworks ||
				refreshBrowse != tc.wantBrowse {
				t.Fatalf("unexpected refresh targets: got c=%t i=%t v=%t n=%t b=%t", refreshContainers, refreshImages, refreshVolumes, refreshNetworks, refreshBrowse)
			}
		})
	}
}
