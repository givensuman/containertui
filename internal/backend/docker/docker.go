// Package docker provides a Docker backend implementation.
package docker

import (
	"archive/tar"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/givensuman/containertui/internal/backend"
)

// DockerBackend implements the Backend interface for Docker.
type DockerBackend struct {
	client *client.Client
}

// New creates a new Docker backend.
func New() (*DockerBackend, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &DockerBackend{client: cli}, nil
}

// Name returns the backend name.
func (d *DockerBackend) Name() string {
	return "docker"
}

// Version returns the Docker version.
func (d *DockerBackend) Version() string {
	ctx := context.Background()
	version, err := d.client.ServerVersion(ctx)
	if err != nil {
		return "unknown"
	}
	return version.Version
}

// Close closes the Docker client connection.
func (d *DockerBackend) Close() error {
	if err := d.client.Close(); err != nil {
		return fmt.Errorf("failed to close Docker client: %w", err)
	}
	return nil
}

// ListContainers lists all containers.
func (d *DockerBackend) ListContainers(ctx context.Context) ([]backend.Container, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]backend.Container, len(containers))
	for i, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		// Parse created time - Docker API uses int64 Unix timestamp for list
		createdTime := time.Unix(c.Created, 0).UTC()

		result[i] = backend.Container{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: createdTime,
		}
	}
	return result, nil
}

// InspectContainer inspects a container.
func (d *DockerBackend) InspectContainer(ctx context.Context, id string) (backend.ContainerDetail, error) {
	c, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return backend.ContainerDetail{}, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Parse created time - Docker API uses RFC3339 string for inspect
	createdTime, _ := time.Parse(time.RFC3339Nano, c.Created)

	// Convert Docker types to backend types
	detail := backend.ContainerDetail{
		Container: backend.Container{
			ID:      c.ID,
			Name:    c.Name,
			Image:   c.Image,
			State:   c.State.Status,
			Status:  c.State.Status,
			Created: createdTime,
		},
		Config: backend.ContainerConfigDetail{
			Hostname:     c.Config.Hostname,
			Domainname:   c.Config.Domainname,
			User:         c.Config.User,
			AttachStdin:  c.Config.AttachStdin,
			AttachStdout: c.Config.AttachStdout,
			AttachStderr: c.Config.AttachStderr,
			Tty:          c.Config.Tty,
			OpenStdin:    c.Config.OpenStdin,
			StdinOnce:    c.Config.StdinOnce,
			Env:          c.Config.Env,
			Cmd:          c.Config.Cmd,
			Image:        c.Config.Image,
			Volumes:      c.Config.Volumes,
			WorkingDir:   c.Config.WorkingDir,
			Entrypoint:   c.Config.Entrypoint,
			Labels:       c.Config.Labels,
			ExposedPorts: convertExposedPorts(c.Config.ExposedPorts),
		},
		HostConfig: convertHostConfig(c.HostConfig),
		NetworkSettings: backend.NetworkSettings{
			Networks: convertNetworks(c.NetworkSettings.Networks),
			Ports:    convertPortMap(c.NetworkSettings.Ports),
		},
		Mounts: convertMounts(c.Mounts),
		Raw:    c,
	}

	return detail, nil
}

// GetContainerState returns the state of a container.
func (d *DockerBackend) GetContainerState(ctx context.Context, id string) (string, error) {
	c, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get container state: %w", err)
	}
	return c.State.Status, nil
}

// CreateContainer creates a new container.
func (d *DockerBackend) CreateContainer(ctx context.Context, config backend.ContainerConfig) (string, error) {
	// Convert backend config to Docker config
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for hostPort, containerPort := range config.Ports {
		natPort, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", fmt.Errorf("invalid port: %w", err)
		}
		portBindings[natPort] = []nat.PortBinding{{HostPort: hostPort}}
		exposedPorts[natPort] = struct{}{}
	}

	containerConfig := &container.Config{
		Image:        config.Image,
		Env:          config.Env,
		ExposedPorts: exposedPorts,
	}

	hostConfig := &container.HostConfig{
		Binds:        config.Volumes,
		PortBindings: portBindings,
		AutoRemove:   config.AutoRemove,
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
	}
	if config.AutoStart {
		hostConfig.RestartPolicy.Name = "always"
	}

	networkConfig := &network.NetworkingConfig{}
	if config.Network != "" {
		networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			config.Network: {},
		}
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, config.Name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}
	return resp.ID, nil
}

// StartContainer starts a container.
func (d *DockerBackend) StartContainer(ctx context.Context, id string) error {
	if err := d.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

// StartContainers starts multiple containers.
func (d *DockerBackend) StartContainers(ctx context.Context, ids []string) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.StartContainer(ctx, id); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// StopContainer stops a container.
func (d *DockerBackend) StopContainer(ctx context.Context, id string) error {
	timeout := 10
	if err := d.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

// StopContainers stops multiple containers.
func (d *DockerBackend) StopContainers(ctx context.Context, ids []string) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.StopContainer(ctx, id); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// RestartContainer restarts a container.
func (d *DockerBackend) RestartContainer(ctx context.Context, id string) error {
	timeout := 10
	if err := d.client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}
	return nil
}

// RestartContainers restarts multiple containers.
func (d *DockerBackend) RestartContainers(ctx context.Context, ids []string) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.RestartContainer(ctx, id); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// PauseContainer pauses a container.
func (d *DockerBackend) PauseContainer(ctx context.Context, id string) error {
	if err := d.client.ContainerPause(ctx, id); err != nil {
		return fmt.Errorf("failed to pause container: %w", err)
	}
	return nil
}

// PauseContainers pauses multiple containers.
func (d *DockerBackend) PauseContainers(ctx context.Context, ids []string) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.PauseContainer(ctx, id); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// UnpauseContainer unpauses a container.
func (d *DockerBackend) UnpauseContainer(ctx context.Context, id string) error {
	if err := d.client.ContainerUnpause(ctx, id); err != nil {
		return fmt.Errorf("failed to unpause container: %w", err)
	}
	return nil
}

// UnpauseContainers unpauses multiple containers.
func (d *DockerBackend) UnpauseContainers(ctx context.Context, ids []string) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.UnpauseContainer(ctx, id); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// RemoveContainer removes a container.
func (d *DockerBackend) RemoveContainer(ctx context.Context, id string, force bool) error {
	if err := d.client.ContainerRemove(ctx, id, container.RemoveOptions{Force: force}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

// RemoveContainers removes multiple containers.
func (d *DockerBackend) RemoveContainers(ctx context.Context, ids []string, force bool) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.RemoveContainer(ctx, id, force); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// RenameContainer renames a container.
func (d *DockerBackend) RenameContainer(ctx context.Context, id, newName string) error {
	if err := d.client.ContainerRename(ctx, id, newName); err != nil {
		return fmt.Errorf("failed to rename container: %w", err)
	}
	return nil
}

// PruneContainers removes all stopped containers.
func (d *DockerBackend) PruneContainers(ctx context.Context) (uint64, error) {
	report, err := d.client.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune containers: %w", err)
	}
	return report.SpaceReclaimed, nil
}

// OpenLogs opens a log stream for a container.
func (d *DockerBackend) OpenLogs(ctx context.Context, id string) (backend.Logs, error) {
	logs, err := d.client.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	})
	if err != nil {
		return backend.Logs{}, fmt.Errorf("failed to open logs: %w", err)
	}

	return backend.Logs{
		Stream: logs,
		Close:  logs.Close,
	}, nil
}

// ExecShell executes a shell in a container.
func (d *DockerBackend) ExecShell(ctx context.Context, id string, shell []string) (io.ReadWriteCloser, error) {
	execConfig := types.ExecConfig{
		Cmd:          shell,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execIDResp, err := d.client.ContainerExecCreate(ctx, id, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := d.client.ContainerExecAttach(ctx, execIDResp.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec: %w", err)
	}

	return resp.Conn, nil
}

// ListImages lists all images.
func (d *DockerBackend) ListImages(ctx context.Context) ([]backend.Image, error) {
	images, err := d.client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	result := make([]backend.Image, len(images))
	for i, img := range images {
		result[i] = backend.Image{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Size:        img.Size,
			Created:     time.Unix(img.Created, 0),
		}
	}
	return result, nil
}

// InspectImage inspects an image.
func (d *DockerBackend) InspectImage(ctx context.Context, id string) (backend.ImageDetail, error) {
	img, _, err := d.client.ImageInspectWithRaw(ctx, id)
	if err != nil {
		return backend.ImageDetail{}, fmt.Errorf("failed to inspect image: %w", err)
	}

	createdTime, _ := time.Parse(time.RFC3339Nano, img.Created)

	detail := backend.ImageDetail{
		Image: backend.Image{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Size:        img.Size,
			Created:     createdTime,
		},
		Author:       img.Author,
		Comment:      img.Comment,
		Architecture: img.Architecture,
		Os:           img.Os,
		Size:         img.Size,
		VirtualSize:  img.VirtualSize,
		RootFS: backend.RootFS{
			Type:   img.RootFS.Type,
			Layers: img.RootFS.Layers,
		},
		Raw: img,
	}

	if img.Config != nil {
		detail.Config = backend.ContainerConfigDetail{
			Hostname:     img.Config.Hostname,
			Domainname:   img.Config.Domainname,
			User:         img.Config.User,
			AttachStdin:  img.Config.AttachStdin,
			AttachStdout: img.Config.AttachStdout,
			AttachStderr: img.Config.AttachStderr,
			Tty:          img.Config.Tty,
			OpenStdin:    img.Config.OpenStdin,
			StdinOnce:    img.Config.StdinOnce,
			Env:          img.Config.Env,
			Cmd:          img.Config.Cmd,
			Image:        img.Config.Image,
			Volumes:      img.Config.Volumes,
			WorkingDir:   img.Config.WorkingDir,
			Entrypoint:   img.Config.Entrypoint,
			Labels:       img.Config.Labels,
			ExposedPorts: convertExposedPorts(img.Config.ExposedPorts),
		}
	}

	return detail, nil
}

// PullImage pulls an image from a registry, sending progress lines to progressChan.
func (d *DockerBackend) PullImage(ctx context.Context, ref string, progressChan chan<- string) error {
	resp, err := d.client.ImagePull(ctx, ref, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		progressChan <- scanner.Text()
	}
	close(progressChan)

	return scanner.Err()
}

// BuildImage builds an image from a Dockerfile.
func (d *DockerBackend) BuildImage(ctx context.Context, dockerfilePath, tag, contextPath string, buildArgs map[string]*string) (io.ReadCloser, error) {
	tarReader, err := createTarArchive(contextPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create build context: %w", err)
	}

	buildOptions := types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       []string{tag},
		BuildArgs:  buildArgs,
		Remove:     true,
	}

	resp, err := d.client.ImageBuild(ctx, tarReader, buildOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}

	return resp.Body, nil
}

// TagImage tags an image.
func (d *DockerBackend) TagImage(ctx context.Context, source, target string) error {
	if err := d.client.ImageTag(ctx, source, target); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}
	return nil
}

// RemoveImage removes an image.
func (d *DockerBackend) RemoveImage(ctx context.Context, id string) error {
	_, err := d.client.ImageRemove(ctx, id, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}
	return nil
}

// RemoveImages removes multiple images.
func (d *DockerBackend) RemoveImages(ctx context.Context, ids []string) error {
	merr := &multiError{}
	for _, id := range ids {
		if err := d.RemoveImage(ctx, id); err != nil {
			merr.Add(err)
		}
	}
	return merr.ToError()
}

// PruneImages removes unused images.
func (d *DockerBackend) PruneImages(ctx context.Context) (uint64, error) {
	report, err := d.client.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune images: %w", err)
	}
	return report.SpaceReclaimed, nil
}

// ListNetworks lists all networks.
func (d *DockerBackend) ListNetworks(ctx context.Context) ([]backend.Network, error) {
	networks, err := d.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	result := make([]backend.Network, len(networks))
	for i, net := range networks {
		result[i] = backend.Network{
			ID:     net.ID,
			Name:   net.Name,
			Driver: net.Driver,
			Scope:  net.Scope,
		}
	}
	return result, nil
}

// InspectNetwork inspects a network.
func (d *DockerBackend) InspectNetwork(ctx context.Context, id string) (backend.NetworkDetail, error) {
	net, err := d.client.NetworkInspect(ctx, id, types.NetworkInspectOptions{})
	if err != nil {
		return backend.NetworkDetail{}, fmt.Errorf("failed to inspect network: %w", err)
	}

	detail := backend.NetworkDetail{
		Network: backend.Network{
			ID:     net.ID,
			Name:   net.Name,
			Driver: net.Driver,
			Scope:  net.Scope,
		},
		EnableIPv6: net.EnableIPv6,
		IPAM: backend.IPAM{
			Driver:  net.IPAM.Driver,
			Options: net.IPAM.Options,
			Config:  convertIPAMConfig(net.IPAM.Config),
		},
		Internal:   net.Internal,
		Attachable: net.Attachable,
		Ingress:    net.Ingress,
		ConfigOnly: net.ConfigOnly,
		Containers: convertEndpointResources(net.Containers),
		Options:    net.Options,
		Labels:     net.Labels,
		Raw:        net,
	}

	if net.ConfigFrom.Network != "" {
		detail.ConfigFrom = backend.ConfigReference{
			Network: net.ConfigFrom.Network,
		}
	}

	return detail, nil
}

// CreateNetwork creates a new network.
func (d *DockerBackend) CreateNetwork(ctx context.Context, name, driver, subnet, gateway string, enableIPv6 bool, labels map[string]string) (string, error) {
	ipamConfig := []network.IPAMConfig{}
	if subnet != "" {
		config := network.IPAMConfig{
			Subnet: subnet,
		}
		if gateway != "" {
			config.Gateway = gateway
		}
		ipamConfig = append(ipamConfig, config)
	}

	options := types.NetworkCreate{
		Driver:     driver,
		EnableIPv6: enableIPv6,
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
		Labels: labels,
	}

	resp, err := d.client.NetworkCreate(ctx, name, options)
	if err != nil {
		return "", fmt.Errorf("failed to create network: %w", err)
	}
	return resp.ID, nil
}

// RemoveNetwork removes a network.
func (d *DockerBackend) RemoveNetwork(ctx context.Context, id string) error {
	if err := d.client.NetworkRemove(ctx, id); err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}
	return nil
}

// PruneNetworks removes unused networks.
func (d *DockerBackend) PruneNetworks(ctx context.Context) (int, error) {
	report, err := d.client.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune networks: %w", err)
	}
	return len(report.NetworksDeleted), nil
}

// ConnectContainerToNetwork attaches a container to a network.
func (d *DockerBackend) ConnectContainerToNetwork(ctx context.Context, containerID, networkID string) error {
	if err := d.client.NetworkConnect(ctx, networkID, containerID, nil); err != nil {
		return fmt.Errorf("failed to connect container to network: %w", err)
	}
	return nil
}

// DisconnectContainerFromNetwork detaches a container from a network.
func (d *DockerBackend) DisconnectContainerFromNetwork(ctx context.Context, containerID, networkID string, force bool) error {
	if err := d.client.NetworkDisconnect(ctx, networkID, containerID, force); err != nil {
		return fmt.Errorf("failed to disconnect container from network: %w", err)
	}
	return nil
}

// ListVolumes lists all volumes.
func (d *DockerBackend) ListVolumes(ctx context.Context) ([]backend.Volume, error) {
	vols, err := d.client.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	result := make([]backend.Volume, len(vols.Volumes))
	for i, vol := range vols.Volumes {
		createdAt := time.Time{}
		if vol.CreatedAt != "" {
			createdAt, _ = time.Parse(time.RFC3339, vol.CreatedAt)
		}

		result[i] = backend.Volume{
			Name:       vol.Name,
			Driver:     vol.Driver,
			Mountpoint: vol.Mountpoint,
			CreatedAt:  createdAt,
		}
	}
	return result, nil
}

// InspectVolume inspects a volume.
func (d *DockerBackend) InspectVolume(ctx context.Context, name string) (backend.VolumeDetail, error) {
	vol, err := d.client.VolumeInspect(ctx, name)
	if err != nil {
		return backend.VolumeDetail{}, fmt.Errorf("failed to inspect volume: %w", err)
	}

	createdAt := time.Time{}
	if vol.CreatedAt != "" {
		createdAt, _ = time.Parse(time.RFC3339, vol.CreatedAt)
	}

	detail := backend.VolumeDetail{
		Volume: backend.Volume{
			Name:       vol.Name,
			Driver:     vol.Driver,
			Mountpoint: vol.Mountpoint,
			CreatedAt:  createdAt,
		},
		Labels:  vol.Labels,
		Scope:   vol.Scope,
		Options: vol.Options,
		Raw:     vol,
	}

	return detail, nil
}

// CreateVolume creates a new volume.
func (d *DockerBackend) CreateVolume(ctx context.Context, name, driver string, labels map[string]string) (string, error) {
	vol, err := d.client.VolumeCreate(ctx, volume.CreateOptions{
		Name:   name,
		Driver: driver,
		Labels: labels,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}
	return vol.Name, nil
}

// RemoveVolume removes a volume.
func (d *DockerBackend) RemoveVolume(ctx context.Context, name string) error {
	if err := d.client.VolumeRemove(ctx, name, true); err != nil {
		return fmt.Errorf("failed to remove volume: %w", err)
	}
	return nil
}

// PruneVolumes removes unused volumes.
func (d *DockerBackend) PruneVolumes(ctx context.Context) (uint64, error) {
	report, err := d.client.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return 0, fmt.Errorf("failed to prune volumes: %w", err)
	}
	return report.SpaceReclaimed, nil
}

// ListServices lists all services (Docker Compose services).
func (d *DockerBackend) ListServices(ctx context.Context) ([]backend.Service, error) {
	// For Docker, we could implement this by querying Docker Compose projects
	// For now, return empty list as this is optional functionality
	return []backend.Service{}, nil
}

// GetContainersUsingImage returns containers using an image.
func (d *DockerBackend) GetContainersUsingImage(ctx context.Context, imageID string) ([]string, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []string
	for _, c := range containers {
		if c.ImageID == imageID || c.Image == imageID {
			name := ""
			if len(c.Names) > 0 {
				name = c.Names[0]
				if len(name) > 0 && name[0] == '/' {
					name = name[1:]
				}
			}
			result = append(result, name)
		}
	}
	return result, nil
}

// GetContainersUsingVolume returns containers using a volume.
func (d *DockerBackend) GetContainersUsingVolume(ctx context.Context, volumeName string) ([]string, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []string
	for _, c := range containers {
		for _, mount := range c.Mounts {
			if mount.Type == "volume" && mount.Name == volumeName {
				name := ""
				if len(c.Names) > 0 {
					name = c.Names[0]
					if len(name) > 0 && name[0] == '/' {
						name = name[1:]
					}
				}
				result = append(result, name)
				break
			}
		}
	}
	return result, nil
}

// GetContainersUsingNetwork returns containers using a network.
func (d *DockerBackend) GetContainersUsingNetwork(ctx context.Context, networkID string) ([]string, error) {
	net, err := d.client.NetworkInspect(ctx, networkID, types.NetworkInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to inspect network: %w", err)
	}

	result := make([]string, 0, len(net.Containers))
	for _, container := range net.Containers {
		result = append(result, container.Name)
	}
	return result, nil
}

// ImageHistory returns the history of an image.
func (d *DockerBackend) ImageHistory(ctx context.Context, imageID string) ([]backend.ImageHistoryItem, error) {
	history, err := d.client.ImageHistory(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image history: %w", err)
	}

	result := make([]backend.ImageHistoryItem, len(history))
	for i, item := range history {
		result[i] = backend.ImageHistoryItem{
			ID:        item.ID,
			Created:   time.Unix(item.Created, 0),
			CreatedBy: item.CreatedBy,
			Tags:      item.Tags,
			Size:      item.Size,
			Comment:   item.Comment,
		}
	}
	return result, nil
}

// GetAllNetworkUsage returns a map of network IDs that are in use by any container.
func (d *DockerBackend) GetAllNetworkUsage(ctx context.Context) (map[string]bool, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make(map[string]bool)
	for _, c := range containers {
		for _, settings := range c.NetworkSettings.Networks {
			if settings.NetworkID != "" {
				result[settings.NetworkID] = true
			}
		}
	}
	return result, nil
}

// GetAllVolumeUsage returns a map of volume names that are in use by any container.
func (d *DockerBackend) GetAllVolumeUsage(ctx context.Context) (map[string]bool, error) {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make(map[string]bool)
	for _, c := range containers {
		for _, mount := range c.Mounts {
			if mount.Type == "volume" && mount.Name != "" {
				result[mount.Name] = true
			}
		}
	}
	return result, nil
}

// Helper functions for type conversion

func convertExposedPorts(ports nat.PortSet) map[string]struct{} {
	result := make(map[string]struct{})
	for port := range ports {
		result[string(port)] = struct{}{}
	}
	return result
}

func convertHostConfig(hc *container.HostConfig) backend.HostConfig {
	if hc == nil {
		return backend.HostConfig{}
	}

	portBindings := make(map[string][]backend.PortBinding)
	for port, bindings := range hc.PortBindings {
		backendBindings := make([]backend.PortBinding, len(bindings))
		for i, b := range bindings {
			backendBindings[i] = backend.PortBinding{
				HostIP:   b.HostIP,
				HostPort: b.HostPort,
			}
		}
		portBindings[string(port)] = backendBindings
	}

	pidsLimit := int64(0)
	if hc.PidsLimit != nil {
		pidsLimit = *hc.PidsLimit
	}

	return backend.HostConfig{
		Binds:           hc.Binds,
		ContainerIDFile: hc.ContainerIDFile,
		NetworkMode:     string(hc.NetworkMode),
		PortBindings:    portBindings,
		RestartPolicy: backend.RestartPolicy{
			Name:              string(hc.RestartPolicy.Name),
			MaximumRetryCount: hc.RestartPolicy.MaximumRetryCount,
		},
		AutoRemove:        hc.AutoRemove,
		Privileged:        hc.Privileged,
		PublishAllPorts:   hc.PublishAllPorts,
		ReadonlyRootfs:    hc.ReadonlyRootfs,
		DNS:               hc.DNS,
		DNSOptions:        hc.DNSOptions,
		DNSSearch:         hc.DNSSearch,
		ExtraHosts:        hc.ExtraHosts,
		CapAdd:            hc.CapAdd,
		CapDrop:           hc.CapDrop,
		CpuShares:         hc.CPUShares,
		Memory:            hc.Memory,
		MemorySwap:        hc.MemorySwap,
		MemoryReservation: hc.MemoryReservation,
		OomKillDisable:    hc.OomKillDisable != nil && *hc.OomKillDisable,
		PidsLimit:         pidsLimit,
	}
}

func convertNetworks(networks map[string]*network.EndpointSettings) map[string]backend.EndpointSettings {
	result := make(map[string]backend.EndpointSettings)
	for name, settings := range networks {
		if settings == nil {
			continue
		}

		es := backend.EndpointSettings{
			Links:               settings.Links,
			Aliases:             settings.Aliases,
			NetworkID:           settings.NetworkID,
			EndpointID:          settings.EndpointID,
			Gateway:             settings.Gateway,
			IPAddress:           settings.IPAddress,
			IPPrefixLen:         settings.IPPrefixLen,
			IPv6Gateway:         settings.IPv6Gateway,
			GlobalIPv6Address:   settings.GlobalIPv6Address,
			GlobalIPv6PrefixLen: settings.GlobalIPv6PrefixLen,
			MacAddress:          settings.MacAddress,
		}

		if settings.IPAMConfig != nil {
			es.IPAMConfig = &backend.EndpointIPAMConfig{
				IPv4Address: settings.IPAMConfig.IPv4Address,
				IPv6Address: settings.IPAMConfig.IPv6Address,
			}
		}

		result[name] = es
	}
	return result
}

func convertPortMap(ports nat.PortMap) map[string][]backend.PortBinding {
	result := make(map[string][]backend.PortBinding)
	for port, bindings := range ports {
		backendBindings := make([]backend.PortBinding, len(bindings))
		for i, b := range bindings {
			backendBindings[i] = backend.PortBinding{
				HostIP:   b.HostIP,
				HostPort: b.HostPort,
			}
		}
		result[string(port)] = backendBindings
	}
	return result
}

func convertMounts(mounts []types.MountPoint) []backend.Mount {
	result := make([]backend.Mount, len(mounts))
	for i, m := range mounts {
		result[i] = backend.Mount{
			Type:        string(m.Type),
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
			RW:          m.RW,
			Propagation: string(m.Propagation),
		}
	}
	return result
}

func convertIPAMConfig(configs []network.IPAMConfig) []backend.IPAMConfig {
	result := make([]backend.IPAMConfig, len(configs))
	for i, c := range configs {
		result[i] = backend.IPAMConfig{
			Subnet:     c.Subnet,
			IPRange:    c.IPRange,
			Gateway:    c.Gateway,
			AuxAddress: c.AuxAddress,
		}
	}
	return result
}

func convertEndpointResources(containers map[string]types.EndpointResource) map[string]backend.EndpointResource {
	result := make(map[string]backend.EndpointResource)
	for id, c := range containers {
		result[id] = backend.EndpointResource{
			Name:        c.Name,
			EndpointID:  c.EndpointID,
			MacAddress:  c.MacAddress,
			IPv4Address: c.IPv4Address,
			IPv6Address: c.IPv6Address,
		}
	}
	return result
}

// multiError collects multiple errors.
type multiError struct {
	errors []error
}

func (m *multiError) Add(err error) {
	if err != nil {
		m.errors = append(m.errors, err)
	}
}

func (m *multiError) ToError() error {
	if len(m.errors) == 0 {
		return nil
	}
	if len(m.errors) == 1 {
		return m.errors[0]
	}
	return fmt.Errorf("multiple errors: %v", m.errors)
}

// createTarArchive creates a tar archive of the build context directory.
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
