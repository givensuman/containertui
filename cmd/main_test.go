package main

import "testing"

func TestResolveStartupTab(t *testing.T) {
	tests := []struct {
		name       string
		configTab  string
		flagTab    string
		subcommand string
		want       string
	}{
		{
			name:       "uses config startup tab when no overrides are provided",
			configTab:  "images",
			flagTab:    "",
			subcommand: "",
			want:       "images",
		},
		{
			name:       "uses --startup-tab flag over config value",
			configTab:  "images",
			flagTab:    "services",
			subcommand: "",
			want:       "services",
		},
		{
			name:       "uses subcommand tab over --startup-tab flag",
			configTab:  "images",
			flagTab:    "services",
			subcommand: "volumes",
			want:       "volumes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveStartupTab(tt.configTab, tt.flagTab, tt.subcommand)
			if got != tt.want {
				t.Fatalf("resolveStartupTab(%q, %q, %q) = %q, want %q", tt.configTab, tt.flagTab, tt.subcommand, got, tt.want)
			}
		})
	}
}
