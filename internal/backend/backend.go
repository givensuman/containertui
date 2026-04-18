// Package backend provides an abstraction layer for container runtime backends (Docker, Podman, etc).
package backend

import (
	"context"
	"io"
)

// Backend defines the interface that all container runtime backends must implement.
// This allows containertui to support multiple backends (Docker, Podman, etc.) with a unified API.
type Backend interface {
	// Metadata
	Name() string    // Returns the backend name (e.g., "docker", "podman")
	Version() string // Returns the backend version
	Close() error    // Closes the backend connection

	// Container operations
	ListContainers(ctx context.Context) ([]Container, error)
	InspectContainer(ctx context.Context, id string) (ContainerDetail, error)
	GetContainerState(ctx context.Context, id string) (string, error)
	CreateContainer(ctx context.Context, config ContainerConfig) (string, error)
	StartContainer(ctx context.Context, id string) error
	StartContainers(ctx context.Context, ids []string) error
	StopContainer(ctx context.Context, id string) error
	StopContainers(ctx context.Context, ids []string) error
	RestartContainer(ctx context.Context, id string) error
	RestartContainers(ctx context.Context, ids []string) error
	PauseContainer(ctx context.Context, id string) error
	PauseContainers(ctx context.Context, ids []string) error
	UnpauseContainer(ctx context.Context, id string) error
	UnpauseContainers(ctx context.Context, ids []string) error
	RemoveContainer(ctx context.Context, id string, force bool) error
	RemoveContainers(ctx context.Context, ids []string, force bool) error
	RenameContainer(ctx context.Context, id, newName string) error
	PruneContainers(ctx context.Context) (uint64, error)

	// Container logs and exec
	OpenLogs(ctx context.Context, id string) (Logs, error)
	ExecShell(ctx context.Context, id string, shell []string) (io.ReadWriteCloser, error)

	// Image operations
	ListImages(ctx context.Context) ([]Image, error)
	InspectImage(ctx context.Context, id string) (ImageDetail, error)
	PullImage(ctx context.Context, ref string, progressChan chan<- string) error
	BuildImage(ctx context.Context, dockerfilePath, tag, contextPath string, buildArgs map[string]*string) (io.ReadCloser, error)
	TagImage(ctx context.Context, source, target string) error
	RemoveImage(ctx context.Context, id string) error
	RemoveImages(ctx context.Context, ids []string) error
	PruneImages(ctx context.Context) (uint64, error)

	// Image history and usage
	ImageHistory(ctx context.Context, imageID string) ([]ImageHistoryItem, error)
	GetAllNetworkUsage(ctx context.Context) (map[string]bool, error)
	GetAllVolumeUsage(ctx context.Context) (map[string]bool, error)

	// Network operations
	ListNetworks(ctx context.Context) ([]Network, error)
	InspectNetwork(ctx context.Context, id string) (NetworkDetail, error)
	CreateNetwork(ctx context.Context, name, driver, subnet, gateway string, enableIPv6 bool, labels map[string]string) (string, error)
	RemoveNetwork(ctx context.Context, id string) error
	PruneNetworks(ctx context.Context) (int, error)

	// Volume operations
	ListVolumes(ctx context.Context) ([]Volume, error)
	InspectVolume(ctx context.Context, name string) (VolumeDetail, error)
	CreateVolume(ctx context.Context, name, driver string, labels map[string]string) (string, error)
	RemoveVolume(ctx context.Context, name string) error
	PruneVolumes(ctx context.Context) (uint64, error)

	// Service operations (Docker Compose, Podman pods, etc.)
	ListServices(ctx context.Context) ([]Service, error)

	// Dependency checking
	GetContainersUsingImage(ctx context.Context, imageID string) ([]string, error)
	GetContainersUsingVolume(ctx context.Context, volumeName string) ([]string, error)
	GetContainersUsingNetwork(ctx context.Context, networkID string) ([]string, error)
}
