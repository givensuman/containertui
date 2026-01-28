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

type FocusState int

const (
	FocusList FocusState = iota
	FocusDetail
	FocusExtra
)

type Pane interface {
	Init() tea.Cmd
	Update(tea.Msg) (Pane, tea.Cmd)
	View() string
	SetSize(width, height int)
}

type ViewportPane struct {
	Viewport viewport.Model
}

func NewViewportPane() *ViewportPane {
	vp := viewport.New()
	vp.SoftWrap = true
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
	return v.Viewport.View()
}

func (v *ViewportPane) SetSize(w, h int) {
	v.Viewport.SetWidth(w)
	v.Viewport.SetHeight(h)
}

func (v *ViewportPane) SetContent(s string) {
	v.Viewport.SetContent(s)
}

// VerticalSplitPane contains two vertically stacked viewports
type VerticalSplitPane struct {
	Top    *ViewportPane
	Bottom *ViewportPane

	topHeight    int
	bottomHeight int
	width        int
	height       int
}

func NewVerticalSplitPane(topRatio float64) *VerticalSplitPane {
	if topRatio <= 0 || topRatio >= 1 {
		topRatio = 0.7 // default to 70/30 split
	}
	return &VerticalSplitPane{
		Top:    NewViewportPane(),
		Bottom: NewViewportPane(),
	}
}

func (v *VerticalSplitPane) Init() tea.Cmd {
	return nil
}

func (v *VerticalSplitPane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	var cmds []tea.Cmd

	// Update top pane
	updatedTop, cmd := v.Top.Update(msg)
	v.Top = updatedTop.(*ViewportPane)
	cmds = append(cmds, cmd)

	// Update bottom pane
	updatedBottom, cmd := v.Bottom.Update(msg)
	v.Bottom = updatedBottom.(*ViewportPane)
	cmds = append(cmds, cmd)

	return v, tea.Batch(cmds...)
}

func (v *VerticalSplitPane) View() string {
	topView := v.Top.View()

	// Create border for bottom panel
	bottomContent := v.Bottom.View()
	bottomStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Muted()).
		Padding(0, 1).
		Width(v.width).
		Height(v.bottomHeight)

	bottomView := bottomStyle.Render(bottomContent)

	return lipgloss.JoinVertical(lipgloss.Left, topView, bottomView)
}

func (v *VerticalSplitPane) SetSize(w, h int) {
	v.width = w
	v.height = h

	// Calculate split: 70% top, 30% bottom
	// Account for border (2 lines for top/bottom border)
	v.bottomHeight = h / 3
	if v.bottomHeight < 3 {
		v.bottomHeight = 3 // minimum for border
	}

	// Bottom pane needs to account for borders (4 lines total: 2 vertical + 2 padding)
	bottomContentHeight := v.bottomHeight - 4
	if bottomContentHeight < 1 {
		bottomContentHeight = 1
	}

	v.topHeight = h - v.bottomHeight - 1 // -1 for spacing
	if v.topHeight < 1 {
		v.topHeight = 1
	}

	v.Top.SetSize(w, v.topHeight)
	v.Bottom.SetSize(w-4, bottomContentHeight) // -4 for left/right borders + padding
}

func (v *VerticalSplitPane) SetTopContent(s string) {
	v.Top.SetContent(s)
}

func (v *VerticalSplitPane) SetBottomContent(s string) {
	v.Bottom.SetContent(s)
}

type SplitView struct {
	List   list.Model
	Detail Pane
	Extra  Pane // Optional third pane below Detail
	Focus  FocusState

	width  int
	height int

	style lipgloss.Style

	focusedDelegate    list.DefaultDelegate
	unfocusedDelegate  list.DefaultDelegate
	hasCachedDelegates bool

	extraRatio float64 // Ratio of height for Extra pane (0 means no extra pane)
}

func NewSplitView(list list.Model, detail Pane) SplitView {
	return SplitView{
		List:       list,
		Detail:     detail,
		Extra:      nil,
		Focus:      FocusList,
		style:      lipgloss.NewStyle(),
		extraRatio: 0.0,
	}
}

func (s *SplitView) SetExtraPane(extra Pane, heightRatio float64) {
	s.Extra = extra
	if heightRatio <= 0 || heightRatio >= 1 {
		heightRatio = 0.3 // default to 30% of height
	}
	s.extraRatio = heightRatio
}

func (s *SplitView) SetDelegates(baseDelegate list.DefaultDelegate) {
	s.focusedDelegate = styles.ChangeDelegateStyles(baseDelegate)
	s.unfocusedDelegate = styles.UnfocusDelegateStyles(baseDelegate)
	s.hasCachedDelegates = true
	s.List.SetDelegate(s.focusedDelegate)
}

func (s SplitView) Init() tea.Cmd {
	return nil
}

func (s SplitView) Update(msg tea.Msg) (SplitView, tea.Cmd) {
	var cmds []tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "tab" && s.List.FilterState() != list.Filtering {
			// Cycle through focus states
			if s.Extra != nil {
				// Three-pane mode: List -> Detail -> Extra -> List
				switch s.Focus {
				case FocusList:
					s.Focus = FocusDetail
					cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: true} })
				case FocusDetail:
					s.Focus = FocusExtra
					cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: true} })
				case FocusExtra:
					s.Focus = FocusList
					cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: false} })
				}
			} else {
				// Two-pane mode: List -> Detail -> List
				if s.Focus == FocusList {
					s.Focus = FocusDetail
					cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: true} })
				} else {
					s.Focus = FocusList
					cmds = append(cmds, func() tea.Msg { return base.MsgFocusChanged{IsDetailsFocused: false} })
				}
			}
			return s, tea.Batch(cmds...)
		}
	}

	if _, ok := msg.(base.MsgFocusChanged); ok {
		if s.Focus == FocusDetail || s.Focus == FocusExtra {
			if s.hasCachedDelegates {
				s.List.SetDelegate(s.unfocusedDelegate)
			} else {
				base := list.NewDefaultDelegate()
				s.List.SetDelegate(styles.UnfocusDelegateStyles(base))
			}
		} else {
			if s.hasCachedDelegates {
				s.List.SetDelegate(s.focusedDelegate)
			} else {
				base := list.NewDefaultDelegate()
				s.List.SetDelegate(styles.ChangeDelegateStyles(base))
			}
		}
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd

	if s.List.FilterState() == list.Filtering {
		s.List, cmd = s.List.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		switch s.Focus {
		case FocusList:
			s.List, cmd = s.List.Update(msg)
			cmds = append(cmds, cmd)
		case FocusDetail:
			updatedDetail, cmd := s.Detail.Update(msg)
			s.Detail = updatedDetail.(Pane)
			cmds = append(cmds, cmd)
		case FocusExtra:
			if s.Extra != nil {
				updatedExtra, cmd := s.Extra.Update(msg)
				s.Extra = updatedExtra.(Pane)
				cmds = append(cmds, cmd)
			}
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *SplitView) SetSize(width, height int) {
	s.width = width
	s.height = height

	layoutManager := layout.NewLayoutManager(width, height)
	masterLayout, detailLayout := layoutManager.CalculateMasterDetail(s.style)

	s.style = s.style.Width(masterLayout.Width).Height(masterLayout.Height)

	s.List.SetWidth(masterLayout.ContentWidth)
	s.List.SetHeight(masterLayout.ContentHeight)

	contentW := detailLayout.Width - 4

	if s.Extra != nil {
		// Split detail area vertically between Detail and Extra panes
		// Total available height
		totalH := detailLayout.Height

		// Calculate box heights (including border+padding)
		extraBoxHeight := int(float64(totalH) * s.extraRatio)
		if extraBoxHeight < 5 {
			extraBoxHeight = 5 // minimum for border + content
		}

		detailBoxHeight := totalH - extraBoxHeight - 1 // -1 for spacing between panes
		if detailBoxHeight < 5 {
			detailBoxHeight = 5
		}

		// Content heights (subtract 4 for border+padding)
		detailContentHeight := detailBoxHeight - 4
		extraContentHeight := extraBoxHeight - 4

		if contentW < 0 {
			contentW = 0
		}
		if detailContentHeight < 0 {
			detailContentHeight = 0
		}
		if extraContentHeight < 0 {
			extraContentHeight = 0
		}

		s.Detail.SetSize(contentW, detailContentHeight)
		s.Extra.SetSize(contentW, extraContentHeight)
	} else {
		// Single detail pane
		contentH := detailLayout.Height - 4

		if contentW < 0 {
			contentW = 0
		}
		if contentH < 0 {
			contentH = 0
		}

		s.Detail.SetSize(contentW, contentH)
	}
}

func (s SplitView) View() string {
	layoutManager := layout.NewLayoutManager(s.width, s.height)
	_, detailLayout := layoutManager.CalculateMasterDetail(s.style)

	listView := s.style.Render(s.List.View())

	contentWidth := detailLayout.Width - 4
	if contentWidth < 0 {
		contentWidth = 0
	}

	if s.Extra != nil {
		// Three-pane layout: List | Detail | Extra (stacked vertically)
		// Total available height (already accounts for outer border/padding)
		totalH := detailLayout.Height
		if totalH < 0 {
			totalH = 0
		}

		// Calculate how much height to allocate to each bordered box
		// Each box needs 4 lines for border+padding, plus content
		extraBoxHeight := int(float64(totalH) * s.extraRatio)
		if extraBoxHeight < 5 {
			extraBoxHeight = 5 // minimum for border + minimal content
		}

		// Remaining height goes to detail, with 1 line spacing between boxes
		detailBoxHeight := totalH - extraBoxHeight - 1 // -1 for spacing
		if detailBoxHeight < 5 {
			detailBoxHeight = 5
		}

		// Content height = box height - border (2) - padding (2)
		detailContentHeight := detailBoxHeight - 4
		extraContentHeight := extraBoxHeight - 4

		if detailContentHeight < 0 {
			detailContentHeight = 0
		}
		if extraContentHeight < 0 {
			extraContentHeight = 0
		}

		// Render detail pane
		detailBorderColor := colors.Muted()
		if s.Focus == FocusDetail {
			detailBorderColor = colors.Primary()
		}

		detailContent := s.Detail.View()
		placedDetailContent := lipgloss.Place(contentWidth, detailContentHeight, lipgloss.Left, lipgloss.Top, detailContent)

		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(detailBorderColor).
			Padding(1)

		detailView := detailStyle.Render(placedDetailContent)

		// Render extra pane
		extraBorderColor := colors.Muted()
		if s.Focus == FocusExtra {
			extraBorderColor = colors.Primary()
		}

		extraContent := s.Extra.View()
		placedExtraContent := lipgloss.Place(contentWidth, extraContentHeight, lipgloss.Left, lipgloss.Top, extraContent)

		extraStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(extraBorderColor).
			Padding(1)

		extraView := extraStyle.Render(placedExtraContent)

		// Stack detail and extra vertically
		rightColumn := lipgloss.JoinVertical(lipgloss.Left, detailView, extraView)

		return lipgloss.JoinHorizontal(lipgloss.Top, listView, rightColumn)
	} else {
		// Two-pane layout: List | Detail
		borderColor := colors.Muted()
		if s.Focus == FocusDetail {
			borderColor = colors.Primary()
		}

		viewportContent := s.Detail.View()

		contentHeight := detailLayout.Height - 4

		if contentHeight < 0 {
			contentHeight = 0
		}

		placedContent := lipgloss.Place(contentWidth, contentHeight, lipgloss.Left, lipgloss.Top, viewportContent)

		detailStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1)

		detailView := detailStyle.Render(placedContent)

		return lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)
	}
}
