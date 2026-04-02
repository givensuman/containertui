package services

import (
	"testing"

	"github.com/givensuman/containertui/internal/client"
)

func TestNewKeybindingsIncludeServiceActions(t *testing.T) {
	b := newKeybindings()

	if b.startService.Help().Desc != "start service" {
		t.Fatalf("start service help = %q, want %q", b.startService.Help().Desc, "start service")
	}

	if b.stopService.Help().Desc != "stop service" {
		t.Fatalf("stop service help = %q, want %q", b.stopService.Help().Desc, "stop service")
	}

	if b.restartService.Help().Desc != "restart service" {
		t.Fatalf("restart service help = %q, want %q", b.restartService.Help().Desc, "restart service")
	}
}

func TestServiceContainerIDs(t *testing.T) {
	svc := client.Service{
		Name: "api",
		Containers: []client.Container{
			{ID: "abc123", State: "running"},
			{ID: "def456", State: "exited"},
		},
	}

	ids := serviceContainerIDs(svc)
	if len(ids) != 2 {
		t.Fatalf("serviceContainerIDs length = %d, want %d", len(ids), 2)
	}

	if ids[0] != "abc123" || ids[1] != "def456" {
		t.Fatalf("serviceContainerIDs = %#v, want %#v", ids, []string{"abc123", "def456"})
	}
}
