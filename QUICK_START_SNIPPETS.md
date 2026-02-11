# Quick Start Code Snippets for Tier 1 Features

Copy-paste ready code for rapid implementation.

## Utility Function (Add to any file that needs it)

```go
// humanizeBytes converts bytes to human-readable format
func humanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

---

## 1. PRUNE OPERATIONS

### Containers Tab - Add to keybindings struct:
```go
pruneContainers key.Binding
```

### Initialize:
```go
pruneContainers: key.NewBinding(
	key.WithKeys("ctrl+p"),
	key.WithHelp("ctrl+p", "prune stopped"),
),
```

### Handler:
```go
func (model *Model) handlePruneContainers() tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		spaceReclaimed, err := state.GetClient().PruneContainers(ctx)
		if err != nil {
			return notifications.ShowError(err)
		}
		msg := fmt.Sprintf("Pruned stopped containers, freed %s", humanizeBytes(spaceReclaimed))
		return tea.Batch(
			notifications.ShowSuccess(msg),
			model.Refresh(),
		)
	}
}
```

### Wire up in Update():
```go
case key.Matches(msg, model.keybindings.pruneContainers):
	cmds = append(cmds, model.handlePruneContainers())
```

### For Images/Volumes/Networks (use 'p' key):
```go
// Images
pruneImages: key.NewBinding(
	key.WithKeys("p"),
	key.WithHelp("p", "prune unused"),
),

// Handler
func (model *Model) handlePruneImages() tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		spaceReclaimed, err := state.GetClient().PruneImages(ctx)
		if err != nil {
			return notifications.ShowError(err)
		}
		msg := fmt.Sprintf("Pruned unused images, freed %s", humanizeBytes(spaceReclaimed))
		return tea.Batch(
			notifications.ShowSuccess(msg),
			model.Refresh(),
		)
	}
}
```

---

## 2. TOGGLE SHOW ALL CONTAINERS

### Add to Model struct:
```go
showAllContainers bool
```

### Initialize in Init():
```go
showAllContainers: true,
```

### Add keybinding:
```go
toggleShowAll: key.NewBinding(
	key.WithKeys("a"),
	key.WithHelp("a", "toggle all/running"),
),
```

### Handler:
```go
func (model *Model) handleToggleShowAll() tea.Cmd {
	model.showAllContainers = !model.showAllContainers
	mode := "all"
	if !model.showAllContainers {
		mode = "running only"
	}
	return tea.Batch(
		notifications.ShowInfo(fmt.Sprintf("Showing %s containers", mode)),
		model.Refresh(),
	)
}
```

### Modify Refresh() to filter:
```go
// In your Refresh() method, after getting containers:
if !model.showAllContainers {
	filtered := make([]client.Container, 0)
	for _, c := range containers {
		if c.State == "running" {
			filtered = append(filtered, c)
		}
	}
	containers = filtered
}
```

---

## 3. FORCE DELETE

### Add to all tabs (containers/images/volumes/networks):
```go
forceRemove: key.NewBinding(
	key.WithKeys("D"),
	key.WithHelp("D", "force delete"),
),
```

### Modify existing remove handler to accept force parameter:
```go
func (model *Model) handleRemoveContainers(force bool) tea.Cmd {
	selectedIDs := model.GetSelectedIDs()
	if len(selectedIDs) == 0 {
		return nil
	}
	
	if !force {
		// Existing confirmation dialog code
		// ... show dialog ...
	} else {
		// Skip confirmation
		spinnerCmd := model.setWorkingState(selectedIDs, true)
		return tea.Batch(
			spinnerCmd,
			PerformContainerOperations(Remove, selectedIDs),
			notifications.ShowWarning(fmt.Sprintf("Force deleting %d containers", len(selectedIDs))),
		)
	}
}
```

### Wire up:
```go
case key.Matches(msg, model.keybindings.removeContainer):
	cmds = append(cmds, model.handleRemoveContainers(false))
case key.Matches(msg, model.keybindings.forceRemove):
	cmds = append(cmds, model.handleRemoveContainers(true))
```

---

## 4. CREATE VOLUME

### Add keybinding:
```go
createVolume: key.NewBinding(
	key.WithKeys("n"),
	key.WithHelp("n", "create volume"),
),
```

### Handler (shows form):
```go
func (model *Model) showCreateVolumeDialog() {
	fields := []components.FormField{
		{
			Key:         "name",
			Label:       "Volume Name",
			Placeholder: "my-volume",
			Required:    true,
		},
		{
			Key:         "driver",
			Label:       "Driver",
			Placeholder: "local",
			Default:     "local",
		},
		{
			Key:         "labels",
			Label:       "Labels (KEY=VALUE, comma-separated)",
			Placeholder: "env=prod,app=web",
		},
	}
	
	dialog := components.NewFormDialog(
		"Create Volume",
		fields,
		func(values map[string]string) tea.Cmd {
			return model.performCreateVolume(values)
		},
	)
	
	model.ShowOverlay(dialog)
}
```

### Perform operation:
```go
func (model *Model) performCreateVolume(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		
		// Parse labels
		labels := make(map[string]string)
		if labelStr := values["labels"]; labelStr != "" {
			pairs := strings.Split(labelStr, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
		
		driver := values["driver"]
		if driver == "" {
			driver = "local"
		}
		
		name, err := state.GetClient().CreateVolume(
			ctx,
			values["name"],
			driver,
			labels,
		)
		
		if err != nil {
			return notifications.ShowError(err)
		}
		
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Created volume: %s", name)),
			model.Refresh(),
		)
	}
}
```

---

## 5. CREATE NETWORK

### Same pattern as volume, different fields:
```go
func (model *Model) showCreateNetworkDialog() {
	fields := []components.FormField{
		{
			Key:         "name",
			Label:       "Network Name",
			Placeholder: "my-network",
			Required:    true,
		},
		{
			Key:         "driver",
			Label:       "Driver",
			Placeholder: "bridge",
			Default:     "bridge",
		},
		{
			Key:         "subnet",
			Label:       "Subnet (CIDR)",
			Placeholder: "172.20.0.0/16",
		},
		{
			Key:         "gateway",
			Label:       "Gateway",
			Placeholder: "172.20.0.1",
		},
		{
			Key:         "ipv6",
			Label:       "Enable IPv6 (yes/no)",
			Default:     "no",
		},
		{
			Key:         "labels",
			Label:       "Labels",
			Placeholder: "env=prod",
		},
	}
	
	dialog := components.NewFormDialog(
		"Create Network",
		fields,
		func(values map[string]string) tea.Cmd {
			return model.performCreateNetwork(values)
		},
	)
	
	model.ShowOverlay(dialog)
}

func (model *Model) performCreateNetwork(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		
		// Parse labels
		labels := make(map[string]string)
		if labelStr := values["labels"]; labelStr != "" {
			pairs := strings.Split(labelStr, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
		
		driver := values["driver"]
		if driver == "" {
			driver = "bridge"
		}
		
		ipv6 := strings.ToLower(values["ipv6"]) == "yes"
		
		networkID, err := state.GetClient().CreateNetwork(
			ctx,
			values["name"],
			driver,
			values["subnet"],
			values["gateway"],
			ipv6,
			labels,
		)
		
		if err != nil {
			return notifications.ShowError(err)
		}
		
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Created network: %s", values["name"])),
			model.Refresh(),
		)
	}
}
```

---

## 6. RUN AND EXEC (Images Tab)

### Add keybinding:
```go
runAndExec: key.NewBinding(
	key.WithKeys("x"),
	key.WithHelp("x", "run & exec"),
),
```

### Handler:
```go
func (model *Model) handleRunAndExec() tea.Cmd {
	selected := model.GetSelectedItem()
	if selected == nil {
		return nil
	}
	
	// Generate temporary container name
	tempName := fmt.Sprintf("temp-%d", time.Now().Unix())
	
	return func() tea.Msg {
		ctx := stdcontext.Background()
		
		// Create ephemeral container
		config := client.CreateContainerConfig{
			Name:      tempName,
			ImageID:   selected.ID,
			Ports:     nil,
			Volumes:   nil,
			Env:       nil,
			AutoStart: true,
			Network:   "bridge",
		}
		
		containerID, err := state.GetClient().CreateContainer(ctx, config)
		if err != nil {
			return notifications.ShowError(err)
		}
		
		// Exec into it
		shell := []string{"/bin/sh"}
		_, err = state.GetClient().ExecShell(ctx, containerID, shell)
		if err != nil {
			// Try bash if sh fails
			shell = []string{"/bin/bash"}
			_, err = state.GetClient().ExecShell(ctx, containerID, shell)
		}
		
		// Note: Container cleanup happens when shell exits
		// You may want to force remove it after
		
		return notifications.ShowInfo("Ephemeral container session ended")
	}
}
```

---

## 7. TAG IMAGE

### Add keybinding:
```go
tagImage: key.NewBinding(
	key.WithKeys("t"),
	key.WithHelp("t", "tag image"),
),
```

### Handler:
```go
func (model *Model) showTagImageDialog() {
	selected := model.GetSelectedItem()
	if selected == nil {
		return
	}
	
	fields := []components.FormField{
		{
			Key:         "tag",
			Label:       "New Tag",
			Placeholder: "myimage:v1.0",
			Required:    true,
		},
	}
	
	dialog := components.NewFormDialog(
		"Tag Image",
		fields,
		func(values map[string]string) tea.Cmd {
			return model.performTagImage(selected.ID, values["tag"])
		},
	)
	
	model.ShowOverlay(dialog)
}

func (model *Model) performTagImage(imageID, newTag string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		err := state.GetClient().TagImage(ctx, imageID, newTag)
		if err != nil {
			return notifications.ShowError(err)
		}
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Tagged as: %s", newTag)),
			model.Refresh(),
		)
	}
}
```

---

## 8. RENAME CONTAINER

### Add keybinding:
```go
renameContainer: key.NewBinding(
	key.WithKeys("f2"),
	key.WithHelp("F2", "rename"),
),
```

### Handler (same pattern as tag):
```go
func (model *Model) showRenameContainerDialog() {
	selected := model.GetSelectedItem()
	if selected == nil {
		return
	}
	
	fields := []components.FormField{
		{
			Key:         "name",
			Label:       "New Name",
			Placeholder: "my-container",
			Required:    true,
		},
	}
	
	dialog := components.NewFormDialog(
		"Rename Container",
		fields,
		func(values map[string]string) tea.Cmd {
			return model.performRenameContainer(selected.ID, values["name"])
		},
	)
	
	model.ShowOverlay(dialog)
}

func (model *Model) performRenameContainer(containerID, newName string) tea.Cmd {
	return func() tea.Msg {
		ctx := stdcontext.Background()
		err := state.GetClient().RenameContainer(ctx, containerID, newName)
		if err != nil {
			return notifications.ShowError(err)
		}
		return tea.Batch(
			notifications.ShowSuccess(fmt.Sprintf("Renamed to: %s", newName)),
			model.Refresh(),
		)
	}
}
```

---

## 9. IMAGE HISTORY (Add to detail panel)

### In image panel builder (internal/ui/components/infopanel/builders/image.go):

```go
// Add to BuildImagePanel function, after existing sections:

// Fetch and display history
ctx := context.Background()
history, err := state.GetClient().ImageHistory(ctx, img.ID)
if err == nil && len(history) > 0 {
	builder.AddSection("Layer History")
	
	for i, layer := range history[:min(len(history), 10)] { // Show max 10 layers
		size := humanizeBytes(uint64(layer.Size))
		created := time.Unix(layer.Created, 0).Format("2006-01-02 15:04")
		
		cmd := strings.TrimSpace(layer.CreatedBy)
		if len(cmd) > 70 {
			cmd = cmd[:67] + "..."
		}
		
		label := fmt.Sprintf("Layer %d", i+1)
		value := fmt.Sprintf("%s │ %s │ %s", created, size, cmd)
		
		builder.AddKeyValue(label, value)
	}
	
	if len(history) > 10 {
		builder.AddKeyValue("...", fmt.Sprintf("(%d more layers)", len(history)-10))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

---

## 10. TAR ARCHIVE FOR BUILD IMAGE

### Replace the stub implementation in internal/client/client.go:

```go
import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func createTarArchive(contextPath string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()
	
	go func() {
		defer pw.Close()
		tw := tar.NewWriter(pw)
		defer tw.Close()
		
		err := filepath.Walk(contextPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			
			// Skip .git directories
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			
			// Create tar header
			header, err := tar.FileInfoHeader(info, path)
			if err != nil {
				return err
			}
			
			// Set relative path
			relPath, err := filepath.Rel(contextPath, path)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)
			
			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			
			// Write file content (if not a directory)
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				
				if _, err := io.Copy(tw, file); err != nil {
					return err
				}
			}
			
			return nil
		})
		
		if err != nil {
			pw.CloseWithError(err)
		}
	}()
	
	return pr, nil
}
```

---

## TESTING TEMPLATE

For each feature:

```go
func TestFeatureName(t *testing.T) {
	// Setup
	ctx := context.Background()
	client, err := client.NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.CloseClient()
	
	// Test
	result, err := client.YourMethod(ctx, params)
	
	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if result != expectedResult {
		t.Errorf("Expected %v, got %v", expectedResult, result)
	}
	
	// Cleanup
	// ... cleanup resources ...
}
```

---

## INTEGRATION TESTING

```bash
# Build the project
just build

# Start some test containers
docker run -d --name test-container nginx
docker run -d --name test-container-2 redis

# Run the TUI
./containertui

# Test each feature:
# 1. Navigate to containers tab (1)
# 2. Press 'a' to toggle view
# 3. Press 'ctrl+p' to prune
# 4. Select container and press 'F2' to rename
# 5. Press 'D' to force delete

# Navigate to images tab (2)
# 1. Press 'p' to prune
# 2. Select image and press 't' to tag
# 3. Press 'x' to run and exec
# 4. View history in detail panel

# Navigate to volumes tab (3)
# 1. Press 'n' to create volume
# 2. Press 'p' to prune
# 3. Press 'D' to force delete

# Navigate to networks tab (4)
# 1. Press 'n' to create network
# 2. Press 'p' to prune
# 3. Press 'D' to force delete

# Cleanup
docker rm -f test-container test-container-2
```

---

## QUICK REFERENCE: All New Keybindings

| Tab | Key | Feature |
|-----|-----|---------|
| Containers | `ctrl+p` | Prune stopped |
| Containers | `a` | Toggle all/running |
| Containers | `D` | Force delete |
| Containers | `F2` | Rename |
| Images | `p` | Prune unused |
| Images | `t` | Tag image |
| Images | `x` | Run & exec |
| Images | `D` | Force delete |
| Volumes | `n` | Create volume |
| Volumes | `p` | Prune unused |
| Volumes | `D` | Force delete |
| Networks | `n` | Create network |
| Networks | `p` | Prune unused |
| Networks | `D` | Force delete |

---

Ready to implement! Start with prune operations (simplest) and work your way through.
