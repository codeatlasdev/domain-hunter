Output web/app/public/docs/check.gif

Set Shell "zsh"
Set FontSize 15
Set Width 800
Set Height 400
Set Theme "Catppuccin Mocha"
Set Padding 16
Set Framerate 24
Set TypingSpeed 35ms
Set CursorBlink false
Type "# Create a names file"
Enter
Type "echo -e 'nexo\nhivo\nzuno\ntevo\nkaze' > names.txt"
Enter
Sleep 1s

Type ""
Enter
Type "# Check against multiple TLDs"
Enter
Type "domh check names.txt --tld com,dev,io --batch --yes"
Enter
Sleep 10s
