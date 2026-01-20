package components

import (
	"sync"
)

// SelectionManager handles selection state for a generic item type T.
// T must be comparable (string, int, etc).
// It tracks both the selected IDs and their list indices for efficient updates.
type SelectionManager[T comparable] struct {
	selections map[T]int
	mu         sync.RWMutex
}

// NewSelectionManager creates a new SelectionManager instance.
func NewSelectionManager[T comparable]() *SelectionManager[T] {
	return &SelectionManager[T]{
		selections: make(map[T]int),
	}
}

// Select marks an item as selected, storing its list index.
func (sm *SelectionManager[T]) Select(id T, index int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selections[id] = index
}

// Unselect removes an item from the selection.
func (sm *SelectionManager[T]) Unselect(id T) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.selections, id)
}

// Toggle toggles the selection state of an item.
// If currently selected, it unselects. If not, it selects.
func (sm *SelectionManager[T]) Toggle(id T, index int) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.selections[id]; exists {
		delete(sm.selections, id)
		return false
	}
	sm.selections[id] = index
	return true
}

// IsSelected checks if an item is currently selected.
func (sm *SelectionManager[T]) IsSelected(id T) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, exists := sm.selections[id]
	return exists
}

// GetSelected returns a slice of all selected IDs.
// Order is not guaranteed.
func (sm *SelectionManager[T]) GetSelected() []T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	ids := make([]T, 0, len(sm.selections))
	for id := range sm.selections {
		ids = append(ids, id)
	}
	return ids
}

// GetIndices returns a slice of all selected item indices.
// Order is not guaranteed.
func (sm *SelectionManager[T]) GetIndices() []int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	indices := make([]int, 0, len(sm.selections))
	for _, idx := range sm.selections {
		indices = append(indices, idx)
	}
	return indices
}

// Clear removes all selections.
func (sm *SelectionManager[T]) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selections = make(map[T]int)
}

// Count returns the number of selected items.
func (sm *SelectionManager[T]) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.selections)
}

// SetSelections bulk updates selections.
// Useful when selecting all items.
func (sm *SelectionManager[T]) SetSelections(selections map[T]int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selections = selections
}
