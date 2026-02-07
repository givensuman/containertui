package client

import "time"

const (
	// DefaultOperationTimeout is the default timeout for Docker operations
	DefaultOperationTimeout = 30 * time.Second
	
	// DefaultRestartTimeout is the default timeout for container restart
	DefaultRestartTimeout = 10 * time.Second
)
