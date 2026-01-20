# Implementation Plan: Lazydocker-Style Raw Inspection Panel

## Project Goal

Replace structured, opinionated information panels with lazydocker-style raw inspection output:
- Summary section with key fields (no borders, simple padding)
- Full YAML/JSON dump of entire `docker inspect` output
- Syntax highlighting
- Clipboard copy support
- Remove CPU stats/graphs

## User Decisions (Final)

1. **Stats Display:** Remove ALL stats (pure lazydocker replication)
2. **Syntax Highlighting:** Map to existing theme colors (Primary/Success/Warning/Muted)
3. **Connected Resources:** Multi-line bulleted format
4. **Error Handling:** Show error message only if marshaling fails
5. **Services Panel:** Apply same raw approach (show compose data)
6. **Implementation Order:** Complete containers first, verify, then other resources

## Current State Analysis

### Already Exists
- ✅ `gopkg.in/yaml.v3` dependency (v3.0.1)
- ✅ `InspectContainer()` method in `internal/client/client.go:644`
- ✅ Formatter utilities in `internal/ui/components/infopanel/formatters.go`
- ✅ Theme system with color config in `internal/config/theme.go`

### Missing
- ❌ No `InspectImage()`, `InspectNetwork()`, `InspectVolume()` methods
- ❌ No YAML marshaling infrastructure
- ❌ No syntax highlighting utilities
- ❌ No `--json` flag in CLI
- ❌ No `inspection-format` config option

## Implementation Phases

---

## PHASE 1: Core Infrastructure ✅

**Status:** COMPLETE
**Completed:** 2026-01-20

### Task 1.1: Create Marshaling & Syntax Highlighting Module ✅

**File:** `internal/ui/components/infopanel/marshal.go` (NEW)

Created with functionality:
- `OutputFormat` type (YAML/JSON)
- `GetOutputFormat()` - reads from config
- `MarshalToFormat()` - converts struct to YAML/JSON
- `ColorizeYAML()` - applies syntax highlighting
- `ColorizeJSON()` - applies syntax highlighting

Implementation uses:
- 2-space indentation (YAML standard)
- Syntax highlighting via regex + Lipgloss
- Color mapping: Keys (Primary), Strings (Success), Numbers (Warning), Booleans/null (Muted)

### Task 1.2: Add Config Support ✅

**File:** `internal/config/config.go` (MODIFIED)

Added to Config struct:
```go
InspectionFormat string `yaml:"inspection-format,omitempty"`
```

Updated `DefaultConfig()`:
```go
InspectionFormat: "yaml",
```

### Task 1.3: Add CLI Flag ✅

**File:** `cmd/main.go` (MODIFIED)

1. Added variable: `var jsonFormat bool`
2. Added flag: `rootCmd.Flags().BoolVar(&jsonFormat, "json", false, "use JSON format for inspection output")`
3. Applied in RunE: `if jsonFormat { cfg.InspectionFormat = "json" }`

---

## PHASE 2: Container Panel (Complete First) ✅

**Status:** COMPLETE
**Completed:** 2026-01-20

### Task 2.1: Rewrite Container Panel Builder ✅

**File:** `internal/ui/components/infopanel/builders/builders.go` (MODIFIED)

**Replaced** `BuildContainerPanel()` function with new implementation:

**Summary Section Fields (in order):**
1. ID (12 chars)
2. Name
3. Image
4. Command
5. Labels (multi-line, indented)
6. Mounts (multi-line, bulleted)
7. Ports (multi-line, bulleted)
8. Networks (multi-line, bulleted) ← Connected Resources
9. Volumes (multi-line, bulleted) ← Connected Resources

**Format:** `"%-12s %s\n", label+":", value`

**Full Details Section:**
- Separator: `"\n\nFull details:\n\n"`
- Marshal entire `types.ContainerJSON` to YAML/JSON
- Apply syntax highlighting

**Helper Functions Added:**
- `formatSummaryField(label, value string) string`
- `formatCommand(c types.ContainerJSON) string`
- `formatLabels(labels map[string]string) string`
- `formatMounts(mounts []types.MountPoint) string`
- `formatPorts(ports nat.PortMap) string`
- `formatBulletList(items []string) string`
- `getConnectedResources(c types.ContainerJSON) (networks, volumes []string)`

### Task 2.2: Remove CPU Stats from Containers View ✅

**File:** `internal/ui/containers/containers.go` (MODIFIED)

**Completed Actions:**
1. ✅ Deleted `formatInspection()` function
2. ✅ Replaced `formatInspection()` calls with direct `builders.BuildContainerPanel()` calls
3. ✅ Removed CPU stats-related fields from Model struct:
   - Removed `cpuHistory []float64`
   - Removed `lastStats client.ContainerStats`
4. ✅ Removed stats ticker and related code:
   - Removed `tickCmd()` function
   - Removed `handleStatsTick()` function
   - Removed `MsgStatsTick` message type
   - Removed `MsgContainerStats` message type
   - Removed stats ticker from `Init()`
   - Removed stats message handlers from `Update()`
5. ✅ Cleaned up imports:
   - Removed unused `time` import
   - Removed unused `strings` import
   - Removed unused `client` import
   - Removed unused `infopanel` import

**Verification:** ✅ Project builds successfully

### Task 2.3: Add Keybindings and UX Improvements ✅

**Files Modified:**
- `internal/ui/containers/containers.go`
- `go.mod` (added clipboard dependency)

**Changes:**

1. **Added Keybindings:**
   - `J` - Toggle between JSON/YAML format
   - `y` - Copy inspection output to clipboard

2. **Added Scroll Position Tracking:**
   - Each container remembers its scroll position
   - When switching containers, scroll position is saved and restored
   - New containers start at the top (YOffset = 0)

3. **Added State Fields:**
   - `scrollPositions map[string]int` - Tracks scroll position per container ID
   - `currentFormat string` - Tracks user's format override (toggleable with 'J')

4. **Added Helper Functions:**
   - `getViewport()` - Gets the viewport from detail pane
   - `saveScrollPosition()` - Saves current scroll position
   - `restoreScrollPosition()` - Restores scroll position for current container
   - `refreshInspectionContent()` - Regenerates content with current format
   - `handleCopyToClipboard()` - Copies raw output to clipboard
   - `handleToggleFormat()` - Toggles between JSON and YAML

5. **Dependencies Added:**
   - `github.com/atotto/clipboard` - For clipboard functionality
   - `charm.land/bubbles/v2/viewport` - For viewport API access

6. **Keybinding Flow:**
   - When detail pane is focused, check for 'J' or 'y' keys
   - 'J' toggles format and refreshes content
   - 'y' marshals data and copies to clipboard
   - Both show notifications on success/error

**Verification:** ✅ Project builds successfully

---

## PHASE 3: Other Resource Types 🔄

**Status:** Pending
**Estimated Time:** 2-3 hours

### Task 3.1: Add Inspect Methods to Client

**File:** `internal/client/client.go` (MODIFY)

Add these methods (after existing `InspectContainer()`):

```go
// InspectImage returns detailed information about an image
func (clientWrapper *ClientWrapper) InspectImage(imageID string) (types.ImageInspect, error)

// InspectNetwork returns detailed information about a network
func (clientWrapper *ClientWrapper) InspectNetwork(networkID string) (types.NetworkResource, error)

// InspectVolume returns detailed information about a volume
func (clientWrapper *ClientWrapper) InspectVolume(volumeName string) (volume.Volume, error)
```

### Task 3.2: Rewrite Image Panel

**File:** `internal/ui/components/infopanel/builders/builders.go` (MODIFY)

**Replace** `BuildImagePanel()` (lines 103-179)

**Summary Fields:**
1. ID (12 chars)
2. Tags (comma-separated or multi-line if many)
3. Size (formatted bytes)
4. Created (timestamp)
5. Architecture
6. OS
7. Labels (multi-line, indented)
8. Used By (multi-line, bulleted) ← Connected Resources

**Full Details:** Marshal entire `types.ImageInspect`

### Task 3.3: Rewrite Network Panel

**File:** `internal/ui/components/infopanel/builders/builders.go` (MODIFY)

**Replace** `BuildNetworkPanel()` (lines 181-249)

**Summary Fields:**
1. ID (12 chars)
2. Name
3. Driver
4. Scope
5. Created
6. Subnet/Gateway (from IPAM config)
7. Labels (multi-line, indented)
8. Connected Containers (multi-line, bulleted with IPs) ← Connected Resources

**Full Details:** Marshal entire `types.NetworkResource`

### Task 3.4: Rewrite Volume Panel

**File:** `internal/ui/components/infopanel/builders/builders.go` (MODIFY)

**Replace** `BuildVolumePanel()` (lines 252-295)

**Summary Fields:**
1. Name
2. Driver
3. Mountpoint
4. Scope
5. Created
6. Labels (multi-line, indented)
7. Mounted By (multi-line, bulleted) ← Connected Resources

**Full Details:** Marshal entire `volume.Volume`

### Task 3.5: Rewrite Service Panel

**File:** `internal/ui/components/infopanel/builders/builders.go` (MODIFY)

**Replace** `BuildServicePanel()` (lines 298-339)

**Summary Fields:**
1. Name
2. Project
3. Replicas (count)
4. Compose File (path)
5. Containers (multi-line, bulleted with status badges)

**Full Details:** Marshal entire `client.Service` struct

---

## PHASE 4: Update View Controllers 🔜

**Status:** Pending
**Estimated Time:** 2 hours

### Task 4.1: Update Images View

**File:** `internal/ui/images/images.go` (MODIFY)

- Locate `updateDetailContent()` (around line 572)
- Replace `BuildSimpleImagePanel()` call with `BuildImagePanel()`
- Ensure it calls new `client.InspectImage()` method

### Task 4.2: Update Networks View

**File:** `internal/ui/networks/networks.go` (MODIFY)

- Locate `updateDetailContent()` (around line 208)
- Replace with call to new `BuildNetworkPanel()`
- Ensure it calls new `client.InspectNetwork()` method

### Task 4.3: Update Volumes View

**File:** `internal/ui/volumes/volumes.go` (MODIFY)

- Locate `updateDetailContent()` (around line 204)
- Replace with call to new `BuildVolumePanel()`
- Ensure it calls new `client.InspectVolume()` method

### Task 4.4: Update Services View

**File:** `internal/ui/services/services.go` (MODIFY)

- Locate `updateDetails()` (around line 197)
- Update to use new `BuildServicePanel()` implementation

---

## PHASE 5: Cleanup 🔜

**Status:** Pending
**Estimated Time:** 1 hour

### Task 5.1: Delete Obsolete Files

**Files to DELETE:**
```bash
rm internal/ui/components/infopanel/infopanel.go
rm internal/ui/components/infopanel/section.go
rm internal/ui/components/infopanel/field.go
rm internal/ui/components/infopanel/statspanel.go
```

**Files to KEEP:**
- `formatters.go` - Still used by summary sections
- `icons.go` - Still needed
- `builders/` directory - Core functionality

### Task 5.2: Update Imports

**Check and clean imports in:**
- `internal/ui/components/infopanel/builders/builders.go`
- `internal/ui/containers/containers.go`
- `internal/ui/images/images.go`
- `internal/ui/networks/networks.go`
- `internal/ui/volumes/volumes.go`
- `internal/ui/services/services.go`

Remove any references to deleted Panel/Section/Field types.

### Task 5.3: Verify Build

```bash
just build
```

Fix any compilation errors.

---

## PHASE 6: Clipboard Support 🔜

**Status:** Pending
**Estimated Time:** 1 hour

### Task 6.1: Add Clipboard Dependency

```bash
go get github.com/atotto/clipboard
```

### Task 6.2: Add Keybinding

**File:** `internal/config/registry.go` or keybindings config

Add action for 'y' key: "Copy inspection output to clipboard"

### Task 6.3: Implement Copy Handler

**In each view controller** (containers, images, networks, volumes, services):

1. Add handler for 'y' key press
2. Get current selected item
3. Call inspect method
4. Marshal to current format (YAML/JSON)
5. Copy to clipboard via `clipboard.WriteAll()`
6. Show notification: "Copied to clipboard"

**Pattern:**
```go
case tea.KeyPressMsg:
    switch msg.String() {
    case "y":
        // Copy logic here
        return m, showNotification("Copied to clipboard")
    }
```

---

## PHASE 7: Testing & Documentation 🔜

**Status:** Pending
**Estimated Time:** 2-3 hours

### Task 7.1: Manual Testing

**Test Matrix:**

| Resource  | Summary OK | YAML OK | JSON OK | Scroll OK | Copy OK | Connected Resources OK |
|-----------|------------|---------|---------|-----------|---------|------------------------|
| Container | [ ]        | [ ]     | [ ]     | [ ]       | [ ]     | [ ]                    |
| Image     | [ ]        | [ ]     | [ ]     | [ ]       | [ ]     | [ ]                    |
| Network   | [ ]        | [ ]     | [ ]     | [ ]       | [ ]     | [ ]                    |
| Volume    | [ ]        | [ ]     | [ ]     | [ ]       | [ ]     | [ ]                    |
| Service   | [ ]        | [ ]     | [ ]     | [ ]       | [ ]     | [ ]                    |

**Edge Cases to Test:**
- [ ] Empty fields (no labels, no mounts)
- [ ] Very long output (1000+ lines)
- [ ] Special characters in strings
- [ ] Missing/null values
- [ ] Invalid resource IDs (error handling)
- [ ] Config file with `inspection-format: json`
- [ ] CLI flag `--json` override

**Verification:**
- [ ] No CPU stats/graphs in container view
- [ ] Syntax highlighting works with current theme
- [ ] Content scrolls without height limits
- [ ] Build completes without warnings

### Task 7.2: Update Documentation

**Files to update:**

1. **`docs/KEYBINDINGS.md`**
   - Add: `y` - Copy inspection output to clipboard

2. **`docs/CONFIGURATION.md`** (create if doesn't exist)
   - Add: `inspection-format: yaml|json` - Format for inspection output (default: yaml)

3. **`docs/CLI_REFERENCE.md`**
   - Add: `--json` - Use JSON format for inspection output (overrides config)

4. **`README.md`** (if applicable)
   - Update feature list if it mentions info panels

---

## File Change Summary

### New Files (1)
- `internal/ui/components/infopanel/marshal.go`

### Modified Files (9)
- `internal/config/config.go`
- `cmd/main.go`
- `internal/ui/components/infopanel/builders/builders.go`
- `internal/ui/containers/containers.go`
- `internal/client/client.go`
- `internal/ui/images/images.go`
- `internal/ui/networks/networks.go`
- `internal/ui/volumes/volumes.go`
- `internal/ui/services/services.go`

### Deleted Files (4)
- `internal/ui/components/infopanel/infopanel.go`
- `internal/ui/components/infopanel/section.go`
- `internal/ui/components/infopanel/field.go`
- `internal/ui/components/infopanel/statspanel.go`

### Documentation Files (3+)
- `docs/KEYBINDINGS.md` (update)
- `docs/CONFIGURATION.md` (update/create)
- `docs/CLI_REFERENCE.md` (update/create)

---

## Timeline Estimate

- **Phase 1:** 2-3 hours (Infrastructure)
- **Phase 2:** 3-4 hours (Container panel complete)
- **Phase 3:** 2-3 hours (Other resources)
- **Phase 4:** 2 hours (View controllers)
- **Phase 5:** 1 hour (Cleanup)
- **Phase 6:** 1 hour (Clipboard)
- **Phase 7:** 2-3 hours (Testing & docs)

**Total:** 13-18 hours

---

## Success Criteria

Implementation is complete when:
- ✅ All resource views show summary + raw YAML/JSON dump
- ✅ User can toggle YAML/JSON via config or `--json` flag
- ✅ Syntax highlighting works using theme colors
- ✅ Content is scrollable with no height limits
- ✅ User can copy output with 'y' keybinding
- ✅ CPU stats are removed from containers view
- ✅ Old panel rendering code is deleted
- ✅ All tests pass
- ✅ Build completes without errors
- ✅ Documentation is updated

---

## Notes & Reminders

### Color Mapping (Theme → Syntax)
- Keys: `colors.Primary` (typically cyan/blue)
- Strings: `colors.Success` (typically green)
- Numbers: `colors.Warning` (typically yellow/orange)
- Booleans/null: `colors.Muted` (typically gray)

### Summary Field Format
```
Label:       value
Label:       value with multi-line
             continuation indented at column 13
```

### Lazydocker Reference
- Summary: 5-10 key fields, simple padding, no borders
- Separator: blank line before "Full details:"
- Full details: entire inspect output, syntax highlighted, scrollable

### Error Philosophy
- Keep it simple: just show error message
- Don't try to be clever with fallbacks
- Let user know clearly what went wrong

---

## Current Phase Status

**STARTING PHASE 1: Core Infrastructure**

**Next Task:** Create `internal/ui/components/infopanel/marshal.go`

---

*Last Updated: 2026-01-20*
*Implementation Mode: ACTIVE*
