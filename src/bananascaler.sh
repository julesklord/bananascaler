#!/usr/bin/env bash
# ==============================================================================
# bananascaler — GPU-Accelerated Neural Video Upscaler
# Real-ESRGAN (Vulkan) + FFmpeg (NVDEC/NVENC or libx265 fallback)
# ==============================================================================
set -euo pipefail

# ── Colors ────────────────────────────────────────────────────────────────────
if [[ -t 1 ]]; then
    C_RESET="\033[0m"; C_BOLD="\033[1m"
    C_YELLOW="\033[33m"; C_CYAN="\033[36m"
    C_GREEN="\033[32m";  C_RED="\033[31m"; C_DIM="\033[2m"
else
    C_RESET=""; C_BOLD=""; C_YELLOW=""; C_CYAN=""
    C_GREEN=""; C_RED=""; C_DIM=""
fi

log_info()  { echo -e "${C_CYAN}${C_BOLD}[INFO]${C_RESET} $*"; }
log_ok()    { echo -e "${C_GREEN}${C_BOLD}[ OK ]${C_RESET} $*"; }
log_warn()  { echo -e "${C_YELLOW}${C_BOLD}[WARN]${C_RESET} $*"; }
log_error() { echo -e "${C_RED}${C_BOLD}[ERR ]${C_RESET} $*" >&2; }
log_step()  { echo -e "\n${C_BOLD}${C_YELLOW}🍌 $*${C_RESET}"; }

# ── Help ──────────────────────────────────────────────────────────────────────
usage() {
    cat <<EOF
${C_BOLD}bananascaler${C_RESET} — GPU-accelerated neural video upscaler

${C_BOLD}USAGE${C_RESET}
  bananascaler.sh <input> [output] [scale]

${C_BOLD}ARGUMENTS${C_RESET}
  input   Path to source video file (required)
  output  Output path (default: <input>_upscaled.mp4)
  scale   Upscale factor: 2, 3, or 4 (default: 2)

${C_BOLD}OPTIONS${C_RESET}
  -h, --help    Show this help and exit
  -g, --gpu N   GPU index for Real-ESRGAN (default: 0, use -1 for CPU)

${C_BOLD}EXAMPLES${C_RESET}
  bananascaler.sh movie.mp4
  bananascaler.sh input.mp4 output_4k.mp4 4
  bananascaler.sh input.mp4 output.mp4 2 --gpu 1
  nohup bananascaler.sh input.mp4 output.mp4 2 > run.log 2>&1 &
EOF
}

# ── Argument parsing ──────────────────────────────────────────────────────────
GPU_INDEX=0
POSITIONAL=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -g|--gpu)     GPU_INDEX="$2"; shift 2 ;;
        -*)           log_error "Unknown option: $1"; usage; exit 1 ;;
        *)            POSITIONAL+=("$1"); shift ;;
    esac
done

if [[ ${#POSITIONAL[@]} -lt 1 ]]; then
    usage; exit 1
fi

INPUT_VIDEO="${POSITIONAL[0]}"
OUTPUT_VIDEO="${POSITIONAL[1]:-}"
SCALE_FACTOR="${POSITIONAL[2]:-2}"

# ── Validate inputs ───────────────────────────────────────────────────────────
if [[ ! -f "$INPUT_VIDEO" ]]; then
    log_error "Input file not found: '$INPUT_VIDEO'"
    exit 1
fi

if [[ ! "$SCALE_FACTOR" =~ ^[234]$ ]]; then
    log_error "Invalid scale factor '$SCALE_FACTOR'. Must be 2, 3, or 4."
    exit 1
fi

if [[ ! "$GPU_INDEX" =~ ^-?[0-9]+$ ]]; then
    log_error "Invalid GPU index '$GPU_INDEX'. Must be an integer (-1 = CPU)."
    exit 1
fi

# ── Auto-name output ──────────────────────────────────────────────────────────
if [[ -z "$OUTPUT_VIDEO" ]]; then
    dir_name="$(dirname "$INPUT_VIDEO")"
    raw_name="$(basename "${INPUT_VIDEO%.*}")"
    OUTPUT_VIDEO="${dir_name}/${raw_name}_upscaled.mp4"
fi

# ── Dependency check (fail-fast before any processing) ───────────────────────
check_dep() {
    if ! command -v "$1" &>/dev/null; then
        log_error "Required dependency not found in PATH: '$1'"
        exit 1
    fi
}
check_dep ffmpeg
check_dep ffprobe
check_dep realesrgan-ncnn-vulkan

# ── Session temp dirs ─────────────────────────────────────────────────────────
SESSION_ID="bananascaler_$(date +%s)_$$"
TEMP_IN="/tmp/${SESSION_ID}_in"
TEMP_OUT="/tmp/${SESSION_ID}_out"
mkdir -p "$TEMP_IN" "$TEMP_OUT"

# ── Cleanup trap (runs on exit, Ctrl+C, or error) ─────────────────────────────
cleanup() {
    local exit_code=$?
    rm -rf "$TEMP_IN" "$TEMP_OUT"
    rm -f "${OUTPUT_VIDEO}.tmp"
    if [[ $exit_code -ne 0 ]]; then
        log_error "bananascaler exited with errors. Temp files cleaned up."
    fi
}
trap cleanup EXIT

# ── Hardware detection ────────────────────────────────────────────────────────
HAS_NVIDIA=false
if command -v nvidia-smi &>/dev/null && nvidia-smi &>/dev/null 2>&1; then
    HAS_NVIDIA=true
fi

if [[ "$HAS_NVIDIA" == true ]]; then
    log_info "NVIDIA GPU detected — enabling NVDEC + NVENC acceleration."
    DEC_FLAGS=(-hwaccel cuda)
    ENC_FLAGS=(-c:v hevc_nvenc -pix_fmt yuv420p)
else
    log_warn "No NVIDIA GPU detected — using CPU (libx265). Expect slower encode."
    DEC_FLAGS=()
    ENC_FLAGS=(-c:v libx265 -preset medium -crf 22 -pix_fmt yuv420p)
fi

# ── Audio stream detection ────────────────────────────────────────────────────
AUDIO_STREAMS=$(ffprobe -v error -select_streams a \
    -show_entries stream=index -of csv=p=0 "$INPUT_VIDEO" | wc -l)

if [[ "$AUDIO_STREAMS" -gt 0 ]]; then
    AUDIO_FLAGS=(-map 1:a -c:a copy)
    log_info "Audio stream detected — will remux without re-encoding."
else
    AUDIO_FLAGS=()
    log_warn "No audio stream found — output will be video-only."
fi

# ── Probe source framerate ────────────────────────────────────────────────────
FRAMERATE=$(ffprobe -v error -select_streams v:0 \
    -show_entries stream=r_frame_rate \
    -of default=noprint_wrappers=1:nokey=1 "$INPUT_VIDEO")

if [[ -z "$FRAMERATE" ]]; then
    log_error "Could not detect framerate from '$INPUT_VIDEO'. Is it a valid video file?"
    exit 1
fi
log_info "Source framerate: ${C_BOLD}${FRAMERATE}${C_RESET} fps  |  Scale: ${C_BOLD}${SCALE_FACTOR}×${C_RESET}  |  GPU index: ${C_BOLD}${GPU_INDEX}${C_RESET}"
echo -e "${C_DIM}Output → ${OUTPUT_VIDEO}${C_RESET}\n"

# ── Stage 1: Extract frames ───────────────────────────────────────────────────
log_step "[1/3] Extracting frames..."
if ! ffmpeg -y -stats -loglevel warning \
        "${DEC_FLAGS[@]}" \
        -i "$INPUT_VIDEO" \
        -f image2 -vcodec mjpeg -q:v 2 \
        "$TEMP_IN/frame_%05d.jpg"; then
    log_error "Frame extraction failed."
    exit 1
fi
FRAME_COUNT=$(find "$TEMP_IN" -name '*.jpg' | wc -l)
log_ok "${FRAME_COUNT} frames extracted."

# ── Stage 2: Neural upscale ───────────────────────────────────────────────────
log_step "[2/3] Neural upscaling (${SCALE_FACTOR}×) via Real-ESRGAN..."
if ! realesrgan-ncnn-vulkan \
        -i "$TEMP_IN" \
        -o "$TEMP_OUT" \
        -n realesr-animevideov3-x2 \
        -s "$SCALE_FACTOR" \
        -g "$GPU_INDEX"; then
    log_error "Real-ESRGAN upscaling failed."
    exit 1
fi
log_ok "Upscaling complete."

# ── Stage 3: Re-encode + mux ──────────────────────────────────────────────────
log_step "[3/3] Re-encoding and muxing audio..."
if ffmpeg -y -stats -loglevel warning \
        -framerate "$FRAMERATE" \
        -i "$TEMP_OUT/frame_%05d.jpg" \
        -i "$INPUT_VIDEO" \
        -map 0:v \
        "${AUDIO_FLAGS[@]}" \
        "${ENC_FLAGS[@]}" \
        "${OUTPUT_VIDEO}.tmp"; then
    mv "${OUTPUT_VIDEO}.tmp" "$OUTPUT_VIDEO"
    log_ok "Done! Output saved to: ${C_BOLD}${OUTPUT_VIDEO}${C_RESET}"
else
    log_error "Final video assembly failed."
    exit 1
fi
