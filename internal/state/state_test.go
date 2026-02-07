package state

import (
	"testing"

	"github.com/givensuman/containertui/internal/config"
)

func TestSetConfig(t *testing.T) {
	cfg := &config.Config{NoNerdFonts: true}
	SetConfig(cfg)
	if GetConfig() != cfg {
		t.Error("SetConfig did not set the config correctly")
	}
}
