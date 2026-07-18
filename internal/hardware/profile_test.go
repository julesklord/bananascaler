package hardware

import (
	"testing"
)

func TestMaxSafeTile(t *testing.T) {
	tests := []struct {
		name     string
		vramMB   int
		model    string
		expected int
	}{
		// 12 GB+ boundaries (vramMB >= 12000)
		{"12GB+ x4plus", 12000, "realesrgan-x4plus", 600},
		{"12GB+ x4plus above boundary", 24000, "realesrgan-x4plus", 600},
		{"12GB+ x4plus-anime", 12000, "realesrgan-x4plus-anime", 500},
		{"12GB+ default", 12000, "realesr-animevideov3-x2", 512},

		// 8-12 GB boundaries (8000 <= vramMB < 12000)
		{"8GB+ x4plus", 8000, "realesrgan-x4plus", 400},
		{"8GB+ x4plus near upper boundary", 11999, "realesrgan-x4plus", 400},
		{"8GB+ x4plus-anime", 8000, "realesrgan-x4plus-anime", 350},
		{"8GB+ x4plus-anime near upper boundary", 11999, "realesrgan-x4plus-anime", 350},
		{"8GB+ default", 8000, "realesr-animevideov3-x2", 400},
		{"8GB+ default near upper boundary", 11999, "realesr-animevideov3-x2", 400},

		// 4-8 GB boundaries (4000 <= vramMB < 8000)
		{"4GB+ x4plus", 4000, "realesrgan-x4plus", 200},
		{"4GB+ x4plus near upper boundary", 7999, "realesrgan-x4plus", 200},
		{"4GB+ x4plus-anime", 4000, "realesrgan-x4plus-anime", 200},
		{"4GB+ x4plus-anime near upper boundary", 7999, "realesrgan-x4plus-anime", 200},
		{"4GB+ default", 4000, "realesr-animevideov3-x2", 300},
		{"4GB+ default near upper boundary", 7999, "realesr-animevideov3-x2", 300},

		// <4 GB boundaries (vramMB < 4000)
		{"<4GB x4plus near upper boundary", 3999, "realesrgan-x4plus", 100},
		{"<4GB x4plus-anime near upper boundary", 3999, "realesrgan-x4plus-anime", 100},
		{"<4GB default near upper boundary", 3999, "realesr-animevideov3-x2", 150},
		{"<4GB default lower bound", 0, "realesr-animevideov3-x2", 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxSafeTile(tt.vramMB, tt.model)
			if result != tt.expected {
				t.Errorf("maxSafeTile(%d, %q) = %d; expected %d", tt.vramMB, tt.model, result, tt.expected)
			}
		})
	}
}
