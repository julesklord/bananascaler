# Output path for the animated GIF
Output docs/demo.gif

# Visual Settings
Set FontSize 14
Set FontFamily "monospace"
Set Width 1000
Set Height 600
Set Theme "TokyoNight"
Set Padding 10

# Command to launch the application (hidden from the viewer)
Hide
Type "PATH=/tmp:\$PATH ./bin/bananascaler input_cut.mp4 -o output_demo.mp4 --scale 2"
Enter
Sleep 1.5s
Show

# --- Interaction Sequence ---
# 1. Wait while frame extraction finishes and upscaling starts
Sleep 4.5s

# 2. Wait while the upscaling progress bar runs (since it's only 2 seconds of video / 60 frames, it should be quick)
Sleep 8s

# 3. Wait for the re-encoding stage to finish and show the "Done!" message
Sleep 4s

# 4. Stay on the completion screen for a moment to let the viewer see it
Sleep 3s
