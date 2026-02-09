package browse

import (
	"time"

	"github.com/givensuman/containertui/internal/registry"
)

// MsgPopularImages is received after fetching popular images.
type MsgPopularImages struct {
	Images []registry.RegistryImage
	Err    error
}

// MsgSearchResults is received after remote search.
type MsgSearchResults struct {
	Query  string
	Images []registry.RegistryImage
	Err    error
}

// MsgImageInspection is received after fetching detailed info.
type MsgImageInspection struct {
	RepoName string
	Detail   registry.RegistryImageDetail
	Err      error
}

// MsgPullProgress provides progress updates during image pull.
type MsgPullProgress struct {
	ImageName string
	Status    string
	DoneChan  <-chan error // Channel to signal completion with error status
}

// MsgPullComplete indicates pull operation completed.
type MsgPullComplete struct {
	ImageName string
	Err       error
}

// MsgRestoreScroll signals to restore scroll position.
type MsgRestoreScroll struct{}

// MsgRefreshItems triggers a refresh of the image list.
type MsgRefreshItems time.Time
