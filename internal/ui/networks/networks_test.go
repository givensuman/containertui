package networks

import "testing"

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
