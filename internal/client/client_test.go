package client

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestImagePruneFiltersIncludeAllUnused(t *testing.T) {
	args := imagePruneFilters()

	if !args.Match("dangling", "false") {
		t.Fatal("expected image prune filters to include dangling=false")
	}
}

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

func TestCreateTarArchiveIncludesBuildContextFiles(t *testing.T) {
	contextDir := t.TempDir()

	dockerfilePath := filepath.Join(contextDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte("FROM alpine\n"), 0o644); err != nil {
		t.Fatalf("failed to write Dockerfile: %v", err)
	}

	nestedDir := filepath.Join(contextDir, "app")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	nestedFilePath := filepath.Join(nestedDir, "main.txt")
	if err := os.WriteFile(nestedFilePath, []byte("hello from context\n"), 0o644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	tarReader, err := createTarArchive(contextDir)
	if err != nil {
		t.Fatalf("createTarArchive returned error: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := tarReader.Close(); closeErr != nil {
			t.Fatalf("failed to close tar reader: %v", closeErr)
		}
	})

	tarBytes, err := io.ReadAll(tarReader)
	if err != nil {
		t.Fatalf("failed to read tar stream: %v", err)
	}

	entries := map[string]string{}
	tarStream := tar.NewReader(bytes.NewReader(tarBytes))
	for {
		hdr, readErr := tarStream.Next()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			t.Fatalf("failed reading tar entry: %v", readErr)
		}

		if hdr.FileInfo().IsDir() {
			continue
		}

		body, readBodyErr := io.ReadAll(tarStream)
		if readBodyErr != nil {
			t.Fatalf("failed reading tar body for %s: %v", hdr.Name, readBodyErr)
		}
		entries[hdr.Name] = string(body)
	}

	if entries["Dockerfile"] != "FROM alpine\n" {
		t.Fatalf("Dockerfile content = %q, want %q", entries["Dockerfile"], "FROM alpine\n")
	}

	if entries["app/main.txt"] != "hello from context\n" {
		t.Fatalf("app/main.txt content = %q, want %q", entries["app/main.txt"], "hello from context\n")
	}

	for name := range entries {
		if strings.HasPrefix(name, "/") {
			t.Fatalf("tar entry %q is absolute path, expected relative path", name)
		}
	}
}

func TestResolveComposeFilePath(t *testing.T) {
	tests := []struct {
		name       string
		workingDir string
		input      string
		want       string
	}{
		{
			name:       "relative file joins with working dir",
			workingDir: "/tmp/project",
			input:      "compose.yml",
			want:       "/tmp/project/compose.yml",
		},
		{
			name:       "absolute file remains absolute",
			workingDir: "/tmp/project",
			input:      "/etc/compose.yml",
			want:       "/etc/compose.yml",
		},
		{
			name:       "empty compose file",
			workingDir: "/tmp/project",
			input:      "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveComposeFilePath(tt.workingDir, tt.input)
			if got != tt.want {
				t.Fatalf("resolveComposeFilePath(%q, %q) = %q, want %q", tt.workingDir, tt.input, got, tt.want)
			}
		})
	}
}
