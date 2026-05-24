Output web/app/public/docs/check.gif

Set Shell "fish"
Set FontSize 14
Set Width 700
Set Height 350
Set Theme "Catppuccin Mocha"
Set Padding 12
Set Framerate 24
Set TypingSpeed 30ms
Set CursorBlink false

Type "echo -e 'nexo\nhivo\nzuno\ntevo\nkaze' > names.txt"
Enter
Sleep 1s
Type "domh check names.txt --tld com,dev,io --batch --yes"
Enter
Sleep 10s
