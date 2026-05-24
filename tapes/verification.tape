Output web/app/public/docs/verification.gif

Set Shell "bash"
Set FontSize 15
Set Width 800
Set Height 400
Set Theme "Catppuccin Mocha"
Set Padding 16
Set Framerate 24
Set TypingSpeed 35ms
Set CursorBlink false

Type "# 4-layer verification: DNS → RDAP → WHOIS → SSL"
Enter
Sleep 1s
Type "# Check with --info to see registrar details"
Enter
Sleep 500ms
Type "domh scan nexo --preset startup --info --batch --yes"
Enter
Sleep 12s
