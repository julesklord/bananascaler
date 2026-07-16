#!/bin/bash
# ==============================================================================
# bananascaler — GPU-Accelerated Neural Video Upscaler
# Utiliza Real-ESRGAN (Vulkan) y FFmpeg (con aceleración de hardware NVIDIA)
# ==============================================================================

# Mostrar ayuda si no hay argumentos
if [ $# -lt 1 ]; then
    echo "Uso: $0 <video_entrada> [video_salida] [factor_escala (2|3|4)]"
    echo "Ejemplo: $0 mi_video.mp4 mi_video_2k.mp4 2"
    exit 1
fi

INPUT_VIDEO="$1"
OUTPUT_VIDEO="$2"
SCALE_FACTOR="${3:-2}" # Por defecto escala 2x (720p -> 1440p / 1080p -> 4K)

# Validar archivo de entrada
if [ ! -f "$INPUT_VIDEO" ]; then
    echo "Error: El archivo de entrada '$INPUT_VIDEO' no existe."
    exit 1
fi

# Definir nombre de salida por defecto si no se especificó
if [ -z "$OUTPUT_VIDEO" ]; then
    dir_name=$(dirname "$INPUT_VIDEO")
    base_name=$(basename "$INPUT_VIDEO")
    ext="${base_name##*.}"
    raw_name="${base_name%.*}"
    OUTPUT_VIDEO="$dir_name/${raw_name}_upscaled.mp4"
fi

# Generar identificador único de sesión para evitar conflictos temporales
SESSION_ID="bananascaler_$(date +%s)_$$"
TEMP_IN="/tmp/${SESSION_ID}_in"
TEMP_OUT="/tmp/${SESSION_ID}_out"

mkdir -p "$TEMP_IN" "$TEMP_OUT"

# 1. Detección Inteligente de Hardware Nvidia GPU
HAS_NVIDIA=false
if command -v nvidia-smi &> /dev/null && nvidia-smi &> /dev/null; then
    HAS_NVIDIA=true
fi

# Configuración de codecs en base a la tarjeta gráfica disponible
if [ "$HAS_NVIDIA" = true ]; then
    echo "[INFO] GPU NVIDIA detectada. Habilitando aceleración por hardware (NVDEC + NVENC)."
    DEC_FLAGS="-hwaccel cuda"
    ENC_FLAGS="-c:v hevc_nvenc -pix_fmt yuv420p"
else
    echo "[WARN] No se detectó GPU NVIDIA. Usando CPU para decodificación y codificación final."
    DEC_FLAGS=""
    ENC_FLAGS="-c:v libx265 -preset medium -crf 22 -pix_fmt yuv420p"
fi

# 2. Verificar que Real-ESRGAN esté instalado
if ! command -v realesrgan-ncnn-vulkan &> /dev/null; then
    echo "Error: 'realesrgan-ncnn-vulkan' no está instalado en el sistema o no está en el PATH."
    echo "Instálalo usando el instalador del repositorio o descárgalo manualmente."
    exit 1
fi

echo "[1/4] Extrayendo frames del video..."
ffmpeg -y -stats -loglevel warning $DEC_FLAGS -i "$INPUT_VIDEO" -f image2 -vcodec mjpeg -q:v 2 "$TEMP_IN/frame_%05d.jpg"

if [ $? -ne 0 ]; then
    echo "Error en la extracción de frames."
    rm -rf "$TEMP_IN" "$TEMP_OUT"
    exit 1
fi

# Detectar framerate original para mantener la sincronía exacta del audio
FRAMERATE=$(ffprobe -v error -select_streams v:0 -show_entries stream=r_frame_rate -of default=noprint_wrappers=1:nokey=1 "$INPUT_VIDEO")
echo "[2/4] Ejecutando escalado por red neuronal (Factor: ${SCALE_FACTOR}x)..."
realesrgan-ncnn-vulkan -i "$TEMP_IN" -o "$TEMP_OUT" -n realesr-animevideov3-x2 -s "$SCALE_FACTOR" -g 0

if [ $? -ne 0 ]; then
    echo "Error durante el escalado por red neuronal."
    rm -rf "$TEMP_IN" "$TEMP_OUT"
    exit 1
fi

echo "[3/4] Re-ensamblando video final y multiplexando audio..."
ffmpeg -y -stats -loglevel warning -framerate "$FRAMERATE" -i "$TEMP_OUT/frame_%05d.jpg" -i "$INPUT_VIDEO" \
       -map 0:v -map 1:a -c:a copy \
       $ENC_FLAGS \
       "$OUTPUT_VIDEO.tmp"

if [ $? -eq 0 ]; then
    mv "$OUTPUT_VIDEO.tmp" "$OUTPUT_VIDEO"
    echo "[4/4] ¡Proceso completado con éxito!"
    echo "Video final guardado en: $OUTPUT_VIDEO"
else
    echo "Error al compilar el video final."
    rm -f "$OUTPUT_VIDEO.tmp"
fi

# Limpieza
rm -rf "$TEMP_IN" "$TEMP_OUT"
