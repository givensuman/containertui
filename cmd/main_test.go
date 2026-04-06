package main

import "testing"

func TestResolveStartupTab(t *testing.T) {
	tests := []struct {
		name       string
		configTab  string
		subcommand string
		want       string
	}{
		{
			name:       "uses config startup tab when no overrides are provided",
			configTab:  "images",
			subcommand: "",
			want:       "images",
		},
		{
			name:       "uses subcommand tab over config value",
			configTab:  "images",
			subcommand: "volumes",
			want:       "volumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveStartupTab(tt.configTab, tt.subcommand)
			if got != tt.want {
				t.Fatalf("resolveStartupTab(%q, %q) = %q, want %q", tt.configTab, tt.subcommand, got, tt.want)
			}
		})
	}
}
