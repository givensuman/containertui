# CONTAINERTUI Architecture

## Table of Contents

1. [Overview](#overview)
2. [MVU Pattern](#mvu-pattern)
3. [Component Architecture](#component-architecture)
4. [View Organization](#view-organization)
5. [State Management](#state-management)
6. [Message Flow](#message-flow)
7. [Adding a New Resource View](#adding-a-new-resource-view)
8. [Best Practices](#best-practices)

---

## Overview

CONTAINERTUI is a terminal-based container management interface built using the Charm stack:
- **Bubble Tea v2** (`charm.land/bubbletea/v2`) - TUI framework implementing the MVU (Model-View-Update) pattern
- **Bubbles v2** (`charm.land/bubbles/v2`) - Reusable UI components (lists, viewports, etc.)
- **Lipgloss v2** (`charm.land/lipgloss/v2`) - Styling and layout
- **Cobra** (`github.com/spf13/cobra`) - CLI framework

The architecture emphasizes:
- **Generic components** for code reusability
- **Consistent UX** across all resource types
- **Message-based communication** for decoupled state management
- **Type safety** through Go generics

---

## MVU Pattern

CONTAINERTUI follows the Model-View-Update (MVU) architectural pattern, also known as The Elm Architecture:

```
┌─────────────────────────────────────────┐
│                                         │
│  ┌──────┐    ┌────────┐    ┌──────┐   │
│  │      │───▶│        │───▶│      │   │
│  │ View │    │ Update │    │ Model│   │
│  │      │◀───│        │◀───│      │   │
│  └──────┘    └────────┘    └──────┘   │
│      │            ▲            │       │
│      │            │            │       │
│      └────────────┴────────────┘       │
│           User Events / Messages       │
│                                         │
└─────────────────────────────────────────┘
```

### Key Concepts

1. **Model**: Immutable state of the application
2. **View**: Pure function that renders the model to a string
3. **Update**: Pure function that receives messages and returns a new model
4. **Commands**: Side effects that produce messages

### MVU in Practice

Every component in CONTAINERTUI implements the `tea.Model` interface:

```go
type Model interface {
    Init() tea.Cmd              // Initialize and return setup commands
    Update(tea.Msg) (tea.Model, tea.Cmd)  // Process messages and return new state
    View() string               // Render the current state
}
```

**Example Message Flow:**
1. User presses "d" to delete a container
2. Update() receives `tea.KeyPressMsg`
3. Update() returns a command to show confirmation dialog
4. Dialog sends `base.SmartConfirmationMessage` when confirmed
5. Update() receives confirmation and returns delete command
6. Delete operation completes and sends result message
7. Update() processes result and refreshes view

---

## Component Architecture

CONTAINERTUI uses a layered component architecture with increasing levels of abstraction:

```
┌─────────────────────────────────────────────────────────┐
│  View Models (containers, images, volumes, etc.)        │
│  - Embed ResourceView                                    │
│  - Add domain-specific logic                            │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  ResourceView[ID, Item] - Generic Resource Manager       │
│  - List/Detail split view pattern                       │
│  - Selection management                                  │
│  - Refresh & state management                           │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│  base.Component - Window Dimension Tracking             │
│  - WindowWidth, WindowHeight                            │
└─────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. `base.Component`
**File**: `internal/ui/base/types.go`

The foundation component that tracks window dimensions:

```go
type Component struct {
    WindowWidth  int
    WindowHeight int
}
```

All views inherit these fields through embedding, ensuring consistent window management.

#### 2. `ResourceView[ID, Item]`
**File**: `internal/ui/components/resource_view.go`

A generic component providing standard list/detail view functionality for any resource type:

```go
type ResourceView[ID comparable, Item list.Item] struct {
    base.Component              // Embedded for window dimensions
    SplitView       SplitView   // List and detail pane layout
    Selections      *SelectionManager[ID]  // Multi-select support
    SessionState    SessionState           // Main view vs overlay
    DetailsKeyBinds DetailsKeybindings     // Standard detail pane keys
    Foreground      any                    // Overlay content (dialogs)
    
    Title          string
    AdditionalHelp []key.Binding
    
    // Resource-specific functions
    LoadItems     func() ([]Item, error)
    GetItemID     func(Item) ID
    GetItemTitle  func(Item) string
    IsItemWorking func(Item) bool
    OnResize      func(w, h int)
}
```

**Key Features:**
- Generic over resource ID type and item type
- Handles list rendering, scrolling, and selection
- Manages split view between list and detail panes
- Provides focus switching between panes
- Supports overlays (dialogs, forms)
- Built-in multi-selection with visual feedback

#### 3. `SplitView`
**File**: `internal/ui/components/splitview.go`

Manages the list/detail split layout:

```go
type SplitView struct {
    list        *list.Model
    rightPane   Pane
    ratio       float64  // List:Detail width ratio
    listFocused bool
    // ...
}
```

**Responsibilities:**
- Calculate and maintain split widths
- Handle focus switching between panes
- Render bordered panes with titles
- Support different right pane types (viewport, custom content)

#### 4. `DetailsPanel`
**File**: `internal/ui/components/details_panel.go`

Reusable component for detail pane functionality:

```go
type DetailsPanel struct {
    currentFormat   string           // "json" or "yaml"
    scrollPositions map[string]int   // Per-resource scroll memory
    currentID       string           // Currently displayed resource
}
```

**Features:**
- Format toggling (JSON ↔ YAML)
- Clipboard copying with error handling
- Scroll position memory per resource
- Consistent behavior across all views

#### 5. `SelectionManager[ID]`
**File**: `internal/ui/components/selection.go`

Handles multi-selection state:

```go
type SelectionManager[ID comparable] struct {
    selections map[ID]bool
}
```

Provides methods for toggling, clearing, and checking selections.

---

## View Organization

All resource views follow a consistent structure:

```
internal/ui/
├── base/               # Base types and messages
│   └── types.go       # Component, messages, action types
├── components/        # Reusable components
│   ├── resource_view.go    # Generic resource view
│   ├── splitview.go        # Split layout
│   ├── details_panel.go    # Detail pane logic
│   ├── selection.go        # Multi-select
│   ├── dialog.go          # Confirmation dialogs
│   └── ...
├── containers/        # Container resource view
│   ├── containers.go  # Model, Update, View
│   ├── messages.go    # View-specific messages
│   └── items.go       # Container list items
├── images/            # Image resource view (same structure)
├── volumes/           # Volume resource view (same structure)
├── networks/          # Network resource view (same structure)
├── services/          # Services resource view (same structure)
├── browse/            # Browse/search view (same structure)
└── ui.go             # Top-level UI coordinator
```

### Resource View Structure

Each resource view follows this pattern:

**1. Model Definition**
```go
type Model struct {
    components.ResourceView[string, ContainerItem]  // Embed ResourceView
    keybindings        *keybindings                  // Custom keybindings
    detailsKeybindings detailsKeybindings            // Detail pane keys
    detailsPanel       components.DetailsPanel       // Detail logic
    inspection         types.ContainerJSON           // Current inspection data
}
```

**2. Constructor** (`New()`)
- Create ResourceView with LoadItems function
- Initialize DetailsPanel
- Set up keybindings
- Configure additional help

**3. Update Method**
- Handle custom keybindings
- Delegate to ResourceView.Update() for standard behavior
- Process domain-specific messages
- Manage dialog interactions

**4. View Method**
- Delegate to ResourceView.View() for rendering

**5. Helper Methods**
- Resource-specific operations (start, stop, delete, etc.)
- Detail content builders
- Inspection fetching

### Common Patterns

All views implement these standard behaviors:

| Feature | Implementation |
|---------|----------------|
| **List Navigation** | ↑/↓ arrows, j/k, page up/down |
| **Focus Switching** | Tab key (list ↔ detail) |
| **Multi-select** | Space to toggle, Shift+D to delete selected |
| **Detail Scrolling** | ↑/↓ when detail focused |
| **Format Toggle** | J key (JSON ↔ YAML) |
| **Clipboard Copy** | y key to copy detail content |
| **Refresh** | r key to reload list |
| **Filter** | / key to filter list |
| **Delete** | d key for single, Shift+D for selected |
| **Dialogs** | Confirmation for destructive actions |

---

## State Management

CONTAINERTUI uses a hybrid state management approach:

### 1. Global State (`internal/state`)

Singleton pattern for application-wide state:

```go
// Global state access
client := state.GetClient()      // Docker/Podman client
config := state.GetConfig()      // User configuration
```

**What's Global:**
- Backend client (Docker/Podman)
- User configuration (theme, keybindings)
- Registry client (Docker Hub)

### 2. Local View State

Each view maintains its own state in its Model struct:

```go
type Model struct {
    ResourceView[...]  // List items, selection, scroll position
    detailsPanel       // Format preference, scroll positions
    inspection         // Current resource details
    // View-specific fields
}
```

**What's Local:**
- List of resources
- Selected item(s)
- Detail pane content
- Scroll positions
- Active dialogs/overlays

### 3. Cross-View Communication

Views communicate through messages rather than shared mutable state:

```go
// A container is created in the Browse tab
return tea.Batch(
    notifications.ShowSuccess("Container created"),
    func() tea.Msg {
        return base.MsgResourceChanged{
            Resource:  base.ResourceContainer,
            Operation: base.OperationCreated,
            IDs:       []string{containerID},
        }
    },
)
```

The top-level UI model receives `MsgResourceChanged` and triggers refreshes on affected tabs.

---

## Message Flow

### Message Types

#### System Messages (from Bubble Tea)
- `tea.KeyPressMsg` - Keyboard input
- `tea.MouseClickMsg` - Mouse clicks
- `tea.MouseMotionMsg` - Mouse movement
- `tea.WindowSizeMsg` - Terminal resize

#### Application Messages

**Base Messages** (`internal/ui/base/types.go`):
```go
// Dialog interaction
type SmartConfirmationMessage struct {
    Action SmartDialogAction
}

type CloseDialogMessage struct{}

// Focus switching
type MsgFocusChanged struct {
    IsDetailsFocused bool
}

// Cross-tab updates
type MsgResourceChanged struct {
    Resource  ResourceType   // container, image, volume, network
    Operation OperationType  // created, deleted, updated, pruned
    IDs       []string
    Metadata  map[string]any
}
```

**View-Specific Messages**:
```go
// Containers view
type MsgContainerInspection struct {
    ID   string
    Data types.ContainerJSON
    Err  error
}

type MsgContainerOperationResult struct {
    Operation Operation  // Start, Stop, Restart, etc.
    ID        string
    Error     error
}
```

### Message Processing Flow

```
User Input (KeyPress)
    │
    ▼
Top-level Update (ui.go)
    │
    ├─▶ Global handlers (quit, tab switch)
    │
    └─▶ Active view Update()
            │
            ├─▶ Handle view-specific keys
            │   └─▶ Return commands (side effects)
            │
            ├─▶ Delegate to ResourceView.Update()
            │   ├─▶ Standard navigation
            │   ├─▶ Selection management
            │   └─▶ Focus switching
            │
            └─▶ Handle async results
                └─▶ Update state, show notifications

Commands execute async
    │
    ▼
Return messages
    │
    ▼
Back to Update() for processing
```

### Example: Deleting a Container

```
1. User presses "d"
   ├─▶ containers.Update() receives tea.KeyPressMsg
   └─▶ Shows confirmation dialog, returns DialogVisible command

2. User presses "y" to confirm
   ├─▶ containers.Update() receives SmartConfirmationMessage
   └─▶ Calls PerformContainerOperation(Remove, containerID)
       └─▶ Returns async command

3. Command executes (calls Docker API)
   └─▶ Returns MsgContainerOperationResult

4. containers.Update() receives MsgContainerOperationResult
   ├─▶ If error: shows error notification
   ├─▶ If success: shows success notification
   ├─▶ Sends MsgResourceChanged to refresh other views
   └─▶ Triggers Refresh() to reload list

5. Top-level Update receives MsgResourceChanged
   └─▶ Triggers refresh on Networks tab (containers affect networks)
```

---

## Adding a New Resource View

Follow these steps to add a new resource type (e.g., "Secrets"):

### Step 1: Define Item Type

Create `internal/ui/secrets/items.go`:

```go
package secrets

import "github.com/givensuman/containertui/internal/client"

type SecretItem struct {
    Secret client.Secret
}

func (s SecretItem) FilterValue() string {
    return s.Secret.Name
}

func (s SecretItem) Title() string {
    return s.Secret.Name
}

func (s SecretItem) Description() string {
    return fmt.Sprintf("Updated: %s", s.Secret.UpdatedAt)
}
```

### Step 2: Create Model

Create `internal/ui/secrets/secrets.go`:

```go
package secrets

import (
    "github.com/givensuman/containertui/internal/ui/components"
    "github.com/givensuman/containertui/internal/ui/base"
)

type Model struct {
    components.ResourceView[string, SecretItem]
    keybindings        *keybindings
    detailsKeybindings detailsKeybindings
    detailsPanel       components.DetailsPanel
    inspection         client.Secret
}

func New() Model {
    // Create fetch function
    fetchSecrets := func() ([]SecretItem, error) {
        secrets, err := state.GetClient().GetSecrets(context.Background())
        if err != nil {
            return []SecretItem{}, nil
        }
        
        items := make([]SecretItem, 0, len(secrets))
        for _, secret := range secrets {
            items = append(items, SecretItem{Secret: secret})
        }
        return items, nil
    }
    
    // Create ResourceView
    resourceView := components.NewResourceView[string, SecretItem](
        "Secrets",
        fetchSecrets,
        func(item SecretItem) string { return item.Secret.ID },
        func(item SecretItem) string { return item.Title() },
        func(w, h int) {
            // Optional resize handler
        },
    )
    
    // Create model
    model := Model{
        ResourceView:       *resourceView,
        keybindings:        newKeybindings(),
        detailsKeybindings: components.NewDetailsKeybindings(),
        detailsPanel:       components.NewDetailsPanel(),
    }
    
    return model
}

func (m Model) Init() tea.Cmd {
    return m.ResourceView.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Handle custom keys
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if m.IsListFocused() {
            switch {
            case key.Matches(msg, m.keybindings.inspect):
                return m, m.inspectSecret()
            case key.Matches(msg, m.keybindings.delete):
                return m, m.showDeleteDialog()
            }
        }
    }
    
    // Delegate to ResourceView
    updatedModel, cmd := m.ResourceView.Update(msg)
    m.ResourceView = updatedModel.(components.ResourceView[string, SecretItem])
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m Model) View() string {
    return m.ResourceView.View()
}
```

### Step 3: Add to Top-Level UI

Edit `internal/ui/ui.go`:

1. Add to Model:
```go
type Model struct {
    // ... existing fields
    secretsModel secrets.Model
}
```

2. Initialize in New():
```go
secretsModel: secrets.New(),
```

3. Add tab in `internal/ui/tabs/tabs.go`:
```go
const (
    // ... existing tabs
    Secrets
)
```

4. Handle in Update/View switches in `ui.go`

### Step 4: Add Keybindings

Create keybindings in `secrets.go`:

```go
type keybindings struct {
    inspect key.Binding
    delete  key.Binding
}

func newKeybindings() *keybindings {
    return &keybindings{
        inspect: key.NewBinding(
            key.WithKeys("enter"),
            key.WithHelp("enter", "inspect"),
        ),
        delete: key.NewBinding(
            key.WithKeys("d"),
            key.WithHelp("d", "delete"),
        ),
    }
}
```

### Step 5: Implement Operations

Add helper methods:

```go
func (m *Model) inspectSecret() tea.Cmd {
    selectedItem := m.GetSelectedItem()
    if selectedItem == nil {
        return nil
    }
    
    return func() tea.Msg {
        secret, err := state.GetClient().InspectSecret(
            context.Background(),
            selectedItem.Secret.ID,
        )
        return MsgSecretInspection{
            ID:   selectedItem.Secret.ID,
            Data: secret,
            Err:  err,
        }
    }
}

func (m *Model) showDeleteDialog() tea.Cmd {
    selectedItem := m.GetSelectedItem()
    if selectedItem == nil {
        return nil
    }
    
    dialog := components.NewSmartDialog(
        "Delete Secret",
        fmt.Sprintf("Delete secret '%s'?", selectedItem.Secret.Name),
        base.SmartDialogAction{
            Type:    "delete_secret",
            Payload: selectedItem.Secret.ID,
        },
    )
    
    m.ShowOverlay(dialog)
    return nil
}
```

### Step 6: Handle Results

Add message handlers:

```go
case MsgSecretInspection:
    m.inspection = msg.Data
    m.refreshSecretDetails()
    
case base.SmartConfirmationMessage:
    if msg.Action.Type == "delete_secret" {
        secretID := msg.Action.Payload.(string)
        return m, m.deleteSecret(secretID)
    }
```

---

## Best Practices

### 1. Follow the ResourceView Pattern

✅ **Do**: Use ResourceView for list/detail views
```go
type Model struct {
    components.ResourceView[string, MyItem]
    // ... custom fields
}
```

❌ **Don't**: Reimplement list/detail logic
```go
type Model struct {
    list      list.Model
    viewport  viewport.Model
    // Manual focus management, etc.
}
```

### 2. Use DetailsPanel for Detail Panes

✅ **Do**: Use DetailsPanel for format toggle and clipboard
```go
model.detailsPanel.HandleToggleFormat()
model.detailsPanel.HandleCopyToClipboard(data)
```

❌ **Don't**: Implement custom format toggle logic
```go
if model.currentFormat == "json" {
    model.currentFormat = "yaml"
} else {
    // ...
}
```

### 3. Send MsgResourceChanged for Cross-Tab Updates

✅ **Do**: Notify other tabs of changes
```go
return tea.Batch(
    notifications.ShowSuccess("Container deleted"),
    func() tea.Msg {
        return base.MsgResourceChanged{
            Resource:  base.ResourceContainer,
            Operation: base.OperationDeleted,
            IDs:       []string{containerID},
        }
    },
)
```

❌ **Don't**: Let views get stale
```go
return notifications.ShowSuccess("Container deleted")
// Networks tab still shows the deleted container's networks
```

### 4. Always Show User Feedback

✅ **Do**: Return notification commands
```go
if err != nil {
    return notifications.ShowError(fmt.Errorf("failed to start: %w", err))
}
return notifications.ShowSuccess("Container started")
```

❌ **Don't**: Fail silently
```go
if err != nil {
    return nil  // User has no idea what happened
}
return nil
```

### 5. Use Proper Error Wrapping

✅ **Do**: Wrap errors with context
```go
if err := client.StartContainer(ctx, id); err != nil {
    return fmt.Errorf("failed to start container %s: %w", id, err)
}
```

❌ **Don't**: Return raw errors
```go
return err  // No context about what failed
```

### 6. Maintain Scroll Position Across Resources

✅ **Do**: Use DetailsPanel scroll memory
```go
model.detailsPanel.SetCurrentID(newID, viewport)
model.detailsPanel.RestoreScrollPosition(viewport)
```

❌ **Don't**: Reset scroll on every selection change
```go
viewport.SetYOffset(0)  // Jumps to top every time
```

### 7. Consistent Keybindings

Follow these conventions:

| Key | Action |
|-----|--------|
| `↑/↓`, `j/k` | Navigate list |
| `enter` | Inspect/Open |
| `tab` | Switch focus |
| `d` | Delete single |
| `shift+d` | Delete selected |
| `r` | Refresh |
| `y` | Copy to clipboard |
| `J` | Toggle JSON/YAML |
| `/` | Filter |
| `space` | Toggle selection |
| `esc` | Close dialog/filter |

### 8. Use Type-Safe Generics

✅ **Do**: Leverage generics for type safety
```go
ResourceView[string, ContainerItem]  // ID is string, Item is ContainerItem
```

❌ **Don't**: Use interface{} or any
```go
ResourceView[interface{}, interface{}]  // Lost type safety
```

### 9. Keep Views Stateless (Where Possible)

✅ **Do**: Fetch data when needed
```go
LoadItems: func() ([]Item, error) {
    return client.GetContainers(ctx)
}
```

❌ **Don't**: Cache stale data
```go
var cachedContainers []Container  // May be outdated
```

### 10. Test Window Resizing

Always test that your view handles:
- Window resize (terminal dimension changes)
- Detail pane width adjustments
- Long content that needs scrolling
- Empty states

---

## Testing Strategy

### Manual Testing Checklist

For each new view, verify:

- [ ] List displays correctly with items
- [ ] Empty state shows placeholder
- [ ] Navigation keys work (↑/↓, j/k, page up/down)
- [ ] Enter key inspects/opens item
- [ ] Tab key switches between list and detail
- [ ] Detail pane shows correct content
- [ ] J key toggles format (JSON ↔ YAML)
- [ ] y key copies to clipboard
- [ ] Clipboard errors show notifications
- [ ] Delete shows confirmation dialog
- [ ] Delete operation works correctly
- [ ] Multi-select with space key works
- [ ] Shift+D deletes all selected items
- [ ] Refresh (r key) reloads list
- [ ] Filter (/ key) filters correctly
- [ ] Window resize adjusts layout properly
- [ ] Scroll position persists when switching items
- [ ] Success/error notifications appear
- [ ] Cross-tab updates work (if applicable)

### Unit Testing

Focus on:
- Item FilterValue/Title/Description methods
- Helper functions (formatters, parsers)
- Message constructors
- Business logic (non-UI code)

Example:
```go
func TestContainerItem_FilterValue(t *testing.T) {
    item := ContainerItem{
        Container: types.Container{
            Names: []string{"/my-container"},
        },
    }
    
    if item.FilterValue() != "my-container" {
        t.Errorf("expected 'my-container', got '%s'", item.FilterValue())
    }
}
```

### Integration Testing Considerations

For integration tests:
- Use Docker/Podman test containers
- Mock the backend client for predictable behavior
- Test message flow through components
- Verify state transitions

---

## Key Architecture Insights

### Strengths

1. **Generic ResourceView Pattern**: Eliminates duplication, consistent UX
2. **MVU Pattern Adherence**: Clean separation of concerns, testable
3. **Message-Based Communication**: Decoupled, easy to reason about
4. **Component Reusability**: DetailsPanel, SplitView, SelectionManager
5. **Type Safety**: Generics prevent many runtime errors

### Trade-offs

1. **Learning Curve**: MVU pattern requires different thinking than imperative UI
2. **Verbosity**: Go's lack of union types means many type switches
3. **Async Complexity**: Commands and messages can be hard to trace
4. **Limited Layout Options**: TUI constraints vs GUI flexibility

### Future Considerations

- **Performance**: Large lists may need virtualization
- **Accessibility**: Screen reader support for TUI
- **Testing**: Consider property-based testing for message handling
- **Documentation**: Keep this document updated as architecture evolves

---

## Additional Resources

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [The Elm Architecture](https://guide.elm-lang.org/architecture/)
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [Project README](./README.md)
- [Refactoring Progress](./REFACTORING_PROGRESS.md)

---

**Document Version**: 1.0  
**Last Updated**: February 13, 2026  
**Maintainer**: CONTAINERTUI Team
