Output demo.mp4
Output demo.gif
Set FontSize 20
Set Width 1920
Set Height 1200
Set TypingSpeed 100ms

# Launch stui
Sleep 1.5s
Type "stui"
Enter
Sleep 2s

# NODES
Sleep 400ms
Type "1"
Sleep 400ms
Type@300ms "jjj"
Sleep 400ms
Type "p"
Sleep 400ms
Down@300ms 2
Sleep 400ms
Enter
Sleep 400ms
Type "s"
Sleep 400ms
Type "dow"
Sleep 400ms
Enter
Sleep 400ms
Type "/"
Sleep 400ms
Type "linux..3"
Sleep 400ms
Enter
Sleep 400ms
Space
Sleep 400ms
Type "j"
Sleep 400ms
Space
Sleep 400ms
Type "j"
Sleep 400ms
Space
Sleep 400ms
Enter
Sleep 400ms
Escape
Sleep 400ms
Type "c"
Sleep 700ms
Type "state=DRAIN"
Sleep 600ms
Enter
Sleep 800ms
Type " reason='maintenance'"
Sleep 400ms
Enter
Sleep 900ms
Escape

Sleep 2s

# JOBS
Sleep 400ms
Type "2"
Sleep 400ms
Type@300ms "jjj"
Sleep 400ms
Type "p"
Sleep 400ms
Down@300ms 2
Sleep 400ms
Enter
Sleep 400ms
Type "s"
Sleep 400ms
Type "run"
Sleep 400ms
Enter
Sleep 400ms
Up
Sleep 400ms
Type "/"
Sleep 400ms
Backspace@80ms 8
Sleep 400ms
Type "spare"
Sleep 400ms
Enter
Sleep 400ms
Space
Sleep 400ms
Type "j"
Sleep 400ms
Space
Sleep 400ms
Type "j"
Sleep 400ms
Space
Sleep 400ms
Enter
Sleep 600ms
Escape
Sleep 400ms
Type "c"
Sleep 400ms
Type "timelimit=1-12:00:00"
Sleep 400ms
Enter
Sleep 900ms
Escape
Sleep 400ms
Escape
Sleep 400ms
Escape

# SACCT
Sleep 400ms
Type "3"
Sleep 400ms
Type@300ms "jjj"
Sleep 400ms
Escape
Sleep 400ms
Type "/"
Sleep 400ms
Type "job-biology-7"
Sleep 400ms
Enter
Sleep 400ms
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms
Space
Sleep 400ms
Enter
Sleep 900ms
Escape
Sleep 900ms

# SACCT MGR
Sleep 400ms
Type "4"
Sleep 400ms
Escape
Sleep 500ms
Down@80ms 5
Sleep 400ms
Type "e"
Sleep 400ms
Down@80ms 10
Sleep 400ms
Enter
Sleep 400ms
Type "/"
Sleep 400ms
Type "john"
Sleep 400ms
Enter
Sleep 400ms
Down@80ms 3

# SDIAG
Sleep 400ms
Type "5"
Sleep 600ms
Down@50ms 10

Sleep 1s
Ctrl+C
Sleep 500ms