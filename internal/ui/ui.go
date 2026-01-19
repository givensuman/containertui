// Package ui implements the terminal user interface.
package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/containers"
	"github.com/givensuman/containertui/internal/ui/images"
	"github.com/givensuman/containertui/internal/ui/networks"
	"github.com/givensuman/containertui/internal/ui/notifications"
	"github.com/givensuman/containertui/internal/ui/services"
	"github.com/givensuman/containertui/internal/ui/tabs"
	"github.com/givensuman/containertui/internal/ui/volumes"
)

// Model represents the top-level Bubbletea UI model.
type Model struct {
	width              int
	height             int
	tabsModel          tabs.Model
	containersModel    containers.Model
	imagesModel        images.Model
	volumesModel       volumes.Model
	networksModel      networks.Model
	servicesModel      services.Model
	notificationsModel notifications.Model
	help               help.Model
}

func NewModel() Model {
	width, height := context.GetWindowSize()

	tabsModel := tabs.New()
	containersModel := containers.New()
	imagesModel := images.New()
	volumesModel := volumes.New()
	networksModel := networks.New()
	servicesModel := services.New()
	notificationsModel := notifications.New()

	helpModel := help.New()

	return Model{
		width:              width,
		height:             height,
		tabsModel:          tabsModel,
		containersModel:    containersModel,
		imagesModel:        imagesModel,
		volumesModel:       volumesModel,
		networksModel:      networksModel,
		servicesModel:      servicesModel,
		notificationsModel: notificationsModel,
		help:               helpModel,
	}
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		model.width = msg.Width
		model.height = msg.Height
		context.SetWindowSize(msg.Width, msg.Height)

		updatedTabs, _ := model.tabsModel.Update(msg)
		model.tabsModel = updatedTabs.(tabs.Model)

		contentHeight := msg.Height - 4
		if contentHeight < 0 {
			contentHeight = 0
		}

		contentMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: contentHeight,
		}

		updatedContainers, _ := model.containersModel.Update(contentMsg)
		model.containersModel = updatedContainers.(containers.Model)

		updatedImages, _ := model.imagesModel.Update(contentMsg)
		model.imagesModel = updatedImages.(images.Model)

		updatedVolumes, _ := model.volumesModel.Update(contentMsg)
		model.volumesModel = updatedVolumes.(volumes.Model)

		updatedNetworks, _ := model.networksModel.Update(contentMsg)
		model.networksModel = updatedNetworks.(networks.Model)

		updatedServices, _ := model.servicesModel.Update(contentMsg)
		model.servicesModel = updatedServices.(services.Model)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return model, tea.Quit
		}

		updatedTabs, tabsCmd := model.tabsModel.Update(msg)
		model.tabsModel = updatedTabs.(tabs.Model)
		if tabsCmd != nil {
			cmds = append(cmds, tabsCmd)
		}

		if msg.String() == "?" {
			model.help.ShowAll = !model.help.ShowAll
		}
	}

	updatedNotifications, notificationsCmd := model.notificationsModel.Update(msg)
	model.notificationsModel = updatedNotifications.(notifications.Model)
	cmds = append(cmds, notificationsCmd)

	switch model.tabsModel.ActiveTab {
	case tabs.Containers:
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedContainers, containersCmd := model.containersModel.Update(msg)
			model.containersModel = updatedContainers.(containers.Model)
			cmds = append(cmds, containersCmd)
		}
	case tabs.Images:
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedImages, imagesCmd := model.imagesModel.Update(msg)
			model.imagesModel = updatedImages.(images.Model)
			cmds = append(cmds, imagesCmd)
		}
	case tabs.Volumes:
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedVolumes, volumesCmd := model.volumesModel.Update(msg)
			model.volumesModel = updatedVolumes.(volumes.Model)
			cmds = append(cmds, volumesCmd)
		}
	case tabs.Networks:
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedNetworks, networksCmd := model.networksModel.Update(msg)
			model.networksModel = updatedNetworks.(networks.Model)
			cmds = append(cmds, networksCmd)
		}
	case tabs.Services:
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedServices, servicesCmd := model.servicesModel.Update(msg)
			model.servicesModel = updatedServices.(services.Model)
			cmds = append(cmds, servicesCmd)
		}
	}

	return model, tea.Batch(cmds...)
}

type helpProvider interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

func (model Model) View() tea.View {
	// Both tabs and content views return tea.View
	// We need to extract the string content to join them vertically
	// For now,  we'll render the tea.View content using fmt.Sprint
	tabsView := fmt.Sprint(model.tabsModel.View().Content)

	// Get the active view
	var contentViewContent tea.Layer
	switch model.tabsModel.ActiveTab {
	case tabs.Containers:
		contentViewContent = model.containersModel.View().Content
	case tabs.Images:
		contentViewContent = model.imagesModel.View().Content
	case tabs.Volumes:
		contentViewContent = model.volumesModel.View().Content
	case tabs.Networks:
		contentViewContent = model.networksModel.View().Content
	case tabs.Services:
		contentViewContent = model.servicesModel.View().Content
	}

	contentViewStr := fmt.Sprint(contentViewContent)

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
	case tabs.Services:
		currentHelp = model.servicesModel
	}

	if currentHelp != nil {
		helpView = model.help.View(currentHelp)
	}

	fullView := lipgloss.JoinVertical(lipgloss.Top, tabsView, contentViewStr)

	if helpView == "" {
		return tea.NewView(fullView)
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
		if cutPoint < 0 {
			cutPoint = 0
		}
		if cutPoint > len(fullLines) {
			cutPoint = len(fullLines)
		}
		topLines := fullLines[:cutPoint]
		return tea.NewView(strings.Join(append(topLines, renderedHelpLines...), "\n"))
	}

	return tea.NewView(fullView)
}

func Start() error {
	model := NewModel()

	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
