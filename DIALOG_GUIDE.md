# Dialog System Guide

## Overview

The dialog system provides a consistent way to display modal overlays throughout CONTAINERTUI. It includes support for:

- **Semantic types**: Info, Success, Warning, Error (with color-coded borders)
- **Multiple sizes**: Small (30%×15%), Medium (40%×20%), Large (60%×40%)
- **Consistent button styling**: All dialogs use centered buttons that won't wrap
- **Proper text alignment**: Messages and buttons are properly centered within the dialog
- **Form dialogs with visible buttons**: FormDialog now displays centered Cancel/Submit buttons

## Dialog Types

### Basic Dialog

The standard `NewDialog()` function creates a medium-sized info dialog:

```go
dialog := components.NewDialog(
    "Are you sure you want to delete this container?",
    []components.DialogButton{
        {Label: "Cancel"},
        {Label: "Delete", Action: base.SmartDialogAction{Type: "DeleteContainer", Payload: containerID}},
    },
)
model.ResourceView.SetOverlay(dialog)
```

### Semantic Dialog Variants

Use semantic constructors for color-coded borders:

#### Info Dialog (Blue border)
```go
dialog := components.NewInfoDialog(
    "This operation will take a few moments...",
    []components.DialogButton{{Label: "OK"}},
)
```

#### Success Dialog (Green border)
```go
dialog := components.NewSuccessDialog(
    "Container created successfully!\n\nContainer ID: abc123",
)
// Note: Success dialogs default to a single "OK" button
```

#### Warning Dialog (Yellow border)
```go
dialog := components.NewWarningDialog(
    "This image is used by 3 containers.\nAre you sure you want to delete it?",
    []components.DialogButton{
        {Label: "Cancel"},
        {Label: "Delete Anyway", Action: deleteAction},
    },
)
```

#### Error Dialog (Red border)
```go
dialog := components.NewErrorDialog(
    "Failed to pull image:\n\nconnection timeout",
)
// Note: Error dialogs default to a single "OK" button
```

### Convenience Constructors

#### Confirmation Dialog
```go
dialog := components.NewConfirmDialog(
    "Are you sure you want to proceed?",
    base.SmartDialogAction{Type: "ConfirmAction"},
)
// Creates a dialog with "Cancel" and "Confirm" buttons
```

#### Delete Confirmation Dialog
```go
dialog := components.NewDeleteDialog(
    fmt.Sprintf("Delete container %s?", name),
    base.SmartDialogAction{Type: "DeleteContainer", Payload: id},
)
// Creates a warning-styled dialog with "Cancel" and "Delete" buttons
```

## Dialog Sizes

### Using Custom Sizes

For dialogs that need more or less space:

```go
// Small dialog (30% width × 15% height)
dialog := components.NewDialogWithOptions(
    "Quick confirmation?",
    buttons,
    components.DialogSizeSmall,
    components.DialogTypeInfo,
)

// Medium dialog (40% width × 20% height) - DEFAULT
dialog := components.NewDialogWithOptions(
    "Standard message",
    buttons,
    components.DialogSizeMedium,
    components.DialogTypeInfo,
)

// Large dialog (60% width × 40% height)
dialog := components.NewDialogWithOptions(
    "Detailed information with lots of content...",
    buttons,
    components.DialogSizeLarge,
    components.DialogTypeInfo,
)
```

## Form Dialogs

FormDialog now includes visible Cancel/Submit buttons and uses consistent styling with Dialog.

### Basic Form Dialog

```go
formDialog := components.NewFormDialog(
    "Pull Image",
    []components.FormField{
        {
            Label:       "Image",
            Placeholder: "nginx:latest",
            Required:    true,
            Validator:   validateImageName,
        },
    },
    base.SmartDialogAction{Type: "PullImage"},
    nil,
)
model.ResourceView.SetOverlay(formDialog)
```

### Form Dialog with Multiple Fields

```go
formDialog := components.NewFormDialog(
    "Create Container",
    []components.FormField{
        {Label: "Name", Placeholder: "my-container", Required: false},
        {Label: "Ports", Placeholder: "8080:80", Required: false, Validator: validatePorts},
        {Label: "Volumes", Placeholder: "/host:/container", Required: false, Validator: validateVolumes},
        {Label: "Environment", Placeholder: "KEY=value", Required: false, Validator: validateEnv},
    },
    base.SmartDialogAction{Type: "CreateContainer", Payload: metadata},
    nil,
)
```

### Form Dialog with Custom Size

```go
// Large form for complex inputs
formDialog := components.NewFormDialogWithSize(
    "Advanced Configuration",
    fields,
    action,
    metadata,
    components.DialogSizeLarge,
)
```

### Form Dialog Navigation

Users can navigate form dialogs using:
- **Tab/Down**: Move to next field or button
- **Shift+Tab/Up**: Move to previous field or button
- **Left/Right** (on buttons): Switch between Cancel and Submit
- **Enter** (on field): Move to next field
- **Enter** (on button): Activate button
- **Esc**: Close dialog

## Button Design

All dialogs now use consistent button styling:

- **Default button**: Muted background, regular text
- **Selected/Hovered button**: Primary color background, bold text
- **Horizontal layout**: Buttons arranged left to right
- **Standardized spacing**: Consistent padding and margins

### Button Order Convention

For destructive actions, follow this pattern:
- **Safe action (Cancel) on the left** - default selected
- **Destructive action (Delete) on the right**

```go
[]components.DialogButton{
    {Label: "Cancel"},              // Left, selected by default (index 0)
    {Label: "Delete", Action: ...}, // Right
}
```

This ensures users must explicitly navigate to the destructive action.

## Dialog Actions

### Using SmartDialogAction

```go
{
    Label: "Confirm",
    Action: base.SmartDialogAction{
        Type:    "ActionType",
        Payload: data,
    },
}
```

### Using DialogAction

```go
{
    Label: "Close",
    Action: base.NewDialogAction[any](base.CloseDialog, nil),
}
```

### Default Behavior

If `Action` is `nil` or not specified, the button will close the dialog:

```go
{Label: "OK"} // Automatically closes dialog
```

## Migration Guide

### Updating Existing Dialogs

#### Before (Old Pattern)
```go
errorDialog := components.NewDialog(
    fmt.Sprintf("Failed to delete:\n\n%v", err),
    []components.DialogButton{{Label: "OK"}},
)
```

#### After (New Pattern - Recommended)
```go
errorDialog := components.NewErrorDialog(
    fmt.Sprintf("Failed to delete:\n\n%v", err),
)
```

#### Before (Old Pattern)
```go
successDialog := components.NewDialog(
    "Operation completed successfully!",
    []components.DialogButton{{Label: "OK"}},
)
```

#### After (New Pattern - Recommended)
```go
successDialog := components.NewSuccessDialog(
    "Operation completed successfully!",
)
```

#### Before (Old Pattern)
```go
confirmDialog := components.NewDialog(
    "Delete this resource?",
    []components.DialogButton{
        {Label: "Cancel"},
        {Label: "Delete", Action: deleteAction},
    },
)
```

#### After (New Pattern - Recommended)
```go
confirmDialog := components.NewDeleteDialog(
    "Delete this resource?",
    deleteAction,
)
```

### Backwards Compatibility

The old `NewDialog()` function still works and is now equivalent to creating a medium-sized info dialog. All existing code will continue to function without changes.

## Best Practices

1. **Use semantic variants**: Choose `NewSuccessDialog`, `NewErrorDialog`, `NewWarningDialog`, or `NewInfoDialog` based on the message type.

2. **Choose appropriate sizes**:
   - Small: Quick confirmations, simple messages
   - Medium: Standard dialogs (default)
   - Large: Complex forms, detailed information

3. **Use convenience constructors**: `NewConfirmDialog` and `NewDeleteDialog` provide consistent patterns for common operations.

4. **Provide clear messages**: Use concise, actionable text. Include newlines (`\n\n`) to separate sections.

5. **Validate form inputs**: Always provide validators for FormDialog fields to catch errors before submission.

6. **Handle window resize**: Dialogs automatically adjust to window size changes.

## Examples

### Progress Dialog
```go
progressDialog := components.NewInfoDialog(
    "Pulling image: nginx\n\nThis may take a few moments...",
    []components.DialogButton{}, // Empty = no buttons, no interaction
)
model.ResourceView.SetOverlay(progressDialog)
```

### Multi-line Warning with List
```go
warningDialog := components.NewWarningDialog(
    fmt.Sprintf("Cannot delete volume %s.\n\nUsed by containers:\n• %s\n• %s",
        volumeName, container1, container2),
    []components.DialogButton{{Label: "OK"}},
)
```

### Confirmation with Context
```go
deleteDialog := components.NewDeleteDialog(
    fmt.Sprintf("Delete %d selected containers?\n\n%s",
        count, strings.Join(names, "\n")),
    base.SmartDialogAction{Type: "DeleteMultiple", Payload: ids},
)
```
