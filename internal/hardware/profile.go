// Package hardware provides runtime detection of GPU capabilities and
// media stream metadata via ffprobe.
package hardware

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
	"os/exec"
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

	// OS process priority for realesrgan-ncnn-vulkan.
	// 0 = normal, 10 = background (like DaVinci Resolve's idle-priority mode).
	// ponytail: Unix nice(2) via syscall; Windows IDLE_PRIORITY_CLASS if ever ported.
	ProcessNice int
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

// DetectGPU queries nvidia-smi for GPU name and VRAM in a single call.
// Returns zero-value GPUInfo if no NVIDIA GPU is found.
func DetectGPU() GPUInfo {
	info := GPUInfo{
		CPUCores: runtime.NumCPU(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ponytail: single nvidia-smi call with two fields instead of two separate spawns
	out, err := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=name,memory.total",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return info
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return info
	}
	parts := strings.SplitN(line, ",", 2)
	if len(parts) < 1 {
		return info
	}
	info.Name = strings.TrimSpace(parts[0])
	if info.Name != "" {
		info.HasNVIDIA = true
	}
	if len(parts) == 2 {
		if vram, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
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

// classifyTier maps VRAM to a hardware tier.
// 6 buckets give the profiler enough granularity without over-engineering.
func classifyTier(info GPUInfo) HardwareTier {
	if !info.HasNVIDIA {
		return TierUnknown
	}
	switch {
	case info.VRAMMB < 3000: // GTX 1050, GTX 960 — very tight
		return TierLowEnd
	case info.VRAMMB < 5000: // GTX 1650, GTX 1060 3GB
		return TierLowEnd
	case info.VRAMMB < 7000: // GTX 1060 6GB, RTX 2060 — mid-low
		return TierMidRange
	case info.VRAMMB < 10000: // RTX 2070, RTX 3070 — true mid
		return TierMidRange
	case info.VRAMMB < 14000: // RTX 3080 10GB, RTX 4070 Ti
		return TierHighEnd
	default: // RTX 3090, RTX 4090, A100
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
		// ── Low-end: ≤5 GB VRAM (GTX 1050 Ti, GTX 1650, GTX 1060 3GB) ──────
		// ProcessNice=15: runs at background priority — desktop stays smooth
		TierLowEnd: {
			PresetFast: {
				Tier:        TierLowEnd,
				Preset:      PresetFast,
				Description: "Fast mode for low-end GPUs (≤5GB). Speed over quality.",
				TileSize:    64,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 4,
				NVEncPreset: "p1",
				X265Preset:  "ultrafast",
				X265CRF:     28,
				MaxScale:    2,
				ProcessNice: 15,
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
				ProcessNice: 15,
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
				ProcessNice: 15,
			},
		},
		// ── Mid-range: 5–10 GB VRAM (GTX 1060 6GB, RTX 2060/2070, RTX 3070) ─
		// ProcessNice=10: noticeable background priority, still fast enough
		TierMidRange: {
			PresetFast: {
				Tier:        TierMidRange,
				Preset:      PresetFast,
				Description: "Fast mode for mid-range GPUs (5-10GB). Quick upscaling.",
				TileSize:    200,
				Model:       "realesr-animevideov3-x2",
				JPEGQuality: 3,
				NVEncPreset: "p3",
				X265Preset:  "fast",
				X265CRF:     26,
				MaxScale:    2,
				ProcessNice: 10,
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
				ProcessNice: 10,
			},
			PresetQuality: {
				Tier:        TierMidRange,
				Preset:      PresetQuality,
				Description: "Quality mode for mid-range GPUs. Best encode settings, 2× upscale.",
				TileSize:    350,
				Model:       "realesrgan-x4plus-anime",
				JPEGQuality: 1,
				NVEncPreset: "p7",
				X265Preset:  "slow",
				X265CRF:     18,
				MaxScale:    2,
				ProcessNice: 10,
			},
		},
		// ── High-end: ≥10 GB VRAM (RTX 3080, RTX 4090, RX 6800 XT) ─────────
		// ProcessNice=5: slight background priority; these GPUs handle it
		TierHighEnd: {
			PresetFast: {
				Tier:        TierHighEnd,
				Preset:      PresetFast,
				Description: "Fast mode for high-end GPUs (10GB+). Quick with heavy models.",
				TileSize:    300,
				Model:       "realesrgan-x4plus-anime",
				JPEGQuality: 2,
				NVEncPreset: "p4",
				X265Preset:  "medium",
				X265CRF:     22,
				MaxScale:    4,
				ProcessNice: 5,
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
				ProcessNice: 5,
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
				ProcessNice: 5,
			},
		},
		// ── Unknown: no NVIDIA GPU (CPU-only / Vulkan via iGPU) ────────────
		// ProcessNice=19: max background; CPU-only is already slow, keep desktop alive
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
				ProcessNice: 19,
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
				ProcessNice: 19,
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
				ProcessNice: 19,
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
