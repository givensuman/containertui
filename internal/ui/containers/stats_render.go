package containers

import (
	"fmt"
	"math"
	"strings"
)

type statsRenderInput struct {
	CPUPercent float64
	MemPercent float64
	CPUSeries  []float64
	MemSeries  []float64
	BarWidth   int
}

func renderStatsGraph(input statsRenderInput) string {
	barWidth := input.BarWidth
	if barWidth <= 0 {
		barWidth = 10
	}

	cpuBar := percentBar(input.CPUPercent, barWidth)
	memBar := percentBar(input.MemPercent, barWidth)

	return fmt.Sprintf(
		"CPU [%s] %.1f%%\nCPU trend %s\nMEM [%s] %.1f%%\nMEM trend %s",
		cpuBar,
		clamp(input.CPUPercent, 0, 100),
		sparkline(input.CPUSeries),
		memBar,
		clamp(input.MemPercent, 0, 100),
		sparkline(input.MemSeries),
	)
}

func percentBar(percent float64, width int) string {
	if width <= 0 {
		return ""
	}

	clamped := clamp(percent, 0, 100)
	filled := int(math.Round((clamped / 100) * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	return strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
}

func sparkline(series []float64) string {
	if len(series) == 0 {
		return ""
	}

	levels := []rune("▁▂▃▄▅▆▇█")
	var out strings.Builder
	out.Grow(len(series) * 3)

	for _, v := range series {
		normalized := clamp(v, 0, 100) / 100
		idx := int(math.Round(normalized * float64(len(levels)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(levels) {
			idx = len(levels) - 1
		}
		out.WriteRune(levels[idx])
	}

	return out.String()
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
