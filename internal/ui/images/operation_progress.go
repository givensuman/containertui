package images

import "fmt"

type stagedOperationProgress struct {
	stages []string
	index  int
}

func newStagedOperationProgress(stages []string) stagedOperationProgress {
	if len(stages) == 0 {
		stages = []string{"Working..."}
	}

	return stagedOperationProgress{stages: stages, index: 0}
}

func (p stagedOperationProgress) status() string {
	if len(p.stages) == 0 {
		return "Working..."
	}

	idx := p.index
	if idx < 0 {
		idx = 0
	}
	if idx >= len(p.stages) {
		idx = len(p.stages) - 1
	}

	return p.stages[idx]
}

func (p stagedOperationProgress) percent() float64 {
	count := len(p.stages)
	if count == 0 {
		return 0
	}

	idx := p.index
	if idx < 0 {
		idx = 0
	}
	if idx >= count {
		idx = count - 1
	}

	return float64(idx+1) / float64(count+1)
}

func (p *stagedOperationProgress) advance() {
	if p.index < len(p.stages)-1 {
		p.index++
	}
}

func createContainerStages(autoStart bool) []string {
	stages := []string{
		"Validating container configuration...",
		"Creating container...",
		"Finalizing container setup...",
	}

	if autoStart {
		stages = append(stages, "Starting container...")
	}

	return stages
}

func pruneStages(resource string) []string {
	return []string{
		fmt.Sprintf("Discovering %s to prune...", resource),
		fmt.Sprintf("Pruning %s...", resource),
		fmt.Sprintf("Finalizing %s prune...", resource),
	}
}
