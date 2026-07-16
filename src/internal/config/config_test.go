package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	// Create a temporary file to use as valid Input
	tmpFile, err := os.CreateTemp("", "bananascaler_test_input_*.mp4")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name        string
		cfg         Config
		expectError bool
		checkOutput string
		checkModel  string
	}{
		{
			name: "valid config with default output",
			cfg: Config{
				Input: tmpFile.Name(),
				Scale: 2,
			},
			expectError: false,
			checkOutput: filepath.Join(filepath.Dir(tmpFile.Name()), filepath.Base(tmpFile.Name())[:len(filepath.Base(tmpFile.Name()))-4]+"_upscaled.mp4"),
			checkModel:  DefaultModel,
		},
		{
			name: "invalid scale low",
			cfg: Config{
				Input: tmpFile.Name(),
				Scale: 1,
			},
			expectError: true,
		},
		{
			name: "invalid scale high",
			cfg: Config{
				Input: tmpFile.Name(),
				Scale: 5,
			},
			expectError: true,
		},
		{
			name: "missing input file",
			cfg: Config{
				Input: "nonexistent_file_12345.mp4",
				Scale: 2,
			},
			expectError: true,
		},
		{
			name: "custom output and model preserved",
			cfg: Config{
				Input:  tmpFile.Name(),
				Output: "/tmp/custom_out.mp4",
				Scale:  3,
				Model:  "realesrgan-x4plus",
			},
			expectError: false,
			checkOutput: "/tmp/custom_out.mp4",
			checkModel:  "realesrgan-x4plus",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.expectError {
				if err == nil {
					t.Errorf("expected validation error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
				if tc.cfg.Output != tc.checkOutput {
					t.Errorf("expected Output %q, got %q", tc.checkOutput, tc.cfg.Output)
				}
				if tc.cfg.Model != tc.checkModel {
					t.Errorf("expected Model %q, got %q", tc.checkModel, tc.cfg.Model)
				}
			}
		})
	}
}
