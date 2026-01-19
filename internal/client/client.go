// Package client exposes a Docker client wrapper for managing containers.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
	client *client.Client
}

// NewClient creates a new ClientWrapper with an initialized Docker client.
func NewClient() (*ClientWrapper, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &ClientWrapper{client: dockerClient}, nil
}

// CloseClient closes the Docker client connection.
func (clientWrapper *ClientWrapper) CloseClient() error {
	return clientWrapper.client.Close()
}

// GetContainers retrieves a list of all Docker containers.
func (clientWrapper *ClientWrapper) GetContainers() ([]Container, error) {
	listOptions := container.ListOptions{
		All: true,
	}

	containers, err := clientWrapper.client.ContainerList(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}

	dockerContainers := make([]Container, 0, len(containers))
	for _, containerItem := range containers {
		dockerContainers = append(dockerContainers, Container{
			ID:    containerItem.ID,
			Name:  containerItem.Names[0][1:],
			Image: containerItem.Image,
			State: containerItem.State,
		})
	}

	return dockerContainers, nil
}

// GetImages retrieves a list of all Docker images.
func (clientWrapper *ClientWrapper) GetImages() ([]Image, error) {
	listOptions := types.ImageListOptions{
		All: true,
	}

	images, err := clientWrapper.client.ImageList(context.Background(), listOptions)
	if err != nil {
		return nil, err
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
func (clientWrapper *ClientWrapper) GetNetworks() ([]Network, error) {
	listOptions := types.NetworkListOptions{}

	networks, err := clientWrapper.client.NetworkList(context.Background(), listOptions)
	if err != nil {
		return nil, err
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
func (clientWrapper *ClientWrapper) GetVolumes() ([]Volume, error) {
	listOptions := volume.ListOptions{}

	volumes, err := clientWrapper.client.VolumeList(context.Background(), listOptions)
	if err != nil {
		return nil, err
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
func (clientWrapper *ClientWrapper) GetContainerState(containerID string) (string, error) {
	inspectResponse, err := clientWrapper.client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return "unknown", err
	}

	return inspectResponse.State.Status, nil
}

// PauseContainer pauses a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) PauseContainer(containerID string) error {
	return clientWrapper.client.ContainerPause(context.Background(), containerID)
}

// PauseContainers pauses multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) PauseContainers(containerIDs []string) error {
	for _, containerID := range containerIDs {
		if err := clientWrapper.PauseContainer(containerID); err != nil {
			return err
		}
	}

	return nil
}

// CreateContainerConfig holds configuration for creating a container.
type CreateContainerConfig struct {
	Name      string
	ImageID   string
	Ports     map[string]string // "hostPort" -> "containerPort"
	Volumes   []string          // "hostPath:containerPath" format
	Env       []string          // "KEY=value" format
	AutoStart bool
	Network   string // Network name (default: "bridge")
}

// CreateContainer creates a new container with the specified configuration.
func (clientWrapper *ClientWrapper) CreateContainer(config CreateContainerConfig) (containerID string, err error) {
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
		NetworkMode:  container.NetworkMode(networkMode),
	}

	// Create the container
	resp, err := clientWrapper.client.ContainerCreate(
		context.Background(),
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
		if err := clientWrapper.StartContainer(resp.ID); err != nil {
			return resp.ID, fmt.Errorf("container created but failed to start: %w", err)
		}
	}

	return resp.ID, nil
}

// UnpauseContainer unpauses a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) UnpauseContainer(containerID string) error {
	return clientWrapper.client.ContainerUnpause(context.Background(), containerID)
}

// UnpauseContainers unpauses multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) UnpauseContainers(containerIDs []string) error {
	for _, containerID := range containerIDs {
		if err := clientWrapper.UnpauseContainer(containerID); err != nil {
			return err
		}
	}

	return nil
}

// StartContainer starts a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) StartContainer(containerID string) error {
	return clientWrapper.client.ContainerStart(context.Background(), containerID, container.StartOptions{})
}

// StartContainers starts multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) StartContainers(containerIDs []string) error {
	for _, containerID := range containerIDs {
		if err := clientWrapper.StartContainer(containerID); err != nil {
			return err
		}
	}

	return nil
}

// StopContainer stops a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) StopContainer(containerID string) error {
	return clientWrapper.client.ContainerStop(context.Background(), containerID, container.StopOptions{})
}

// StopContainers stops multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) StopContainers(containerIDs []string) error {
	for _, containerID := range containerIDs {
		if err := clientWrapper.StopContainer(containerID); err != nil {
			return err
		}
	}

	return nil
}

// RemoveContainer removes a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) RemoveContainer(containerID string) error {
	removeOptions := container.RemoveOptions{
		Force: true,
	}

	return clientWrapper.client.ContainerRemove(context.Background(), containerID, removeOptions)
}

// RemoveContainers removes multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) RemoveContainers(containerIDs []string) error {
	for _, containerID := range containerIDs {
		if err := clientWrapper.RemoveContainer(containerID); err != nil {
			return err
		}
	}

	return nil
}

// RestartContainer restarts a specific Docker container by its ID.
func (clientWrapper *ClientWrapper) RestartContainer(containerID string) error {
	timeout := 10 // seconds
	return clientWrapper.client.ContainerRestart(
		context.Background(),
		containerID,
		container.StopOptions{Timeout: &timeout},
	)
}

// RestartContainers restarts multiple Docker containers by their IDs.
func (clientWrapper *ClientWrapper) RestartContainers(containerIDs []string) error {
	for _, containerID := range containerIDs {
		if err := clientWrapper.RestartContainer(containerID); err != nil {
			return err
		}
	}

	return nil
}

// Service represents a Docker Compose service.
type Service struct {
	Name        string
	Replicas    int
	Containers  []Container
	ComposeFile string
}

// GetServices retrieves services based on docker-compose labels from containers.
func (clientWrapper *ClientWrapper) GetServices() ([]Service, error) {
	containers, err := clientWrapper.GetContainers()
	if err != nil {
		return nil, err
	}

	servicesMap := make(map[string]*Service)

	for _, container := range containers {
		// We need to inspect to get labels
		details, err := clientWrapper.InspectContainer(container.ID)
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
						// The label might contain multiple files, we take the first one?
						// Or check workingDir?
						// Often config_files is absolute path.
						composeFile = files[0]
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
					Name:        serviceName,
					Replicas:    0,
					Containers:  []Container{},
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

// Logs represents the response from Moby's ContainerLogs.
type Logs io.ReadCloser

// OpenLogs streams logs from a Docker container.
func (clientWrapper *ClientWrapper) OpenLogs(containerID string) (Logs, error) {
	logsOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
	}

	reader, err := clientWrapper.client.ContainerLogs(context.Background(), containerID, logsOptions)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// ExecShell starts an interactive shell (e.g., /bin/sh or /bin/bash) in the container with a TTY.
// Returns an io.ReadWriteCloser for bi-directional communication, or error.
func (clientWrapper *ClientWrapper) ExecShell(containerID string, shell []string) (io.ReadWriteCloser, error) {
	execCreateOptions := types.ExecConfig{
		Cmd:          shell,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execResp, err := clientWrapper.client.ContainerExecCreate(context.Background(), containerID, execCreateOptions)
	if err != nil {
		return nil, err
	}

	execAttachOptions := types.ExecStartCheck{
		Tty: true,
	}

	attachResp, err := clientWrapper.client.ContainerExecAttach(context.Background(), execResp.ID, execAttachOptions)
	if err != nil {
		return nil, err
	}

	return attachResp.Conn, nil // Attaches to socket, full duplex.
}

// RemoveImage removes a specific Docker image by its ID.
func (clientWrapper *ClientWrapper) RemoveImage(imageID string) error {
	options := types.ImageRemoveOptions{
		Force:         false,
		PruneChildren: true,
	}

	_, err := clientWrapper.client.ImageRemove(context.Background(), imageID, options)
	return err
}

// RemoveVolume removes a specific Docker volume by its name.
func (clientWrapper *ClientWrapper) RemoveVolume(volumeName string) error {
	return clientWrapper.client.VolumeRemove(context.Background(), volumeName, false)
}

// RemoveNetwork removes a specific Docker network by its ID.
func (clientWrapper *ClientWrapper) RemoveNetwork(networkID string) error {
	return clientWrapper.client.NetworkRemove(context.Background(), networkID)
}

// PruneImages removes all unused images.
func (clientWrapper *ClientWrapper) PruneImages() (uint64, error) {
	report, err := clientWrapper.client.ImagesPrune(context.Background(), filters.Args{})
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

// PruneVolumes removes all unused volumes.
func (clientWrapper *ClientWrapper) PruneVolumes() (uint64, error) {
	report, err := clientWrapper.client.VolumesPrune(context.Background(), filters.Args{})
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

// PruneNetworks removes all unused networks.
func (clientWrapper *ClientWrapper) PruneNetworks() error {
	_, err := clientWrapper.client.NetworksPrune(context.Background(), filters.Args{})
	return err
}

// GetContainersUsingImage returns a list of container names that are using the specified image ID.
func (clientWrapper *ClientWrapper) GetContainersUsingImage(imageID string) ([]string, error) {
	containers, err := clientWrapper.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
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
func (clientWrapper *ClientWrapper) GetContainersUsingVolume(volumeName string) ([]string, error) {
	containers, err := clientWrapper.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
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
func (clientWrapper *ClientWrapper) GetContainersUsingNetwork(networkID string) ([]string, error) {
	containers, err := clientWrapper.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
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

// GetContainerStats retrieves the current CPU and memory usage of a container.
func (clientWrapper *ClientWrapper) GetContainerStats(containerID string) (ContainerStats, error) {
	stats, err := clientWrapper.client.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return ContainerStats{}, err
	}
	defer func() {
		_ = stats.Body.Close()
	}()

	var statsJSON types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return ContainerStats{}, err
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
func (clientWrapper *ClientWrapper) InspectContainer(containerID string) (types.ContainerJSON, error) {
	return clientWrapper.client.ContainerInspect(context.Background(), containerID)
}

// PullImage pulls an image from a registry.
// If progressChan is provided, progress messages will be sent to it.
// The channel will be closed when the pull completes or fails.
func (clientWrapper *ClientWrapper) PullImage(imageName string, progressChan chan<- string) error {
	out, err := clientWrapper.client.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		if progressChan != nil {
			close(progressChan)
		}
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()

	// Read progress JSON stream
	decoder := json.NewDecoder(out)
	for {
		var event struct {
			Status   string `json:"status"`
			Progress string `json:"progress"`
			ID       string `json:"id"`
			Error    string `json:"error"`
		}

		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			if progressChan != nil {
				close(progressChan)
			}
			return fmt.Errorf("error reading pull progress: %w", err)
		}

		// Check for error in response
		if event.Error != "" {
			if progressChan != nil {
				close(progressChan)
			}
			return fmt.Errorf("pull failed: %s", event.Error)
		}

		if progressChan != nil {
			progressMsg := event.Status
			if event.Progress != "" {
				progressMsg += ": " + event.Progress
			}
			progressChan <- progressMsg
		}
	}

	if progressChan != nil {
		close(progressChan)
	}

	return nil
}
