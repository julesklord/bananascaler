# Usage Tutorial — bananascaler

This tutorial walks you through using **bananascaler** in both CLI mode (for direct execution and scripting) and TUI mode (for interactive folder browsing and configuration).

---

## 1. TUI Mode (Interactive File Explorer & Upscaling)

TUI Mode allows you to browse your filesystem, configure upscale parameters interactively, and launch processes without typing complex terminal commands.

### Step 1: Launch the TUI
Open a terminal in the folder containing your videos and run:
```bash
bananascaler tui
```
The file selection screen will load, displaying files and folders in your current working directory.

### Step 2: Navigate and Browse
- **Move Cursor**: Use `↑` / `↓` or `j` / `k` to move the highlighted cursor bar up and down.
- **Enter Folder / Select File**: Press `Enter` or `→` on a folder (rendered in **blue** with a `📁` icon) to open it. Press `Enter` on a video file (rendered in **green** with a `🎥` icon) to choose it and start the upscaling process.
- **Go Back**: Press `Backspace`, `h`, or `←` to navigate to the parent directory.
- **Quit**: Press `q` or `Esc` to close the application.

### Step 3: Cycle Upscale Settings
Before hitting `Enter` on a video file, you can customize the processing settings using keyboard shortcuts:
- **`s` (Scale)**: Cycle the scaling factor (`2×` ➔ `3×` ➔ `4×` ➔ `2×`).
- **`g` (GPU)**: Cycle between detected GPUs or CPU mode (`GPU 0` ➔ `GPU 1` ➔ `CPU` ➔ `GPU 0`).
- **`m` (Model)**: Cycle the Real-ESRGAN model (`realesr-animevideov3-x2` ➔ `realesrgan-x4plus` ➔ `realesrgan-x4plus-anime`).

### Step 4: Monitor Upscaling
Once you select a video file, the TUI switches to the **Pipeline Dashboard**:
- Monitor progress of each of the 3 stages in real-time.
- Press **`v`** to toggle verbose mode (shows raw stdout/stderr from `ffmpeg` and `realesrgan-ncnn-vulkan` inside the TUI logs).
- Press **`q`** or **`Esc`** at any time to cancel. The process terminates immediately and automatically sweeps all intermediate temporary frames, preventing disk waste.

---

## 2. CLI Mode (Direct Command Execution)

To bypass the interactive file manager and start upscaling a specific video file directly, run:

```bash
bananascaler <input_file> [flags]
```

By default, this will launch the TUI dashboard for that single video file.

### Available Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `<input>_upscaled.mp4` | Custom output file path |
| `--scale` | `-s` | `2` | Upscale factor: `2`, `3`, or `4` |
| `--gpu` | `-g` | `0` | GPU device index (`-1` for CPU) |
| `--model` | `-m` | `realesr-animevideov3-x2` | Real-ESRGAN model name |
| `--verbose` | `-v` | `false` | Stream raw tool outputs to the terminal |
| `--no-tui` | | `false` | Disable the interactive dashboard (uses standard text logger) |

### CLI Examples

#### Example 1: Basic 2× Upscaling
```bash
bananascaler vacation.mp4
```
*Result: Upscales `vacation.mp4` to `vacation_upscaled.mp4` at 2× scale factor on GPU 0.*

#### Example 2: High-Quality 4× Upscale on Custom GPU
```bash
bananascaler intro.mkv --output highres_intro.mkv --scale 4 --gpu 1 --model realesrgan-x4plus
```
*Result: Upscales `intro.mkv` to `highres_intro.mkv` at 4× scale using the `realesrgan-x4plus` model on GPU 1.*

---

## 3. Automation and Scripting

If you are running **bananascaler** in automated shell scripts, cron jobs, or on remote servers (via SSH), you should disable the interactive UI to prevent terminal redraw conflicts.

### Headless execution
Pass the `--no-tui` flag to output plain sequential log lines suitable for log files:
```bash
bananascaler clip.mp4 --no-tui --scale 3 > upscale.log 2>&1
```

### Background Processing (`nohup`)
For long-running tasks, run the command in the background:
```bash
nohup bananascaler long_movie.mp4 --no-tui --scale 2 --verbose > run.log 2>&1 &
```
You can monitor progress by tailing the log file:
```bash
tail -f run.log
```
