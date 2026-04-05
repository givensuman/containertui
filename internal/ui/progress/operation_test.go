package progress

import "testing"

func TestStagePercent(t *testing.T) {
	if got := StagePercent(0, 3); got <= 0 || got >= 1 {
		t.Fatalf("expected stage percent between 0 and 1, got %f", got)
	}

	if got := StagePercent(100, 3); got >= 1 {
		t.Fatalf("expected capped staged percent below 1, got %f", got)
	}
}
