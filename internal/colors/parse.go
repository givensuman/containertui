package colors

import (
	"fmt"
	"strings"

	"github.com/givensuman/containertui/internal/config"
)

// ParseColors parses color overrides from a slice of strings
// Format: ["primary=#b4befe'", "warning=#f9e2af", "success=#a6e3a1"]
func ParseColors(colorStrings []string) (*config.ThemeConfig, error) {
	if len(colorStrings) == 0 {
		return &config.ThemeConfig{}, nil
	}

	themeConfig := &config.ThemeConfig{}
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

		// Validate that value doesn't contain another '='
		if strings.Contains(value, "=") {
			return nil, fmt.Errorf("invalid color value: %s (values cannot contain '=')", value)
		}

		switch key {
		case "primary":
			themeConfig.Primary = config.ConfigString(value)
		case "border":
			themeConfig.Border = config.ConfigString(value)
		case "text":
			themeConfig.Text = config.ConfigString(value)
		case "muted":
			themeConfig.Muted = config.ConfigString(value)
		case "selected":
			themeConfig.Selected = config.ConfigString(value)
		case "success":
			themeConfig.Success = config.ConfigString(value)
		case "warning":
			themeConfig.Warning = config.ConfigString(value)
		case "error":
			themeConfig.Error = config.ConfigString(value)
		default:
			return nil, fmt.Errorf("unknown theme key: %s", key)
		}
	}

	return themeConfig, nil
}
