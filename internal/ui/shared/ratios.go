package shared

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

var RatioContainerLogs WindowRatio = WindowRatio{
	width:  0.1,
	height: 0.1,
}
