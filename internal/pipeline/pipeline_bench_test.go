package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCountFiles(b *testing.B) {
	dir, err := os.MkdirTemp("", "bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create 10,000 files to simulate a real directory
	for i := 1; i <= 10000; i++ {
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("frame_%05d.png", i)), nil, 0644)
	}

	b.Run("ReadDir", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			countFiles(dir, ".png")
		}
	})

	b.Run("Stat_Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			prev := 10000
			for {
				if _, err := os.Stat(filepath.Join(dir, fmt.Sprintf("frame_%05d.png", prev+1))); err == nil {
					prev++
				} else {
					break
				}
			}
		}
	})
}
