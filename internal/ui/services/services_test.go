package services

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components"
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

func TestNewConfiguresComposeExtraPane(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	rv := components.ResourceView[string, ServiceItem]{
		SplitView: components.NewSplitView(listModel, components.NewViewportPane()),
	}

	configureServiceSplitView(&rv)

	if rv.SplitView.Extra == nil {
		t.Fatal("expected extra pane to be configured")
	}
}

func TestUpdateDetailContentClearsBothPanesWhenNoSelection(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	splitView := components.NewSplitView(listModel, components.NewViewportPane())
	splitView.SetExtraPane(components.NewViewportPane(), 0.4)

	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")

	model := Model{
		ResourceView: components.ResourceView[string, ServiceItem]{
			SplitView: splitView,
		},
		detailsPanel: detailsPanel,
	}

	model.SetContent("inspect old")
	model.SetExtraContent("compose old")

	_ = model.updateDetailContent()

	vp, ok := model.SplitView.Detail.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected detail pane viewport")
	}
	vp.SetSize(80, 8)
	if !strings.Contains(vp.View(), "No service selected") {
		t.Fatalf("expected no service selected content, got %q", vp.View())
	}

	extraVp, ok := model.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane viewport")
	}
	extraVp.SetSize(80, 8)
	if !strings.Contains(extraVp.View(), "No compose file available") {
		t.Fatalf("expected no compose file content, got %q", extraVp.View())
	}
}

func TestHandleToggleFormatRefreshesInspectAndKeepsComposePane(t *testing.T) {
	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")

	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	rv := components.ResourceView[string, ServiceItem]{
		SplitView: components.NewSplitView(listModel, components.NewViewportPane()),
	}
	configureServiceSplitView(&rv)

	model := Model{ResourceView: rv, detailsPanel: detailsPanel}
	svc := client.Service{Name: "api", ComposeFile: ""}
	model.refreshServiceDetails(svc)

	beforeExtra, ok := model.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane viewport")
	}
	beforeExtra.SetSize(80, 8)
	before := beforeExtra.View()

	_ = model.handleToggleFormat()
	model.refreshServiceDetails(svc)

	afterExtra, ok := model.SplitView.Extra.(*components.ViewportPane)
	if !ok {
		t.Fatal("expected extra pane viewport")
	}
	afterExtra.SetSize(80, 8)
	after := afterExtra.View()

	if before == "" || after == "" {
		t.Fatal("expected non-empty compose view before/after format toggle")
	}
	if !strings.Contains(after, "No compose file available") {
		t.Fatalf("expected compose fallback after toggle, got %q", after)
	}
}

func TestShortHelpHidesToggleJSONWhenComposePaneFocused(t *testing.T) {
	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	rv := components.ResourceView[string, ServiceItem]{
		SplitView: components.NewSplitView(listModel, components.NewViewportPane()),
	}
	configureServiceSplitView(&rv)

	model := Model{
		ResourceView:       rv,
		detailsKeybindings: newDetailsKeybindings(),
	}
	model.SplitView.Focus = components.FocusExtra

	help := model.ShortHelp()
	for _, binding := range help {
		if binding.Help().Desc == "toggle JSON/YAML" {
			t.Fatal("did not expect toggle JSON/YAML help in compose pane")
		}
	}
}

func TestUpdateIgnoresToggleJSONWhenComposePaneFocused(t *testing.T) {
	detailsPanel := components.NewDetailsPanel()
	detailsPanel.SetCurrentFormat("yaml")

	listModel := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 20)
	rv := components.ResourceView[string, ServiceItem]{
		SplitView: components.NewSplitView(listModel, components.NewViewportPane()),
	}
	configureServiceSplitView(&rv)

	model := Model{
		ResourceView:       rv,
		detailsPanel:       detailsPanel,
		detailsKeybindings: newDetailsKeybindings(),
	}
	model.SplitView.Focus = components.FocusExtra

	updated, _ := model.Update(tea.KeyPressMsg{Code: 'J', Text: "J"})
	if updated.detailsPanel.GetCurrentFormat() != "yaml" {
		t.Fatalf("expected format to remain yaml, got %q", updated.detailsPanel.GetCurrentFormat())
	}
}

func TestComposeClipboardContentReturnsFileContent(t *testing.T) {
	dir := t.TempDir()
	composePath := dir + "/compose.yaml"
	want := "services:\n  api:\n    image: nginx:latest\n"

	if err := os.WriteFile(composePath, []byte(want), 0o644); err != nil {
		t.Fatal(err)
	}

	service := client.Service{Name: "api", ComposeFile: composePath}
	got, err := composeClipboardContent(service)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != want {
		t.Fatalf("unexpected compose clipboard content: got %q want %q", got, want)
	}
}

func TestServiceTitleDoesNotWrapNameWithANSI(t *testing.T) {
	state.SetConfig(config.DefaultConfig())
	name := "api-service"
	item := ServiceItem{Service: client.Service{Name: name, Containers: []client.Container{{State: "running"}}}}
	title := item.Title()

	if regexp.MustCompile("\\x1b\\[[0-9;]*m" + regexp.QuoteMeta(name) + "\\x1b\\[[0-9;]*m").MatchString(title) {
		t.Fatalf("expected service name to be plain text, got %q", title)
	}
	if strings.Contains(title, "\x1b[") {
		t.Fatalf("expected fully plain title without ANSI, got %q", title)
	}
}
