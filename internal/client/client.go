// Package client exposes a Docker client wrapper for managing containers.
package client

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Container represents a Docker container with essential details.
type Container struct {
	container.ContainerJSONBase
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	Image string `json:"Image"`
	State string `json:"State"`
}

// ClientWrapper wraps the Docker client to provide container management functionalities.
type ClientWrapper struct {
	client *client.Client
}

// NewClient initializes a new Docker client.
func (cw *ClientWrapper) NewClient() {
	var err error
	cw.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err.Error())
	}
}

// CloseClient closes the Docker client connection.
func (cw *ClientWrapper) CloseClient() {
	cw.client.Close()
}

// GetContainers retrieves a list of all Docker containers.
func (cw *ClientWrapper) GetContainers() []Container {
	containers, err := cw.client.ContainerList(
		context.Background(),
		container.ListOptions{All: true},
	)
	if err != nil {
		panic(err)
	}
	var dockerContainers []Container
	for _, container := range containers {
		dockerContainers = append(dockerContainers, Container{
			ID:    container.ID,
			Name:  container.Names[0][1:],
			Image: container.Image,
			State: container.State,
		})
	}
	return dockerContainers
}

// GetContainerState retrieves the current state of a specific Docker container by its ID.
func (cw *ClientWrapper) GetContainerState(id string) string {
	container, _ := cw.client.ContainerInspect(context.Background(), id)
	return container.State.Status
}

// PauseContainer pauses a specific Docker container by its ID.
func (cw *ClientWrapper) PauseContainer(id string) {
	cw.client.ContainerPause(context.Background(), id)
}

// PauseContainers pauses multiple Docker containers by their IDs.
func (cw *ClientWrapper) PauseContainers(ids []string) {
	for _, id := range ids {
		cw.PauseContainer(id)
	}
}

// UnpauseContainer unpauses a specific Docker container by its ID.
func (cw *ClientWrapper) UnpauseContainer(id string) {
	cw.client.ContainerUnpause(context.Background(), id)
}

// UnpauseContainers unpauses multiple Docker containers by their IDs.
func (cw *ClientWrapper) UnpauseContainers(ids []string) {
	for _, id := range ids {
		cw.UnpauseContainer(id)
	}
}

// StartContainer starts a specific Docker container by its ID.
func (cw *ClientWrapper) StartContainer(id string) {
	cw.client.ContainerStart(context.Background(), id, container.StartOptions{})
}

// StartContainers starts multiple Docker containers by their IDs.
func (cw *ClientWrapper) StartContainers(ids []string) {
	for _, id := range ids {
		cw.StartContainer(id)
	}
}

// StopContainer stops a specific Docker container by its ID.
func (cw *ClientWrapper) StopContainer(id string) {
	cw.client.ContainerStop(context.Background(), id, container.StopOptions{})
}

// StopContainers stops multiple Docker containers by their IDs.
func (cw *ClientWrapper) StopContainers(ids []string) {
	for _, id := range ids {
		cw.StopContainer(id)
	}
}
