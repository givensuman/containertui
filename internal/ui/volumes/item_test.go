package volumes

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/givensuman/containertui/internal/backend"
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
	item := VolumeItem{Volume: backend.Volume{Name: long}, IsMounted: false}

	title := item.Title()
	if !strings.Contains(title, "volume-name-that-is-way-too-l...") {
		t.Fatalf("expected truncated title content, got %q", title)
	}
	if strings.Contains(title, "normal-list-display") {
		t.Fatalf("expected full name to be truncated, got %q", title)
	}
}

func TestTruncateVolumeNameIsRuneSafe(t *testing.T) {
	name := "volume-東京-very-long-name-for-truncation"
	got := truncateVolumeName(name, 16)

	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
	if !utf8.ValidString(got) {
		t.Fatalf("expected valid UTF-8 after truncation, got %q", got)
	}
}

func TestTitleDoesNotContainBrokenANSIFragments(t *testing.T) {
	long := "volume-name-that-is-way-too-long-for-normal-list-display"
	item := VolumeItem{Volume: backend.Volume{Name: long}, IsMounted: true}

	title := item.Title()
	if strings.Contains(title, "[93m") || strings.Contains(title, "[0m") {
		t.Fatalf("expected no broken ANSI fragments, got %q", title)
	}
}

func TestTitleDoesNotApplyANSIToVolumeName(t *testing.T) {
	item := VolumeItem{Volume: backend.Volume{Name: "my-volume-name"}, IsMounted: false}
	title := item.Title()

	lastSpace := strings.LastIndex(title, " ")
	if lastSpace == -1 {
		t.Fatalf("expected icon and name in title, got %q", title)
	}
	namePart := title[lastSpace+1:]
	if strings.Contains(namePart, "\x1b[") {
		t.Fatalf("expected plain volume name without ANSI styling, got %q", namePart)
	}
	if strings.Contains(title, "\x1b[") {
		t.Fatalf("expected fully plain title without ANSI, got %q", title)
	}
}
