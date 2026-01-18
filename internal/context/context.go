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
	// Configuration file/runtime instance
	configInstance *config.Config
	// Window width and height
	windowSize struct {
		Width  int
		Height int
	}
	// ～(^з^)~~☆
	once sync.Once
)

// InitializeClient initializes the shared client instance.
func InitializeClient() error {
	var err error
	once.Do(func() {
		clientInstance, err = client.NewClient()
	})
	return err
}

// GetClient returns the shared client instance.
func GetClient() *client.ClientWrapper {
	return clientInstance
}

// CloseClient closes the shared client instance.
func CloseClient() error {
	if clientInstance != nil {
		return clientInstance.CloseClient()
	}

	return nil
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
	windowSize.Width = width
	windowSize.Height = height
}

// GetWindowSize returns the current window size.
func GetWindowSize() (int, int) {
	return windowSize.Width, windowSize.Height
}
