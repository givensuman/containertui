package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/ui/components"
)

func TestBuildComposeContentReturnsFallbackWhenNoComposeFile(t *testing.T) {
	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")
	model := Model{detailsPanel: detailsPanel}
	service := client.Service{Name: "api", ComposeFile: ""}

	content := model.buildComposeContent(service, 80)
	if !strings.Contains(content, "No compose file available") {
		t.Fatalf("expected no compose fallback, got %q", content)
	}
}

func TestBuildComposeContentRendersComposeYAML(t *testing.T) {
	dir := t.TempDir()
	composePath := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(composePath, []byte("services:\n  api:\n    image: nginx:latest\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")
	model := Model{detailsPanel: detailsPanel}
	service := client.Service{Name: "api", ComposeFile: composePath}

	content := model.buildComposeContent(service, 80)
	if !strings.Contains(content, "services:") {
		t.Fatalf("expected compose yaml in output, got %q", content)
	}
	if !strings.Contains(content, "api:") {
		t.Fatalf("expected service name in output, got %q", content)
	}
}

func TestBuildInspectContentReturnsPanelData(t *testing.T) {
	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")
	model := Model{detailsPanel: detailsPanel}
	service := client.Service{Name: "api"}

	content := model.buildInspectContent(service, 80)
	if content == "" {
		t.Fatal("expected inspect content to be non-empty")
	}
}

func TestBuildInspectContentIncludesComposeSummary(t *testing.T) {
	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")
	model := Model{detailsPanel: detailsPanel}
	service := client.Service{
		Project:  "demo",
		Name:     "api",
		Replicas: 2,
		Containers: []client.Container{
			{State: "running"},
			{State: "exited"},
		},
	}

	content := model.buildInspectContent(service, 80)
	if !strings.Contains(content, "Project: demo") {
		t.Fatalf("expected compose summary in inspect content, got %q", content)
	}
	if !strings.Contains(content, "Health: 1 running / 1 stopped") {
		t.Fatalf("expected health summary in inspect content, got %q", content)
	}
}
