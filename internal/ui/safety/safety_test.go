package safety

import (
	"strings"
	"testing"
)

func TestDeleteConfirmationContainsDestructiveGuidance(t *testing.T) {
	msg := DeleteConfirmation("image", "sha256:abc")
	if msg == "" {
		t.Fatal("expected non-empty delete confirmation message")
	}
	if !strings.Contains(msg, "destructive") || !strings.Contains(msg, "cannot be undone") {
		t.Fatalf("expected destructive guidance in message, got %q", msg)
	}
}

func TestForceDeleteInUseConfirmationContainsDisruptionGuidance(t *testing.T) {
	msg := ForceDeleteInUseConfirmation("Volume", "data", 2, []string{"c1", "c2"})
	if !strings.Contains(msg, "disrupt") {
		t.Fatalf("expected disruption guidance in force delete message, got %q", msg)
	}
}

func TestForceDeleteInUseConfirmationContainsContainerCount(t *testing.T) {
	msg := ForceDeleteInUseConfirmation("Image", "sha256:abc", 3, []string{"a", "b", "c"})
	if !strings.Contains(msg, "used by 3 containers") {
		t.Fatalf("expected container count guidance in force delete message, got %q", msg)
	}
}

func TestPruneConfirmationContainsCountAndSamples(t *testing.T) {
	msg := PruneConfirmation("images", 4, []string{"nginx:latest", "redis:7", "alpine:3.20"})
	if !strings.Contains(msg, "Prune 4 images") {
		t.Fatalf("expected prune count in confirmation, got %q", msg)
	}
	if !strings.Contains(msg, "nginx:latest") || !strings.Contains(msg, "redis:7") {
		t.Fatalf("expected sample names in confirmation, got %q", msg)
	}
	if !strings.Contains(msg, "destructive") {
		t.Fatalf("expected destructive guidance in confirmation, got %q", msg)
	}
}
