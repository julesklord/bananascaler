// Package config defines the runtime configuration for a bananascaler run.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultModel = "realesr-animevideov3-x2"

// Config holds all settings parsed from CLI flags and arguments.
type Config struct {
	Input   string
	Output  string
	Scale   int
	GPU     int
	Model   string
	Verbose bool
	NoTUI   bool
}

// Validate checks the config for invalid values and applies defaults.
func (c *Config) Validate() error {
	if c.Scale < 2 || c.Scale > 4 {
		return fmt.Errorf("invalid scale factor %d: must be 2, 3, or 4", c.Scale)
	}
	if _, err := os.Stat(c.Input); err != nil {
		return fmt.Errorf("input file not found: %q", c.Input)
	}
	if c.Output == "" {
		dir := filepath.Dir(c.Input)
		base := strings.TrimSuffix(filepath.Base(c.Input), filepath.Ext(c.Input))
		c.Output = filepath.Join(dir, base+"_upscaled.mp4")
	}
	if c.Model == "" {
		c.Model = DefaultModel
	}
	return nil
}
