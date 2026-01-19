package shared

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestAdjustedWidth(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		ratio    WindowRatio
		expected int
	}{
		{"100 width at 50%", 100, WindowRatio{0.5, 0.5}, 50},
		{"100 width at 100%", 100, WindowRatio{1.0, 1.0}, 100},
		{"99 width at 80%", 99, WindowRatio{0.8, 0.8}, 79},
		{"0 width at 50%", 0, WindowRatio{0.5, 0.5}, 0},
		{"1 width at 100%", 1, WindowRatio{1.0, 1.0}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AdjustedWidth(tt.width, tt.ratio)
			if result != tt.expected {
				t.Errorf("AdjustedWidth(%d, %v) = %d; want %d", tt.width, tt.ratio, result, tt.expected)
			}
		})
	}
}

func TestAdjustedHeight(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		ratio    WindowRatio
		expected int
	}{
		{"100 height at 50%", 100, WindowRatio{0.5, 0.5}, 50},
		{"100 height at 100%", 100, WindowRatio{1.0, 1.0}, 100},
		{"99 height at 80%", 99, WindowRatio{0.8, 0.8}, 79},
		{"0 height at 50%", 0, WindowRatio{0.5, 0.5}, 0},
		{"1 height at 100%", 1, WindowRatio{1.0, 1.0}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AdjustedHeight(tt.height, tt.ratio)
			if result != tt.expected {
				t.Errorf("AdjustedHeight(%d, %v) = %d; want %d", tt.height, tt.ratio, result, tt.expected)
			}
		})
	}
}

func TestLayoutManagerCalculate(t *testing.T) {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1)

	tests := []struct {
		name          string
		windowWidth   int
		windowHeight  int
		ratio         WindowRatio
		checkWidth    bool
		checkHeight   bool
		checkOffsetX  bool
		checkOffsetY  bool
		expectedWidth int
	}{
		{
			name:          "100x50 window fullscreen",
			windowWidth:   100,
			windowHeight:  50,
			ratio:         RatioFullscreen,
			checkWidth:    true,
			checkHeight:   true,
			expectedWidth: 100,
		},
		{
			name:          "100x50 window modal 40%",
			windowWidth:   100,
			windowHeight:  50,
			ratio:         RatioModal,
			checkWidth:    true,
			checkHeight:   true,
			expectedWidth: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := NewLayoutManager(tt.windowWidth, tt.windowHeight)
			dims := lm.Calculate(tt.ratio, style)

			if tt.checkWidth && dims.Width != tt.expectedWidth {
				t.Errorf("Width = %d; want %d", dims.Width, tt.expectedWidth)
			}
			if tt.checkHeight && dims.Height != int(float64(tt.windowHeight)*tt.ratio.height) {
				t.Errorf("Height = %d; want %d", dims.Height, int(float64(tt.windowHeight)*tt.ratio.height))
			}
			if tt.checkOffsetX && dims.OffsetX != (tt.windowWidth-dims.Width)/2 {
				t.Errorf("OffsetX = %d; expected %d", dims.OffsetX, (tt.windowWidth-dims.Width)/2)
			}
			if tt.checkOffsetY && dims.OffsetY != (tt.windowHeight-dims.Height)/2 {
				t.Errorf("OffsetY = %d; expected %d", dims.OffsetY, (tt.windowHeight-dims.Height)/2)
			}
		})
	}
}

func TestLayoutManagerUpdateDimensions(t *testing.T) {
	lm := NewLayoutManager(100, 50)

	if lm.windowWidth != 100 || lm.windowHeight != 50 {
		t.Errorf("Initial dimensions = (%d, %d); want (100, 50)", lm.windowWidth, lm.windowHeight)
	}

	lm.UpdateDimensions(200, 100)

	if lm.windowWidth != 200 || lm.windowHeight != 100 {
		t.Errorf("Updated dimensions = (%d, %d); want (200, 100)", lm.windowWidth, lm.windowHeight)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{-1, 1, -1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		result := Min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Min(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{-1, 1, 1},
		{0, 0, 0},
	}

	for _, tt := range tests {
		result := Max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Max(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestRatioConstants(t *testing.T) {
	if RatioFullscreen.width != 1.0 || RatioFullscreen.height != 1.0 {
		t.Errorf("RatioFullscreen = (%f, %f); want (1.0, 1.0)", RatioFullscreen.width, RatioFullscreen.height)
	}

	if RatioModal.width != 0.4 || RatioModal.height != 0.2 {
		t.Errorf("RatioModal = (%f, %f); want (0.4, 0.2)", RatioModal.width, RatioModal.height)
	}

	if RatioLargeOverlay.width != 0.8 || RatioLargeOverlay.height != 0.8 {
		t.Errorf("RatioLargeOverlay = (%f, %f); want (0.8, 0.8)", RatioLargeOverlay.width, RatioLargeOverlay.height)
	}
}
