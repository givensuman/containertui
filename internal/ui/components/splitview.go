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

type SplitView struct {
	List   list.Model
	Detail Pane
	Focus  FocusState

	width  int
	height int

	style lipgloss.Style

	focusedDelegate    list.DefaultDelegate
	unfocusedDelegate  list.DefaultDelegate
	hasCachedDelegates bool
}

func NewSplitView(list list.Model, detail Pane) SplitView {
	return SplitView{
		List:   list,
		Detail: detail,
		Focus:  FocusList,
		style:  lipgloss.NewStyle(),
	}
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

	if _, ok := msg.(base.MsgFocusChanged); ok {
		if s.Focus == FocusDetail {
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
		if s.Focus == FocusList {
			s.List, cmd = s.List.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			updatedDetail, cmd := s.Detail.Update(msg)
			s.Detail = updatedDetail.(Pane)
			cmds = append(cmds, cmd)
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

	listView := s.style.Render(s.List.View())

	borderColor := colors.Muted()
	if s.Focus == FocusDetail {
		borderColor = colors.Primary()
	}

	viewportContent := s.Detail.View()

	contentWidth := detailLayout.Width - 4
	contentHeight := detailLayout.Height - 4

	if contentWidth < 0 {
		contentWidth = 0
	}
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
