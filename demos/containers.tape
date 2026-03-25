# Container management demo
Output ../assets/demo-containers.gif

Set Theme "Catppuccin Mocha"
Set Width 1200
Set Height 600
Set FontSize 14
Set Padding 10
Set TypingSpeed 50ms

Type "go run ./cmd/main.go"
Enter
Sleep 3s

# Navigate through container list
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

# Inspect a container
Type "i"
Sleep 2s
Type "q"
Sleep 1s

# Show logs
Type "l"
Sleep 2s
Type "q"
Sleep 1s

# Toggle container state
Type " "
Sleep 2s

# Quit
Type "q"
Sleep 500ms
