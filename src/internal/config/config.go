// Package config defines the runtime configuration for a bananascaler run.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/julesklord/bananascaler/internal/hardware"
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

	// Profile system
	Profile    *hardware.UpscaleProfile // Active profile (nil = use legacy defaults)
	AutoDetect bool                     // Auto-detect hardware and pick profile
	PresetStr  string                   // User-requested preset ("fast"/"balanced"/"quality")
}

// ResolveProfile applies profile logic: auto-detect > user preset > explicit profile > legacy defaults.
// Must be called after CLI flags are parsed but before Validate().
func (c *Config) ResolveProfile() error {
	switch {
	case c.AutoDetect:
		preset := hardware.PresetBalanced
		if c.PresetStr != "" {
			var err error
			preset, err = hardware.ParsePreset(c.PresetStr)
			if err != nil {
				return err
			}
		}
		profile, _ := hardware.AutoProfileWithPreset(preset)
		c.Profile = profile
		c.applyProfileDefaults()
		return nil

	case c.PresetStr != "":
		preset, err := hardware.ParsePreset(c.PresetStr)
		if err != nil {
			return err
		}
		// Auto-detect tier but use user's preset
		profile, _ := hardware.AutoProfileWithPreset(preset)
		c.Profile = profile
		c.applyProfileDefaults()
		return nil

	default:
		// No profile selected — use legacy behavior with optional profile
		// If a model was explicitly set by user, keep it
		return nil
	}
}

// applyProfileDefaults sets Config fields from the active profile,
// but only if the user hasn't explicitly overridden them via CLI flags.
func (c *Config) applyProfileDefaults() {
	if c.Profile == nil {
		return
	}
	p := c.Profile

	// Model: use profile default unless user explicitly set one
	if c.Model == "" || c.Model == DefaultModel {
		c.Model = p.Model
	}

	// Scale: cap at profile's max
	if c.Scale > p.MaxScale {
		c.Scale = p.MaxScale
	}
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

	// Enforce profile's max scale
	if c.Profile != nil && c.Scale > c.Profile.MaxScale {
		return fmt.Errorf("scale factor %d exceeds profile maximum (%d). Use a higher-tier profile or reduce --scale", c.Scale, c.Profile.MaxScale)
	}

	return nil
}

// ProfileSummary returns a one-line description of the active profile for logging.
func (c *Config) ProfileSummary() string {
	if c.Profile == nil {
		return fmt.Sprintf("legacy (model=%s, scale=%dx)", c.Model, c.Scale)
	}
	return c.Profile.String()
}
