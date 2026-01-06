// Package containers defines the containers list component.
package containers

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type keybindings struct {
	pauseContainer   key.Binding
	unpauseContainer key.Binding
	startContainer   key.Binding
	stopContainer    key.Binding
	toggleSelect     key.Binding
}

func newKeybindings() *keybindings {
	return &keybindings{
		pauseContainer: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause container"),
		),
		unpauseContainer: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "unpause container"),
		),
		startContainer: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start container"),
		),
		stopContainer: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stop container"),
		),
		toggleSelect: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp(tea.KeySpace.String(), "toggle selection"),
		),
	}
}
