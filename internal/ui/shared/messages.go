package shared

// MsgFocusChanged is sent when focus switches between list and details views.
type MsgFocusChanged struct {
	IsDetailsFocused bool
}
