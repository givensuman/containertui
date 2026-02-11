# Tier 1 Feature Implementation Guide

This document provides detailed implementation guidance for all Tier 1 features.

## Status Overview

### Backend (✅ Complete)
All backend methods have been added to `internal/client/client.go`:
- ✅ `PruneContainers(ctx) (uint64, error)` 
- ✅ `PruneImages(ctx) (uint64, error)` - Already existed
- ✅ `PruneVolumes(ctx) (uint64, error)` - Already existed
- ✅ `PruneNetworks(ctx) error` - Already existed
- ✅ `CreateVolume(ctx, name, driver, labels) (string, error)`
- ✅ `CreateNetwork(ctx, name, driver, subnet, gateway, ipv6, labels) (string, error)`
- ✅ `TagImage(ctx, imageID, newTag) error`
- ✅ `RenameContainer(ctx, containerID, newName) error`
- ✅ `ImageHistory(ctx, imageID) ([]image.HistoryResponseItem, error)`
- ⚠️  `BuildImage(ctx, dockerfile, tag, context, buildArgs) (io.ReadCloser, error)` - Needs tar implementation

### Frontend (❌ To Be Implemented)
The following UI changes need to be made to each tab.

---

## Feature 1: Prune Operations

### Containers Tab
**Keybinding:** `ctrl+p` (to avoid conflict with pause='p')

**Implementation Steps:**
1. Add keybinding to `keybindings` struct:
```go
pruneContainers key.Binding
```

2. Initialize in `newKeybindings()`:
```go
pruneContainers: key.NewBinding(
    key.WithKeys("ctrl+p"),
    key.WithHelp("ctrl+p", "prune stopped containers"),
),
```

3. Add to `AdditionalHelp` in `Init()`

4. Handle in `Update()` method (in list focus section):
```go
case key.Matches(msg, model.keybindings.pruneContainers):
    cmds = append(cmds, model.handlePruneContainers())
```

5. Add handler method:
```go
func (model *Model) handlePruneContainers() tea.Cmd {
    return func() tea.Msg {
        ctx := stdcontext.Background()
        spaceReclaimed, err := state.GetClient().PruneContainers(ctx)
        if err != nil {
            return notifications.ShowError(err)
        }
        msg := fmt.Sprintf("Pruned stopped containers, freed %s", 
            humanizeBytes(spaceReclaimed))
        return tea.Batch(
            notifications.ShowSuccess(msg),
            model.Refresh(),
        )
    }
}

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

### Images Tab
**Keybinding:** `p`

Same pattern as containers, but simpler since 'p' isn't used:
```go
pruneImages: key.NewBinding(
    key.WithKeys("p"),
    key.WithHelp("p", "prune unused images"),
),
```

### Volumes Tab
**Keybinding:** `p`

Same pattern:
```go
pruneVolumes: key.NewBinding(
    key.WithKeys("p"),
    key.WithHelp("p", "prune unused volumes"),
),
```

### Networks Tab
**Keybinding:** `p`

Same pattern:
```go
pruneNetworks: key.NewBinding(
    key.WithKeys("p"),
    key.WithHelp("p", "prune unused networks"),
),
```

---

## Feature 2: Toggle Show All Containers

**Location:** `internal/ui/containers/containers.go`
**Keybinding:** `a`

**Implementation Steps:**

1. Add state field to `Model`:
```go
showAllContainers bool  // true = show all, false = show running only
```

2. Initialize in `Init()`:
```go
showAllContainers: true,  // Default to showing all
```

3. Add keybinding:
```go
toggleShowAll: key.NewBinding(
    key.WithKeys("a"),
    key.WithHelp("a", "toggle show all/running"),
),
```

4. Modify `Refresh()` to respect the filter:
```go
func (model Model) Refresh() tea.Cmd {
    return func() tea.Msg {
        ctx := stdcontext.Background()
        containers, err := state.GetClient().GetContainers(ctx)
        if err != nil {
            return notifications.ShowError(err)
        }
        
        // Filter if needed
        if !model.showAllContainers {
            filtered := make([]client.Container, 0)
            for _, c := range containers {
                if c.State == "running" {
                    filtered = append(filtered, c)
                }
            }
            containers = filtered
        }
        
        items := make([]ContainerItem, 0, len(containers))
        for _, container := range containers {
            items = append(items, ContainerItem{
                ID:    container.ID,
                Name:  container.Name,
                Image: container.Image,
                State: container.State,
            })
        }
        return base.RefreshItemsMsg[ContainerItem]{Items: items}
    }
}
```

5. Add toggle handler:
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

6. Update title/status bar to show current mode

---

## Feature 3: Force Delete Option

**All Tabs**
**Keybinding:** `D` (Shift+d)

**Implementation Steps:**

For each tab (containers, images, volumes, networks):

1. Add keybinding:
```go
forceRemove: key.NewBinding(
    key.WithKeys("D"),
    key.WithHelp("D", "force delete"),
),
```

2. Modify existing remove handlers to accept a `force` parameter:
```go
func (model *Model) handleRemoveContainers(force bool) {
    selectedIDs := model.GetSelectedIDs()
    if len(selectedIDs) == 0 {
        return
    }
    
    if !force {
        // Show existing confirmation dialog
        // ... existing code ...
    } else {
        // Skip confirmation, delete immediately
        spinnerCmd := model.setWorkingState(selectedIDs, true)
        return tea.Batch(
            spinnerCmd,
            PerformContainerOperations(Remove, selectedIDs),
            notifications.ShowWarning("Force deleting containers..."),
        )
    }
}
```

3. Add handler for force delete:
```go
case key.Matches(msg, model.keybindings.forceRemove):
    model.handleRemoveContainers(true)  // force=true
```

---

## Feature 4: Create Volume

**Location:** `internal/ui/volumes/volumes.go`
**Keybinding:** `n`

**Implementation Steps:**

1. Add keybinding:
```go
createVolume: key.NewBinding(
    key.WithKeys("n"),
    key.WithHelp("n", "create volume"),
),
```

2. Create form dialog fields:
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
            return performCreateVolume(values)
        },
    )
    
    model.ShowOverlay(dialog)
}
```

3. Create volume operation:
```go
func performCreateVolume(values map[string]string) tea.Cmd {
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
        
        name, err := state.GetClient().CreateVolume(
            ctx,
            values["name"],
            values["driver"],
            labels,
        )
        
        if err != nil {
            return notifications.ShowError(err)
        }
        
        return tea.Batch(
            notifications.ShowSuccess(fmt.Sprintf("Created volume: %s", name)),
            // Trigger refresh message
        )
    }
}
```

---

## Feature 5: Create Network

**Location:** `internal/ui/networks/networks.go`
**Keybinding:** `n`

**Implementation Steps:**

Similar to Create Volume, but with network-specific fields:

```go
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
        Placeholder: "no",
        Default:     "no",
    },
    {
        Key:         "labels",
        Label:       "Labels (KEY=VALUE, comma-separated)",
        Placeholder: "env=prod",
    },
}
```

---

## Feature 6: Run and Exec (Quick Start)

**Location:** `internal/ui/images/images.go`
**Keybinding:** `x`

**Implementation Steps:**

1. Add keybinding:
```go
runAndExec: key.NewBinding(
    key.WithKeys("x"),
    key.WithHelp("x", "run & exec"),
),
```

2. Handler:
```go
func (model *Model) handleRunAndExec() tea.Cmd {
    selected := model.GetSelectedItem()
    if selected == nil {
        return nil
    }
    
    return func() tea.Msg {
        ctx := stdcontext.Background()
        
        // Create ephemeral container
        config := client.CreateContainerConfig{
            Name:      fmt.Sprintf("temp-%d", time.Now().Unix()),
            ImageID:   selected.ID,
            Ports:     nil,  // No ports
            Volumes:   nil,  // No volumes
            Env:       nil,  // No env vars
            AutoStart: true,
            Network:   "bridge",
        }
        
        containerID, err := state.GetClient().CreateContainer(ctx, config)
        if err != nil {
            return notifications.ShowError(err)
        }
        
        // Immediately exec into it
        shell := []string{"/bin/sh"}
        _, err = state.GetClient().ExecShell(ctx, containerID, shell)
        if err != nil {
            return notifications.ShowError(err)
        }
        
        return notifications.ShowSuccess("Ephemeral container created and attached")
    }
}
```

Note: The container should be created with `--rm` flag, but this needs to be added to CreateContainerConfig.

---

## Feature 7: Tag Image

**Location:** `internal/ui/images/images.go`
**Keybinding:** `t`

**Implementation Steps:**

1. Add keybinding and form dialog:
```go
tagImage: key.NewBinding(
    key.WithKeys("t"),
    key.WithHelp("t", "tag image"),
),
```

2. Form dialog:
```go
fields := []components.FormField{
    {
        Key:         "tag",
        Label:       "New Tag",
        Placeholder: "myimage:v1.0",
        Required:    true,
    },
}
```

3. Handler:
```go
func performTagImage(imageID, newTag string) tea.Cmd {
    return func() tea.Msg {
        ctx := stdcontext.Background()
        err := state.GetClient().TagImage(ctx, imageID, newTag)
        if err != nil {
            return notifications.ShowError(err)
        }
        return tea.Batch(
            notifications.ShowSuccess(fmt.Sprintf("Tagged image as: %s", newTag)),
            // Refresh images list
        )
    }
}
```

---

## Feature 8: Rename Container

**Location:** `internal/ui/containers/containers.go`
**Keybinding:** `F2`

**Implementation Steps:**

Similar to tag image:

```go
renameContainer: key.NewBinding(
    key.WithKeys("f2"),
    key.WithHelp("F2", "rename container"),
),
```

Form with single field for new name, then call:
```go
err := state.GetClient().RenameContainer(ctx, containerID, newName)
```

---

## Feature 9: Image History

**Location:** `internal/ui/components/infopanel/builders/image.go`

**Implementation Steps:**

1. Fetch history when building image panel:
```go
func BuildImagePanel(img types.ImageInspect, width int, format Format) string {
    // ... existing code ...
    
    // Add history section
    ctx := context.Background()
    history, err := state.GetClient().ImageHistory(ctx, img.ID)
    if err == nil && len(history) > 0 {
        panel.AddSection("History")
        for i, layer := range history {
            size := humanizeBytes(uint64(layer.Size))
            created := time.Unix(layer.Created, 0).Format("2006-01-02")
            cmd := strings.TrimSpace(layer.CreatedBy)
            if len(cmd) > 60 {
                cmd = cmd[:57] + "..."
            }
            panel.AddKeyValue(fmt.Sprintf("Layer %d", i), 
                fmt.Sprintf("%s | %s | %s", created, size, cmd))
        }
    }
    
    // ... rest of code ...
}
```

---

## Feature 10: In-TUI Log Viewer

**This is the most complex feature and requires creating a new component.**

**Location:** Create `internal/ui/components/logviewer/logviewer.go`

**Key Requirements:**
- Use `viewport` for scrolling
- Stream logs from Docker API
- Support tail -f (follow mode)
- Search capability
- Toggle timestamps
- Copy to clipboard

**Component Structure:**
```go
type LogViewer struct {
    viewport   viewport.Model
    logs       []string
    following  bool
    showTimestamps bool
    containerID string
    maxLines   int
}
```

**This needs significant implementation and should be done carefully.**

---

## Feature 11: Build Image from Dockerfile

**This is complex and requires proper tar archive creation.**

**Location:** `internal/ui/images/images.go`
**Keybinding:** `b`

**Key Challenges:**
1. Need to properly create tar archive of build context
2. Need to stream build output
3. Need progress indication

**Recommended Approach:**
1. Use `archive/tar` package to create proper tar
2. Walk the build context directory
3. Honor .dockerignore files
4. Stream build output to progress dialog

**Basic tar implementation needed in client.go:**
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
            
            header, err := tar.FileInfoHeader(info, path)
            if err != nil {
                return err
            }
            
            relPath, err := filepath.Rel(contextPath, path)
            if err != nil {
                return err
            }
            header.Name = relPath
            
            if err := tw.WriteHeader(header); err != nil {
                return err
            }
            
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

## Testing Checklist

For each feature, test:
- [ ] Keybinding works
- [ ] Dialog displays correctly
- [ ] Operation succeeds
- [ ] Error handling works
- [ ] UI refreshes after operation
- [ ] Notification shows result
- [ ] Multi-selection works (where applicable)
- [ ] Help panel shows new keybinding

---

## Summary

**Completed:**
- ✅ All backend methods (except tar implementation for build)

**Remaining:**
- ❌ UI implementations for all 11 features
- ❌ Form dialogs for create/tag/rename operations
- ❌ Log viewer component (complex)
- ❌ Build image tar implementation (complex)
- ❌ Help panel updates
- ❌ Testing

**Estimated Remaining Effort:** 30-40 hours

**Recommended Next Steps:**
1. Implement prune operations first (simplest, high value)
2. Implement create volume/network (moderate complexity)
3. Implement tag/rename (simple dialogs)
4. Implement toggle show all (state management)
5. Implement force delete (modify existing)
6. Implement run & exec (reuse existing)
7. Implement image history (panel addition)
8. Implement log viewer (complex, time-consuming)
9. Implement build image (complex, needs tar)
10. Update all help panels and documentation
