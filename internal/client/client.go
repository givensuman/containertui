// Package client exposes a Docker client wrapper for managing containers.
package client

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/givensuman/containertui/internal/registry"
)

// ContainerStats represents the CPU and memory usage of a container.
type ContainerStats struct {
	CPUPercent float64
	MemUsage   float64
	MemLimit   float64
	NetRx      float64
	NetTx      float64
}

// Container represents a Docker container with essential details.
type Container struct {
	container.Config
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	Image string `json:"Image"`
	State string `json:"State"`
}

// Image represents a Docker image.
type Image struct {
	ID       string   `json:"Id"`
	RepoTags []string `json:"RepoTags"`
	Size     int64    `json:"Size"`
	Created  int64    `json:"Created"`
}

// Network represents a Docker network.
type Network struct {
	ID     string `json:"Id"`
	Name   string `json:"Name"`
	Driver string `json:"Driver"`
	Scope  string `json:"Scope"`
}

// Volume represents a Docker volume.
type Volume struct {
	Name       string `json:"Name"`
	Driver     string `json:"Driver"`
	Mountpoint string `json:"Mountpoint"`
}

// ClientWrapper wraps the Docker client to provide container management functionalities.
type ClientWrapper struct {
	client             *client.Client
	registryClient     *registry.Client
	quayRegistryClient *registry.QuayClient
}

func imagePruneFilters() filters.Args {
	args := filters.NewArgs()
	args.Add("dangling", "false")
	return args
}

// NewClient creates a new ClientWrapper with an initialized Docker client.
func NewClient() (*ClientWrapper, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &ClientWrapper{
		client:             dockerClient,
		registryClient:     registry.NewClient(),
		quayRegistryClient: registry.NewQuayClient(),
	}, nil
}

// CloseClient closes the Docker client connection.
func (clientWrapper *ClientWrapper) CloseClient() error {
	if err := clientWrapper.client.Close(); err != nil {
		return fmt.Errorf("failed to close Docker client: %w", err)
	}
	return nil
}

// GetContainers retrieves a list of all Docker containers.
func (clientWrapper *ClientWrapper) GetContainers(ctx context.Context) ([]Container, error) {
	listOptions := container.ListOptions{
		All: true,
	}

	containers, err := clientWrapper.client.ContainerList(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	dockerContainers := make([]Container, 0, len(containers))
	for _, containerItem := range containers {
		// Bounds checking for Names array
		name := ""
		if len(containerItem.Names) > 0 {
			name = containerItem.Names[0]
			// Remove leading slash if present
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		dockerContainers = append(dockerContainers, Container{
			ID:    containerItem.ID,
			Name:  name,
			Image: containerItem.Image,
			State: containerItem.State,
		})
	}

	return dockerContainers, nil
}

// GetImages retrieves a list of all Docker images.
func (clientWrapper *ClientWrapper) GetImages(ctx context.Context) ([]Image, error) {
	listOptions := types.ImageListOptions{
		All: true,
	}

	images, err := clientWrapper.client.ImageList(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	dockerImages := make([]Image, 0, len(images))
	for _, imageItem := range images {
		dockerImages = append(dockerImages, Image{
			ID:       imageItem.ID,
			RepoTags: imageItem.RepoTags,
			Size:     imageItem.Size,
			Created:  imageItem.Created,
		})
	}

	return dockerImages, nil
}

// GetNetworks retrieves a list of all Docker networks.
func (clientWrapper *ClientWrapper) GetNetworks(ctx context.Context) ([]Network, error) {
	listOptions := types.NetworkListOptions{}

	networks, err := clientWrapper.client.NetworkList(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	dockerNetworks := make([]Network, 0, len(networks))
	for _, networkItem := range networks {
		dockerNetworks = append(dockerNetworks, Network{
			ID:     networkItem.ID,
			Name:   networkItem.Name,
			Driver: networkItem.Driver,
			Scope:  networkItem.Scope,
		})
	}

	return dockerNetworks, nil
}

// GetVolumes retrieves a list of all Docker volumes.
func (clientWrapper *ClientWrapper) GetVolumes(ctx context.Context) ([]Volume, error) {
	listOptions := volume.ListOptions{}

	volumes, err := clientWrapper.client.VolumeList(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	dockerVolumes := make([]Volume, 0, len(volumes.Volumes))
	for _, volumeItem := range volumes.Volumes {
		dockerVolumes = append(dockerVolumes, Volume{
			Name:       volumeItem.Name,
			Driver:     volumeItem.Driver,
			Mountpoint: volumeItem.Mountpoint,
		})
	}

	return dockerVolumes, nil
}

// GetContainerState retrieves the current state of a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) GetContainerState(ctx context.Context, containerID string) (string, error) {
	inspectResponse, err := clientWrapper.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return "unknown", fmt.Errorf("failed to get state for container %s: %w", containerID, err)
	}

	return inspectResponse.State.Status, nil
}

// PauseContainer pauses a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) PauseContainer(ctx context.Context, containerID string) error {
	if err := clientWrapper.client.ContainerPause(ctx, containerID); err != nil {
		return fmt.Errorf("failed to pause container %s: %w", containerID, err)
	}
	return nil
}

// PauseContainers pauses multiple Docker containers by their IDs.
// Returns a MultiError containing all failures, or nil if all operations succeeded.
func (clientWrapper *ClientWrapper) PauseContainers(ctx context.Context, containerIDs []string) error {
	var multiErr MultiError
	for _, containerID := range containerIDs {
		if err := clientWrapper.PauseContainer(ctx, containerID); err != nil {
			multiErr.Add(containerID, err)
		}
	}
	return multiErr.ToError()
}

// CreateContainerConfig holds configuration for creating a container.
type CreateContainerConfig struct {
	Name       string
	ImageID    string
	Ports      map[string]string // "hostPort" -> "containerPort"
	Volumes    []string          // "hostPath:containerPath" format
	Env        []string          // "KEY=value" format
	AutoStart  bool
	AutoRemove bool
	Network    string // Network name (default: "bridge")
}

// CreateContainer creates a new container with the specified configuration.
func (clientWrapper *ClientWrapper) CreateContainer(ctx context.Context, config CreateContainerConfig) (containerID string, err error) {
	// Parse ports into nat.PortMap and nat.PortSet
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for hostPort, containerPort := range config.Ports {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", fmt.Errorf("invalid container port %s: %w", containerPort, err)
		}
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
	}

	// Create container config
	containerConfig := &container.Config{
		Image:        config.ImageID,
		Env:          config.Env,
		ExposedPorts: exposedPorts,
	}

	// Set network mode, default to bridge
	networkMode := config.Network
	if networkMode == "" {
		networkMode = "bridge"
	}

	// Create host config
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        config.Volumes,
		AutoRemove:   config.AutoRemove,
		NetworkMode:  container.NetworkMode(networkMode),
	}

	// Create the container
	resp, err := clientWrapper.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil, // NetworkingConfig
		nil, // Platform
		config.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Auto-start if requested
	if config.AutoStart {
		if err := clientWrapper.StartContainer(ctx, resp.ID); err != nil {
			return resp.ID, fmt.Errorf("container created but failed to start: %w", err)
		}
	}

	return resp.ID, nil
}

// UnpauseContainer unpauses a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) UnpauseContainer(ctx context.Context, containerID string) error {
	if err := clientWrapper.client.ContainerUnpause(ctx, containerID); err != nil {
		return fmt.Errorf("failed to unpause container %s: %w", containerID, err)
	}
	return nil
}

// UnpauseContainers unpauses multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) UnpauseContainers(ctx context.Context, containerIDs []string) error {
	var multiErr MultiError
	for _, containerID := range containerIDs {
		if err := clientWrapper.UnpauseContainer(ctx, containerID); err != nil {
			multiErr.Add(containerID, err)
		}
	}
	return multiErr.ToError()
}

// StartContainer starts a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) StartContainer(ctx context.Context, containerID string) error {
	if err := clientWrapper.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %w", containerID, err)
	}
	return nil
}

// StartContainers starts multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) StartContainers(ctx context.Context, containerIDs []string) error {
	var multiErr MultiError
	for _, containerID := range containerIDs {
		if err := clientWrapper.StartContainer(ctx, containerID); err != nil {
			multiErr.Add(containerID, err)
		}
	}
	return multiErr.ToError()
}

// StopContainer stops a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) StopContainer(ctx context.Context, containerID string) error {
	if err := clientWrapper.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}
	return nil
}

// StopContainers stops multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) StopContainers(ctx context.Context, containerIDs []string) error {
	var multiErr MultiError
	for _, containerID := range containerIDs {
		if err := clientWrapper.StopContainer(ctx, containerID); err != nil {
			multiErr.Add(containerID, err)
		}
	}
	return multiErr.ToError()
}

// RemoveContainer removes a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	removeOptions := container.RemoveOptions{
		Force: force,
	}

	if err := clientWrapper.client.ContainerRemove(ctx, containerID, removeOptions); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}
	return nil
}

// RemoveContainers removes multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) RemoveContainers(ctx context.Context, containerIDs []string, force bool) error {
	var multiErr MultiError
	for _, containerID := range containerIDs {
		if err := clientWrapper.RemoveContainer(ctx, containerID, force); err != nil {
			multiErr.Add(containerID, err)
		}
	}
	return multiErr.ToError()
}

// RestartContainer restarts a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) RestartContainer(ctx context.Context, containerID string) error {
	timeout := int(DefaultRestartTimeout.Seconds())
	if err := clientWrapper.client.ContainerRestart(
		ctx,
		containerID,
		container.StopOptions{Timeout: &timeout},
	); err != nil {
		return fmt.Errorf("failed to restart container %s: %w", containerID, err)
	}
	return nil
}

// RestartContainers restarts multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) RestartContainers(ctx context.Context, containerIDs []string) error {
	var multiErr MultiError
	for _, containerID := range containerIDs {
		if err := clientWrapper.RestartContainer(ctx, containerID); err != nil {
			multiErr.Add(containerID, err)
		}
	}
	return multiErr.ToError()
}

// Service represents a Docker Compose service.
type Service struct {
	Project     string
	Name        string
	Replicas    int
	Containers  []Container
	WorkingDir  string
	ComposeFile string
}

func resolveComposeFilePath(workingDir, composeFile string) string {
	trimmed := strings.TrimSpace(composeFile)
	if trimmed == "" {
		return ""
	}

	if filepath.IsAbs(trimmed) || strings.TrimSpace(workingDir) == "" {
		return trimmed
	}

	return filepath.Join(workingDir, trimmed)
}

// GetServices retrieves services based on docker-compose labels from containers.
func (clientWrapper *ClientWrapper) GetServices(ctx context.Context) ([]Service, error) {
	containers, err := clientWrapper.GetContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	servicesMap := make(map[string]*Service)

	for _, container := range containers {
		// We need to inspect to get labels
		details, err := clientWrapper.InspectContainer(ctx, container.ID)
		if err != nil {
			continue
		}

		projectName := details.Config.Labels["com.docker.compose.project"]
		serviceName := details.Config.Labels["com.docker.compose.service"]
		workingDir := details.Config.Labels["com.docker.compose.project.working_dir"]
		configFiles := details.Config.Labels["com.docker.compose.project.config_files"]

		if projectName != "" && serviceName != "" {
			key := projectName + "/" + serviceName
			if _, exists := servicesMap[key]; !exists {
				composeFile := ""
				if configFiles != "" {
					files := strings.Split(configFiles, ",")
					if len(files) > 0 {
						composeFile = resolveComposeFilePath(workingDir, files[0])
					}
				}
				if composeFile == "" && workingDir != "" {
					// Fallback to trying standard names in working dir
					possiblePaths := []string{
						fmt.Sprintf("%s/docker-compose.yml", workingDir),
						fmt.Sprintf("%s/docker-compose.yaml", workingDir),
						fmt.Sprintf("%s/compose.yml", workingDir),
						fmt.Sprintf("%s/compose.yaml", workingDir),
					}
					for _, p := range possiblePaths {
						if _, err := os.Stat(p); err == nil {
							composeFile = p
							break
						}
					}
				}

				servicesMap[key] = &Service{
					Project:     projectName,
					Name:        serviceName,
					Replicas:    0,
					Containers:  []Container{},
					WorkingDir:  workingDir,
					ComposeFile: composeFile,
				}
			}
			servicesMap[key].Replicas++
			servicesMap[key].Containers = append(servicesMap[key].Containers, container)
		}
	}

	var services []Service
	for _, s := range servicesMap {
		services = append(services, *s)
	}
	return services, nil
}

// RunComposeCommand runs a docker compose command for a project context.
func (clientWrapper *ClientWrapper) RunComposeCommand(ctx context.Context, workingDir, composeFile string, args ...string) error {
	commandArgs := []string{"compose"}
	if strings.TrimSpace(composeFile) != "" {
		commandArgs = append(commandArgs, "-f", composeFile)
	}
	commandArgs = append(commandArgs, args...)

	cmd := exec.CommandContext(ctx, "docker", commandArgs...)
	if strings.TrimSpace(workingDir) != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed == "" {
			return fmt.Errorf("failed to run docker %s: %w", strings.Join(commandArgs, " "), err)
		}
		return fmt.Errorf("failed to run docker %s: %w: %s", strings.Join(commandArgs, " "), err, trimmed)
	}

	return nil
}

// Logs represents the response from Moby's ContainerLogs.
type Logs io.ReadCloser

// OpenLogs streams logs from a Docker container.
func (clientWrapper *ClientWrapper) OpenLogs(ctx context.Context, containerID string) (Logs, error) {
	logsOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
	}

	reader, err := clientWrapper.client.ContainerLogs(ctx, containerID, logsOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to open logs for container %s: %w", containerID, err)
	}

	return reader, nil
}

// ExecShell starts an interactive shell (e.g., /bin/sh or /bin/bash) in the container with a TTY.
// Returns an io.ReadWriteCloser for bi-directional communication, or error.
func (clientWrapper *ClientWrapper) ExecShell(ctx context.Context, containerID string, shell []string) (io.ReadWriteCloser, error) {
	execCreateOptions := types.ExecConfig{
		Cmd:          shell,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execResp, err := clientWrapper.client.ContainerExecCreate(ctx, containerID, execCreateOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec for container %s: %w", containerID, err)
	}

	execAttachOptions := types.ExecStartCheck{
		Tty: true,
	}

	attachResp, err := clientWrapper.client.ContainerExecAttach(ctx, execResp.ID, execAttachOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec for container %s: %w", containerID, err)
	}

	return attachResp.Conn, nil // Attaches to socket, full duplex.
}

// RemoveImage removes a specific Docker image by its ID.
func (clientWrapper *ClientWrapper) RemoveImage(ctx context.Context, imageID string, force bool) error {
	options := types.ImageRemoveOptions{
		Force:         force,
		PruneChildren: true,
	}

	_, err := clientWrapper.client.ImageRemove(ctx, imageID, options)
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", imageID, err)
	}
	return nil
}

// RemoveVolume removes a specific Docker volume by its name.
func (clientWrapper *ClientWrapper) RemoveVolume(ctx context.Context, volumeName string, force bool) error {
	if err := clientWrapper.client.VolumeRemove(ctx, volumeName, force); err != nil {
		return fmt.Errorf("failed to remove volume %s: %w", volumeName, err)
	}
	return nil
}

// RemoveNetwork removes a specific Docker network by its ID.
func (clientWrapper *ClientWrapper) RemoveNetwork(ctx context.Context, networkID string) error {
	if err := clientWrapper.client.NetworkRemove(ctx, networkID); err != nil {
		return fmt.Errorf("failed to remove network %s: %w", networkID, err)
	}
	return nil
}

// PruneImages removes all unused images.
func (clientWrapper *ClientWrapper) PruneImages(ctx context.Context) (uint64, error) {
	report, err := clientWrapper.client.ImagesPrune(ctx, imagePruneFilters())
	if err != nil {
		return 0, fmt.Errorf("failed to prune images: %w", err)
	}
	return report.SpaceReclaimed, nil
}

// PruneVolumes removes all unused volumes.
func (clientWrapper *ClientWrapper) PruneVolumes(ctx context.Context) (uint64, error) {
	report, err := clientWrapper.client.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune volumes: %w", err)
	}
	return report.SpaceReclaimed, nil
}

// PruneNetworks removes all unused networks.
func (clientWrapper *ClientWrapper) PruneNetworks(ctx context.Context) (int, error) {
	report, err := clientWrapper.client.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune networks: %w", err)
	}
	return len(report.NetworksDeleted), nil
}

// PruneContainers removes all stopped containers.
func (clientWrapper *ClientWrapper) PruneContainers(ctx context.Context) (uint64, error) {
	report, err := clientWrapper.client.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune containers: %w", err)
	}
	return report.SpaceReclaimed, nil
}

// GetContainersUsingImage returns a list of container names that are using the specified image ID.
func (clientWrapper *ClientWrapper) GetContainersUsingImage(ctx context.Context, imageID string) ([]string, error) {
	containers, err := clientWrapper.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for image usage check: %w", err)
	}

	var usedBy []string
	for _, containerItem := range containers {
		if containerItem.ImageID == imageID {
			// Name usually comes with a slash, e.g., "/my-container".
			name := containerItem.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
			usedBy = append(usedBy, name)
		}
	}
	return usedBy, nil
}

// GetContainersUsingVolume returns a list of container names that are using the specified volume name.
func (clientWrapper *ClientWrapper) GetContainersUsingVolume(ctx context.Context, volumeName string) ([]string, error) {
	containers, err := clientWrapper.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for volume usage check: %w", err)
	}

	var usedBy []string
	for _, containerItem := range containers {
		for _, mount := range containerItem.Mounts {
			if mount.Name == volumeName || mount.Source == volumeName {
				name := containerItem.Names[0]
				if len(name) > 0 && name[0] == '/' {
					name = name[1:]
				}
				usedBy = append(usedBy, name)
				break // Found usage in this container, move to next container.
			}
		}
	}
	return usedBy, nil
}

// GetContainersUsingNetwork returns a list of container names that are attached to the specified network ID.
func (clientWrapper *ClientWrapper) GetContainersUsingNetwork(ctx context.Context, networkID string) ([]string, error) {
	containers, err := clientWrapper.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for network usage check: %w", err)
	}

	var usedBy []string
	for _, containerItem := range containers {
		if containerItem.NetworkSettings != nil {
			for _, network := range containerItem.NetworkSettings.Networks {
				if network.NetworkID == networkID {
					name := containerItem.Names[0]
					if len(name) > 0 && name[0] == '/' {
						name = name[1:]
					}
					usedBy = append(usedBy, name)
					break
				}
			}
		}
	}
	return usedBy, nil
}

// GetAllNetworkUsage returns a map of network IDs/names to a boolean indicating if they are in use.
// This is more efficient than calling GetContainersUsingNetwork for each network individually.
func (clientWrapper *ClientWrapper) GetAllNetworkUsage(ctx context.Context) (map[string]bool, error) {
	containers, err := clientWrapper.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for network usage check: %w", err)
	}

	networkUsage := make(map[string]bool)
	for _, containerItem := range containers {
		if containerItem.NetworkSettings != nil {
			for networkName, network := range containerItem.NetworkSettings.Networks {
				networkUsage[networkName] = true
				networkUsage[network.NetworkID] = true
			}
		}
	}
	return networkUsage, nil
}

// GetAllVolumeUsage returns a map of volume names to a boolean indicating if they are in use.
// This is more efficient than calling GetContainersUsingVolume for each volume individually.
func (clientWrapper *ClientWrapper) GetAllVolumeUsage(ctx context.Context) (map[string]bool, error) {
	containers, err := clientWrapper.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers for volume usage check: %w", err)
	}

	volumeUsage := make(map[string]bool)
	for _, containerItem := range containers {
		for _, mount := range containerItem.Mounts {
			if mount.Name != "" {
				volumeUsage[mount.Name] = true
			}
			if mount.Source != "" {
				volumeUsage[mount.Source] = true
			}
		}
	}
	return volumeUsage, nil
}

// GetContainerStats retrieves the current CPU and memory usage of a container.
func (clientWrapper *ClientWrapper) GetContainerStats(ctx context.Context, containerID string) (ContainerStats, error) {
	stats, err := clientWrapper.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return ContainerStats{}, fmt.Errorf("failed to get stats for container %s: %w", containerID, err)
	}
	defer func() {
		if closeErr := stats.Body.Close(); closeErr != nil {
			// Log the close error but don't override the return error
			fmt.Fprintf(os.Stderr, "warning: failed to close stats body: %v\n", closeErr)
		}
	}()

	var statsJSON types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return ContainerStats{}, fmt.Errorf("failed to decode stats for container %s: %w", containerID, err)
	}

	var cpuPercent float64
	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage) - float64(statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage) - float64(statsJSON.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(statsJSON.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	// Calculate memory usage.
	// MemUsage is statsJSON.MemoryStats.Usage - statsJSON.MemoryStats.Stats["cache"].
	var memUsage float64
	if statsJSON.MemoryStats.Usage > 0 {
		memUsage = float64(statsJSON.MemoryStats.Usage)
		if cache, ok := statsJSON.MemoryStats.Stats["cache"]; ok {
			memUsage -= float64(cache)
		}
	}

	// Calculate network I/O.
	var rx, tx float64
	for _, network := range statsJSON.Networks {
		rx += float64(network.RxBytes)
		tx += float64(network.TxBytes)
	}

	return ContainerStats{
		CPUPercent: cpuPercent,
		MemUsage:   memUsage,
		MemLimit:   float64(statsJSON.MemoryStats.Limit),
		NetRx:      rx,
		NetTx:      tx,
	}, nil
}

// InspectContainer returns the detailed inspection information for a container.
func (clientWrapper *ClientWrapper) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	inspect, err := clientWrapper.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerJSON{}, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}
	return inspect, nil
}

// GetRegistryClient returns the registry client for Docker Hub operations.
func (clientWrapper *ClientWrapper) GetRegistryClient() *registry.Client {
	return clientWrapper.registryClient
}

// GetQuayRegistryClient returns the registry client for Quay search operations.
func (clientWrapper *ClientWrapper) GetQuayRegistryClient() *registry.QuayClient {
	return clientWrapper.quayRegistryClient
}

// PullImageFromRegistry pulls an image from a registry (Docker Hub by default).
func (clientWrapper *ClientWrapper) PullImageFromRegistry(ctx context.Context, imageName string, progressChan chan<- string) error {
	reader, err := clientWrapper.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()
	if progressChan != nil {
		defer close(progressChan)
	}

	// Stream progress to the channel
	if progressChan != nil {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			progressChan <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read pull progress: %w", err)
		}
	}
	return nil
}

// CreateVolume creates a new Docker volume with the specified configuration.
func (clientWrapper *ClientWrapper) CreateVolume(ctx context.Context, name, driver string, labels map[string]string) (string, error) {
	volumeCreateBody := volume.CreateOptions{
		Name:   name,
		Driver: driver,
		Labels: labels,
	}

	vol, err := clientWrapper.client.VolumeCreate(ctx, volumeCreateBody)
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	return vol.Name, nil
}

// CreateNetwork creates a new Docker network with the specified configuration.
func (clientWrapper *ClientWrapper) CreateNetwork(ctx context.Context, name, driver, subnet, gateway string, ipv6 bool, labels map[string]string) (string, error) {
	networkCreate := types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         driver,
		EnableIPv6:     ipv6,
		Labels:         labels,
	}

	// Configure IPAM if subnet or gateway is provided
	if subnet != "" || gateway != "" {
		ipamConfig := network.IPAMConfig{
			Subnet:  subnet,
			Gateway: gateway,
		}
		networkCreate.IPAM = &network.IPAM{
			Config: []network.IPAMConfig{ipamConfig},
		}
	}

	response, err := clientWrapper.client.NetworkCreate(ctx, name, networkCreate)
	if err != nil {
		return "", fmt.Errorf("failed to create network: %w", err)
	}

	return response.ID, nil
}

// TagImage creates a new tag for an image.
func (clientWrapper *ClientWrapper) TagImage(ctx context.Context, imageID, newTag string) error {
	if err := clientWrapper.client.ImageTag(ctx, imageID, newTag); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}
	return nil
}

// RenameContainer renames a container.
func (clientWrapper *ClientWrapper) RenameContainer(ctx context.Context, containerID, newName string) error {
	if err := clientWrapper.client.ContainerRename(ctx, containerID, newName); err != nil {
		return fmt.Errorf("failed to rename container: %w", err)
	}
	return nil
}

// ImageHistory retrieves the history of an image (layers and commands).
func (clientWrapper *ClientWrapper) ImageHistory(ctx context.Context, imageID string) ([]image.HistoryResponseItem, error) {
	history, err := clientWrapper.client.ImageHistory(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image history: %w", err)
	}
	return history, nil
}

// BuildImage builds a Docker image from a Dockerfile.
func (clientWrapper *ClientWrapper) BuildImage(ctx context.Context, dockerfilePath, tag, contextPath string, buildArgs map[string]*string) (io.ReadCloser, error) {
	// Create tar archive of build context
	tarReader, err := createTarArchive(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create build context: %w", err)
	}

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{tag},
		Dockerfile: dockerfilePath,
		BuildArgs:  buildArgs,
		Remove:     true, // Remove intermediate containers
	}

	response, err := clientWrapper.client.ImageBuild(ctx, tarReader, buildOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}

	return response.Body, nil
}

// createTarArchive creates a tar archive of the build context directory
func createTarArchive(contextPath string) (io.ReadCloser, error) {
	cleanContextPath := filepath.Clean(contextPath)

	contextInfo, err := os.Stat(cleanContextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat build context: %w", err)
	}
	if !contextInfo.IsDir() {
		return nil, fmt.Errorf("build context must be a directory: %s", cleanContextPath)
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		tarWriter := tar.NewWriter(pw)
		defer tarWriter.Close()

		walkErr := filepath.Walk(cleanContextPath, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			relPath, err := filepath.Rel(cleanContextPath, path)
			if err != nil {
				return fmt.Errorf("failed to compute relative path for %s: %w", path, err)
			}
			if relPath == "." {
				return nil
			}

			archivePath := filepath.ToSlash(relPath)

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return fmt.Errorf("failed to create tar header for %s: %w", path, err)
			}
			header.Name = archivePath

			if info.IsDir() && !strings.HasSuffix(header.Name, "/") {
				header.Name += "/"
			}

			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header for %s: %w", path, err)
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("failed to write file %s to tar: %w", path, err)
			}

			return nil
		})

		if walkErr != nil {
			pw.CloseWithError(walkErr)
		}
	}()

	return pr, nil
}

// InspectImage retrieves detailed information about a Docker image.
func (clientWrapper *ClientWrapper) InspectImage(ctx context.Context, imageID string) (types.ImageInspect, error) {
	imageInfo, _, err := clientWrapper.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return types.ImageInspect{}, fmt.Errorf("failed to inspect image %s: %w", imageID, err)
	}
	return imageInfo, nil
}

// PullImage pulls a Docker image from a registry.
func (clientWrapper *ClientWrapper) PullImage(ctx context.Context, imageName string, progressChan chan<- string) error {
	reader, err := clientWrapper.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()
	if progressChan != nil {
		defer close(progressChan)
	}

	// Stream progress to the channel if provided
	if progressChan != nil {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			progressChan <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read pull progress: %w", err)
		}
	} else {
		// Still need to consume the reader even if no channel
		_, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read pull output: %w", err)
		}
	}
	return nil
}

// InspectNetwork retrieves detailed information about a Docker network.
func (clientWrapper *ClientWrapper) InspectNetwork(ctx context.Context, networkID string) (types.NetworkResource, error) {
	networkInfo, err := clientWrapper.client.NetworkInspect(ctx, networkID, types.NetworkInspectOptions{})
	if err != nil {
		return types.NetworkResource{}, fmt.Errorf("failed to inspect network %s: %w", networkID, err)
	}
	return networkInfo, nil
}

// InspectVolume retrieves detailed information about a Docker volume.
func (clientWrapper *ClientWrapper) InspectVolume(ctx context.Context, volumeName string) (volume.Volume, error) {
	volumeInfo, err := clientWrapper.client.VolumeInspect(ctx, volumeName)
	if err != nil {
		return volume.Volume{}, fmt.Errorf("failed to inspect volume %s: %w", volumeName, err)
	}
	return volumeInfo, nil
}
