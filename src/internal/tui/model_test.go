package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/julesklord/bananascaler/internal/config"
)

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"video.mp4", true},
		{"movie.MKV", true},
		{"clip.avi", true},
		{"recording.MOV", true},
		{"stream.webm", true},
		{"file.txt", false},
		{"image.png", false},
		{"no_extension", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isVideoFile(tc.name)
			if got != tc.expected {
				t.Errorf("isVideoFile(%q) = %v; want %v", tc.name, got, tc.expected)
			}
		})
	}
}

func TestShortModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"realesr-animevideov3-x2", "animevideov3-x2"},
		{"realesrgan-x4plus-anime", "x4plus-anime"},
		{"custom-model", "custom-model"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := shortModel(tc.input)
			if got != tc.expected {
				t.Errorf("shortModel(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestModelExplorerKeybinds(t *testing.T) {
	cfg := &config.Config{
		Scale: 2,
		GPU:   0,
		Model: "realesr-animevideov3-x2",
	}

	m := NewModel(cfg)
	m.state = stateSelectFile

	// Test cycling scale factor (s)
	// Default is 2. 2 -> 3 -> 4 -> 2
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updatedModel.(Model)
	if m.cfg.Scale != 3 {
		t.Errorf("expected Scale to cycle to 3, got %d", m.cfg.Scale)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updatedModel.(Model)
	if m.cfg.Scale != 4 {
		t.Errorf("expected Scale to cycle to 4, got %d", m.cfg.Scale)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updatedModel.(Model)
	if m.cfg.Scale != 2 {
		t.Errorf("expected Scale to cycle back to 2, got %d", m.cfg.Scale)
	}

	// Test cycling GPU (g)
	// Default is 0. 0 -> 1 -> -1 -> 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(Model)
	if m.cfg.GPU != 1 {
		t.Errorf("expected GPU to cycle to 1, got %d", m.cfg.GPU)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(Model)
	if m.cfg.GPU != -1 {
		t.Errorf("expected GPU to cycle to -1, got %d", m.cfg.GPU)
	}

	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	m = updatedModel.(Model)
	if m.cfg.GPU != 0 {
		t.Errorf("expected GPU to cycle back to 0, got %d", m.cfg.GPU)
	}

	// Test cycling model (m)
	// Default is realesr-animevideov3-x2.
	// modelNames = ["realesr-animevideov3-x2", "realesrgan-x4plus", "realesrgan-x4plus-anime"]
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	m = updatedModel.(Model)
	if m.cfg.Model != "realesrgan-x4plus" {
		t.Errorf("expected Model to cycle to realesrgan-x4plus, got %q", m.cfg.Model)
	}
}

func TestModelExplorerNavigation(t *testing.T) {
	cfg := &config.Config{
		Scale: 2,
		GPU:   0,
		Model: "realesr-animevideov3-x2",
	}

	m := NewModel(cfg)
	m.state = stateSelectFile

	// Mock file entries
	tempDir, err := os.MkdirTemp("", "bananascaler_nav_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	m.currentDir = tempDir
	err = os.WriteFile(filepath.Join(tempDir, "video1.mp4"), []byte("data"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "video2.mp4"), []byte("data"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	// Read directory
	msg := m.readDirCmd()()
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(Model)

	if len(m.files) != 2 {
		t.Fatalf("expected 2 files in explorer, got %d", len(m.files))
	}

	// Cursor defaults to 0
	if m.cursor != 0 {
		t.Errorf("expected cursor to start at 0, got %d", m.cursor)
	}

	// Press down (j) -> cursor should be 1
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor to move to 1, got %d", m.cursor)
	}

	// Press down (j) again -> cursor should stay at 1 (bounds checking)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updatedModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor to stay at 1, got %d", m.cursor)
	}

	// Press up (k) -> cursor should be 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor to move to 0, got %d", m.cursor)
	}

	// Press up (k) again -> cursor should stay at 0
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}
}
