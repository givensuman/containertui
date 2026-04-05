// Package safety provides shared wording helpers for destructive operation dialogs.
package safety

import "fmt"

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
