-- Feats
Expand functionality (https://github.com/ajayd-san/gomanagedocker).
Logs should open in a modal.
Shell should open in a modal (https://github.com/charmbracelet/x/xpty).
Improve informational panel.
Add graphs to containers tab.
Implement browse tab.
Support other clients.
Better color handling (lipgloss.AdaptiveColor), less usage of Muted()
Scrollbar in informational panel/viewports
Syntax highlighting in YAML/JSON of informational panel

-- Other
Reconsider usage of nerd fonts.
Expand configuration options.
Better Cobra commands for launching into tabs, possibly more.
Overhaul CI/CD.

-- Later
Test in multiple terminal types
Improve styling re: list items
Decide what goes in informational panel
  - ties into improve styling, top bit can be shown in list w/ icons and panel can be pure JSON/YAML

-- Bugs
Fix synced spinners when stopping multiple containers
"Pull image" functionality should update selected image; use spinner like containers
