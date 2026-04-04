package networks

import (
	"regexp"
	"strings"
	"testing"

	"github.com/givensuman/containertui/internal/client"
)

func TestHandleCreateNetworkCompleteSuccess(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateNetworkComplete(MsgCreateNetworkComplete{NetworkID: "123456789012345"})
	if cmd == nil {
		t.Fatal("expected command for successful create network completion")
	}
}

func TestHandleCreateNetworkCompleteError(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateNetworkComplete(MsgCreateNetworkComplete{Err: errTestCreateNetwork})
	if cmd == nil {
		t.Fatal("expected command for failed create network completion")
	}
}

func TestWithCreateNetworkDialogShowsOverlay(t *testing.T) {
	model := Model{}

	model = model.withCreateNetworkDialog()

	if !model.IsOverlayVisible() {
		t.Fatal("expected create network dialog to be visible")
	}
}

var errTestCreateNetwork = testError("create network failed")

type testError string

func (e testError) Error() string {
	return string(e)
}

func TestNetworkTitleDoesNotWrapNameWithANSI(t *testing.T) {
	name := "very-long-network-name"
	item := NetworkItem{Network: client.Network{Name: name}, IsActive: true}
	title := item.Title()

	if regexp.MustCompile("\\x1b\\[[0-9;]*m" + regexp.QuoteMeta(name) + "\\x1b\\[[0-9;]*m").MatchString(title) {
		t.Fatalf("expected network name to be plain text, got %q", title)
	}
	if strings.Contains(title, "\x1b[") {
		t.Fatalf("expected fully plain title without ANSI, got %q", title)
	}
}
