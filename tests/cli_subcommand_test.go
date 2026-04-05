package tests

import (
	"testing"

	"github.com/givensuman/containertui/internal/ui/tabs"
)

func TestTabFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected tabs.Tab
		valid    bool
	}{
		{"containers", tabs.Containers, true},
		{"images", tabs.Images, true},
		{"volumes", tabs.Volumes, true},
		{"networks", tabs.Networks, true},
		{"services", tabs.Tab(-1), false},
		{"browse", tabs.Browse, true},
		{"Containers", tabs.Containers, true},
		{"IMAGES", tabs.Images, true},
		{"  volumes  ", tabs.Volumes, true},
		{"invalid", tabs.Tab(-1), false},
		{"", tabs.Tab(-1), false},
	}

	for _, tt := range tests {
		result := tabs.TabFromString(tt.input)
		if result != tt.expected {
			t.Errorf("TabFromString(%q) = %v, want %v", tt.input, result, tt.expected)
		}

		valid := tabs.IsValidTab(tt.input)
		if valid != tt.valid {
			t.Errorf("IsValidTab(%q) = %v, want %v", tt.input, valid, tt.valid)
		}
	}
}

func TestAllTabNames(t *testing.T) {
	names := tabs.AllTabNames()
	if len(names) != 5 {
		t.Errorf("AllTabNames() returned %d tabs, expected 5", len(names))
	}

	expectedNames := map[string]bool{
		"containers": true,
		"images":     true,
		"volumes":    true,
		"networks":   true,
		"browse":     true,
	}

	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("AllTabNames() returned unexpected tab: %q", name)
		}
	}
}
