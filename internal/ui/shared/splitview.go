package shared

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
)

// FocusState defines which pane is currently active.
type FocusState int

const (
	FocusList FocusState = iota
	FocusDetail
)

// Pane is a component that can be placed in the right side of a SplitView.
// It must accept size updates and standard bubbletea events.
type Pane interface {
	Init() tea.Cmd
	Update(tea.Msg) (Pane, tea.Cmd)
	View() string
	SetSize(width, height int)
}

// ViewportPane is a default implementation of Pane that wraps a text viewport.
type ViewportPane struct {
	Viewport viewport.Model
}

func NewViewportPane() *ViewportPane {
	return &ViewportPane{
		Viewport: viewport.New(),
	}
}

func (v *ViewportPane) Init() tea.Cmd {
	return nil
}

func (v *ViewportPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	var cmd tea.Cmd
	v.Viewport, cmd = v.Viewport.Update(msg)
	return v, cmd
}

func (v *ViewportPane) View() string {
	return v.Viewport.View()
}

func (v *ViewportPane) SetSize(w, h int) {
	v.Viewport = viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))
}

func (v *ViewportPane) SetContent(s string) {
	v.Viewport.SetContent(s)
}

// SplitView manages a left-side List and a right-side Pane.
type SplitView struct {
	List   list.Model
	Detail Pane
	Focus  FocusState

	// width and height of the entire component
	width  int
	height int

	style lipgloss.Style
}

func NewSplitView(list list.Model, detail Pane) SplitView {
	return SplitView{
		List:   list,
		Detail: detail,
		Focus:  FocusList,
		style:  lipgloss.NewStyle(), // Base style, will be updated on resize
	}
}

func (s SplitView) Init() tea.Cmd {
	return nil
}

func (s SplitView) Update(msg tea.Msg) (SplitView, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global tab switching for focus
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "tab" && s.List.FilterState() != list.Filtering {
			if s.Focus == FocusList {
				s.Focus = FocusDetail
				cmds = append(cmds, func() tea.Msg { return MsgFocusChanged{IsDetailsFocused: true} })
			} else {
				s.Focus = FocusList
				cmds = append(cmds, func() tea.Msg { return MsgFocusChanged{IsDetailsFocused: false} })
			}
			return s, tea.Batch(cmds...)
		}
	}

	// Handle Focus Changes
	if _, ok := msg.(MsgFocusChanged); ok {
		// Note: list.Model does not export Delegate().
		// We cannot get the current delegate to cast it.
		// However, we can construct new delegates since we know the types we use in this app.
		// We assume the caller initializes the List with a DefaultDelegate.

		if s.Focus == FocusDetail {
			// We need to construct a base delegate and unfocus it.
			// Since we can't read the existing one, we create a new default and apply unfocused styles.
			// This assumes standard behavior.
			base := list.NewDefaultDelegate()
			// We might lose specific settings like ShowDescription if they were changed dynamically.
			// But for this app, they are usually static per view.
			s.List.SetDelegate(UnfocusDelegateStyles(base))
		} else {
			// Refocus
			base := list.NewDefaultDelegate()
			s.List.SetDelegate(ChangeDelegateStyles(base))
		}
	}

	// Handle Resize
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.SetSize(msg.Width, msg.Height)
	}

	// Forward events based on Focus
	// Note: We always update the List slightly so it can handle filter inputs even if not fully focused?
	// Actually, standard behavior:
	// If FocusList: List gets keys.
	// If FocusDetail: Detail gets keys.
	// Both get WindowSizeMsg (handled above via SetSize).

	var cmd tea.Cmd

	// If filtering, List needs input regardless of our internal "Focus" state
	// (though usually we are in FocusList if filtering).
	if s.List.FilterState() == list.Filtering {
		s.List, cmd = s.List.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		// Normal navigation
		if s.Focus == FocusList {
			s.List, cmd = s.List.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Update Detail
			updatedDetail, cmd := s.Detail.Update(msg)
			s.Detail = updatedDetail.(Pane) // Type assertion back to interface
			cmds = append(cmds, cmd)
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *SplitView) SetSize(width, height int) {
	s.width = width
	s.height = height

	// Use existing LayoutManager
	layoutManager := NewLayoutManager(width, height)
	masterLayout, detailLayout := layoutManager.CalculateMasterDetail(s.style)

	s.style = s.style.Width(masterLayout.Width).Height(masterLayout.Height)

	// Resize List
	// Note: masterLayout.ContentWidth/Height accounts for padding/borders if applied to the container
	s.List.SetWidth(masterLayout.ContentWidth)
	s.List.SetHeight(masterLayout.ContentHeight)

	// Resize Detail Pane
	// We calculate the inner size available for the pane
	// The detail view usually has a border.
	// We need to account for that border in the Pane size or the container.
	// In the original code:
	// detailStyle := lipgloss.NewStyle().Width(detailLayout.Width - 2).Height(detailLayout.Height).Border(...).Padding(1)
	// viewportWidth := detailLayout.Width - 4 (2 for border, 2 for padding)

	// To keep SplitView generic, we will let the Pane handle its own internal sizing,
	// BUT we must determine if SplitView renders the border or the Pane does.
	//
	// Strategy: SplitView renders the layout structure (two side-by-side blocks).
	// It passes the explicit dimensions to the children.
	//
	// However, the "Focus Border" (blue when active) was part of the original renderMainView logic.
	// If we move that into SplitView, SplitView needs to know about borders.

	// Let's have SplitView calculate the "Container" size for the right side.
	// The Pane receives the size MINUS the border/padding that SplitView draws.
	// SplitView will draw the border to indicate focus.

	// Original code:
	// detailStyle... Width(detailLayout.Width - 2).Height(detailLayout.Height)... Padding(1)
	// viewportWidth := detailLayout.Width - 4
	// viewportHeight := detailLayout.Height - 2

	// We'll mimic this:
	// The "Container" for the detail view is (W-2, H).
	// It has Padding(1), effectively reducing content area by another 2 in width and 2 in height.
	// So content area is (W-4, H-2).

	contentW := detailLayout.Width - 4
	contentH := detailLayout.Height - 2

	if contentW < 0 {
		contentW = 0
	}
	if contentH < 0 {
		contentH = 0
	}

	s.Detail.SetSize(contentW, contentH)
}

func (s SplitView) View() string {
	layoutManager := NewLayoutManager(s.width, s.height)
	_, detailLayout := layoutManager.CalculateMasterDetail(s.style)

	// 1. Render List
	listView := s.style.Render(s.List.View())

	// 2. Render Detail Wrapper (Border + Focus color)
	borderColor := colors.Muted()
	if s.Focus == FocusDetail {
		borderColor = colors.Primary()
	}

	detailStyle := lipgloss.NewStyle().
		Width(detailLayout.Width - 2). // -2 for the border itself
		Height(detailLayout.Height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1) // Padding inside the border

	detailView := detailStyle.Render(s.Detail.View())

	// 3. Join
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}
