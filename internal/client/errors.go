package client

import (
	"fmt"
	"strings"
)

// OperationError represents an error for a single resource operation.
type OperationError struct {
	ID  string
	Err error
}

// MultiError collects multiple operation errors.
type MultiError struct {
	Errors []OperationError
}

// Error implements the error interface.
func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return fmt.Sprintf("operation failed for %s: %v", m.Errors[0].ID, m.Errors[0].Err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d operations failed:\n", len(m.Errors)))
	for _, e := range m.Errors {
		sb.WriteString(fmt.Sprintf("  - %s: %v\n", e.ID, e.Err))
	}
	return sb.String()
}

// HasErrors returns true if there are any errors.
func (m *MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

// Add adds an operation error.
func (m *MultiError) Add(id string, err error) {
	if err != nil {
		m.Errors = append(m.Errors, OperationError{ID: id, Err: err})
	}
}

// ToError returns nil if no errors, otherwise returns the MultiError.
func (m *MultiError) ToError() error {
	if !m.HasErrors() {
		return nil
	}
	return m
}
