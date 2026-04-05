package images

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type buildProgressLine struct {
	Stream         string `json:"stream"`
	Status         string `json:"status"`
	ID             string `json:"id"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
	ErrorDetail struct {
		Message string `json:"message"`
	} `json:"errorDetail"`
}

var stepPattern = regexp.MustCompile(`Step\s+(\d+)/(\d+)`)

func parseBuildStatusMessage(raw string) string {
	var line buildProgressLine
	if err := json.Unmarshal([]byte(raw), &line); err != nil {
		return "Building image..."
	}

	if line.ErrorDetail.Message != "" {
		return line.ErrorDetail.Message
	}
	if line.Error != "" {
		return line.Error
	}

	if s := strings.TrimSpace(line.Stream); s != "" {
		return s
	}

	if line.Status != "" && line.ID != "" {
		if line.Progress != "" {
			return fmt.Sprintf("%s (%s) %s", line.Status, line.ID, strings.TrimSpace(line.Progress))
		}
		return fmt.Sprintf("%s (%s)", line.Status, line.ID)
	}

	if line.Status != "" {
		return line.Status
	}

	return "Building image..."
}

func estimateBuildPercent(raw string, currentMax float64) (float64, bool) {
	var line buildProgressLine
	if err := json.Unmarshal([]byte(raw), &line); err != nil {
		return 0, false
	}

	if line.ProgressDetail.Total > 0 {
		percent := float64(line.ProgressDetail.Current) / float64(line.ProgressDetail.Total)
		if percent > 0.95 {
			percent = 0.95
		}
		if percent < currentMax {
			percent = currentMax
		}
		return percent, true
	}

	stream := strings.TrimSpace(line.Stream)
	if stream == "" {
		return 0, false
	}

	matches := stepPattern.FindStringSubmatch(stream)
	if len(matches) == 3 {
		var step, total int
		_, _ = fmt.Sscanf(matches[0], "Step %d/%d", &step, &total)
		if total > 0 {
			percent := (float64(step) / float64(total)) * 0.9
			if percent < currentMax {
				percent = currentMax
			}
			return percent, true
		}
	}

	return 0, false
}
