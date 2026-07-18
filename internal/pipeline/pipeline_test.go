package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/julesklord/bananascaler/internal/config"
)

// mockLogger implements Logger to capture progress updates during testing
type mockLogger struct {
	infos    []string
	progress []string
	stages   []int
	done     bool
}

func (m *mockLogger) Info(msg string) { m.infos = append(m.infos, msg) }
func (m *mockLogger) OK(msg string)   { m.infos = append(m.infos, "OK: "+msg) }
func (m *mockLogger) Warn(msg string) { m.infos = append(m.infos, "WARN: "+msg) }
func (m *mockLogger) Step(msg string) { m.infos = append(m.infos, "STEP: "+msg) }
func (m *mockLogger) Err(msg string)  { m.infos = append(m.infos, "ERR: "+msg) }
func (m *mockLogger) Progress(stage, current, total int, eta time.Duration) {
	m.stages = append(m.stages, stage)
	if current == total {
		m.done = true
	}
}

func TestPipelineIntegration(t *testing.T) {
	// Create temp directory for bin mocks
	binDir, err := os.MkdirTemp("", "bananascaler_test_bin_*")
	if err != nil {
		t.Fatalf("failed to create bin temp dir: %v", err)
	}
	defer os.RemoveAll(binDir)

	// Write mock ffprobe script
	ffprobeScript := `#!/bin/sh
if echo "$@" | grep -q "stream=r_frame_rate"; then
  echo "24/1"
  exit 0
fi
if echo "$@" | grep -q "select_streams a"; then
  echo "1"
  exit 0
fi
exit 0
`
	err = os.WriteFile(filepath.Join(binDir, "ffprobe"), []byte(ffprobeScript), 0755)
	if err != nil {
		t.Fatalf("failed to write mock ffprobe: %v", err)
	}

	// Write mock ffmpeg script
	ffmpegScript := `#!/bin/sh
if echo "$@" | grep -q "image2"; then
  # Extraction mode
  # The last argument is the path pattern like /tmp/.../frame_%05d.jpg
  out_pattern=""
  for arg do out_pattern="$arg"; done
  out_dir=$(dirname "$out_pattern")
  mkdir -p "$out_dir"
  touch "$out_dir/frame_00001.jpg"
  touch "$out_dir/frame_00002.jpg"
  exit 0
fi

# Assembly/encoding mode
out_file=""
for arg do out_file="$arg"; done
mkdir -p "$(dirname "$out_file")"
touch "$out_file"
exit 0
`
	err = os.WriteFile(filepath.Join(binDir, "ffmpeg"), []byte(ffmpegScript), 0755)
	if err != nil {
		t.Fatalf("failed to write mock ffmpeg: %v", err)
	}

	// Write mock realesrgan-ncnn-vulkan script
	realesrganScript := `#!/bin/sh
in_dir=""
out_dir=""
while [ $# -gt 0 ]; do
  case "$1" in
    -i) in_dir="$2"; shift ;;
    -o) out_dir="$2"; shift ;;
  esac
  shift
done

mkdir -p "$out_dir"
for f in "$in_dir"/*.jpg; do
  name=$(basename "$f" .jpg)
  touch "$out_dir/$name.png"
done
exit 0
`
	err = os.WriteFile(filepath.Join(binDir, "realesrgan-ncnn-vulkan"), []byte(realesrganScript), 0755)
	if err != nil {
		t.Fatalf("failed to write mock realesrgan: %v", err)
	}

	// Mock nvidia-smi command to simulate NVIDIA GPU presence
	nvidiaSmiScript := `#!/bin/sh
exit 0
`
	err = os.WriteFile(filepath.Join(binDir, "nvidia-smi"), []byte(nvidiaSmiScript), 0755)
	if err != nil {
		t.Fatalf("failed to write mock nvidia-smi: %v", err)
	}

	// Prepend binDir to PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+":"+oldPath)

	// Create dummy input file
	inputDir, err := os.MkdirTemp("", "bananascaler_test_input_*")
	if err != nil {
		t.Fatalf("failed to create input temp dir: %v", err)
	}
	defer os.RemoveAll(inputDir)

	inputFile := filepath.Join(inputDir, "input_video.mp4")
	err = os.WriteFile(inputFile, []byte("fake video content"), 0644)
	if err != nil {
		t.Fatalf("failed to write dummy input file: %v", err)
	}

	outputFile := filepath.Join(inputDir, "output_video.mp4")

	cfg := &config.Config{
		Input:  inputFile,
		Output: outputFile,
		Scale:  2,
		GPU:    0,
		Model:  "realesr-animevideov3-x2",
	}

	ml := &mockLogger{}
	err = Run(cfg, ml)
	if err != nil {
		t.Fatalf("pipeline Run failed unexpectedly: %v", err)
	}

	// Verify that the output file was successfully created (atomic rename completed)
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("output file %q was not created", outputFile)
	}

	// Check if stage progress was logged correctly
	if len(ml.stages) == 0 {
		t.Errorf("expected progress updates, got none")
	}

	// Verify that stage 1, 2, and 3 were reached
	stage1Seen, stage2Seen, stage3Seen := false, false, false
	for _, s := range ml.stages {
		switch s {
		case 1:
			stage1Seen = true
		case 2:
			stage2Seen = true
		case 3:
			stage3Seen = true
		}
	}

	if !stage1Seen || !stage2Seen || !stage3Seen {
		t.Errorf("did not see all pipeline stages: stage1=%v, stage2=%v, stage3=%v",
			stage1Seen, stage2Seen, stage3Seen)
	}
}

func TestCountFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bananascaler_count_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create dummy files
	os.WriteFile(filepath.Join(tempDir, "file1.jpg"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.jpg"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file3.png"), []byte("data"), 0644)
	os.Mkdir(filepath.Join(tempDir, "subdir.jpg"), 0755)

	count, err := countFiles(tempDir, ".jpg")
	if err != nil {
		t.Fatalf("unexpected error counting files: %v", err)
	}

	if count != 2 {
		t.Errorf("expected to count 2 .jpg files, got %d", count)
	}
}
