package containers

import (
	"strings"
	"testing"
)

func TestRenderStatsGraphIncludesBarsAndSparkline(t *testing.T) {
	input := statsRenderInput{
		CPUPercent: 50,
		MemPercent: 25,
		CPUSeries:  []float64{0, 50, 100},
		MemSeries:  []float64{0, 25, 50},
		BarWidth:   10,
	}

	rendered := renderStatsGraph(input)

	if !strings.Contains(rendered, "[=====-----]") {
		t.Fatalf("expected CPU bar in output, got %q", rendered)
	}

	if !strings.Contains(rendered, "[===-------]") {
		t.Fatalf("expected memory bar in output, got %q", rendered)
	}

	if !strings.Contains(rendered, "▁▅█") {
		t.Fatalf("expected sparkline glyphs in output, got %q", rendered)
	}
}

func TestSparklineMapsSeriesToGlyphs(t *testing.T) {
	series := []float64{0, 25, 50, 75, 100}

	got := sparkline(series)

	if got != "▁▃▅▆█" {
		t.Fatalf("expected sparkline %q, got %q", "▁▃▅▆█", got)
	}
}

func TestRenderStatsGraphPreservesHighCPUText(t *testing.T) {
	input := statsRenderInput{
		CPUPercent: 175,
		MemPercent: 40,
		CPUSeries:  []float64{120, 140, 175},
		MemSeries:  []float64{20, 30, 40},
		BarWidth:   10,
	}

	rendered := renderStatsGraph(input)

	if !strings.Contains(rendered, "175.0%") {
		t.Fatalf("expected CPU text to preserve high value, got %q", rendered)
	}
}

func TestSparklineRetainsVariationAbove100(t *testing.T) {
	series := []float64{100, 120, 140, 160}

	got := sparkline(series)

	if got != "▅▆▇█" {
		t.Fatalf("expected varied high-cpu sparkline %q, got %q", "▅▆▇█", got)
	}
}

func TestPercentBarClampsRange(t *testing.T) {
	if got := percentBar(-20, 10); got != "----------" {
		t.Fatalf("expected lower clamp bar %q, got %q", "----------", got)
	}

	if got := percentBar(140, 10); got != "==========" {
		t.Fatalf("expected upper clamp bar %q, got %q", "==========", got)
	}
}
