package components

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/colors"
	"github.com/givensuman/containertui/internal/ui/base"
	"github.com/givensuman/containertui/internal/ui/layout"
	"github.com/givensuman/containertui/internal/ui/styles"
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
	vp := viewport.New()
	vp.SoftWrap = true // Enable text wrapping for long lines
	return &ViewportPane{
		Viewport: vp,
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
	// Let the viewport handle text wrapping internally
	// lipgloss.Place() in SplitView will constrain the final output dimensions
	return v.Viewport.View()
}

func (v *ViewportPane) SetSize(w, h int) {
	v.Viewport.SetWidth(w)
	v.Viewport.SetHeight(h)
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

	// Store delegates to preserve custom UpdateFunc and other settings
	focusedDelegate    list.DefaultDelegate
	unfocusedDelegate  list.DefaultDelegate
	hasCachedDelegates bool
}

func NewSplitView(list list.Model, detail Pane) SplitView {
	return SplitView{
		List:   list,
		Detail: detail,
		Focus:  FocusList,
		style:  lipgloss.NewStyle(), // Base style, will be updated on resize
	}
}

// SetDelegates stores the focused and unfocused versions of a delegate
// This should be called after initializing the list with a custom delegate
func (s *SplitView) SetDelegates(baseDelegate list.DefaultDelegate) {
	s.focusedDelegate = styles.ChangeDelegateStyles(baseDelegate)
	s.unfocusedDelegate = styles.UnfocusDelegateStyles(baseDelegate)
	s.hasCachedDelegates = true
	// Apply focused delegate initially
	s.List.SetDelegate(s.focusedDelegate)
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
				cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: true} })
			} else {
				s.Focus = FocusList
				cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: false} })
			}
			return s, tea.Batch(cmds...)
		}
	}

	// Handle Focus Changes
	if _, ok := msg.(base.MsgFocusChanged); ok {
		// Use stored delegates if available, otherwise fall back to creating new ones
		if s.Focus == FocusDetail {
			if s.hasCachedDelegates {
				// We have a stored unfocused delegate
				s.List.SetDelegate(s.unfocusedDelegate)
			} else {
				// Fallback: create new delegate and apply unfocused styles
				base := list.NewDefaultDelegate()
				s.List.SetDelegate(styles.UnfocusDelegateStyles(base))
			}
		} else {
			if s.hasCachedDelegates {
				// We have a stored focused delegate
				s.List.SetDelegate(s.focusedDelegate)
			} else {
				// Fallback: create new delegate and apply focused styles
				base := list.NewDefaultDelegate()
				s.List.SetDelegate(styles.ChangeDelegateStyles(base))
			}
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
	layoutManager := layout.NewLayoutManager(width, height)
	masterLayout, detailLayout := layoutManager.CalculateMasterDetail(s.style)

	s.style = s.style.Width(masterLayout.Width).Height(masterLayout.Height)

	// Resize List
	// Note: masterLayout.ContentWidth/Height accounts for padding/borders if applied to the container
	s.List.SetWidth(masterLayout.ContentWidth)
	s.List.SetHeight(masterLayout.ContentHeight)

	// Resize Detail Pane
	// The detail view has a border (2) and padding (2)
	// Border: 1 top + 1 bottom = 2
	// Padding: 1 top + 1 bottom = 2
	// Total vertical frame: 4
	contentW := detailLayout.Width - 4
	contentH := detailLayout.Height - 4

	if contentW < 0 {
		contentW = 0
	}
	if contentH < 0 {
		contentH = 0
	}

	s.Detail.SetSize(contentW, contentH)
}

func (s SplitView) View() string {
	layoutManager := layout.NewLayoutManager(s.width, s.height)
	_, detailLayout := layoutManager.CalculateMasterDetail(s.style)

	// 1. Render List
	listView := s.style.Render(s.List.View())

	// 2. Render Detail Wrapper (Border + Focus color)
	// Use lipgloss.Place to strictly constrain the viewport to exact dimensions
	borderColor := colors.Muted()
	if s.Focus == FocusDetail {
		borderColor = colors.Primary()
	}

	// Get viewport output (with text wrapping enabled via SoftWrap)
	viewportContent := s.Detail.View()

	// Calculate the exact dimensions for content after accounting for border and padding
	// Border adds 2 to width and 2 to height
	// Padding adds 2 to width and 2 to height (inside the border)
	// Total: width - 4, height - 4
	contentWidth := detailLayout.Width - 4
	contentHeight := detailLayout.Height - 4

	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Use lipgloss.Place to strictly constrain content to exact dimensions
	// The viewport wraps text at contentWidth, so Place will preserve wrapped lines
	// and truncate any height overflow
	placedContent := lipgloss.Place(contentWidth, contentHeight, lipgloss.Left, lipgloss.Top, viewportContent)

	// Apply border and padding WITHOUT setting explicit Width/Height
	// This lets the border wrap the placed content exactly
	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1) // Padding inside the border

	detailView := detailStyle.Render(placedContent)

	// 3. Join
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
}
