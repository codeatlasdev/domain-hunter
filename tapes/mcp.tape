Output web/app/public/docs/mcp.gif

Set Shell "zsh"
Set FontSize 15
Set Width 800
Set Height 400
Set Theme "Catppuccin Mocha"
Set Padding 16
Set Framerate 24
Set TypingSpeed 35ms
Set CursorBlink false
Type "# MCP server for AI agents"
Enter
Type "cat .mcp.json"
Enter
Sleep 2s

Type ""
Enter
Type "# Start MCP server (stdio mode)"
Enter
Type "domh mcp &"
Enter
Sleep 2s
Type "kill %1 2>/dev/null"
Enter
Sleep 1s
