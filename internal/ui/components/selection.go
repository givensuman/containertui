package components

import (
	"sync"
)

type SelectionManager[T comparable] struct {
	selections map[T]int
	mu         sync.RWMutex
}

func NewSelectionManager[T comparable]() *SelectionManager[T] {
	return &SelectionManager[T]{
		selections: make(map[T]int),
	}
}

func (sm *SelectionManager[T]) Select(id T, index int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selections[id] = index
}

func (sm *SelectionManager[T]) Unselect(id T) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.selections, id)
}

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

func (sm *SelectionManager[T]) IsSelected(id T) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, exists := sm.selections[id]
	return exists
}

func (sm *SelectionManager[T]) GetSelected() []T {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	ids := make([]T, 0, len(sm.selections))
	for id := range sm.selections {
		ids = append(ids, id)
	}
	return ids
}

func (sm *SelectionManager[T]) GetIndices() []int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	indices := make([]int, 0, len(sm.selections))
	for _, idx := range sm.selections {
		indices = append(indices, idx)
	}
	return indices
}

func (sm *SelectionManager[T]) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selections = make(map[T]int)
}

func (sm *SelectionManager[T]) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.selections)
}

func (sm *SelectionManager[T]) SetSelections(selections map[T]int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.selections = selections
}
