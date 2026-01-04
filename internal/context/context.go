// Package context provides a context for shared application state.
package context

import (
	"sync"

	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/config"
)

var (
	clientInstance *client.ClientWrapper
	configInstance *config.Config
	once           sync.Once
)

func InitializeClient() {
	once.Do(func() {
		clientInstance = client.NewClient()
	})
}

// GetClient returns the shared client instance.
func GetClient() *client.ClientWrapper {
	return clientInstance
}

// SetConfig sets the shared config instance.
func SetConfig(cfg *config.Config) {
	configInstance = cfg
}

// GetConfig returns the shared config instance.
func GetConfig() *config.Config {
	return configInstance
}

// CloseClient closes the shared client instance.
func CloseClient() {
	if clientInstance != nil {
		clientInstance.CloseClient()
	}
}
