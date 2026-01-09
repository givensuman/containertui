package colors

import (
	"fmt"
	"strings"

	"github.com/givensuman/containertui/internal/config"
)

// ParseColors parses color overrides from a slice of strings
// Format: ["primary=#b4befe'", "yellow=#f9e2af", "green=#a6e3a1"]
func ParseColors(colorStrings []string) (*config.ColorConfig, error) {
	if len(colorStrings) == 0 {
		return &config.ColorConfig{}, nil
	}

	colorConfig := &config.ColorConfig{}
	allPairs := []string{}

	// Collect all pairs from all strings
	for _, colorString := range colorStrings {
		colorString = strings.TrimSpace(colorString)
		if colorString == "" {
			continue
		}
		// Split each string by commas to handle cases where users might still use commas
		pairs := strings.Split(colorString, ",")
		allPairs = append(allPairs, pairs...)
	}

	for _, pair := range allPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid color format: %s (expected key=value)", pair)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		// Validate that value doesn't contain '=' (invalid format)
		if strings.Contains(value, "=") {
			return nil, fmt.Errorf("invalid color value: %s (values cannot contain '=')", value)
		}

		switch key {
		case "primary":
			colorConfig.Primary = config.ConfigString(value)
		case "yellow":
			colorConfig.Yellow = config.ConfigString(value)
		case "green":
			colorConfig.Green = config.ConfigString(value)
		case "gray":
			colorConfig.Gray = config.ConfigString(value)
		case "blue":
			colorConfig.Blue = config.ConfigString(value)
		case "white":
			colorConfig.White = config.ConfigString(value)
		case "black":
			colorConfig.Black = config.ConfigString(value)
		case "red":
			colorConfig.Red = config.ConfigString(value)
		case "magenta":
			colorConfig.Magenta = config.ConfigString(value)
		case "cyan":
			colorConfig.Cyan = config.ConfigString(value)
		case "bright-black":
			colorConfig.BrightBlack = config.ConfigString(value)
		case "bright-red":
			colorConfig.BrightRed = config.ConfigString(value)
		case "bright-green":
			colorConfig.BrightGreen = config.ConfigString(value)
		case "bright-yellow":
			colorConfig.BrightYellow = config.ConfigString(value)
		case "bright-blue":
			colorConfig.BrightBlue = config.ConfigString(value)
		case "bright-magenta":
			colorConfig.BrightMagenta = config.ConfigString(value)
		case "bright-cyan":
			colorConfig.BrightCyan = config.ConfigString(value)
		case "bright-white":
			colorConfig.BrightWhite = config.ConfigString(value)
		default:
			return nil, fmt.Errorf("unknown color key: %s", key)
		}
	}

	return colorConfig, nil
}
