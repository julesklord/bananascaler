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
