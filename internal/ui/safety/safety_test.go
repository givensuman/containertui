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
