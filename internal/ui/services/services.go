package services

import (
	stdcontext "context"
	"os"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/components"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
	"github.com/givensuman/containertui/internal/ui/components/infopanel/builders"
)

type MsgRefreshServices time.Time

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
	switchTab key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		switchTab: key.NewBinding(
			key.WithKeys("1", "2", "3", "4", "5"),
			key.WithHelp("1-5", "switch tab"),
		),
	}
}

type Model struct {
	components.ResourceView[string, ServiceItem]
	keybindings        *keybindings
	detailsKeybindings detailsKeybindings
	currentServiceName string
	currentFormat      string
}

func New() Model {
	serviceKeybindings := newKeybindings()

	fetchServices := func() ([]ServiceItem, error) {
		services, err := state.GetClient().GetServices(stdcontext.Background())
		if err != nil {
			return []ServiceItem{}, nil
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

	// Set custom delegate
	delegate := newDefaultDelegate()
	resourceView.SetDelegate(delegate)

	model := Model{
		ResourceView:       *resourceView,
		keybindings:        serviceKeybindings,
		detailsKeybindings: newDetailsKeybindings(),
		currentFormat:      "",
	}

	// Add custom keybindings to help
	model.AdditionalHelp = []key.Binding{
		serviceKeybindings.switchTab,
	}

	return model
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return MsgRefreshServices(t)
	})
}

func (model Model) Init() tea.Cmd {
	return tea.Batch(model.ResourceView.Init(), tickCmd())
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
		// Refresh the services list via ResourceView
		cmds = append(cmds, model.Refresh())
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
			}
		}
	} else if !model.IsOverlayVisible() && !model.IsListFocused() {
		// Detail pane is focused
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, model.detailsKeybindings.ToggleJSON):
				cmd := model.handleToggleFormat()
				cmds = append(cmds, cmd)
			case key.Matches(msg, model.detailsKeybindings.CopyOutput):
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
		model.SetContent(lipgloss.NewStyle().Foreground(colors.Muted()).Render("No service selected."))
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
	// Determine format to use
	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	// Use the panel builder with selected format
	panelContent := builders.BuildServicePanel(service, model.GetContentWidth(), false, format)

	// Add compose file content if available
	if service.ComposeFile != "" {
		data, err := os.ReadFile(service.ComposeFile)
		if err == nil {
			panelContent += "\n\n"
			sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(colors.Primary()).Underline(true).MarginTop(1).MarginBottom(0)
			panelContent += sectionHeader.Render("Compose File Content") + "\n"
			panelContent += string(data)
		}
	}

	model.SetContent(panelContent)
}

// handleCopyToClipboard copies the current service details to clipboard
func (model *Model) handleCopyToClipboard() tea.Cmd {
	selectedItem := model.GetSelectedItem()
	if selectedItem == nil {
		return nil
	}

	// Determine which format to use
	format := infopanel.GetOutputFormat()
	if model.currentFormat != "" {
		if model.currentFormat == "json" {
			format = infopanel.FormatJSON
		} else {
			format = infopanel.FormatYAML
		}
	}

	// Marshal the data without syntax highlighting
	data, err := infopanel.MarshalToFormat(selectedItem.Service, format)
	if err != nil {
		return nil
	}

	// Copy to clipboard
	if err := clipboard.WriteAll(string(data)); err != nil {
		return nil
	}

	return nil
}

// handleToggleFormat toggles between JSON and YAML format
func (model *Model) handleToggleFormat() tea.Cmd {
	// Determine current effective format
	currentFormat := model.currentFormat
	if currentFormat == "" {
		cfg := state.GetConfig()
		currentFormat = cfg.InspectionFormat
		if currentFormat == "" {
			currentFormat = "yaml"
		}
	}

	// Toggle to the opposite format
	if currentFormat == "json" {
		model.currentFormat = "yaml"
	} else {
		model.currentFormat = "json"
	}

	// Refresh content with new format
	selectedItem := model.GetSelectedItem()
	if selectedItem != nil {
		model.refreshServiceDetails(selectedItem.Service)
	}

	return nil
}

func (model Model) View() string {
	return model.ResourceView.View()
}

func (model Model) IsFiltering() bool {
	return model.ResourceView.IsFiltering()
}

func (model Model) ShortHelp() []key.Binding {
	if !model.IsListFocused() {
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
