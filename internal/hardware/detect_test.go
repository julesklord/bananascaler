package hardware

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckDeps(t *testing.T) {
	tempDir := t.TempDir()

	deps := []string{"ffmpeg", "ffprobe", "realesrgan-ncnn-vulkan"}
	for _, dep := range deps {
		file := filepath.Join(tempDir, dep)
		err := os.WriteFile(file, []byte("dummy"), 0755)
		if err != nil {
			t.Fatalf("failed to create mock binary: %v", err)
		}
	}

	t.Run("All dependencies found", func(t *testing.T) {
		origPath := os.Getenv("PATH")
		defer os.Setenv("PATH", origPath)

		os.Setenv("PATH", tempDir)

		err := CheckDeps()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("Missing dependency", func(t *testing.T) {
		origPath := os.Getenv("PATH")
		defer os.Setenv("PATH", origPath)

		emptyDir := t.TempDir()
		os.Setenv("PATH", emptyDir)

		err := CheckDeps()
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "required dependency not found in PATH") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
