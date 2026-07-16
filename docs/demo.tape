# bananascaler demo — TUI with live progress bars
Output docs/demo.gif

Set FontSize 14
Set FontFamily "monospace"
Set Width 1000
Set Height 600
Set Theme "TokyoNight"
Set Padding 10
Set Margin 0

# Start the TUI (hidden so the viewer only sees the rendered output)
Hide
Type "PATH=/tmp:$PATH ./bin/bananascaler /tmp/demo_video.mp4 -o /tmp/demo_video_out.mp4 --scale 2 --profile balanced"
Enter
Sleep 1s
Show

# Let the TUI render: extraction (~2s) → upscaling (~6s) → re-encode (~1s)
Sleep 12s
