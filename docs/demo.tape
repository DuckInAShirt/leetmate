Output docs/demo.gif
Output docs/demo.mp4
Set Shell "bash"
Set FontSize 18
Set Width 1400
Set Height 780
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type "source ./docs/demo-setup.sh || exit 1"
Enter
Type "clear"
Enter
Show

Type "# LeetMate — 不给答案，只给提示。"
Enter
Sleep 800ms
Type "# Hints, not handouts. Practice now, review before you forget."
Enter
Sleep 1.5s
Hide
Type "clear && ./leetmate"
Enter
Show
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
Sleep 4s

# Ask for a non-answer Hint and wait for the streamed coach reply.
Type "1"
Sleep 6s

# Hold the two-pane Hint view before opening the full response.
Sleep 1s

# Open detail mode, then switch from the statement to the full coach reply.
Type "o"
Sleep 500ms
Tab
Sleep 3s

Ctrl+C
Sleep 500ms

Hide
Type "cleanup_demo"
Enter
