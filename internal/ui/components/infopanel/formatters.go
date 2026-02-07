package infopanel

import (
	"fmt"
	"strings"
	"time"
)

// FormatBytes formats bytes into human-readable size (e.g., "1.5GB", "256MB").
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatBytesShort formats bytes into short human-readable size (e.g., "1.5G", "256M").
func FormatBytesShort(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"K", "M", "G", "T", "P"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(div), units[exp])
}

// FormatDuration formats a time.Duration into human-readable format (e.g., "2h 34m", "5d 3h").
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// FormatTimestamp formats a Unix timestamp into relative time (e.g., "2 hours ago", "3 days ago").
func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return FormatTimeAgo(t)
}

// FormatTimeAgo formats a time.Time into relative time string.
func FormatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < 7*24*time.Hour {
		days := int(duration.Hours()) / 24
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if duration < 30*24*time.Hour {
		weeks := int(duration.Hours()) / 24 / 7
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
	if duration < 365*24*time.Hour {
		months := int(duration.Hours()) / 24 / 30
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := int(duration.Hours()) / 24 / 365
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// TruncateString truncates a string to maxLen characters, adding ellipsis if needed.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// TruncateID truncates a container/image ID to 12 characters (standard Docker short ID).
func TruncateID(id string) string {
	// Remove sha256: prefix if present
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// FormatPercentage formats a float as a percentage string.
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// FormatList formats a list of strings with bullet points.
func FormatList(items []string, maxItems int) string {
	if len(items) == 0 {
		return "  (none)"
	}

	var builder strings.Builder
	displayCount := len(items)
	if maxItems > 0 && len(items) > maxItems {
		displayCount = maxItems
	}

	for i := 0; i < displayCount; i++ {
		builder.WriteString("  • ")
		builder.WriteString(items[i])
		builder.WriteString("\n")
	}

	if maxItems > 0 && len(items) > maxItems {
		remaining := len(items) - maxItems
		builder.WriteString(fmt.Sprintf("  ... (%d more)\n", remaining))
	}

	return strings.TrimRight(builder.String(), "\n")
}

// WrapText wraps text to fit within a specified width.
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// StatusBadge returns a styled status badge for a container state.
func StatusBadge(state string, isRunning bool) string {
	if isRunning {
		return "✓ " + state
	}
	return "✗ " + state
}
