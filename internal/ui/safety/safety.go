// Package safety provides shared wording helpers for destructive operation dialogs.
package safety

import (
	"fmt"
	"strings"
)

func DeleteConfirmation(resource, identifier string) string {
	return fmt.Sprintf(
		"Delete %s %s?\n\nThis action is destructive and cannot be undone.",
		resource,
		identifier,
	)
}

func ForceDeleteInUseConfirmation(resource, identifier string, inUseCount int, inUseBy any) string {
	return fmt.Sprintf(
		"%s %s is used by %d containers (%v).\n\nForce delete anyway? This may disrupt dependent workloads.",
		resource,
		identifier,
		inUseCount,
		inUseBy,
	)
}

func PruneConfirmation(resourcePlural string, count int, samples []string) string {
	message := fmt.Sprintf("Prune %d %s?\n\nThis action is destructive and cannot be undone.", count, resourcePlural)

	if len(samples) == 0 {
		return message
	}

	var b strings.Builder
	b.WriteString(message)
	b.WriteString("\n\nExamples:\n")
	for _, sample := range samples {
		trimmed := strings.TrimSpace(sample)
		if trimmed == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(trimmed)
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
