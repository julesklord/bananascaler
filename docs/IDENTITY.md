# Identity: bananascaler

**Name**: bananascaler  
**Type**: CLI tool (Go + Bubbletea TUI)  
**Version**: 0.4.0  
**Author**: julesklord  
**License**: MIT  
**Repository**: https://github.com/julesklord/bananascaler  

## Purpose

A GPU-accelerated neural video upscaler written in Go. Combines Real-ESRGAN (Vulkan) and FFmpeg (NVENC/NVDEC) into an atomic, fault-tolerant processing chain with an interactive Bubbletea TUI dashboard.

## Stack

- **Language**: Go ≥ 1.22
- **CLI**: Cobra
- **TUI**: Bubbletea + Lipgloss
- **External tools**: ffmpeg, ffprobe, realesrgan-ncnn-vulkan
- **Build**: Makefile (`make build`)
