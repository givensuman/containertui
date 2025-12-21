package client

import (
	"context"
	"log"

	"github.com/moby/moby/client"
)

var CLI *client.Client

func init() {
	var err error
	CLI, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Test connection
	_, err = CLI.Ping(context.Background(), client.PingOptions{})
	if err != nil {
		log.Fatalf("Failed to ping Docker daemon: %v", err)
	}
}
