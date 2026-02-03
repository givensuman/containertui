package layout

import "math"

type WindowRatio struct {
	width  float64
	height float64
}

func AdjustedWidth(width int, ratio WindowRatio) int {
	widthFloat := float64(width)
	return int(math.Floor(widthFloat * ratio.width))
}

func AdjustedHeight(height int, ratio WindowRatio) int {
	heightFloat := float64(height)
	return int(math.Floor(heightFloat * ratio.height))
}

var (
	RatioFullscreen = WindowRatio{1.0, 1.0}
	RatioSmall      = WindowRatio{0.3, 0.15}
	RatioMedium     = WindowRatio{0.4, 0.2}
	RatioLarge      = WindowRatio{0.6, 0.4}

	// Legacy aliases for backwards compatibility
	RatioModal        = RatioMedium
	RatioLargeOverlay = WindowRatio{0.8, 0.8}
)
