Output docs/demo.gif
Set Shell "zsh"
Set FontSize 18
Set Width 1400
Set Height 780
Set Theme "Catppuccin Mocha"

Type "./leetmate"
Enter
Sleep 1.5s

# Open Study Plans from the home screen.
Down
Sleep 500ms
Enter
Sleep 800ms

# Open the first built-in plan: Hot 100.
Enter
Sleep 800ms

# Open the first due problem in Hot 100.
Enter
Sleep 2s

# Ask for a non-answer Hint and wait for the streamed coach reply.
Type "1"
Sleep 4s

# Focus the coach panel and expand the full coach reply.
Tab
Sleep 500ms
Type "o"
Sleep 3s

Ctrl+C
