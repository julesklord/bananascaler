// Package hardware provides runtime detection of GPU capabilities and
// media stream metadata via ffprobe.
package hardware

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// PresetLevel defines the speed/quality tradeoff.
type PresetLevel string

const (
	PresetFast     PresetLevel = "fast"
	PresetBalanced PresetLevel = "balanced"
	PresetQuality  PresetLevel = "quality"
)

// AllPresets returns the three available preset levels.
func AllPresets() []PresetLevel {
	return []PresetLevel{PresetFast, PresetBalanced, PresetQuality}
}

// ParsePreset converts a string to a PresetLevel.
func ParsePreset(s string) (PresetLevel, error) {
	switch strings.ToLower(s) {
	case "fast", "f":
		return PresetFast, nil
	case "balanced", "b", "bal":
		return PresetBalanced, nil
	case "quality", "q", "qual":
		return PresetQuality, nil
	default:
		return "", fmt.Errorf("unknown preset %q: must be fast, balanced, or quality", s)
	}
}

// HardwareTier classifies the system's GPU capability.
type HardwareTier string

const (
	TierUnknown  HardwareTier = "unknown"
	TierLowEnd   HardwareTier = "low-end"
	TierMidRange HardwareTier = "mid-range"
	TierHighEnd  HardwareTier = "high-end"
)

// GPUInfo holds detected GPU information.
type GPUInfo struct {
	Name      string
	VRAMMB    int
	HasNVIDIA bool
	CPUCores  int
}

// UpscaleProfile contains all tunable parameters for the upscale pipeline.
type UpscaleProfile struct {
	Tier        HardwareTier
	Preset      PresetLevel
	Description string

	// Real-ESRGAN
	TileSize int
	Model    string

	// Frame extraction (ffmpeg -q:v)
	JPEGQuality int

	// NVENC encoding preset (p1-p7)
	NVEncPreset string

	// CPU encoding fallback (libx265)
	X265Preset string
	X265CRF    int

	// Scale limits
	MaxScale int
}

// String returns a short human-readable summary.
func (p *UpscaleProfile) String() string {
	return fmt.Sprintf("%s / %s (tile=%d, model=%s)",
		p.Tier, p.Preset, p.TileSize, shortModelName(p.Model))
}

func shortModelName(m string) string {
	m = strings.TrimPrefix(m, "realesr-")
	m = strings.TrimPrefix(m, "realesrgan-")
	return m
}

// ── Hardware detection ───────────────────────────────────────────────────────

// DetectGPU queries nvidia-smi for GPU name and VRAM.
// Returns zero-value GPUInfo if no NVIDIA GPU is found.
func DetectGPU() GPUInfo {
	info := GPUInfo{
		CPUCores: runtime.NumCPU(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query GPU name
	if out, err := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=name", "--format=csv,noheader",
	).Output(); err == nil {
		name := strings.TrimSpace(string(out))
		if name != "" {
			info.Name = name
			info.HasNVIDIA = true
		}
	}

	if !info.HasNVIDIA {
		return info
	}

	// Query VRAM in MB
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if out, err := exec.CommandContext(ctx2, "nvidia-smi",
		"--query-gpu=memory.total", "--format=csv,noheader,nounits",
	).Output(); err == nil {
		vramStr := strings.TrimSpace(string(out))
		if vram, err := strconv.Atoi(vramStr); err == nil {
			info.VRAMMB = vram
		}
	}

	return info
}

// DetectGPUCount returns the number of NVIDIA GPUs.
func DetectGPUCount() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=index", "--format=csv,noheader",
	).Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			count++
		}
	}
	return count
}

// ── Tier classification ──────────────────────────────────────────────────────

func classifyTier(info GPUInfo) HardwareTier {
	if !info.HasNVIDIA {
		return TierUnknown
	}
	switch {
	case info.VRAMMB < 4000:
		return TierLowEnd
	case info.VRAMMB < 8000:
		return TierMidRange
	default:
		return TierHighEnd
	}
}

// ── Profile database ─────────────────────────────────────────────────────────
//
// Tile size / model VRAM guide (empirical):
//
//   realesr-animevideov3-x2  — lightweight: safe up to tile 400 on 4GB
//   realesrgan-x4plus-anime  — medium:      needs ≤ tile 400 on 10GB
//   realesrgan-x4plus        — heavy:       needs ≤ tile 256 on 8GB, ≤ tile 512 on 12GB
//
// Rule of thumb: heavier model → smaller tile for same VRAM budget.
// Always keep a 1–2 GB margin for desktop compositor + OS.
// Mid-range tier (4-8GB) is limited to the lightweight model only —
// heavier models at 4× scale cause SEGV on 6GB VRAM even at tile=200.

func profileDB() map[HardwareTier]map[PresetLevel]*UpscaleProfile {
	return map[HardwareTier]map[PresetLevel]*UpscaleProfile{
		// ── Low-end: ≤4 GB VRAM (GTX 1050 Ti, GTX 1650, RX 570) ──────────
		TierLowEnd: {
			PresetFast: {
				Tier:        TierLowEnd,
				Preset:      PresetFast,
				Description: "Fast mode for low-end GPUs (≤4GB). Speed over quality.",
				TileSize:    64,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 4,
				NVEncPreset: "p1",
				X265Preset:  "ultrafast",
				X265CRF:     28,
				MaxScale:    2,
			},
			PresetBalanced: {
				Tier:        TierLowEnd,
				Preset:      PresetBalanced,
				Description: "Balanced mode for low-end GPUs. Good speed/quality ratio.",
				TileSize:    100,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 3,
				NVEncPreset: "p3",
				X265Preset:  "fast",
				X265CRF:     26,
				MaxScale:    2,
			},
			PresetQuality: {
				Tier:        TierLowEnd,
				Preset:      PresetQuality,
				Description: "Quality mode for low-end GPUs. Best possible at this tier.",
				TileSize:    150,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 2,
				NVEncPreset: "p5",
				X265Preset:  "medium",
				X265CRF:     24,
				MaxScale:    3,
			},
		},
		// ── Mid-range: 4–8 GB VRAM (GTX 1060 6GB, RTX 2060, RX 5700 XT) ──
		TierMidRange: {
			PresetFast: {
				Tier:        TierMidRange,
				Preset:      PresetFast,
				Description: "Fast mode for mid-range GPUs (4-8GB). Quick upscaling.",
				TileSize:    150,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 3,
				NVEncPreset: "p3",
				X265Preset:  "fast",
				X265CRF:     26,
				MaxScale:    2,
			},
			PresetBalanced: {
				Tier:        TierMidRange,
				Preset:      PresetBalanced,
				Description: "Balanced mode for mid-range GPUs. Recommended for most users.",
				TileSize:    300,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 2,
				NVEncPreset: "p5",
				X265Preset:  "medium",
				X265CRF:     22,
				MaxScale:    2,
			},
			PresetQuality: {
				Tier:        TierMidRange,
				Preset:      PresetQuality,
				Description: "Quality mode for mid-range GPUs. Best encode settings, 2× upscale.",
				TileSize:    300,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 1,
				NVEncPreset: "p7",
				X265Preset:  "slow",
				X265CRF:     18,
				MaxScale:    2,
			},
		},
		// ── High-end: ≥8 GB VRAM (RTX 3080, RTX 4090, RX 6800 XT) ────────
		TierHighEnd: {
			PresetFast: {
				Tier:        TierHighEnd,
				Preset:      PresetFast,
				Description: "Fast mode for high-end GPUs (8GB+). Quick with heavy models.",
				TileSize:    300,
				Model:       "realesrgan-x4plus-anime",
				JPEGQuality: 2,
				NVEncPreset: "p4",
				X265Preset:  "medium",
				X265CRF:     22,
				MaxScale:    4,
			},
			PresetBalanced: {
				Tier:        TierHighEnd,
				Preset:      PresetBalanced,
				Description: "Balanced mode for high-end GPUs. Excellent quality and speed.",
				TileSize:    400,
				Model:       "realesrgan-x4plus",
				JPEGQuality: 2,
				NVEncPreset: "p6",
				X265Preset:  "slow",
				X265CRF:     20,
				MaxScale:    4,
			},
			PresetQuality: {
				Tier:        TierHighEnd,
				Preset:      PresetQuality,
				Description: "Quality mode for high-end GPUs. Maximum quality, no compromises.",
				TileSize:    512,
				Model:       "realesrgan-x4plus",
				JPEGQuality: 1,
				NVEncPreset: "p7",
				X265Preset:  "veryslow",
				X265CRF:     18,
				MaxScale:    4,
			},
		},
		// ── Unknown: no NVIDIA GPU (CPU-only / Vulkan via iGPU) ────────────
		TierUnknown: {
			PresetFast: {
				Tier:        TierUnknown,
				Preset:      PresetFast,
				Description: "Fast fallback (no NVIDIA GPU detected). CPU-only mode.",
				TileSize:    64,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 4,
				NVEncPreset: "",
				X265Preset:  "ultrafast",
				X265CRF:     28,
				MaxScale:    2,
			},
			PresetBalanced: {
				Tier:        TierUnknown,
				Preset:      PresetBalanced,
				Description: "Balanced fallback (no NVIDIA GPU detected). CPU-only mode.",
				TileSize:    100,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 3,
				NVEncPreset: "",
				X265Preset:  "fast",
				X265CRF:     24,
				MaxScale:    3,
			},
			PresetQuality: {
				Tier:        TierUnknown,
				Preset:      PresetQuality,
				Description: "Quality fallback (no NVIDIA GPU detected). CPU-only mode.",
				TileSize:    150,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 2,
				NVEncPreset: "",
				X265Preset:  "medium",
				X265CRF:     22,
				MaxScale:    3,
			},
		},
	}
}

// GetProfile returns a specific profile by tier and preset.
func GetProfile(tier HardwareTier, preset PresetLevel) *UpscaleProfile {
	db := profileDB()
	if tierProfiles, ok := db[tier]; ok {
		if p, ok := tierProfiles[preset]; ok {
			return p
		}
	}
	// Fallback: unknown tier, balanced preset
	return db[TierUnknown][PresetBalanced]
}

// GetDefaultProfile returns the "balanced" profile for a given tier.
func GetDefaultProfile(tier HardwareTier) *UpscaleProfile {
	return GetProfile(tier, PresetBalanced)
}

// AutoProfile detects hardware and returns the recommended profile.
// For unknown GPUs it falls back to TierUnknown + PresetBalanced.
func AutoProfile() (*UpscaleProfile, GPUInfo) {
	info := DetectGPU()
	tier := classifyTier(info)
	profile := GetProfile(tier, PresetBalanced)
	return profile, info
}

// AutoProfileWithPreset detects hardware and returns a profile with
// the user's preferred preset, adapted to detected tier.
func AutoProfileWithPreset(preset PresetLevel) (*UpscaleProfile, GPUInfo) {
	info := DetectGPU()
	tier := classifyTier(info)
	profile := GetProfile(tier, preset)
	return profile, info
}

// ProfileSummary returns a one-line summary for display in the TUI/CLI.
func ProfileSummary(info GPUInfo, p *UpscaleProfile) string {
	var parts []string
	if info.HasNVIDIA {
		name := info.Name
		if name == "" {
			name = "NVIDIA GPU"
		}
		vram := fmt.Sprintf("%dMB", info.VRAMMB)
		if info.VRAMMB >= 1024 {
			vram = fmt.Sprintf("%.1fGB", float64(info.VRAMMB)/1024.0)
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", name, vram))
	} else {
		parts = append(parts, fmt.Sprintf("CPU-only (%d cores)", info.CPUCores))
	}
	parts = append(parts, fmt.Sprintf("tier=%s", p.Tier))
	parts = append(parts, fmt.Sprintf("preset=%s", p.Preset))
	return strings.Join(parts, " | ")
}

// ProfileDisplay returns a multi-line profile summary for verbose output.
func ProfileDisplay(info GPUInfo, p *UpscaleProfile) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Hardware: %s\n", ProfileSummary(info, p)))
	b.WriteString(fmt.Sprintf("  Tile size:     %d\n", p.TileSize))
	b.WriteString(fmt.Sprintf("  Model:         %s\n", p.Model))
	b.WriteString(fmt.Sprintf("  JPEG quality:  %d\n", p.JPEGQuality))
	if p.NVEncPreset != "" {
		b.WriteString(fmt.Sprintf("  NVENC preset:  %s\n", p.NVEncPreset))
	} else {
		b.WriteString("  NVENC:         not available (CPU encoding)\n")
	}
	b.WriteString(fmt.Sprintf("  x265 preset:   %s (CRF %d)\n", p.X265Preset, p.X265CRF))
	b.WriteString(fmt.Sprintf("  Max scale:     %dx\n", p.MaxScale))
	return b.String()
}

// ── Safety validation ────────────────────────────────────────────────────────

// maxSafeTile returns the maximum tile size considered safe for a given
// VRAM and model. These are conservative heuristics — not benchmarks.
func maxSafeTile(vramMB int, model string) int {
	switch {
	case vramMB >= 12000:
		// 12 GB+: generous
		switch {
		case strings.Contains(model, "x4plus") && !strings.Contains(model, "anime"):
			return 600
		case strings.Contains(model, "x4plus-anime"):
			return 500
		default: // animevideov3
			return 512
		}
	case vramMB >= 8000:
		// 8–12 GB
		switch {
		case strings.Contains(model, "x4plus") && !strings.Contains(model, "anime"):
			return 400
		case strings.Contains(model, "x4plus-anime"):
			return 350
		default:
			return 400
		}
	case vramMB >= 4000:
		// 4–8 GB
		switch {
		case strings.Contains(model, "x4plus") && !strings.Contains(model, "anime"):
			return 200
		case strings.Contains(model, "x4plus-anime"):
			return 200
		default:
			return 300
		}
	default:
		// <4 GB
		switch {
		case strings.Contains(model, "x4plus"):
			return 100
		default:
			return 150
		}
	}
}

// CheckTileSafety returns a warning if the tile size may cause OOM on the
// detected GPU. Returns empty string if safe.
func CheckTileSafety(info GPUInfo, tile int, model string) string {
	if !info.HasNVIDIA {
		return ""
	}
	safe := maxSafeTile(info.VRAMMB, model)
	if tile > safe {
		return fmt.Sprintf(
			"tile=%d may exceed safe limit (≤%d) for %s on %s (%dMB). Consider reducing --tile or using a lighter model to avoid GPU crash.",
			tile, safe, shortModelName(model), info.Name, info.VRAMMB)
	}
	return ""
}
