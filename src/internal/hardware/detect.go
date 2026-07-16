// Package hardware provides runtime detection of GPU capabilities and
// media stream metadata via ffprobe.
package hardware

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckDeps verifies all required external binaries are available in PATH.
// Returns an error on the first missing dependency.
func CheckDeps() error {
	for _, dep := range []string{"ffmpeg", "ffprobe", "realesrgan-ncnn-vulkan"} {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("required dependency not found in PATH: %q\nInstall it and ensure it is accessible via $PATH", dep)
		}
	}
	return nil
}

// HasNVIDIA returns true if nvidia-smi is present and exits cleanly,
// indicating a functioning NVIDIA driver stack.
func HasNVIDIA() bool {
	return exec.Command("nvidia-smi").Run() == nil
}

// Framerate returns the rational frame rate of the first video stream
// (e.g. "24000/1001", "30/1"). Returns an error if detection fails or
// the file contains no video stream.
func Framerate(input string) (string, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=r_frame_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	).Output()
	if err != nil {
		return "", fmt.Errorf("ffprobe framerate detection: %w", err)
	}
	fps := strings.TrimSpace(string(out))
	if fps == "" {
		return "", fmt.Errorf("no video stream detected in %q", input)
	}
	return fps, nil
}

// HasAudio returns true if the input file contains at least one audio stream.
func HasAudio(input string) (bool, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		input,
	).Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe audio detection: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}
