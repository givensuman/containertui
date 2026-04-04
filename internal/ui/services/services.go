package services

import (
	stdcontext "context"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

type MsgRefreshServices time.Time
type MsgServicesLoaded struct {
	Items []ServiceItem
}

type detailsKeybindings struct {
	Up         key.Binding
	Down       key.Binding
	Switch     key.Binding
	ToggleJSON key.Binding
	CopyOutput key.Binding
}

func newDetailsKeybindings() detailsKeybindings {
	return detailsKeybindings{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Switch: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		ToggleJSON: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "toggle JSON/YAML"),
		),
		CopyOutput: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy to clipboard"),
		),
	}
}

type keybindings struct {
	switchTab      key.Binding
	startService   key.Binding
	stopService    key.Binding
	restartService key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5", "6"),
			key.WithHelp("1-6", "switch tab"),
		),
		startService: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start service"),
		),
		stopService: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop service"),
		),
		restartService: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart service"),
		),
	}
}

func serviceContainerIDs(service client.Service) []string {
	ids := make([]string, 0, len(service.Containers))
	for _, container := range service.Containers {
		if strings.TrimSpace(container.ID) == "" {
			continue
		}
		ids = append(ids, container.ID)
	}
	return ids
}

func (model *Model) serviceActionCmd(actionName string, action func(stdcontext.Context, []string) error) tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	containerIDs := serviceContainerIDs(selectedItem.Service)
	if len(containerIDs) == 0 {
		return notifications.ShowError(fmt.Errorf("service %q has no containers", selectedItem.Service.Name))
	}

	serviceName := selectedItem.Service.Name

	return func() tea.Msg {
		if err := action(stdcontext.Background(), containerIDs); err != nil {
			return notifications.ShowError(fmt.Errorf("failed to %s service %q: %w", actionName, serviceName, err))
		}

		actionLabel := map[string]string{
			"start":   "Started",
			"stop":    "Stopped",
			"restart": "Restarted",
		}[actionName]
		if actionLabel == "" {
			actionLabel = "Updated"
		}

		return tea.Batch(notifications.ShowSuccess(fmt.Sprintf("%s service %q", actionLabel, serviceName)), model.refreshServicesCmd())
	}
}

type Model struct {
	components.ResourceView[string, ServiceItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	detailsPanel       components.DetailsPanel
	currentServiceName string
}

func New() Model {
	serviceKeybindings := newKeybindings()

	fetchServices := func() ([]ServiceItem, error) {
		services, err := state.GetClient().GetServices(stdcontext.Background())
		if err != nil {
			return nil, err
		}
		items := make([]ServiceItem, 0, len(services))
		for _, service := range services {
			items = append(items, ServiceItem{Service: service})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, ServiceItem](
		"Services",
		fetchServices,
		func(item ServiceItem) string { return item.Service.Name },
		func(item ServiceItem) string { return item.Title() },
		func(w, h int) {
			// Window resize handled by base component
		},
	)

	configureServiceSplitView(resourceView)

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        serviceKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		detailsPanel:       components.NewDetailsPanel(),
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		serviceKeybindings.switchTab,
		serviceKeybindings.startService,
		serviceKeybindings.stopService,
		serviceKeybindings.restartService,
	}

	return model
}

func configureServiceSplitView(resourceView *components.ResourceView[string, ServiceItem]) {
	resourceView.SplitView.SetDetailTitle("Inspect")
	extraPane := components.NewViewportPane()
	extraPane.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No compose file available"))
	resourceView.SplitView.SetExtraPane(extraPane, 0.4)
	resourceView.SplitView.SetExtraTitle("Compose File")
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return MsgRefreshServices(t)
	})
}

func (model Model) Init() tea.Cmd {
	return tea.Batch(model.ResourceView.Init(), tickCmd(), model.refreshServicesCmd())
}

func (model Model) refreshServicesCmd() tea.Cmd {
	return func() tea.Msg {
		services, err := state.GetClient().GetServices(stdcontext.Background())
		if err != nil {
			return notifications.ShowError(fmt.Errorf("failed to load services: %w", err))
		}

		items := make([]ServiceItem, 0, len(services))
		for _, service := range services {
			items = append(items, ServiceItem{Service: service})
		}

		return MsgServicesLoaded{Items: items}
	}
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Forward messages to ResourceView first
	updatedView, viewCmd := model.ResourceView.Update(msg)
	model.ResourceView = updatedView
	cmds = append(cmds, viewCmd)

	// Handle specific messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.WindowWidth = msg.Width
		model.WindowHeight = msg.Height
		model.UpdateWindowDimensions(msg)

	case MsgRefreshServices:
		cmds = append(cmds, tickCmd())
		cmds = append(cmds, model.refreshServicesCmd())

	case MsgServicesLoaded:
		listItems := make([]list.Item, len(msg.Items))
		for i, item := range msg.Items {
			listItems[i] = item
		}
		cmds = append(cmds, model.SplitView.List.SetItems(listItems))
	}

	// Main View Logic (only when no overlay)
	if !model.IsOverlayVisible() && model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.ResourceView.IsFiltering() {
				break
			}

			switch {
			case key.Matches(msg, model.keybindings.switchTab):
				return model, tea.Batch(cmds...) // Handled by parent
			case key.Matches(msg, model.keybindings.startService):
				if cmd := model.serviceActionCmd("start", state.GetClient().StartContainers); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)
			case key.Matches(msg, model.keybindings.stopService):
				if cmd := model.serviceActionCmd("stop", state.GetClient().StopContainers); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)
			case key.Matches(msg, model.keybindings.restartService):
				if cmd := model.serviceActionCmd("restart", state.GetClient().RestartContainers); cmd != nil {
					cmds = append(cmds, cmd)
				}
				return model, tea.Batch(cmds...)
			}
		}
	} else if !model.IsOverlayVisible() && !model.IsListFocused() {
		// Detail or compose pane is focused
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case model.IsDetailFocused() && key.Matches(msg, model.detailsKeybindings.ToggleJSON):
				cmd := model.handleToggleFormat()
				cmds = append(cmds, cmd)
			case (model.IsDetailFocused() || model.IsExtraFocused()) && key.Matches(msg, model.detailsKeybindings.CopyOutput):
				cmd := model.handleCopyToClipboard()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	// Update Detail Content
	detailCmd := model.updateDetailContent()
	if detailCmd != nil {
		cmds = append(cmds, detailCmd)
	}

	return model, tea.Batch(cmds...)
}

func (model *Model) updateDetailContent() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No service selected"))
		model.SetExtraContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No compose file available"))
		return nil
	}

	serviceName := selectedItem.Service.Name
	if serviceName != model.currentServiceName {
		model.currentServiceName = serviceName
		model.refreshServiceDetails(selectedItem.Service)
	}

	return nil
}

func (model *Model) refreshServiceDetails(service client.Service) {
	model.SetContent(model.buildInspectContent(service, model.GetContentWidth()))
	model.SetExtraContent(model.buildComposeContent(service, model.getComposeContentWidth()))
}

func (model *Model) getComposeContentWidth() int {
	if vp, ok := model.SplitView.Extra.(*components.ViewportPane); ok {
		return vp.Viewport.Width()
	}
	return model.GetContentWidth()
}

func (model *Model) buildInspectContent(service client.Service, width int) string {
	format := model.detailsPanel.GetFormatForDisplay()
	return builders.BuildServicePanel(service, width, false, format)
}

func (model *Model) buildComposeContent(service client.Service, width int) string {
	if strings.TrimSpace(service.ComposeFile) == "" {
		return lipgloss.NewStyle().Foreground(colors.Muted()).Render("No compose file available")
	}

	data, err := os.ReadFile(service.ComposeFile)
	if err != nil {
		return lipgloss.NewStyle().Foreground(colors.Muted()).Render(
			fmt.Sprintf("Failed to read compose file: %s", service.ComposeFile),
		)
	}

	composeMarkdown := fmt.Sprintf("```yaml\n%s\n```", string(data))
	rendered, renderErr := infopanel.RenderMarkdown(composeMarkdown, width)
	if renderErr != nil {
		return composeMarkdown
	}

	return rendered
}

// handleCopyToClipboard copies the current service details to clipboard
func (model *Model) handleCopyToClipboard() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	if model.IsExtraFocused() {
		content, err := composeClipboardContent(selectedItem.Service)
		if err != nil {
			return notifications.ShowError(err)
		}
		if err := clipboard.WriteAll(content); err != nil {
			return notifications.ShowError(err)
		}
		return notifications.ShowSuccess("Copied compose file to clipboard")
	}

	return model.detailsPanel.HandleCopyToClipboard(selectedItem.Service)
}

func composeClipboardContent(service client.Service) (string, error) {
	if strings.TrimSpace(service.ComposeFile) == "" {
		return "", fmt.Errorf("service %q has no compose file", service.Name)
	}

	data, err := os.ReadFile(service.ComposeFile)
	if err != nil {
		return "", fmt.Errorf("failed to read compose file: %w", err)
	}

	return string(data), nil
}

// handleToggleFormat toggles between JSON and YAML format
func (model *Model) handleToggleFormat() tea.Cmd {
	// Use detailsPanel to handle format toggle
	_, cmd := model.detailsPanel.HandleToggleFormat()

	// Refresh content with new format
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil {
		model.refreshServiceDetails(selectedItem.Service)
	}

	return cmd
}

func (model Model) View() string {
	return model.ResourceView.View()
}

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model Model) ShortHelp() []key.Binding {
	if !model.IsListFocused() {
		if model.IsExtraFocused() {
			return []key.Binding{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
				model.detailsKeybindings.CopyOutput,
			}
		}
		return []key.Binding{
			model.detailsKeybindings.Up,
			model.detailsKeybindings.Down,
			model.detailsKeybindings.Switch,
			model.detailsKeybindings.ToggleJSON,
			model.detailsKeybindings.CopyOutput,
		}
	}
	return model.ResourceView.ShortHelp()
}

func (model Model) FullHelp() [][]key.Binding {
	if !model.IsListFocused() {
		if model.IsExtraFocused() {
			return [][]key.Binding{
				{
					model.detailsKeybindings.Up,
					model.detailsKeybindings.Down,
					model.detailsKeybindings.Switch,
				},
				{
					model.detailsKeybindings.CopyOutput,
				},
			}
		}
		return [][]key.Binding{
			{
				model.detailsKeybindings.Up,
				model.detailsKeybindings.Down,
				model.detailsKeybindings.Switch,
			},
			{
				model.detailsKeybindings.ToggleJSON,
				model.detailsKeybindings.CopyOutput,
			},
		}
	}
	return model.ResourceView.FullHelp()
}
