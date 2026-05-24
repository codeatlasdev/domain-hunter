# domh demo recording
Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 20
Set Framerate 30
Set TypingSpeed 40ms
Set CursorBlink false

# Show help
Type "domh --help"
Enter
Sleep 3s

# Clear and show presets
Type "clear"
Enter
Sleep 500ms
Type "domh presets"
Enter
Sleep 3s

# Clear and do a dry-run
Type "clear"
Enter
Sleep 500ms
Type "domh scan coolname --preset startup --dry-run"
Enter
Sleep 3s

# Clear and run a real scan
Type "clear"
Enter
Sleep 500ms
Type "domh scan --tld com --length 4 --pattern cvcv --workers 50"
Enter
Sleep 20s

# Quit
Type "q"
Sleep 2s
