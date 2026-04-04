package volumes

import (
	"strings"
	"testing"

	"github.com/givensuman/containertui/internal/client"
)

func TestTruncateVolumeName(t *testing.T) {
	short := "short-name"
	if got := truncateVolumeName(short, 20); got != short {
		t.Fatalf("expected unchanged short name, got %q", got)
	}

	long := "volume-name-that-is-way-too-long-for-normal-list-display"
	got := truncateVolumeName(long, 20)
	if got != "volume-name-that-..." {
		t.Fatalf("unexpected truncation result: %q", got)
	}

	if got := truncateVolumeName(long, 3); got != long {
		t.Fatalf("expected unchanged name when maxLen <= 3, got %q", got)
	}
}

func TestTitleUsesTruncatedVolumeName(t *testing.T) {
	long := "volume-name-that-is-way-too-long-for-normal-list-display"
	item := VolumeItem{Volume: client.Volume{Name: long}, IsMounted: false}

	title := item.Title()
	if !strings.Contains(title, "volume-name-that-is-way-too-l...") {
		t.Fatalf("expected truncated title content, got %q", title)
	}
	if strings.Contains(title, "normal-list-display") {
		t.Fatalf("expected full name to be truncated, got %q", title)
	}
}
