package infopanel

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/givensuman/containertui/internal/context"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the format for inspection output.
type OutputFormat string

const (
	// FormatYAML represents YAML output format.
	FormatYAML OutputFormat = "yaml"
	// FormatJSON represents JSON output format.
	FormatJSON OutputFormat = "json"
)

// GetOutputFormat returns the configured output format for inspection data.
// Defaults to YAML if not configured.
func GetOutputFormat() OutputFormat {
	cfg := context.GetConfig()
	if cfg == nil {
		return FormatYAML
	}

	format := strings.ToLower(cfg.InspectionFormat)
	switch format {
	case "json":
		return FormatJSON
	case "yaml", "":
		return FormatYAML
	default:
		return FormatYAML
	}
}

// MarshalToFormat converts a struct to YAML or JSON format.
// For YAML, it first marshals to JSON (respecting JSON tags), then converts to YAML.
func MarshalToFormat(data interface{}, format OutputFormat) (string, error) {
	switch format {
	case FormatYAML:
		return marshalToYAML(data)
	case FormatJSON:
		return marshalToJSON(data)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// marshalToYAML converts data to YAML format.
// Uses JSON as intermediate format to respect JSON struct tags.
func marshalToYAML(data interface{}) (string, error) {
	// First marshal to JSON to respect JSON tags
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	// Then unmarshal JSON to generic map/slice structure
	var intermediate interface{}
	if err := json.Unmarshal(jsonBytes, &intermediate); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Finally marshal to YAML with proper formatting
	yamlBytes, err := yaml.Marshal(intermediate)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// marshalToJSON converts data to pretty-printed JSON format.
func marshalToJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// ColorizeYAML applies syntax highlighting to YAML text using theme colors.
func ColorizeYAML(text string) string {
	cfg := context.GetConfig()
	if cfg == nil {
		return text
	}

	// Get theme colors
	theme := cfg.Theme
	primaryColor := string(theme.Primary)
	successColor := string(theme.Success)
	warningColor := string(theme.Warning)
	mutedColor := string(theme.Muted)

	// Create styles
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(primaryColor))
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(successColor))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(warningColor))
	boolNullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor))

	lines := strings.Split(text, "\n")
	var result strings.Builder

	// Regex patterns for YAML syntax
	keyPattern := regexp.MustCompile(`^(\s*)([a-zA-Z_][\w-]*):`)
	stringPattern := regexp.MustCompile(`:\s*"([^"]*)"`)
	stringPattern2 := regexp.MustCompile(`:\s*'([^']*)'`)
	numberPattern := regexp.MustCompile(`:\s*(-?\d+\.?\d*)`)
	boolPattern := regexp.MustCompile(`:\s*(true|false|null)`)

	for _, line := range lines {
		colorizedLine := line

		// Colorize keys (must be first to avoid conflicts)
		if keyPattern.MatchString(line) {
			matches := keyPattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				indent := matches[1]
				key := matches[2]
				rest := strings.TrimPrefix(line, indent+key)
				colorizedLine = indent + keyStyle.Render(key) + rest
			}
		}

		// Colorize quoted strings
		colorizedLine = stringPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := stringPattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return `: ` + stringStyle.Render(`"`+parts[1]+`"`)
			}
			return match
		})

		colorizedLine = stringPattern2.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := stringPattern2.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return `: ` + stringStyle.Render(`'`+parts[1]+`'`)
			}
			return match
		})

		// Colorize booleans and null (before numbers to avoid conflicts)
		colorizedLine = boolPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := boolPattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return `: ` + boolNullStyle.Render(parts[1])
			}
			return match
		})

		// Colorize numbers
		colorizedLine = numberPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := numberPattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return `: ` + numberStyle.Render(parts[1])
			}
			return match
		})

		result.WriteString(colorizedLine)
		result.WriteString("\n")
	}

	return strings.TrimRight(result.String(), "\n")
}

// ColorizeJSON applies syntax highlighting to JSON text using theme colors.
func ColorizeJSON(text string) string {
	cfg := context.GetConfig()
	if cfg == nil {
		return text
	}

	// Get theme colors
	theme := cfg.Theme
	primaryColor := string(theme.Primary)
	successColor := string(theme.Success)
	warningColor := string(theme.Warning)
	mutedColor := string(theme.Muted)

	// Create styles
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(primaryColor))
	stringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(successColor))
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(warningColor))
	boolNullStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(mutedColor))

	lines := strings.Split(text, "\n")
	var result strings.Builder

	// Regex patterns for JSON syntax
	keyPattern := regexp.MustCompile(`"([^"]+)":`)
	stringPattern := regexp.MustCompile(`:\s*"([^"]*)"`)
	numberPattern := regexp.MustCompile(`:\s*(-?\d+\.?\d*)([,\s}])`)
	boolNullPattern := regexp.MustCompile(`:\s*(true|false|null)([,\s}])`)

	for _, line := range lines {
		colorizedLine := line

		// Colorize keys first
		colorizedLine = keyPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := keyPattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return keyStyle.Render(`"`+parts[1]+`"`) + `:`
			}
			return match
		})

		// Colorize string values
		colorizedLine = stringPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := stringPattern.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return `: ` + stringStyle.Render(`"`+parts[1]+`"`)
			}
			return match
		})

		// Colorize booleans and null (before numbers)
		colorizedLine = boolNullPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := boolNullPattern.FindStringSubmatch(match)
			if len(parts) >= 3 {
				return `: ` + boolNullStyle.Render(parts[1]) + parts[2]
			}
			return match
		})

		// Colorize numbers
		colorizedLine = numberPattern.ReplaceAllStringFunc(colorizedLine, func(match string) string {
			parts := numberPattern.FindStringSubmatch(match)
			if len(parts) >= 3 {
				return `: ` + numberStyle.Render(parts[1]) + parts[2]
			}
			return match
		})

		result.WriteString(colorizedLine)
		result.WriteString("\n")
	}

	return strings.TrimRight(result.String(), "\n")
}
