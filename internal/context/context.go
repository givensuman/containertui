// Package context provides a context for shared application state.
package context

import (
	"sync"

	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/config"
)

var (
	// Shared Moby client instance
	clientInstance *client.ClientWrapper
	clientMu       sync.Mutex

	// Configuration file/runtime instance
	configInstance *config.Config
	configMu       sync.RWMutex

	// Window width and height
	windowSize struct {
		mu     sync.RWMutex
		Width  int
		Height int
	}

	// Ensures client is only initialized once
	clientOnce sync.Once
)

// InitializeClient initializes the shared client instance.
func InitializeClient() error {
	var err error
	clientOnce.Do(func() {
		clientMu.Lock()
		defer clientMu.Unlock()
		clientInstance, err = client.NewClient()
	})
	return err
}

// GetClient returns the shared client instance.
func GetClient() *client.ClientWrapper {
	clientMu.Lock()
	defer clientMu.Unlock()
	return clientInstance
}

// CloseClient closes the shared client instance.
func CloseClient() error {
	clientMu.Lock()
	defer clientMu.Unlock()
	if clientInstance != nil {
		return clientInstance.CloseClient()
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
