# Tab Launch Subcommands Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable users to launch containertui directly into specific tabs via CLI subcommands (e.g., `containertui images`), with configuration file support for default startup tab.

**Architecture:** 
- Restructure CLI to use Cobra subcommands while preserving backward compatibility
- Add `StartupTab` field to Config struct with YAML support
- Create tab string conversion utilities in tabs package
- Pass startup tab from CLI/config down through UI initialization
- Implement validation and helpful error messages for invalid tab names

**Tech Stack:** Cobra CLI framework, YAML config parsing, Go standard library

---

## File Structure

**Files to Create:**
- None (all changes are to existing files)

**Files to Modify:**
1. `cmd/main.go` - Restructure Cobra commands to support subcommands
2. `internal/config/config.go` - Add StartupTab field to Config struct
3. `internal/ui/tabs/tabs.go` - Add TabFromString() and IsValidTab() helper functions
4. `internal/ui/ui.go` - Accept and use startup tab parameter

**Testing:**
- Unit tests for TabFromString() validation
- Integration test for CLI subcommand parsing

---

## Task 1: Add StartupTab Field to Config

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add StartupTab field to Config struct**

In `internal/config/config.go`, update the Config struct to include:

```go
// Add this field to the Config struct (around line 13-16)
StartupTab string `yaml:"startup-tab,omitempty"`
```

The full struct should look like:
```go
type Config struct {
	NoNerdFonts      ConfigBool  `yaml:"no-nerd-fonts"`
	Theme            ThemeConfig `yaml:"colors,omitempty"`
	InspectionFormat string      `yaml:"inspection-format,omitempty"`
	StartupTab       string      `yaml:"startup-tab,omitempty"`
}
```

- [ ] **Step 2: Update DefaultConfig to include default StartupTab**

Update the `DefaultConfig()` function to set the default:

```go
func DefaultConfig() *Config {
	return &Config{
		NoNerdFonts:      false,
		Theme:            emptyThemeConfig(),
		InspectionFormat: "yaml",
		StartupTab:       "containers",  // Add this line
	}
}
```

- [ ] **Step 3: Verify config.go still compiles**

Run: `go build ./cmd/`

Expected: No compilation errors

- [ ] **Step 4: Commit config changes**

```bash
git add internal/config/config.go
git commit -m "feat: add StartupTab field to Config struct"
```

---

## Task 2: Add Tab String Conversion Utilities

**Files:**
- Modify: `internal/ui/tabs/tabs.go`

- [ ] **Step 1: Add TabFromString function**

In `internal/ui/tabs/tabs.go`, add this function after the `String()` method (around line 34):

```go
// TabFromString converts a string to a Tab, returns -1 if invalid
func TabFromString(s string) Tab {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "containers":
		return Containers
	case "images":
		return Images
	case "volumes":
		return Volumes
	case "networks":
		return Networks
	case "services":
		return Services
	case "browse":
		return Browse
	default:
		return -1
	}
}
```

- [ ] **Step 2: Add IsValidTab function**

After the TabFromString function, add:

```go
// IsValidTab checks if a tab string is valid
func IsValidTab(s string) bool {
	return TabFromString(s) != -1
}
```

- [ ] **Step 3: Add AllTabNames function for help text**

After IsValidTab, add:

```go
// AllTabNames returns all valid tab names
func AllTabNames() []string {
	return []string{"containers", "images", "volumes", "networks", "services", "browse"}
}
```

- [ ] **Step 4: Update New() to accept startup tab parameter**

Modify the `New()` function signature and body:

```go
func New(startupTab Tab) Model {
	return Model{
		ActiveTab: startupTab,
		Tabs:      []Tab{Containers, Images, Volumes, Networks, Services, Browse},
		KeyMap:    NewKeyMap(),
	}
}
```

- [ ] **Step 5: Verify tabs.go compiles**

Run: `go build ./cmd/`

Expected: Will fail because NewModel in ui.go still uses old tabs.New() signature (will fix in next task)

- [ ] **Step 6: Commit tabs changes**

```bash
git add internal/ui/tabs/tabs.go
git commit -m "feat: add TabFromString and tab validation utilities"
```

---

## Task 3: Update UI Model to Accept Startup Tab

**Files:**
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Update NewModel signature to accept startup tab**

In `internal/ui/ui.go` around line 40, change:

```go
func NewModel() Model {
```

to:

```go
func NewModel(startupTab tabs.Tab) Model {
```

- [ ] **Step 2: Update tabs.New() call to use startupTab parameter**

In the NewModel function (around line 43), change:

```go
tabsModel := tabs.New()
```

to:

```go
tabsModel := tabs.New(startupTab)
```

- [ ] **Step 3: Update previousTab to use startupTab**

In NewModel around line 57, change:

```go
previousTab:        tabs.Containers,
```

to:

```go
previousTab:        startupTab,
```

- [ ] **Step 4: Update Start function to handle startup tab from config**

In `internal/ui/ui.go`, find the `Start()` function and update it. First read the full Start function to see its current implementation, then modify it to:

1. Parse the startup tab from config
2. Validate it
3. Pass it to NewModel

The function should look like:
```go
func Start() error {
	cfg := state.GetConfig()
	
	// Determine startup tab
	startupTab := tabs.Containers // default
	if cfg.StartupTab != "" {
		if !tabs.IsValidTab(cfg.StartupTab) {
			fmt.Fprintf(os.Stderr, "warning: invalid startup tab '%s', using 'containers' instead\n", cfg.StartupTab)
		} else {
			startupTab = tabs.TabFromString(cfg.StartupTab)
		}
	}
	
	model := NewModel(startupTab)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

(You'll need to add `"fmt"` and `"os"` imports if they're not already there)

- [ ] **Step 5: Verify ui.go compiles with all changes**

Run: `go build ./cmd/`

Expected: Compilation errors about main.go passing no argument to NewModel (will fix in next task)

- [ ] **Step 6: Commit ui changes**

```bash
git add internal/ui/ui.go
git commit -m "feat: update UI model to accept and use startup tab"
```

---

## Task 4: Restructure CLI to Support Subcommands

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: Create a shared run function**

In `cmd/main.go`, add this helper function before `main()`:

```go
func runContainertui(cmd *cobra.Command, tabName string, noNerdFonts bool, configPath string, colorsFlag []string, jsonFormat bool) error {
	var cfg *config.Config
	var err error
	if configPath != "" {
		cfg, err = config.LoadFromFile(configPath)
		if err != nil {
			return err
		}
	} else {
		cfg = config.DefaultConfig()
	}

	if noNerdFonts {
		cfg.NoNerdFonts = true
	}

	if jsonFormat {
		cfg.InspectionFormat = "json"
	}

	// Set startup tab if provided via subcommand
	if tabName != "" {
		cfg.StartupTab = tabName
	}

	if len(colorsFlag) > 0 {
		colorOverrides, err := colors.ParseColors(colorsFlag)
		if err != nil {
			return fmt.Errorf("failed to parse colors: %w", err)
		}

		if colorOverrides.Primary.IsAssigned() {
			cfg.Theme.Primary = colorOverrides.Primary
		}
		if colorOverrides.Border.IsAssigned() {
			cfg.Theme.Border = colorOverrides.Border
		}
		if colorOverrides.Text.IsAssigned() {
			cfg.Theme.Text = colorOverrides.Text
		}
		if colorOverrides.Muted.IsAssigned() {
			cfg.Theme.Muted = colorOverrides.Muted
		}
		if colorOverrides.Selected.IsAssigned() {
			cfg.Theme.Selected = colorOverrides.Selected
		}
		if colorOverrides.Success.IsAssigned() {
			cfg.Theme.Success = colorOverrides.Success
		}
		if colorOverrides.Warning.IsAssigned() {
			cfg.Theme.Warning = colorOverrides.Warning
		}
		if colorOverrides.Error.IsAssigned() {
			cfg.Theme.Error = colorOverrides.Error
		}
	}

	state.SetConfig(cfg)

	// Initialize the shared Docker client
	if err := state.InitializeClient(); err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	defer func() {
		if err := state.CloseClient(); err != nil {
			log.Printf("error closing Docker client: %v", err)
		}
	}()

	state.InitializeLog()

	// Start the UI
	if err := ui.Start(); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}

	return nil
}
```

- [ ] **Step 2: Restructure main() with root and subcommands**

Replace the entire `main()` function with:

```go
func main() {
	var noNerdFonts bool
	var configPath string
	var colorsFlag []string
	var jsonFormat bool
	var startupTab string

	// Create subcommand runner factory
	makeSubcommand := func(tabName string, use string, short string) *cobra.Command {
		return &cobra.Command{
			Use:   use,
			Short: short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runContainertui(cmd, tabName, noNerdFonts, configPath, colorsFlag, jsonFormat)
			},
		}
	}

	rootCmd := &cobra.Command{
		Use:   "containertui",
		Short: "a tui for managing container lifecycles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContainertui(cmd, "", noNerdFonts, configPath, colorsFlag, jsonFormat)
		},
	}

	// Add subcommands
	rootCmd.AddCommand(makeSubcommand("containers", "containers", "launch containertui to the containers tab"))
	rootCmd.AddCommand(makeSubcommand("images", "images", "launch containertui to the images tab"))
	rootCmd.AddCommand(makeSubcommand("volumes", "volumes", "launch containertui to the volumes tab"))
	rootCmd.AddCommand(makeSubcommand("networks", "networks", "launch containertui to the networks tab"))
	rootCmd.AddCommand(makeSubcommand("services", "services", "launch containertui to the services tab"))
	rootCmd.AddCommand(makeSubcommand("browse", "browse", "launch containertui to the browse tab"))

	// Add global flags to root command
	rootCmd.PersistentFlags().BoolVar(&noNerdFonts, "no-nerd-fonts", false, "disable nerd fonts")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")
	rootCmd.PersistentFlags().StringSliceVar(&colorsFlag, "colors", nil, "color overrides (format: --colors 'primary=#b4befe' --colors 'warning=#f9e2af,success=#a6e3a1')")
	rootCmd.PersistentFlags().BoolVar(&jsonFormat, "json", false, "use JSON format for inspection output")
	rootCmd.PersistentFlags().StringVar(&startupTab, "startup-tab", "", "startup tab (containers, images, volumes, networks, services, browse)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: Verify main.go compiles**

Run: `go build ./cmd/`

Expected: No compilation errors

- [ ] **Step 4: Test the build**

Run: `go build -o containertui ./cmd/`

Expected: Binary created successfully

- [ ] **Step 5: Commit CLI changes**

```bash
git add cmd/main.go
git commit -m "feat: restructure CLI to support tab subcommands"
```

---

## Task 5: Integration Testing

**Files:**
- Create: `tests/cli_test.go` or add to existing test file

- [ ] **Step 1: Create integration test file**

Create `tests/cli_subcommand_test.go` with basic structure:

```go
package tests

import (
	"testing"
	
	"github.com/givensuman/containertui/internal/ui/tabs"
)

func TestTabFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected tabs.Tab
		valid    bool
	}{
		{"containers", tabs.Containers, true},
		{"images", tabs.Images, true},
		{"volumes", tabs.Volumes, true},
		{"networks", tabs.Networks, true},
		{"services", tabs.Services, true},
		{"browse", tabs.Browse, true},
		{"Containers", tabs.Containers, true},
		{"IMAGES", tabs.Images, true},
		{"  volumes  ", tabs.Volumes, true},
		{"invalid", tabs.Tab(-1), false},
		{"", tabs.Tab(-1), false},
	}

	for _, tt := range tests {
		result := tabs.TabFromString(tt.input)
		if result != tt.expected {
			t.Errorf("TabFromString(%q) = %v, want %v", tt.input, result, tt.expected)
		}

		valid := tabs.IsValidTab(tt.input)
		if valid != tt.valid {
			t.Errorf("IsValidTab(%q) = %v, want %v", tt.input, valid, tt.valid)
		}
	}
}

func TestAllTabNames(t *testing.T) {
	names := tabs.AllTabNames()
	if len(names) != 6 {
		t.Errorf("AllTabNames() returned %d tabs, expected 6", len(names))
	}

	expectedNames := map[string]bool{
		"containers": true,
		"images":     true,
		"volumes":    true,
		"networks":   true,
		"services":   true,
		"browse":     true,
	}

	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("AllTabNames() returned unexpected tab: %q", name)
		}
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./tests -v`

Expected: All tests pass

- [ ] **Step 3: Test CLI manually**

Test each subcommand works (don't actually run the full TUI, just verify no panic on startup):

```bash
./containertui --help
./containertui containers --help
./containertui images --help
```

Expected: Help text displays correctly for each

- [ ] **Step 4: Commit test file**

```bash
git add tests/cli_subcommand_test.go
git commit -m "test: add tests for tab conversion and validation"
```

---

## Task 6: Update Documentation

**Files:**
- Modify: `README.md` (if usage section exists)

- [ ] **Step 1: Review current README**

Read the README to understand current usage documentation

- [ ] **Step 2: Add subcommand usage to README**

Add a new section documenting the new feature. Example:

```markdown
### Launching Specific Tabs

You can launch containertui directly to a specific tab using subcommands:

\`\`\`bash
containertui containers    # Launch to containers tab (default)
containertui images        # Launch to images tab
containertui volumes       # Launch to volumes tab
containertui networks      # Launch to networks tab
containertui services      # Launch to services tab
containertui browse        # Launch to browse tab
\`\`\`

All existing flags continue to work with subcommands:

\`\`\`bash
containertui images --config /path/to/config --no-nerd-fonts
\`\`\`

You can also set a default startup tab in your config file:

\`\`\`yaml
# ~/.config/containertui/config.yaml
startup-tab: images
\`\`\`
```

- [ ] **Step 3: Commit documentation**

```bash
git add README.md
git commit -m "docs: add subcommand usage documentation"
```

---

## Task 7: Final Verification

- [ ] **Step 1: Build the project**

Run: `go build -o containertui ./cmd/`

Expected: Binary builds successfully

- [ ] **Step 2: Verify no compilation warnings**

Run: `go build -o containertui ./cmd/ 2>&1 | grep -i warning || echo "No warnings"`

Expected: No output (no warnings)

- [ ] **Step 3: Run all tests**

Run: `go test ./...`

Expected: All tests pass

- [ ] **Step 4: Check git history**

Run: `git log --oneline -7`

Expected: Shows all commits from the implementation

- [ ] **Step 5: Create summary of changes**

Review all commits made:

Run: `git log --oneline HEAD~6..HEAD`

Expected: Shows 6-7 commits for the feature implementation

---

## Success Criteria

✅ `containertui containers` launches containertui to containers tab
✅ `containertui images` launches containertui to images tab  
✅ All 6 tabs accessible via subcommands
✅ `containertui` with no args defaults to containers (backward compatible)
✅ All existing CLI flags work with subcommands
✅ Config file `startup-tab` field works
✅ Help text clearly documents subcommands
✅ Invalid tab names show helpful error message
✅ All tests pass
✅ No compilation errors or warnings
