# Development Guide

## Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| `bash` | ≥ 4.0 | Script runtime |
| `ffmpeg` | Any recent | Frame I/O, encoding |
| `realesrgan-ncnn-vulkan` | v0.2.5.0+ | Neural upscaling |
| `ffprobe` | Bundled with ffmpeg | Framerate detection |
| NVIDIA drivers + CUDA | Optional | Hardware acceleration |

## Local Setup

```bash
# 1. Clone
git clone https://github.com/julesklord/video-upscaler-gpu.git
cd video-upscaler-gpu

# 2. Make script executable
chmod +x src/upscale.sh

# 3. Install dependencies (Arch Linux / CachyOS)
sudo pacman -S ffmpeg

# Real-ESRGAN
mkdir -p ~/.local/share/realesrgan && cd ~/.local/share/realesrgan
curl -sL -O "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-ubuntu.zip"
unzip realesrgan-ncnn-vulkan-20220424-ubuntu.zip && rm *.zip
chmod +x realesrgan-ncnn-vulkan
ln -sf ~/.local/share/realesrgan/realesrgan-ncnn-vulkan ~/.local/bin/realesrgan-ncnn-vulkan
```

## Useful Commands

```bash
# Run with defaults (2× scale, auto output name)
src/upscale.sh input.mp4

# Run with explicit output and scale
src/upscale.sh input.mp4 output_4k.mp4 4

# Background run with log capture
nohup src/upscale.sh input.mp4 output.mp4 2 > upscale.log 2>&1 &

# Verify dependencies
command -v ffmpeg && command -v realesrgan-ncnn-vulkan && echo "OK"
```

## No Build Step

This is a Bash script. There is nothing to compile. `chmod +x` is the only setup.
