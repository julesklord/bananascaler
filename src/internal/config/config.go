// Package config defines the runtime configuration for a bananascaler run.
package config

// Config holds all settings parsed from CLI flags and arguments.
type Config struct {
	// Input is the path to the source video file. Required.
	Input string
	// Output is the destination path. Auto-generated if empty.
	Output string
	// Scale is the upscale factor (2, 3, or 4). Default: 2.
	Scale int
	// GPU is the GPU device index passed to realesrgan-ncnn-vulkan. -1 = CPU.
	GPU int
	// Model is the Real-ESRGAN model name.
	Model string
	// Verbose enables forwarding of ffmpeg/realesrgan stdout to the terminal.
	Verbose bool
}
