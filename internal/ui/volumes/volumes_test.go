package volumes

import "testing"

func TestHandleCreateVolumeCompleteSuccess(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateVolumeComplete(MsgCreateVolumeComplete{VolumeName: "my-volume"})
	if cmd == nil {
		t.Fatal("expected command for successful create volume completion")
	}
}

func TestHandleCreateVolumeCompleteError(t *testing.T) {
	model := Model{}

	_, cmd := model.handleCreateVolumeComplete(MsgCreateVolumeComplete{Err: errTestCreateVolume})
	if cmd == nil {
		t.Fatal("expected command for failed create volume completion")
	}
}

func TestWithCreateVolumeDialogShowsOverlay(t *testing.T) {
	model := Model{}

	model = model.withCreateVolumeDialog()

	if !model.IsOverlayVisible() {
		t.Fatal("expected create volume dialog to be visible")
	}
}

var errTestCreateVolume = testError("create volume failed")

type testError string

func (e testError) Error() string {
	return string(e)
}
