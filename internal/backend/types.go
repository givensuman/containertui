package backend

import (
	"io"
	"time"
)

// Container represents a backend-agnostic container.
type Container struct {
	ID      string
	Name    string
	Image   string
	State   string
	Status  string
	Created time.Time
}

// ContainerDetail contains detailed information about a container.
type ContainerDetail struct {
	Container
	Config          ContainerConfigDetail
	HostConfig      HostConfig
	NetworkSettings NetworkSettings
	Mounts          []Mount
	Raw             interface{} // Backend-specific raw data
}

// ContainerConfigDetail contains container configuration details.
type ContainerConfigDetail struct {
	Hostname     string
	Domainname   string
	User         string
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Tty          bool
	OpenStdin    bool
	StdinOnce    bool
	Env          []string
	Cmd          []string
	Image        string
	Volumes      map[string]struct{}
	WorkingDir   string
	Entrypoint   []string
	Labels       map[string]string
	ExposedPorts map[string]struct{}
}

// HostConfig contains host configuration for a container.
type HostConfig struct {
	Binds             []string
	ContainerIDFile   string
	NetworkMode       string
	PortBindings      map[string][]PortBinding
	RestartPolicy     RestartPolicy
	AutoRemove        bool
	Privileged        bool
	PublishAllPorts   bool
	ReadonlyRootfs    bool
	DNS               []string
	DNSOptions        []string
	DNSSearch         []string
	ExtraHosts        []string
	CapAdd            []string
	CapDrop           []string
	CpuShares         int64
	Memory            int64
	MemorySwap        int64
	MemoryReservation int64
	OomKillDisable    bool
	PidsLimit         int64
}

// PortBinding represents a port binding.
type PortBinding struct {
	HostIP   string
	HostPort string
}

// RestartPolicy represents a container restart policy.
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// NetworkSettings contains network settings for a container.
type NetworkSettings struct {
	Networks map[string]EndpointSettings
	Ports    map[string][]PortBinding
}

// EndpointSettings contains network endpoint settings.
type EndpointSettings struct {
	IPAMConfig          *EndpointIPAMConfig
	Links               []string
	Aliases             []string
	NetworkID           string
	EndpointID          string
	Gateway             string
	IPAddress           string
	IPPrefixLen         int
	IPv6Gateway         string
	GlobalIPv6Address   string
	GlobalIPv6PrefixLen int
	MacAddress          string
}

// EndpointIPAMConfig contains IPAM configuration for an endpoint.
type EndpointIPAMConfig struct {
	IPv4Address string
	IPv6Address string
}

// Mount represents a volume mount.
type Mount struct {
	Type        string
	Source      string
	Destination string
	Mode        string
	RW          bool
	Propagation string
}

// Image represents a backend-agnostic image.
type Image struct {
	ID          string
	RepoTags    []string
	RepoDigests []string
	Size        int64
	Created     time.Time
}

// ImageDetail contains detailed information about an image.
type ImageDetail struct {
	Image
	Author       string
	Comment      string
	Config       ContainerConfigDetail
	Architecture string
	Os           string
	Size         int64
	VirtualSize  int64
	RootFS       RootFS
	Raw          interface{} // Backend-specific raw data
}

// RootFS contains root filesystem information.
type RootFS struct {
	Type   string
	Layers []string
}

// Network represents a backend-agnostic network.
type Network struct {
	ID     string
	Name   string
	Driver string
	Scope  string
}

// NetworkDetail contains detailed information about a network.
type NetworkDetail struct {
	Network
	EnableIPv6 bool
	IPAM       IPAM
	Internal   bool
	Attachable bool
	Ingress    bool
	ConfigFrom ConfigReference
	ConfigOnly bool
	Containers map[string]EndpointResource
	Options    map[string]string
	Labels     map[string]string
	Raw        interface{} // Backend-specific raw data
}

// IPAM represents IP Address Management configuration.
type IPAM struct {
	Driver  string
	Options map[string]string
	Config  []IPAMConfig
}

// IPAMConfig contains IPAM configuration.
type IPAMConfig struct {
	Subnet     string
	IPRange    string
	Gateway    string
	AuxAddress map[string]string
}

// ConfigReference represents a network config reference.
type ConfigReference struct {
	Network string
}

// EndpointResource represents a network endpoint resource.
type EndpointResource struct {
	Name        string
	EndpointID  string
	MacAddress  string
	IPv4Address string
	IPv6Address string
}

// Volume represents a backend-agnostic volume.
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	CreatedAt  time.Time
}

// VolumeDetail contains detailed information about a volume.
type VolumeDetail struct {
	Volume
	Labels  map[string]string
	Scope   string
	Options map[string]string
	Raw     interface{} // Backend-specific raw data
}

// Service represents a service (Docker Compose service, Podman pod, etc.).
type Service struct {
	ID    string
	Name  string
	State string
}

// Logs represents a log stream from a container.
type Logs struct {
	Stream io.ReadCloser
	Close  func() error
}

// ContainerConfig holds configuration for creating a container.
type ContainerConfig struct {
	Name       string
	Image      string
	Ports      map[string]string // "hostPort" -> "containerPort"
	Volumes    []string          // "hostPath:containerPort" format
	Env        []string          // "KEY=value" format
	Cmd        []string          // optional command override; nil = use image default
	Tty        bool              // allocate a pseudo-TTY (keeps shell-based containers alive when detached)
	OpenStdin  bool              // keep stdin open (required for interactive shells)
	AutoStart  bool
	AutoRemove bool
	Network    string // Network name (default: "bridge")
}

// ImageHistoryItem represents a single layer in an image's history.
type ImageHistoryItem struct {
	ID        string
	Created   time.Time
	CreatedBy string
	Tags      []string
	Size      int64
	Comment   string
}
