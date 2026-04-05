package services

import (
	stdcontext "context"
	"fmt"
	"os"
	"slices"
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
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
	"github.com/givensuman/containertui/internal/ui/notifications"
)

type MsgRefreshServices time.Time
type MsgServicesLoaded struct {
	Items []ServiceItem
}

type MsgServiceActionResult struct {
	Action       string
	ServiceNames []string
	ServiceIDs   []string
	Err          error
}

type MsgServiceActionStart struct {
	Action       string
	ServiceNames []string
	ServiceIDs   []string
	ContainerIDs []string
}

type MsgComposeActionStart struct {
	Action       string
	Project      string
	ServiceNames []string
	ServiceIDs   []string
	Args         []string
	WorkingDir   string
	ComposeFile  string
}

type MsgComposeActionResult struct {
	Action       string
	Project      string
	ServiceNames []string
	ServiceIDs   []string
	Err          error
}

func serviceSelectionID(service client.Service) string {
	project := strings.TrimSpace(service.Project)
	name := strings.TrimSpace(service.Name)
	if project == "" {
		return name
	}

	return project + "/" + name
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
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab/shift+tab", "switch focus"),
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
	switchTab       key.Binding
	startService    key.Binding
	stopService     key.Binding
	restartService  key.Binding
	composeUp       key.Binding
	composeDown     key.Binding
	composePull     key.Binding
	composeRebuild  key.Binding
	scaleService    key.Binding
	toggleSelection key.Binding
	toggleAll       key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "switch tab"),
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
		composeUp: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "compose up"),
		),
		composeDown: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "compose down"),
		),
		composePull: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "compose pull"),
		),
		composeRebuild: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "compose build"),
		),
		scaleService: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "scale service"),
		),
		toggleSelection: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle selection"),
		),
		toggleAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle selection of all"),
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

	return func() tea.Msg {
		return MsgServiceActionStart{
			Action:       actionName,
			ServiceNames: []string{selectedItem.Service.Name},
			ServiceIDs:   []string{serviceSelectionID(selectedItem.Service)},
			ContainerIDs: containerIDs,
		}
	}
}

type Model struct {
	components.ResourceView[string, ServiceItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	detailsPanel       components.DetailsPanel
	currentServiceID   string
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
			items = append(items, ServiceItem{Service: service, spinner: newSpinner()})
		}
		return items, nil
	}

	resourceView := components.NewResourceView[string, ServiceItem](
		"Services",
		fetchServices,
		func(item ServiceItem) string { return serviceSelectionID(item.Service) },
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
	model.IsItemWorking = func(item ServiceItem) bool { return item.isWorking }

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		serviceKeybindings.switchTab,
		serviceKeybindings.startService,
		serviceKeybindings.stopService,
		serviceKeybindings.restartService,
		serviceKeybindings.composeUp,
		serviceKeybindings.composeDown,
		serviceKeybindings.composePull,
		serviceKeybindings.composeRebuild,
		serviceKeybindings.scaleService,
		serviceKeybindings.toggleSelection,
		serviceKeybindings.toggleAll,
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
			items = append(items, ServiceItem{Service: service, spinner: newSpinner()})
		}

		return MsgServicesLoaded{Items: items}
	}
}

func (model Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && !model.IsOverlayVisible() && model.IsListFocused() && !model.IsFiltering() {
		switch {
		case key.Matches(keyMsg, model.keybindings.startService):
			if cmd := model.serviceActionCmd("start", state.GetClient().StartContainers); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.stopService):
			if cmd := model.serviceActionCmd("stop", state.GetClient().StopContainers); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.restartService):
			if cmd := model.serviceActionCmd("restart", state.GetClient().RestartContainers); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.composeUp):
			if cmd := model.composeProjectCommand("up", "up", "-d"); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.composeDown):
			if cmd := model.composeProjectCommand("down", "down"); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.composePull):
			if cmd := model.composeProjectCommand("pull", "pull"); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.composeRebuild):
			if cmd := model.composeProjectCommand("build", "build"); cmd != nil {
				return model, cmd
			}
		case key.Matches(keyMsg, model.keybindings.scaleService):
			model.showScaleDialog()
			return model, nil
		}
	}

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
		model.applyServicesLoaded(msg.Items)
		listItems := make([]list.Item, len(model.GetItems()))
		for i, item := range model.GetItems() {
			listItems[i] = item
		}
		cmds = append(cmds, model.SplitView.List.SetItems(listItems))

	case MsgServiceActionResult:
		if cmd := model.setWorkingState(msg.ServiceIDs, false); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Err != nil {
			cmds = append(cmds, notifications.ShowError(msg.Err))
			break
		}

		actionLabel := map[string]string{
			"start":   "Started",
			"stop":    "Stopped",
			"restart": "Restarted",
		}[msg.Action]
		if actionLabel == "" {
			actionLabel = "Updated"
		}

		display := strings.Join(msg.ServiceNames, ", ")
		if len(msg.ServiceNames) > 1 {
			display = fmt.Sprintf("%d services", len(msg.ServiceNames))
		}
		cmds = append(cmds, notifications.ShowSuccess(fmt.Sprintf("%s %s", actionLabel, display)), model.refreshServicesCmd())

	case MsgServiceActionStart:
		if cmd := model.setWorkingState(msg.ServiceIDs, true); cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, model.performServiceAction(msg))

	case MsgComposeActionStart:
		if cmd := model.setWorkingState(msg.ServiceIDs, true); cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, model.performComposeAction(msg))

	case MsgComposeActionResult:
		if cmd := model.setWorkingState(msg.ServiceIDs, false); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Err != nil {
			cmds = append(cmds, notifications.ShowError(msg.Err))
			break
		}
		cmds = append(cmds,
			notifications.ShowSuccess(fmt.Sprintf("Compose %s completed for project %q", msg.Action, msg.Project)),
			model.refreshServicesCmd(),
		)

	case base.SmartConfirmationMessage:
		if msg.Action.Type == "ScaleServiceAction" {
			payload, ok := msg.Action.Payload.(map[string]any)
			if !ok {
				cmds = append(cmds, notifications.ShowError(fmt.Errorf("invalid scale payload")))
				break
			}
			values, ok := payload["values"].(map[string]string)
			if !ok {
				cmds = append(cmds, notifications.ShowError(fmt.Errorf("invalid scale form values")))
				break
			}
			selected := model.GetSelectedItem()
			if selected == nil {
				cmds = append(cmds, notifications.ShowError(fmt.Errorf("no service selected")))
				break
			}
			replicas := strings.TrimSpace(values["Replicas"])
			if replicas == "" {
				cmds = append(cmds, notifications.ShowError(fmt.Errorf("replicas is required")))
				break
			}
			model.CloseOverlay()
			cmds = append(cmds, model.composeProjectCommand(
				"scale",
				"up",
				"--scale",
				fmt.Sprintf("%s=%s", selected.Service.Name, replicas),
				"-d",
			))
		}
	}

	// Main View Logic (only when no overlay)
	if !model.IsOverlayVisible() && model.IsListFocused() {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			if model.IsFiltering() {
				break
			}

			switch {
			case key.Matches(msg, model.keybindings.switchTab):
				return model, tea.Batch(cmds...) // Handled by parent
			case key.Matches(msg, model.keybindings.toggleSelection):
				model.handleToggleSelection()
			case key.Matches(msg, model.keybindings.toggleAll):
				model.handleToggleAll()
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

	serviceID := serviceSelectionID(selectedItem.Service)
	if serviceID != model.currentServiceID {
		model.currentServiceID = serviceID
		model.refreshServiceDetails(selectedItem.Service)
	}

	return nil
}

func (model *Model) refreshServiceDetails(service client.Service) {
	model.SetContent(model.buildInspectContent(service, model.GetContentWidth()))
	model.SetExtraContent(model.buildComposeContent(service, model.getComposeContentWidth()))
}

func composeSummary(service client.Service) string {
	running := 0
	stopped := 0
	for _, c := range service.Containers {
		if c.State == "running" {
			running++
		} else {
			stopped++
		}
	}

	project := service.Project
	if strings.TrimSpace(project) == "" {
		project = "unknown"
	}

	return fmt.Sprintf(
		"Project: %s\nService: %s\nReplicas: %d\nHealth: %d running / %d stopped",
		project,
		service.Name,
		service.Replicas,
		running,
		stopped,
	)
}

func (model *Model) getComposeContentWidth() int {
	if vp, ok := model.SplitView.Extra.(*components.ViewportPane); ok {
		return vp.Viewport.Width()
	}
	return model.GetContentWidth()
}

func (model *Model) buildInspectContent(service client.Service, width int) string {
	format := model.detailsPanel.GetFormatForDisplay()
	panel := builders.BuildServicePanel(service, width, false, format)
	return composeSummary(service) + "\n\n" + panel
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

func (model *Model) composeProjectCommand(actionLabel string, args ...string) tea.Cmd {
	targets := model.selectedServices()
	if len(targets) == 0 {
		return nil
	}

	project := strings.TrimSpace(targets[0].Project)
	workingDir := strings.TrimSpace(targets[0].WorkingDir)
	composeFile := strings.TrimSpace(targets[0].ComposeFile)
	serviceNames := make([]string, 0, len(targets))
	serviceIDs := make([]string, 0, len(targets))
	for _, service := range targets {
		if strings.TrimSpace(service.Project) == "" {
			return notifications.ShowError(fmt.Errorf("service %q is not part of a compose project", service.Name))
		}
		if service.Project != project {
			return notifications.ShowError(fmt.Errorf("selected services must belong to the same compose project"))
		}
		serviceNames = append(serviceNames, service.Name)
		serviceIDs = append(serviceIDs, serviceSelectionID(service))
	}

	return func() tea.Msg {
		return MsgComposeActionStart{
			Action:       actionLabel,
			Project:      project,
			ServiceNames: serviceNames,
			ServiceIDs:   serviceIDs,
			Args:         args,
			WorkingDir:   workingDir,
			ComposeFile:  composeFile,
		}
	}
}

func (model *Model) showScaleDialog() {
	targets := model.selectedServices()
	if len(targets) != 1 {
		return
	}
	selected := targets[0]

	fields := []components.FormField{
		{
			Label:       "Replicas",
			Placeholder: "e.g. 3",
			Value:       fmt.Sprintf("%d", selected.Replicas),
			Required:    true,
		},
	}

	dialog := components.NewFormDialog(
		"Scale Service",
		fields,
		base.SmartDialogAction{Type: "ScaleServiceAction"},
		nil,
	)
	model.SetOverlay(dialog)
}

func (model *Model) selectedServices() []client.Service {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) > 0 {
		items := model.GetItems()
		services := make([]client.Service, 0, len(selectedIDs))
		for _, item := range items {
			if !slices.Contains(selectedIDs, serviceSelectionID(item.Service)) {
				continue
			}
			services = append(services, item.Service)
		}
		return services
	}

	selected := model.GetSelectedItem()
	if selected == nil {
		return nil
	}

	return []client.Service{selected.Service}
}

func (model *Model) setWorkingState(serviceIDs []string, working bool) tea.Cmd {
	var cmds []tea.Cmd
	items := model.GetItems()
	for i, item := range items {
		if !slices.Contains(serviceIDs, serviceSelectionID(item.Service)) {
			continue
		}
		item.isWorking = working
		if working {
			item.spinner = newSpinner()
			spin := item.spinner
			cmds = append(cmds, func() tea.Msg { return spin.Tick() })
		}
		model.SetItem(i, item)
	}
	return tea.Batch(cmds...)
}

func (model *Model) applyServicesLoaded(fresh []ServiceItem) {
	current := model.GetItems()
	byName := make(map[string]ServiceItem, len(current))
	for _, item := range current {
		byName[serviceSelectionID(item.Service)] = item
	}

	updated := make([]ServiceItem, 0, len(fresh))
	for _, item := range fresh {
		prev, ok := byName[serviceSelectionID(item.Service)]
		if ok {
			item.isWorking = prev.isWorking
			item.isSelected = prev.isSelected
			item.spinner = prev.spinner
		} else {
			item.spinner = newSpinner()
		}
		updated = append(updated, item)
	}

	listItems := make([]list.Item, len(updated))
	for i, item := range updated {
		listItems[i] = item
	}
	_ = model.SplitView.List.SetItems(listItems)
}

func (model *Model) performComposeAction(msg MsgComposeActionStart) tea.Cmd {
	return func() tea.Msg {
		err := state.GetClient().RunComposeCommand(stdcontext.Background(), msg.WorkingDir, msg.ComposeFile, msg.Args...)
		if err != nil {
			return MsgComposeActionResult{
				Action:       msg.Action,
				Project:      msg.Project,
				ServiceNames: msg.ServiceNames,
				ServiceIDs:   msg.ServiceIDs,
				Err:          fmt.Errorf("compose %s failed for project %q: %w", msg.Action, msg.Project, err),
			}
		}

		return MsgComposeActionResult{
			Action:       msg.Action,
			Project:      msg.Project,
			ServiceNames: msg.ServiceNames,
			ServiceIDs:   msg.ServiceIDs,
		}
	}
}

func (model *Model) performServiceAction(msg MsgServiceActionStart) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch msg.Action {
		case "start":
			err = state.GetClient().StartContainers(stdcontext.Background(), msg.ContainerIDs)
		case "stop":
			err = state.GetClient().StopContainers(stdcontext.Background(), msg.ContainerIDs)
		case "restart":
			err = state.GetClient().RestartContainers(stdcontext.Background(), msg.ContainerIDs)
		default:
			err = fmt.Errorf("unsupported service action %q", msg.Action)
		}

		if err != nil {
			return MsgServiceActionResult{
				Action:       msg.Action,
				ServiceNames: msg.ServiceNames,
				ServiceIDs:   msg.ServiceIDs,
				Err:          fmt.Errorf("failed to %s services: %w", msg.Action, err),
			}
		}

		return MsgServiceActionResult{
			Action:       msg.Action,
			ServiceNames: msg.ServiceNames,
			ServiceIDs:   msg.ServiceIDs,
		}
	}
}

func (model *Model) handleToggleSelection() {
	selected := model.GetSelectedItem()
	if selected == nil || selected.isWorking {
		return
	}

	model.HandleToggleSelection()
	index := model.GetSelectedIndex()
	if item := model.GetSelectedItem(); item != nil {
		item.isSelected = model.Selections.IsSelected(serviceSelectionID(item.Service))
		model.SetItem(index, *item)
	}
}

func (model *Model) handleToggleAll() {
	model.HandleToggleAll()
	items := model.GetItems()
	for i, item := range items {
		item.isSelected = model.Selections.IsSelected(serviceSelectionID(item.Service))
		model.SetItem(i, item)
	}
}

func (model Model) View() string {
	return lipgloss.NewStyle().MarginTop(1).Render(model.ResourceView.View())
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
