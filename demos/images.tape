# Image management demo
Output ../assets/demo-images.gif

Set Theme "Catppuccin Mocha"
Set Width 1200
Set Height 600
Set FontSize 14
Set Padding 10
Set TypingSpeed 50ms

Type "/tmp/containertui"
Enter
Sleep 5s

# Switch to images tab
Type "2"
Sleep 2s

# Navigate through image list
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "k"
Sleep 500ms

# Inspect an image
Type "i"
Sleep 2s
Type "q"
Sleep 1s

# Show image history
Type "h"
Sleep 2s
Type "q"
Sleep 1s

# Quit
Type "q"
Sleep 500ms
