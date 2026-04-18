// Package state provides shared application state management.
package state

import (
	"sync"

	"github.com/givensuman/containertui/internal/backend"
	dockerbackend "github.com/givensuman/containertui/internal/backend/docker"
	"github.com/givensuman/containertui/internal/config"
	"github.com/givensuman/containertui/internal/registry"
)

var (
	// Shared backend instance
	backendInstance backend.Backend
	backendMu       sync.Mutex

	// Shared registry client instances
	registryClient     *registry.Client
	quayRegistryClient *registry.QuayClient

	// Configuration file/runtime instance
	configInstance *config.Config
	configMu       sync.RWMutex

	// Window width and height
	windowSize struct {
		mu     sync.RWMutex
		Width  int
		Height int
	}

	// Ensures backend is only initialized once
	clientOnce sync.Once
)

// InitializeClient initializes the shared backend and registry client instances.
func InitializeClient() error {
	var err error
	clientOnce.Do(func() {
		backendMu.Lock()
		defer backendMu.Unlock()

		var b *dockerbackend.DockerBackend
		b, err = dockerbackend.New()
		if err != nil {
			return
		}
		backendInstance = b
		registryClient = registry.NewClient()
		quayRegistryClient = registry.NewQuayClient()
	})
	return err
}

// GetBackend returns the shared backend instance.
func GetBackend() backend.Backend {
	backendMu.Lock()
	defer backendMu.Unlock()
	return backendInstance
}

// GetRegistryClient returns the shared DockerHub registry client.
func GetRegistryClient() *registry.Client {
	backendMu.Lock()
	defer backendMu.Unlock()
	return registryClient
}

// GetQuayRegistryClient returns the shared Quay registry client.
func GetQuayRegistryClient() *registry.QuayClient {
	backendMu.Lock()
	defer backendMu.Unlock()
	return quayRegistryClient
}

// CloseClient closes the shared backend instance.
func CloseClient() error {
	backendMu.Lock()
	defer backendMu.Unlock()
	if backendInstance != nil {
		return backendInstance.Close()
	}
	return nil
}

// SetConfig sets the shared config instance.
func SetConfig(cfg *config.Config) {
	configMu.Lock()
	defer configMu.Unlock()
	configInstance = cfg
}

// GetConfig returns the shared config instance.
func GetConfig() *config.Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return configInstance
}

// SetWindowSize sets the current window size.
func SetWindowSize(width, height int) {
	windowSize.mu.Lock()
	defer windowSize.mu.Unlock()
	windowSize.Width = width
	windowSize.Height = height
}

// GetWindowSize returns the current window size.
func GetWindowSize() (int, int) {
	windowSize.mu.RLock()
	defer windowSize.mu.RUnlock()
	return windowSize.Width, windowSize.Height
}
