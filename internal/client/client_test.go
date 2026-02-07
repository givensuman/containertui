package client

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.client == nil {
		t.Error("client.client is nil")
	}
}

func TestContainer(t *testing.T) {
	c := Container{
		ID:    "test-id",
		Name:  "test-name",
		Image: "test-image",
		State: "running",
	}
	if c.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", c.ID)
	}
	if c.Name != "test-name" {
		t.Errorf("expected Name test-name, got %s", c.Name)
	}
	if c.Image != "test-image" {
		t.Errorf("expected Image test-image, got %s", c.Image)
	}
	if c.State != "running" {
		t.Errorf("expected State running, got %s", c.State)
	}
}

func TestGetContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}
	defer func() {
		err := client.CloseClient()
		if err != nil {
			t.Fatalf("failed to close client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := client.GetContainers(ctx)
	if err != nil {
		t.Fatalf("failed to get containers: %v", err)
	}
	_ = containers
}

func TestGetContainerState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}
	defer func() {
		err := client.CloseClient()
		if err != nil {
			t.Fatalf("failed to close client: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state, err := client.GetContainerState(ctx, "nonexistent")
	if err == nil && state != "unknown" {
		t.Errorf("expected 'unknown' for nonexistent container, got %s", state)
	}
}

func TestMultiError(t *testing.T) {
	tests := []struct {
		name     string
		errors   []OperationError
		hasError bool
		errCount int
	}{
		{
			name:     "no errors",
			errors:   []OperationError{},
			hasError: false,
			errCount: 0,
		},
		{
			name: "single error",
			errors: []OperationError{
				{ID: "container1", Err: context.DeadlineExceeded},
			},
			hasError: true,
			errCount: 1,
		},
		{
			name: "multiple errors",
			errors: []OperationError{
				{ID: "container1", Err: context.DeadlineExceeded},
				{ID: "container2", Err: context.Canceled},
			},
			hasError: true,
			errCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			me := MultiError{Errors: tt.errors}
			if me.HasErrors() != tt.hasError {
				t.Errorf("HasErrors() = %v, want %v", me.HasErrors(), tt.hasError)
			}
			if len(me.Errors) != tt.errCount {
				t.Errorf("error count = %v, want %v", len(me.Errors), tt.errCount)
			}
			if tt.hasError {
				if me.ToError() == nil {
					t.Error("ToError() returned nil for errors")
				}
			} else {
				if me.ToError() != nil {
					t.Error("ToError() returned non-nil for no errors")
				}
			}
		})
	}
}
