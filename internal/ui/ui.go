// Package ui implements the terminal user interface.
package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"

	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/context"
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
	tabsModel          tabs.Model
	containersModel    containers.Model
	imagesModel        images.Model
	volumesModel       volumes.Model
	networksModel      networks.Model
	notificationsModel notifications.Model
	overlayModel       *overlay.Model
	help               help.Model
}

func NewModel() Model {
	width, height := context.GetWindowSize()

	tabsModel := tabs.New()
	containersModel := containers.New()
	imagesModel := images.New()
	volumesModel := volumes.New()
	networksModel := networks.New()
	notificationsModel := notifications.New()

	overlayModel := overlay.New(notificationsModel, containersModel, overlay.Right, overlay.Top, 0, 0)
	helpModel := help.New()

	return Model{
		width:              width,
		height:             height,
		tabsModel:          tabsModel,
		containersModel:    containersModel,
		imagesModel:        imagesModel,
		volumesModel:       volumesModel,
		networksModel:      networksModel,
		notificationsModel: notificationsModel,
		overlayModel:       overlayModel,
		help:               helpModel,
	}
}

func (model Model) Init() tea.Cmd {
	return nil
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	overlayMsg := msg

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
		overlayMsg = contentMsg

		updatedContainers, _ := model.containersModel.Update(contentMsg)
		model.containersModel = updatedContainers.(containers.Model)

		updatedImages, _ := model.imagesModel.Update(contentMsg)
		model.imagesModel = updatedImages.(images.Model)

		updatedVolumes, _ := model.volumesModel.Update(contentMsg)
		model.volumesModel = updatedVolumes.(volumes.Model)

		updatedNetworks, _ := model.networksModel.Update(contentMsg)
		model.networksModel = updatedNetworks.(networks.Model)

		model.help.Width = msg.Width

	case tea.KeyMsg:
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

	var activeView tea.Model
	switch model.tabsModel.ActiveTab {
	case tabs.Containers:
		activeView = model.containersModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedContainers, containersCmd := model.containersModel.Update(msg)
			model.containersModel = updatedContainers.(containers.Model)
			cmds = append(cmds, containersCmd)
			activeView = model.containersModel
		}
	case tabs.Images:
		activeView = model.imagesModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedImages, imagesCmd := model.imagesModel.Update(msg)
			model.imagesModel = updatedImages.(images.Model)
			cmds = append(cmds, imagesCmd)
			activeView = model.imagesModel
		}
	case tabs.Volumes:
		activeView = model.volumesModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedVolumes, volumesCmd := model.volumesModel.Update(msg)
			model.volumesModel = updatedVolumes.(volumes.Model)
			cmds = append(cmds, volumesCmd)
			activeView = model.volumesModel
		}
	case tabs.Networks:
		activeView = model.networksModel
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			updatedNetworks, networksCmd := model.networksModel.Update(msg)
			model.networksModel = updatedNetworks.(networks.Model)
			cmds = append(cmds, networksCmd)
			activeView = model.networksModel
		}
	}

	model.overlayModel.Foreground = model.notificationsModel
	model.overlayModel.Background = activeView

	updatedOverlay, overlayCmd := model.overlayModel.Update(overlayMsg)
	if ov, ok := updatedOverlay.(*overlay.Model); ok {
		model.overlayModel = ov
	}
	cmds = append(cmds, overlayCmd)

	return model, tea.Batch(cmds...)
}

type helpProvider interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

func (model Model) View() string {
	tabsView := model.tabsModel.View()
	contentView := model.overlayModel.View()

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
	}

	if currentHelp != nil {
		helpView = model.help.View(currentHelp)
	}

	fullView := lipgloss.JoinVertical(lipgloss.Top, tabsView, contentView)

	if helpView == "" {
		return fullView
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
		return strings.Join(append(topLines, renderedHelpLines...), "\n")
	}

	return fullView
}

func Start() error {
	model := NewModel()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
