# Context Migration Guide

## Status

The client package is being migrated to use `context.Context` for all Docker API calls.

## Completed Methods

- ✅ GetContainers
- ✅ GetImages  
- ✅ GetNetworks
- ✅ GetVolumes
- ✅ GetContainerState
- ✅ PauseContainer
- ✅ PauseContainers (now returns MultiError)
- ✅ UnpauseContainer
- ✅ UnpauseContainers (now returns MultiError)

## Methods Still Needing Context Parameter

The following methods still use `context.Background()` and need to be updated:

- StartContainer
- StartContainers
- StopContainer
- StopContainers
- RemoveContainer
- RemoveContainers
- RestartContainer
- RestartContainers
- GetServices
- OpenLogs
- ExecShell
- RemoveImage
- RemoveVolume
- RemoveNetwork
- PruneImages
- PruneVolumes
- PruneNetworks
- GetContainersUsing* methods
- GetContainerStats
- Inspect* methods
- PullImage
- CreateContainer

## Pattern to Follow

### Before:
```go
func (cw *ClientWrapper) SomeMethod(id string) error {
    err := cw.client.SomeCall(context.Background(), id)
    if err != nil {
        return fmt.Errorf("failed: %w", err)
    }
    return nil
}
```

### After:
```go
func (cw *ClientWrapper) SomeMethod(ctx context.Context, id string) error {
    err := cw.client.SomeCall(ctx, id)
    if err != nil {
        return fmt.Errorf("failed: %w", err)
    }
    return nil
}
```

### For Batch Operations:
```go
func (cw *ClientWrapper) SomeMethods(ctx context.Context, ids []string) error {
    var multiErr MultiError
    for _, id := range ids {
        if err := cw.SomeMethod(ctx, id); err != nil {
            multiErr.Add(id, err)
        }
    }
    return multiErr.ToError()
}
```

## Calling with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), client.DefaultOperationTimeout)
defer cancel()

err := clientWrapper.SomeMethod(ctx, id)
```
