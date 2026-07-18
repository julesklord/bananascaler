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
		t.Setenv("PATH", tempDir)

		err := CheckDeps()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("Missing dependency", func(t *testing.T) {
		emptyDir := t.TempDir()
		t.Setenv("PATH", emptyDir)

		err := CheckDeps()
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "required dependency not found in PATH") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestFramerate(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Happy path",
			script:      "#!/bin/sh\necho '24000/1001'\n",
			expected:    "24000/1001",
			expectError: false,
		},
		{
			name:        "Empty output",
			script:      "#!/bin/sh\necho ''\n",
			expected:    "",
			expectError: true,
			errorMsg:    "no video stream detected",
		},
		{
			name:        "Execution error",
			script:      "#!/bin/sh\nexit 1\n",
			expected:    "",
			expectError: true,
			errorMsg:    "ffprobe framerate detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			mockPath := filepath.Join(tempDir, "ffprobe")
			err := os.WriteFile(mockPath, []byte(tt.script), 0755)
			if err != nil {
				t.Fatalf("failed to create mock binary: %v", err)
			}

			t.Setenv("PATH", tempDir+string(os.PathListSeparator)+os.Getenv("PATH"))

			fps, err := Framerate("dummy.mp4")

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if fps != tt.expected {
					t.Errorf("expected fps %q, got %q", tt.expected, fps)
				}
			}
		})
	}
}
