// Package context provides a context for shared application state.
package context

import (
	"sync"

	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/config"
)

type WindowSize struct{ width, height int }

var (
	// Shared Moby client instance
	clientInstance *client.ClientWrapper
	// Configuration file/runtime instance
	configInstance *config.Config
	// Window width and height
	windowSize WindowSize
	once       sync.Once
)

// InitializeClient initializes the shared client instance.
func InitializeClient() {
	once.Do(func() {
		clientInstance = client.NewClient()
	})
}

// GetClient returns the shared client instance.
func GetClient() *client.ClientWrapper {
	return clientInstance
}

// CloseClient closes the shared client instance.
func CloseClient() {
	if clientInstance != nil {
		clientInstance.CloseClient()
	}
}

// SetConfig sets the shared config instance.
func SetConfig(cfg *config.Config) {
	configInstance = cfg
}

// GetConfig returns the shared config instance.
func GetConfig() *config.Config {
	return configInstance
}

// SetWindowSize sets the current window size.
func SetWindowSize(width, height int) {
	windowSize.width = width
	windowSize.height = height
}

// GetWindowSize returns the current window size.
func GetWindowSize() (int, int) {
	return windowSize.width, windowSize.height
}
