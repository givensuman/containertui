# Services view demo
Output ../assets/demo-services.gif

Set Theme "Catppuccin Mocha"
Set Width 1200
Set Height 600
Set FontSize 14
Set Padding 10
Set TypingSpeed 50ms

Type "go run ./cmd"
Enter
Sleep 3s

# Switch to services tab
Type "5"
Sleep 2s

# Navigate through services if any exist
Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

# Inspect if available
Type "i"
Sleep 2s
Type "q"
Sleep 1s

# Quit
Type "q"
Sleep 500ms
