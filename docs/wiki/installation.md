# Installation Guide — bananascaler

This guide details how to install **bananascaler** and its system dependencies for GPU-accelerated video upscaling.

---

## 1. System Requirements

For GPU acceleration, you need:
- An **NVIDIA GPU** (Kepler or newer for NVENC/NVDEC; Maxwell/RTX recommended).
- **NVIDIA proprietary drivers** installed and running.
- **Vulkan drivers** installed (usually bundled with NVIDIA drivers on Linux/Windows).

*Note: If no NVIDIA GPU is detected or CPU mode is forced (`-g -1`), the tool automatically falls back to CPU decoding/encoding (`libx265`).*

---

## 2. Installing External Dependencies

**bananascaler** orchestrates two main external tools: `ffmpeg` and `realesrgan-ncnn-vulkan`. Both must be executable and available in your system's `PATH`.

### Linux (Arch Linux / CachyOS / Ubuntu)

#### FFmpeg
```bash
# Arch Linux / CachyOS
sudo pacman -S ffmpeg

# Ubuntu / Debian
sudo apt update && sudo apt install -y ffmpeg
```

#### Real-ESRGAN (ncnn-vulkan)
1. Download the pre-built Linux binary release:
   ```bash
   mkdir -p ~/.local/share/realesrgan && cd ~/.local/share/realesrgan
   curl -sL -O "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-ubuntu.zip"
   unzip realesrgan-ncnn-vulkan-20220424-ubuntu.zip
   rm realesrgan-ncnn-vulkan-20220424-ubuntu.zip
   chmod +x realesrgan-ncnn-vulkan
   ```
2. Create a symlink to a folder in your `PATH` (e.g., `~/.local/bin` or `/usr/local/bin`):
   ```bash
   mkdir -p ~/.local/bin
   ln -sf ~/.local/share/realesrgan/realesrgan-ncnn-vulkan ~/.local/bin/realesrgan-ncnn-vulkan
   ```

### macOS

#### FFmpeg
Install via Homebrew:
```bash
brew install ffmpeg
```

#### Real-ESRGAN (ncnn-vulkan)
1. Download the macOS release:
   ```bash
   mkdir -p /usr/local/share/realesrgan && cd /usr/local/share/realesrgan
   curl -sL -O "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-macos.zip"
   unzip realesrgan-ncnn-vulkan-20220424-macos.zip
   chmod +x realesrgan-ncnn-vulkan
   ```
2. Symlink to `/usr/local/bin`:
   ```bash
   ln -sf /usr/local/share/realesrgan/realesrgan-ncnn-vulkan /usr/local/bin/realesrgan-ncnn-vulkan
   ```

---

## 3. Installing bananascaler

### Option A: System-wide Installation (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/julesklord/bananascaler.git
   cd bananascaler
   ```
2. Compile and install:
   ```bash
   sudo make install
   ```
   This builds the binary and copies it to `/usr/local/bin/bananascaler`.

   To customize the prefix location (e.g. to install in `/usr/bin`):
   ```bash
   sudo PREFIX=/usr make install
   ```

### Option B: Local Build from Source

If you do not have root privileges or prefer to keep the binary local:
```bash
git clone https://github.com/julesklord/bananascaler.git
cd bananascaler
make build
```
The compiled executable will be located at `./bin/bananascaler`.

### Option C: Build from the `src` directory

You can compile directly inside the Go module folder:
```bash
cd src
make build
```
The executable is generated at `../bin/bananascaler`.

---

## 4. Verification

To verify that the installation succeeded and all dependencies are correctly in `PATH`, run:

```bash
bananascaler tui
```

If any dependency is missing, **bananascaler** will exit immediately with an error message indicating which tool needs to be installed.
