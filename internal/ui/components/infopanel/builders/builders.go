// Package builders provides panel builders for different resource types.
package builders

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/givensuman/containertui/internal/backend"
	"github.com/givensuman/containertui/internal/client"
	"github.com/givensuman/containertui/internal/registry"
	"github.com/givensuman/containertui/internal/ui/components/infopanel"
)

// BuildContainerPanel builds a raw inspection panel for a container.
func BuildContainerPanel(container backend.ContainerDetail, width int, expandEnv bool, format infopanel.OutputFormat) string {
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}

	// Marshal the container data
	rawData, err := infopanel.MarshalToFormat(container, format)
	if err != nil {
		return fmt.Sprintf("Error marshaling data: %v", err)
	}

	// Wrap in markdown code block with syntax highlighting
	rendered, err := infopanel.WrapInMarkdownCodeBlock(rawData, format, width)
	if err != nil {
		// Fallback to basic colorization if markdown rendering fails
		if format == infopanel.FormatYAML {
			return infopanel.ColorizeYAML(rawData)
		}
		return infopanel.ColorizeJSON(rawData)
	}

	return rendered
}

// BuildBrowsePanel builds a panel for Docker Hub registry image details.
func BuildBrowsePanel(detail registry.RegistryImageDetail, width int) string {
	// Get description
	description := detail.FullDescription
	if description == "" {
		description = detail.Description
	}
	if description == "" {
		description = "No description available."
	}

	description = normalizeReadmeWhitespace(description)

	// Render the markdown description
	rendered, err := infopanel.RenderMarkdown(description, width)
	if err != nil {
		// Fallback to plain text if markdown rendering fails
		return description
	}

	return rendered
}

func normalizeReadmeWhitespace(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	joined := strings.Join(lines, "\n")
	return strings.TrimSpace(joined)
}

// BuildImagePanel builds an informational panel for an image with raw YAML/JSON output.
func BuildImagePanel(image types.ImageInspect, width int, format infopanel.OutputFormat) string {
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}

	// Marshal the image data
	rawData, err := infopanel.MarshalToFormat(image, format)
	if err != nil {
		return fmt.Sprintf("Error marshaling data: %v", err)
	}

	// Wrap in markdown code block with syntax highlighting
	rendered, err := infopanel.WrapInMarkdownCodeBlock(rawData, format, width)
	if err != nil {
		// Fallback to basic colorization if markdown rendering fails
		if format == infopanel.FormatYAML {
			return infopanel.ColorizeYAML(rawData)
		}
		return infopanel.ColorizeJSON(rawData)
	}

	return rendered
}

// BuildNetworkPanel builds an informational panel for a network with raw YAML/JSON output.
func BuildNetworkPanel(network types.NetworkResource, width int, format infopanel.OutputFormat) string {
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}

	// Marshal the network data
	rawData, err := infopanel.MarshalToFormat(network, format)
	if err != nil {
		return fmt.Sprintf("Error marshaling data: %v", err)
	}

	// Wrap in markdown code block with syntax highlighting
	rendered, err := infopanel.WrapInMarkdownCodeBlock(rawData, format, width)
	if err != nil {
		// Fallback to basic colorization if markdown rendering fails
		if format == infopanel.FormatYAML {
			return infopanel.ColorizeYAML(rawData)
		}
		return infopanel.ColorizeJSON(rawData)
	}

	return rendered
}

// BuildVolumePanel builds an informational panel for a volume with raw YAML/JSON output.
func BuildVolumePanel(vol volume.Volume, width int, format infopanel.OutputFormat) string {
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}

	// Marshal the volume data
	rawData, err := infopanel.MarshalToFormat(vol, format)
	if err != nil {
		return fmt.Sprintf("Error marshaling data: %v", err)
	}

	// Wrap in markdown code block with syntax highlighting
	rendered, err := infopanel.WrapInMarkdownCodeBlock(rawData, format, width)
	if err != nil {
		// Fallback to basic colorization if markdown rendering fails
		if format == infopanel.FormatYAML {
			return infopanel.ColorizeYAML(rawData)
		}
		return infopanel.ColorizeJSON(rawData)
	}

	return rendered
}

// BuildServicePanel builds an informational panel for a service with raw YAML/JSON output.
func BuildServicePanel(service client.Service, width int, showFullCompose bool, format infopanel.OutputFormat) string {
	// If format is empty, use default from config
	if format == "" {
		format = infopanel.GetOutputFormat()
	}

	// Marshal the service data
	rawData, err := infopanel.MarshalToFormat(service, format)
	if err != nil {
		return fmt.Sprintf("Error marshaling data: %v", err)
	}

	// Wrap in markdown code block with syntax highlighting
	rendered, err := infopanel.WrapInMarkdownCodeBlock(rawData, format, width)
	if err != nil {
		// Fallback to basic colorization if markdown rendering fails
		if format == infopanel.FormatYAML {
			return infopanel.ColorizeYAML(rawData)
		}
		return infopanel.ColorizeJSON(rawData)
	}

	return rendered
}
