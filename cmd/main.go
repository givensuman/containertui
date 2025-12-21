package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/givensuman/oci-tui/internal/client"
	"github.com/givensuman/oci-tui/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
