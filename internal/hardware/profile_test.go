package hardware

import (
	"testing"
)

func BenchmarkMaxSafeTile(b *testing.B) {
	// These combinations cover the different branches
	cases := []struct {
		vram  int
		model string
	}{
		{16000, "realesrgan-x4plus"},
		{16000, "realesrgan-x4plus-anime"},
		{16000, "realesr-animevideov3-x2"},
		{10000, "realesrgan-x4plus"},
		{10000, "realesrgan-x4plus-anime"},
		{10000, "realesr-animevideov3-x2"},
		{6000, "realesrgan-x4plus"},
		{6000, "realesrgan-x4plus-anime"},
		{6000, "realesr-animevideov3-x2"},
		{2000, "realesrgan-x4plus"},
		{2000, "realesrgan-x4plus-anime"},
		{2000, "realesr-animevideov3-x2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range cases {
			maxSafeTile(c.vram, c.model)
		}
	}
}

func TestParsePreset(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    PresetLevel
		expectError bool
	}{
		// Valid inputs for PresetFast
		{"fast lower", "fast", PresetFast, false},
		{"fast short lower", "f", PresetFast, false},
		{"fast upper", "FAST", PresetFast, false},
		{"fast short upper", "F", PresetFast, false},

		// Valid inputs for PresetBalanced
		{"balanced lower", "balanced", PresetBalanced, false},
		{"balanced short lower", "b", PresetBalanced, false},
		{"balanced mid lower", "bal", PresetBalanced, false},
		{"balanced upper", "BALANCED", PresetBalanced, false},
		{"balanced short upper", "B", PresetBalanced, false},
		{"balanced mid upper", "BAL", PresetBalanced, false},
		{"balanced mixed case", "Bal", PresetBalanced, false},

		// Valid inputs for PresetQuality
		{"quality lower", "quality", PresetQuality, false},
		{"quality short lower", "q", PresetQuality, false},
		{"quality mid lower", "qual", PresetQuality, false},
		{"quality upper", "QUALITY", PresetQuality, false},
		{"quality short upper", "Q", PresetQuality, false},
		{"quality mid upper", "QUAL", PresetQuality, false},
		{"quality mixed case", "Qual", PresetQuality, false},

		// Invalid inputs
		{"empty string", "", "", true},
		{"unknown preset", "unknown", "", true},
		{"typo", "fastt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePreset(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got none", tt.input)
				}
				if result != "" {
					t.Errorf("expected empty result on error, got %q", result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}
