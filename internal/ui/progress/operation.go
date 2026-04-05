package progress

import (
	"bufio"
	"fmt"
	"io"
)

// StagePercent returns a monotonic staged percentage in range [0, 1).
func StagePercent(stageIndex, stageCount int) float64 {
	if stageCount <= 0 {
		return 0
	}

	if stageIndex < 0 {
		stageIndex = 0
	}
	if stageIndex >= stageCount {
		stageIndex = stageCount - 1
	}

	return float64(stageIndex+1) / float64(stageCount+1)
}

// StreamLines scans reader line-by-line and returns progress and completion channels.
func StreamLines(reader io.ReadCloser) (<-chan string, <-chan error) {
	progressChan := make(chan string, 100)
	doneChan := make(chan error, 1)

	go func() {
		defer close(progressChan)
		defer close(doneChan)
		defer reader.Close()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			progressChan <- scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			doneChan <- fmt.Errorf("failed to read stream output: %w", err)
			return
		}

		doneChan <- nil
	}()

	return progressChan, doneChan
}
