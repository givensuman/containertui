// Package client exposes a Docker client wrapper for managing containers.
package client

import (
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
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
