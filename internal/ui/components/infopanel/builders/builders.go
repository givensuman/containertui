// Package builders provides panel builders for different resource types.
package builders

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/context"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
)

// BuildContainerPanel builds a raw inspection panel for a container in lazydocker style.
func BuildContainerPanel(container types.ContainerJSON, width int, expandEnv bool, format infopanel.OutputFormat) string {
	var output strings.Builder

	// === SUMMARY SECTION (lazydocker style) ===
	output.WriteString(formatSummaryField("ID", infopanel.TruncateID(container.ID)))
	output.WriteString(formatSummaryField("Name", container.Name))
	output.WriteString(formatSummaryField("Image", container.Config.Image))
	output.WriteString(formatSummaryField("Command", formatCommand(container)))

	// Labels (multi-line, indented)
	if len(container.Config.Labels) > 0 {
		output.WriteString(formatSummaryField("Labels", formatLabels(container.Config.Labels)))
	}

	// Mounts (multi-line, bulleted)
	if len(container.Mounts) > 0 {
		output.WriteString(formatSummaryField("Mounts", formatMounts(container.Mounts)))
	}

	// Ports (multi-line, bulleted)
	if container.NetworkSettings != nil && len(container.NetworkSettings.Ports) > 0 {
		portsStr := formatPorts(container.NetworkSettings.Ports)
		if portsStr != "" {
			output.WriteString(formatSummaryField("Ports", portsStr))
		}
	}

	// Connected Resources (our addition)
	networks, volumes := getConnectedResources(container)
	if len(networks) > 0 {
		output.WriteString(formatSummaryField("Networks", formatBulletList(networks)))
	}
	if len(volumes) > 0 {
		output.WriteString(formatSummaryField("Volumes", formatBulletList(volumes)))
	}

	// === FULL DETAILS SECTION ===
	output.WriteString("\n\nFull details:\n\n")

	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}
	rawData, err := infopanel.MarshalToFormat(container, format)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error marshaling data: %v", err))
		return output.String()
	}

	// Apply syntax highlighting
	if format == infopanel.FormatYAML {
		output.WriteString(infopanel.ColorizeYAML(rawData))
	} else {
		output.WriteString(infopanel.ColorizeJSON(rawData))
	}

	return output.String()
}

// formatSummaryField formats a field in lazydocker style: "Label:     Value"
func formatSummaryField(label, value string) string {
	return fmt.Sprintf("%-12s %s\n", label+":", value)
}

// formatCommand formats the command from container config
func formatCommand(c types.ContainerJSON) string {
	if len(c.Config.Cmd) == 0 {
		return "(none)"
	}
	return strings.Join(c.Config.Cmd, " ")
}

// formatLabels formats labels for summary section (multi-line, indented)
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "(none)"
	}
	var lines []string
	first := true
	for k, v := range labels {
		if first {
			lines = append(lines, fmt.Sprintf("%s=%s", k, v))
			first = false
		} else {
			lines = append(lines, fmt.Sprintf("\n             %s=%s", k, v))
		}
	}
	return strings.Join(lines, "")
}

// formatMounts formats mounts for summary section (multi-line, bulleted)
func formatMounts(mounts []types.MountPoint) string {
	if len(mounts) == 0 {
		return "(none)"
	}
	var lines []string
	for i, m := range mounts {
		mountStr := fmt.Sprintf("%s: %s", m.Type, m.Name)
		if m.Name == "" {
			// For bind mounts without a name
			mountStr = fmt.Sprintf("%s: %s", m.Type, infopanel.TruncateString(m.Source, 40))
		}
		if i == 0 {
			lines = append(lines, mountStr)
		} else {
			lines = append(lines, "\n             "+mountStr)
		}
	}
	return strings.Join(lines, "")
}

// formatPorts formats ports for summary section (multi-line, bulleted)
func formatPorts(ports nat.PortMap) string {
	var lines []string
	first := true
	for port, bindings := range ports {
		if len(bindings) == 0 {
			continue
		}
		for _, binding := range bindings {
			portStr := fmt.Sprintf("%s:%s → %s", binding.HostIP, binding.HostPort, port)
			if first {
				lines = append(lines, portStr)
				first = false
			} else {
				lines = append(lines, "\n             "+portStr)
			}
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "")
}

// formatBulletList formats a list of strings with bullets (multi-line)
func formatBulletList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	var lines []string
	for i, item := range items {
		if i == 0 {
			lines = append(lines, "• "+item)
		} else {
			lines = append(lines, "\n             • "+item)
		}
	}
	return strings.Join(lines, "")
}

// getConnectedResources extracts network and volume names from a container
func getConnectedResources(c types.ContainerJSON) (networks, volumes []string) {
	// Extract network names
	if c.NetworkSettings != nil {
		for name := range c.NetworkSettings.Networks {
			networks = append(networks, name)
		}
	}

	// Extract volume names
	for _, mount := range c.Mounts {
		if mount.Type == "volume" && mount.Name != "" {
			volumes = append(volumes, mount.Name)
		}
	}

	return networks, volumes
}

// formatTags formats image tags for summary section
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "(none)"
	}
	return tags[0]
}

// formatImageLabels formats image labels for summary section (multi-line, indented)
func formatImageLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "(none)"
	}
	var lines []string
	first := true
	for k, v := range labels {
		labelStr := fmt.Sprintf("%s=%s", k, v)
		if first {
			lines = append(lines, labelStr)
			first = false
		} else {
			lines = append(lines, "\n             "+labelStr)
		}
	}
	return strings.Join(lines, "")
}

// BuildImagePanel builds an informational panel for an image with raw YAML/JSON output.
func BuildImagePanel(image types.ImageInspect, width int, format infopanel.OutputFormat) string {
	var output strings.Builder

	// Summary section
	output.WriteString(formatSummaryField("ID", infopanel.TruncateID(image.ID)))
	output.WriteString(formatSummaryField("Tags", formatTags(image.RepoTags)))
	output.WriteString(formatSummaryField("Size", infopanel.FormatBytes(image.Size)))
	output.WriteString(formatSummaryField("Created", image.Created))
	output.WriteString(formatSummaryField("Architecture", image.Architecture))
	output.WriteString(formatSummaryField("OS", image.Os))

	// Labels section (if present)
	if image.Config != nil && len(image.Config.Labels) > 0 {
		output.WriteString(formatSummaryField("Labels", formatImageLabels(image.Config.Labels)))
	}

	// Connected Resources - containers using this image
	usedBy, err := context.GetClient().GetContainersUsingImage(image.ID)
	if err == nil && len(usedBy) > 0 {
		output.WriteString(formatSummaryField("Used By", formatBulletList(usedBy)))
	}

	// Full details separator
	output.WriteString("\n\nFull details:\n\n")

	// Marshal and colorize full inspect output
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}
	rawData, err := infopanel.MarshalToFormat(image, format)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error marshaling data: %v", err))
		return output.String()
	}

	if format == infopanel.FormatYAML {
		output.WriteString(infopanel.ColorizeYAML(rawData))
	} else {
		output.WriteString(infopanel.ColorizeJSON(rawData))
	}

	return output.String()
}

// BuildNetworkPanel builds an informational panel for a network with raw YAML/JSON output.
func BuildNetworkPanel(network types.NetworkResource, width int, format infopanel.OutputFormat) string {
	var output strings.Builder

	// Summary section
	output.WriteString(formatSummaryField("ID", infopanel.TruncateID(network.ID)))
	output.WriteString(formatSummaryField("Name", network.Name))
	output.WriteString(formatSummaryField("Driver", network.Driver))
	output.WriteString(formatSummaryField("Scope", network.Scope))
	output.WriteString(formatSummaryField("Created", infopanel.FormatTimestamp(network.Created.Unix())))

	// Subnet/Gateway from IPAM config
	if len(network.IPAM.Config) > 0 {
		for _, config := range network.IPAM.Config {
			if config.Subnet != "" {
				output.WriteString(formatSummaryField("Subnet", config.Subnet))
			}
			if config.Gateway != "" {
				output.WriteString(formatSummaryField("Gateway", config.Gateway))
			}
			break // Only show first config in summary
		}
	}

	// Labels section (if present)
	if len(network.Labels) > 0 {
		output.WriteString(formatSummaryField("Labels", formatNetworkLabels(network.Labels)))
	}

	// Connected Resources - containers using this network
	usedBy, err := context.GetClient().GetContainersUsingNetwork(network.ID)
	if err == nil && len(usedBy) > 0 {
		// Build container list with IPs
		var containerLines []string
		for _, name := range usedBy {
			ipAddr := ""
			if ep, exists := network.Containers[name]; exists {
				if ep.IPv4Address != "" {
					ipAddr = " (" + ep.IPv4Address + ")"
				}
			}
			containerLines = append(containerLines, name+ipAddr)
		}
		output.WriteString(formatSummaryField("Connected", formatBulletList(containerLines)))
	}

	// Full details separator
	output.WriteString("\n\nFull details:\n\n")

	// Marshal and colorize full inspect output
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}
	rawData, err := infopanel.MarshalToFormat(network, format)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error marshaling data: %v", err))
		return output.String()
	}

	if format == infopanel.FormatYAML {
		output.WriteString(infopanel.ColorizeYAML(rawData))
	} else {
		output.WriteString(infopanel.ColorizeJSON(rawData))
	}

	return output.String()
}

// formatNetworkLabels formats network labels for summary section (multi-line, indented)
func formatNetworkLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "(none)"
	}
	var lines []string
	first := true
	for k, v := range labels {
		labelStr := fmt.Sprintf("%s=%s", k, v)
		if first {
			lines = append(lines, labelStr)
			first = false
		} else {
			lines = append(lines, "\n             "+labelStr)
		}
	}
	return strings.Join(lines, "")
}

// BuildVolumePanel builds an informational panel for a volume with raw YAML/JSON output.
func BuildVolumePanel(vol volume.Volume, width int, format infopanel.OutputFormat) string {
	var output strings.Builder

	// Summary section
	output.WriteString(formatSummaryField("Name", vol.Name))
	output.WriteString(formatSummaryField("Driver", vol.Driver))
	output.WriteString(formatSummaryField("Mountpoint", infopanel.TruncateString(vol.Mountpoint, 50)))
	output.WriteString(formatSummaryField("Scope", vol.Scope))
	output.WriteString(formatSummaryField("Created", vol.CreatedAt))

	// Labels section (if present)
	if len(vol.Labels) > 0 {
		output.WriteString(formatSummaryField("Labels", formatVolumeLabels(vol.Labels)))
	}

	// Connected Resources - containers using this volume
	usedBy, err := context.GetClient().GetContainersUsingVolume(vol.Name)
	if err == nil && len(usedBy) > 0 {
		output.WriteString(formatSummaryField("Mounted By", formatBulletList(usedBy)))
	}

	// Full details separator
	output.WriteString("\n\nFull details:\n\n")

	// Marshal and colorize full inspect output
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}
	rawData, err := infopanel.MarshalToFormat(vol, format)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error marshaling data: %v", err))
		return output.String()
	}

	if format == infopanel.FormatYAML {
		output.WriteString(infopanel.ColorizeYAML(rawData))
	} else {
		output.WriteString(infopanel.ColorizeJSON(rawData))
	}

	return output.String()
}

// formatVolumeLabels formats volume labels for summary section (multi-line, indented)
func formatVolumeLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "(none)"
	}
	var lines []string
	first := true
	for k, v := range labels {
		labelStr := fmt.Sprintf("%s=%s", k, v)
		if first {
			lines = append(lines, labelStr)
			first = false
		} else {
			lines = append(lines, "\n             "+labelStr)
		}
	}
	return strings.Join(lines, "")
}

// BuildServicePanel builds an informational panel for a service with raw YAML/JSON output.
func BuildServicePanel(service client.Service, width int, showFullCompose bool, format infopanel.OutputFormat) string {
	var output strings.Builder

	// Summary section
	output.WriteString(formatSummaryField("Name", service.Name))
	output.WriteString(formatSummaryField("Project", service.Name))
	output.WriteString(formatSummaryField("Replicas", fmt.Sprintf("%d", service.Replicas)))
	output.WriteString(formatSummaryField("Compose File", infopanel.TruncateString(service.ComposeFile, 50)))

	// Containers section with status badges
	if len(service.Containers) > 0 {
		var containerLines []string
		for _, c := range service.Containers {
			status := infopanel.StatusBadge(c.State, c.State == "running")
			containerLines = append(containerLines, fmt.Sprintf("%s %s", c.Name, status))
		}
		output.WriteString(formatSummaryField("Containers", formatBulletList(containerLines)))
	}

	// Full details separator
	output.WriteString("\n\nFull details:\n\n")

	// Marshal and colorize full service struct
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}
	rawData, err := infopanel.MarshalToFormat(service, format)
	if err != nil {
		output.WriteString(fmt.Sprintf("Error marshaling data: %v", err))
		return output.String()
	}

	if format == infopanel.FormatYAML {
		output.WriteString(infopanel.ColorizeYAML(rawData))
	} else {
		output.WriteString(infopanel.ColorizeJSON(rawData))
	}

	return output.String()
}
