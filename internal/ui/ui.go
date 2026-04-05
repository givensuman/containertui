// Package ui implements the terminal user interface.
package ui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/state"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/browse"
	"github.com/givensuman/containertui/internal/ui/containers"
	"github.com/givensuman/containertui/internal/ui/images"
	"github.com/givensuman/containertui/internal/ui/networks"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/tabs"
	"github.com/givensuman/containertui/internal/ui/volumes"
)

// Model represents the top-level Bubbletea UI model.
type Model struct {
	width              int
	height             int
	previousTab        tabs.Tab
	tabsModel          tabs.Model
	containersModel    containers.Model
	imagesModel        images.Model
	volumesModel       volumes.Model
	networksModel      networks.Model
	browseModel        browse.Model
	notificationsModel notifications.Model
	help               help.Model
}

func NewModel(startupTab tabs.Tab) Model {
	width, height := state.GetWindowSize()

	tabsModel := tabs.New(startupTab)
	containersModel := containers.New()
	imagesModel := images.New()
	volumesModel := volumes.New()
	networksModel := networks.New()
	browseModel := browse.New()
	notificationsModel := notifications.New()

	helpModel := help.New()

	return Model{
		width:              width,
		height:             height,
		previousTab:        startupTab,
		tabsModel:          tabsModel,
		containersModel:    containersModel,
		imagesModel:        imagesModel,
		volumesModel:       volumesModel,
		networksModel:      networksModel,
		browseModel:        browseModel,
		notificationsModel: notificationsModel,
		help:               helpModel,
	}
}

func (model Model) Init() tea.Cmd {
	return tea.Batch(
		model.containersModel.Init(),
		model.imagesModel.Init(),
		model.volumesModel.Init(),
		model.networksModel.Init(),
		model.browseModel.Init(),
	)
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	refreshContainers, refreshImages, refreshVolumes, refreshNetworks, refreshBrowse := crossTabRefreshTargets(msg)

	if refreshContainers {
		var containersCmd tea.Cmd
		model.containersModel, containersCmd = model.containersModel.Update(msg)
		cmds = append(cmds, containersCmd)
	}
	if refreshImages {
		var imagesCmd tea.Cmd
		model.imagesModel, imagesCmd = model.imagesModel.Update(msg)
		cmds = append(cmds, imagesCmd)
	}
	if refreshVolumes {
		var volumesCmd tea.Cmd
		model.volumesModel, volumesCmd = model.volumesModel.Update(msg)
		cmds = append(cmds, volumesCmd)
	}
	if refreshNetworks {
		var networksCmd tea.Cmd
		model.networksModel, networksCmd = model.networksModel.Update(msg)
		cmds = append(cmds, networksCmd)
	}
	if refreshBrowse {
		var browseCmd tea.Cmd
		model.browseModel, browseCmd = model.browseModel.Update(msg)
		cmds = append(cmds, browseCmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		state.SetWindowSize(msg.Width, msg.Height)

		var tabsCmd tea.Cmd
		model.tabsModel, tabsCmd = model.tabsModel.Update(msg)
		cmds = append(cmds, tabsCmd)

		contentHeight := max(0, msg.Height-4)

		contentMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: contentHeight,
		}

		var containersCmd tea.Cmd
		model.containersModel, containersCmd = model.containersModel.Update(contentMsg)
		cmds = append(cmds, containersCmd)

		var imagesCmd tea.Cmd
		model.imagesModel, imagesCmd = model.imagesModel.Update(contentMsg)
		cmds = append(cmds, imagesCmd)

		var volumesCmd tea.Cmd
		model.volumesModel, volumesCmd = model.volumesModel.Update(contentMsg)
		cmds = append(cmds, volumesCmd)

		var networksCmd tea.Cmd
		model.networksModel, networksCmd = model.networksModel.Update(contentMsg)
		cmds = append(cmds, networksCmd)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return model, tea.Quit
		}

		// Check if the current view is filtering or has an overlay before processing quit and tab switches
		isFiltering := false
		hasOverlay := false
		switch model.tabsModel.ActiveTab {
		case tabs.Containers:
			isFiltering = model.containersModel.IsFiltering()
			hasOverlay = model.containersModel.IsOverlayVisible()
		case tabs.Images:
			isFiltering = model.imagesModel.IsFiltering()
			hasOverlay = model.imagesModel.IsOverlayVisible()
		case tabs.Volumes:
			isFiltering = model.volumesModel.IsFiltering()
			hasOverlay = model.volumesModel.IsOverlayVisible()
		case tabs.Networks:
			isFiltering = model.networksModel.IsFiltering()
			hasOverlay = model.networksModel.IsOverlayVisible()
		case tabs.Browse:
			isFiltering = model.browseModel.IsFiltering()
			hasOverlay = model.browseModel.IsOverlayVisible()
		}

		// Allow "q" to quit only when not filtering and no overlay is visible
		if msg.String() == "q" && !isFiltering && !hasOverlay {
			return model, tea.Quit
		}

		// Only process tab switching keypresses if not filtering and no overlay is visible
		if !isFiltering && !hasOverlay {
			var tabsCmd tea.Cmd
			model.tabsModel, tabsCmd = model.tabsModel.Update(msg)
			if tabsCmd != nil {
				cmds = append(cmds, tabsCmd)
			}
		}

		if msg.String() == "?" {
			if hasOverlay {
				model.help.ShowAll = false
			} else {
				model.help.ShowAll = !model.help.ShowAll
			}
		}
	}

	updatedNotifications, notificationsCmd := model.notificationsModel.Update(msg)
	if m, ok := updatedNotifications.(notifications.Model); ok {
		model.notificationsModel = m
	}
	cmds = append(cmds, notificationsCmd)

	// Detect tab changes and trigger update on newly active tab
	if model.tabsModel.ActiveTab != model.previousTab {
		model.previousTab = model.tabsModel.ActiveTab

		// Send a window size message to trigger update on the newly active tab
		// This ensures updateDetailContent() gets called
		dummyMsg := tea.WindowSizeMsg{Width: model.width, Height: model.height - 4}
		switch model.tabsModel.ActiveTab {
		case tabs.Images:
			var imagesCmd tea.Cmd
			model.imagesModel, imagesCmd = model.imagesModel.Update(dummyMsg)
			cmds = append(cmds, imagesCmd)
		case tabs.Networks:
			var networksCmd tea.Cmd
			model.networksModel, networksCmd = model.networksModel.Update(dummyMsg)
			cmds = append(cmds, networksCmd)
		case tabs.Volumes:
			var volumesCmd tea.Cmd
			model.volumesModel, volumesCmd = model.volumesModel.Update(dummyMsg)
			cmds = append(cmds, volumesCmd)
		case tabs.Browse:
			var browseCmd tea.Cmd
			model.browseModel, browseCmd = model.browseModel.Update(dummyMsg)
			cmds = append(cmds, browseCmd)
		}
	}

	switch model.tabsModel.ActiveTab {
	case tabs.Containers:
		if !refreshContainers {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var containersCmd tea.Cmd
				model.containersModel, containersCmd = model.containersModel.Update(msg)
				cmds = append(cmds, containersCmd)
			}
		}
	case tabs.Images:
		if !refreshImages {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var imagesCmd tea.Cmd
				model.imagesModel, imagesCmd = model.imagesModel.Update(msg)
				cmds = append(cmds, imagesCmd)
			}
		}
	case tabs.Volumes:
		if !refreshVolumes {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var volumesCmd tea.Cmd
				model.volumesModel, volumesCmd = model.volumesModel.Update(msg)
				cmds = append(cmds, volumesCmd)
			}
		}
	case tabs.Networks:
		if !refreshNetworks {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var networksCmd tea.Cmd
				model.networksModel, networksCmd = model.networksModel.Update(msg)
				cmds = append(cmds, networksCmd)
			}
		}
	case tabs.Browse:
		if !refreshBrowse {
			if _, ok := msg.(tea.WindowSizeMsg); !ok {
				var browseCmd tea.Cmd
				model.browseModel, browseCmd = model.browseModel.Update(msg)
				cmds = append(cmds, browseCmd)
			}
		}
	}

	if model.containersModel.IsOverlayVisible() ||
		model.imagesModel.IsOverlayVisible() ||
		model.volumesModel.IsOverlayVisible() ||
		model.networksModel.IsOverlayVisible() ||
		model.browseModel.IsOverlayVisible() {
		model.help.ShowAll = false
	}

	return model, tea.Batch(cmds...)
}

func crossTabRefreshTargets(msg tea.Msg) (refreshContainers, refreshImages, refreshVolumes, refreshNetworks, refreshBrowse bool) {
	switch typed := msg.(type) {
	case base.MsgContainerCreated:
		_ = typed
		return true, false, false, false, false
	case base.MsgImagePulled:
		_ = typed
		return false, true, false, false, false
	case base.MsgResourceChanged:
		switch typed.Resource {
		case base.ResourceContainer:
			return true, false, false, false, false
		case base.ResourceImage:
			return false, true, false, false, false
		case base.ResourceVolume:
			return false, false, true, false, false
		case base.ResourceNetwork:
			return false, false, false, true, false
		}
	case containers.MsgContainersRefreshed:
		_ = typed
		return true, false, false, false, false
	case containers.MsgRefreshContainers:
		_ = typed
		return true, false, false, false, false
	}

	return false, false, false, false, false
}

type helpProvider interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

// overlayNotifications overlays notifications on top of the content using lipgloss layers
func (model Model) overlayNotifications(content string) string {
	notificationsContent := model.notificationsModel.ViewString()
	if notificationsContent == "" {
		return content
	}

	if model.width == 0 || model.height == 0 {
		return content
	}

	// Create background layer with the full content
	bgLayer := lipgloss.NewLayer(content).Width(model.width).Height(model.height)

	// Create foreground layer with notifications
	fgLayer := lipgloss.NewLayer(notificationsContent)

	// Position notifications in the top-right corner
	fgWidth := fgLayer.GetWidth()
	// Position with some padding from the right edge and top
	x := model.width - fgWidth - 2
	y := 1 // Small padding from the top

	if x < 0 {
		x = 0
	}

	fgLayer = fgLayer.X(x).Y(y).Z(1)

	// Create canvas with both layers
	canvas := lipgloss.NewCanvas(bgLayer, fgLayer)

	return canvas.Render()
}

func (model Model) View() tea.View {
	var view tea.View

	tabsView := model.tabsModel.View()

	// Get the active view content based on active tab
	var contentViewContent string
	switch model.tabsModel.ActiveTab {
	case tabs.Containers:
		contentViewContent = model.containersModel.View()
	case tabs.Images:
		contentViewContent = model.imagesModel.View()
	case tabs.Volumes:
		contentViewContent = model.volumesModel.View()
	case tabs.Networks:
		contentViewContent = model.networksModel.View()
	case tabs.Browse:
		contentViewContent = model.browseModel.View()
	}

	contentViewStr := contentViewContent

	var helpView string
	var currentHelp helpProvider

	switch model.tabsModel.ActiveTab {
	case tabs.Containers:
		currentHelp = model.containersModel
	case tabs.Images:
		currentHelp = model.imagesModel
	case tabs.Volumes:
		currentHelp = model.volumesModel
	case tabs.Networks:
		currentHelp = model.networksModel
	case tabs.Browse:
		currentHelp = model.browseModel
	}

	if currentHelp != nil {
		helpView = model.help.View(currentHelp)
	}

	fullView := lipgloss.JoinVertical(lipgloss.Top, tabsView, contentViewStr)

	// Apply notification overlay helper
	fullView = model.overlayNotifications(fullView)

	if helpView == "" {
		view = tea.NewView(fullView)
		view.AltScreen = true
		view.MouseMode = tea.MouseModeCellMotion

		return view
	}

	helpStyle := lipgloss.NewStyle().Width(model.width)

	if model.help.ShowAll {
		helpStyle = helpStyle.
			Border(lipgloss.ASCIIBorder(), true, false, false, false).
			BorderForeground(colors.Muted())
	}

	renderedHelpView := helpStyle.Render(helpView)
	renderedHelpLines := strings.Split(renderedHelpView, "\n")
	renderedHelpHeight := len(renderedHelpLines)

	fullLines := strings.Split(fullView, "\n")

	if len(fullLines) >= renderedHelpHeight {
		for len(fullLines) < model.height {
			fullLines = append(fullLines, "")
		}

		cutPoint := model.height - renderedHelpHeight
		cutPoint = max(cutPoint, 0)
		cutPoint = min(cutPoint, len(fullLines))
		topLines := fullLines[:cutPoint]

		view = tea.NewView(strings.Join(append(topLines, renderedHelpLines...), "\n"))
		view.AltScreen = true
		view.MouseMode = tea.MouseModeCellMotion

		return view
	}

	view = tea.NewView(fullView)
	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion

	return view
}

func Start() error {
	cfg := state.GetConfig()

	// Determine startup tab
	startupTab := tabs.Containers // default
	if cfg.StartupTab != "" {
		if !tabs.IsValidTab(cfg.StartupTab) {
			fmt.Fprintf(os.Stderr, "warning: invalid startup tab '%s', using 'containers' instead\n", cfg.StartupTab)
		} else {
			startupTab = tabs.TabFromString(cfg.StartupTab)
		}
	}

	model := NewModel(startupTab)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
