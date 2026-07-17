// Package hardware provides runtime detection of GPU capabilities and
// media stream metadata via ffprobe.
package hardware

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CheckDeps verifies all required external binaries are available in PATH.
func CheckDeps() error {
	for _, dep := range []string{"ffmpeg", "ffprobe", "realesrgan-ncnn-vulkan"} {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("required dependency not found in PATH: %q. Install it and ensure it is accessible via $PATH", dep)
		}
	}
	return nil
}

// HasNVIDIA returns true if nvidia-smi is present and exits cleanly.
func HasNVIDIA() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "nvidia-smi").Run() == nil
}

// Framerate returns the rational frame rate of the first video stream.
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

// FrameCount returns the total number of frames in the first video stream.
// Falls back to 0 if ffprobe cannot determine the count (e.g. container has no nb_frames).
func FrameCount(input string) int {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=nb_frames",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	).Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}
