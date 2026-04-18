# Architecture Cleanup Implementation Plan

Goal: Remove dead code (services tab, unused backend abstraction), complete the backend migration so all tabs use backend.Backend instead of concrete *client.ClientWrapper, and fix smaller structural issues (async refresh, duplicate pull method, key conflict, Component naming).

Architecture: internal/backend already defines Backend interface + full DockerBackend implementation — neither is wired in. Plan: extend the interface to cover all UI needs, swap internal/state to expose backend.Backend, update every tab, delete internal/client. Services tab and all dead references removed first.

Tech Stack: Go, Bubbletea v2, Moby Docker SDK, Lipgloss, Bubbles

---

## Status

- [x] Task 1: Delete Services Tab Dead Code
- [x] Task 2: Rename base.Component → base.WindowSize
- [x] Task 3: Async ResourceView.Refresh()
- [ ] Task 4: Deduplicate PullImage / PullImageFromRegistry
- [ ] Task 5: Extend Backend Interface to Cover All UI Needs
- [ ] Task 6: Wire Backend into internal/state
- [ ] Task 7: Migrate Containers Tab
- [ ] Task 8: Migrate Images Tab
- [ ] Task 9: Migrate Networks Tab
- [ ] Task 10: Migrate Volumes Tab
- [ ] Task 11: Migrate Browse Tab
- [ ] Task 12: Update Infopanel Builders
- [ ] Task 13: Final Build, Test, Delete internal/client

---

## Task 4: Deduplicate PullImage / PullImageFromRegistry
Files: internal/client/client.go, internal/ui/browse/browse.go
- [ ] Delete PullImageFromRegistry method from client.go (lines 838–860)
- [ ] In browse.go: s/PullImageFromRegistry/PullImage/g
- [ ] Run: go build ./... && go test ./...
Commit: `remove duplicate PullImageFromRegistry, consolidate to PullImage`

---

## Task 5: Extend Backend Interface to Cover All UI Needs
Files: internal/backend/backend.go, internal/backend/types.go, internal/backend/docker/docker.go
- [ ] Add ImageHistoryItem to types.go:
    type ImageHistoryItem struct {
      ID        string
      Created   time.Time
      CreatedBy string
      Tags      []string
      Size      int64
      Comment   string
    }
- [ ] Add AutoRemove bool to ContainerConfig in types.go
- [ ] Confirm then delete PullOptions, PullResult, BuildOptions, BuildResult from types.go (grep -rn "PullOptions\|PullResult\|BuildOptions\|BuildResult" internal/)
- [ ] Update backend.go interface — change/add these signatures:
    PullImage(ctx context.Context, ref string, progressChan chan<- string) error
    BuildImage(ctx context.Context, dockerfilePath, tag, contextPath string, buildArgs map[string]*string) (io.ReadCloser, error)
    RemoveContainer(ctx context.Context, id string, force bool) error
    RemoveContainers(ctx context.Context, ids []string, force bool) error
    ImageHistory(ctx context.Context, imageID string) ([]ImageHistoryItem, error)
    GetAllNetworkUsage(ctx context.Context) (map[string]bool, error)
    GetAllVolumeUsage(ctx context.Context) (map[string]bool, error)
- [ ] Implement all new/changed methods in docker/docker.go:
  - PullImage — ImagePull → scan to channel → close channel
  - BuildImage — copy createTarArchive helper verbatim from client/client.go:955–1026, then ImageBuild, return resp.Body
  - RemoveContainer(ctx, id, force) — ContainerRemove with Force: force
  - RemoveContainers(ctx, ids, force) — loop + collect errors
  - ImageHistory — client.ImageHistory → map to backend.ImageHistoryItem
  - GetAllNetworkUsage — ContainerList → iterate NetworkSettings.Networks
  - GetAllVolumeUsage — ContainerList → iterate Mounts
  - Update CreateContainer to pass AutoRemove: cfg.AutoRemove to HostConfig
- [ ] Run: go build ./internal/backend/... && go test ./internal/backend/...
Commit: `extend Backend interface to cover all UI needs, implement in DockerBackend`

---

## Task 6: Wire Backend into internal/state
Files: internal/state/state.go
- [ ] Rewrite state.go to hold backend.Backend (from dockerbackend.New()) plus separate *registry.Client and *registry.QuayClient fields. Expose:
  - InitializeClient() error — creates DockerBackend + registry clients
  - GetBackend() backend.Backend
  - GetRegistryClient() *registry.Client
  - GetQuayRegistryClient() *registry.QuayClient
  - CloseClient() error — calls backend.Close()
  - Keep SetConfig, GetConfig, SetWindowSize, GetWindowSize unchanged
- [ ] Run: go build ./internal/state/...
Commit: `swap state package to backend.Backend`

---

## Task 7: Migrate Containers Tab
Files: internal/ui/containers/item.go, messages.go, containers.go
- [ ] item.go: client.Container → backend.Container (check for container.Config field access — backend.Container doesn't embed it; any such fields move to inspection path)
- [ ] messages.go: state.GetClient() → state.GetBackend(), RemoveContainer call already has force param
- [ ] containers.go:
  - MsgContainerInspection.Container: types.ContainerJSON → backend.ContainerDetail
  - GetContainers → ListContainers
  - All state.GetClient() → state.GetBackend()
  - OpenLogs result is now backend.Logs{Stream, Close} — update logs.go accordingly
- [ ] Run: go build ./internal/ui/containers/... && go test ./internal/ui/containers/...
Commit: `migrate containers tab to backend.Backend`

---

## Task 8: Migrate Images Tab
Files: internal/ui/images/item.go, images.go
- [ ] item.go: client.Image → backend.Image (Created is now time.Time not int64 — update any display code)
- [ ] images.go: All GetClient() → GetBackend(), method renames (GetImages → ListImages), ImageHistory returns []backend.ImageHistoryItem, update inspection message type to backend.ImageDetail
- [ ] Run: go build ./internal/ui/images/... && go test ./internal/ui/images/...
Commit: `migrate images tab to backend.Backend`

---

## Task 9: Migrate Networks Tab
Files: internal/ui/networks/item.go, networks.go
- [ ] item.go: client.Network → backend.Network (identical fields, no access changes)
- [ ] networks.go: GetNetworks → ListNetworks, GetAllNetworkUsage stays same name, inspection type → backend.NetworkDetail, all GetClient() → GetBackend()
- [ ] Run: go build ./internal/ui/networks/... && go test ./internal/ui/networks/...
Commit: `migrate networks tab to backend.Backend`

---

## Task 10: Migrate Volumes Tab
Files: internal/ui/volumes/item.go, volumes.go
- [ ] item.go: client.Volume → backend.Volume (backend.Volume adds CreatedAt time.Time)
- [ ] volumes.go: GetVolumes → ListVolumes, all GetClient() → GetBackend(), inspection type → backend.VolumeDetail
- [ ] Run: go build ./internal/ui/volumes/... && go test ./internal/ui/volumes/...
Commit: `migrate volumes tab to backend.Backend`

---

## Task 11: Migrate Browse Tab
Files: internal/ui/browse/browse.go
- [ ] state.GetClient().GetRegistryClient() → state.GetRegistryClient()
- [ ] state.GetClient().GetQuayRegistryClient() → state.GetQuayRegistryClient()
- [ ] state.GetClient().PullImage(...) → state.GetBackend().PullImage(...)
- [ ] Run: go build ./internal/ui/browse/... && go test ./internal/ui/browse/...
Commit: `migrate browse tab to backend.Backend`

---

## Task 12: Update Infopanel Builders
Files: internal/ui/components/infopanel/builders/builders.go
- [ ] Update all function signatures to accept backend.*Detail types:
  - BuildContainerPanel(container backend.ContainerDetail, ...)
  - BuildImagePanel(image backend.ImageDetail, ...)
  - BuildNetworkPanel(network backend.NetworkDetail, ...)
  - BuildVolumePanel(vol backend.VolumeDetail, ...)
- [ ] Field names in backend.*Detail match Docker SDK types closely — verify each builder's field accesses compile
- [ ] Run: go build ./internal/ui/components/... && go test ./internal/ui/components/...
Commit: `update infopanel builders to accept backend detail types`

---

## Task 13: Final Build, Test, Delete internal/client
- [ ] go build ./... — must be zero errors
- [ ] grep -rn "givensuman/containertui/internal/client" . --include="*.go" — must be zero results
- [ ] rm -rf internal/client/
- [ ] go build ./... — must still pass
- [ ] go test ./... — all pass
Commit: `delete internal/client after completing backend migration`
