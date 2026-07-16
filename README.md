# GPU Video Upscaler

Un script de consola en Bash rápido y optimizado para el escalado de videos mediante redes neuronales de súper-resolución. Combina la velocidad de procesamiento gráfico de **Real-ESRGAN (Vulkan)** con la flexibilidad de **FFmpeg** y aceleración de hardware por GPU NVIDIA.

## Requisitos

El script detectará de forma automática tu configuración de hardware y utilizará aceleración por GPU si tienes una tarjeta gráfica NVIDIA apta.

*   **FFmpeg** (con soporte para codificación `hevc_nvenc` preferiblemente)
*   **Real-ESRGAN** (`realesrgan-ncnn-vulkan` en tu PATH del sistema)
*   **Controladores NVIDIA y CUDA** (opcional, pero altamente recomendado para máxima velocidad)

## Instalación rápida (Arch Linux / CachyOS)

```bash
# Instalar FFmpeg
sudo pacman -S ffmpeg

# Descargar Real-ESRGAN y agregarlo al PATH local
mkdir -p ~/.local/share/realesrgan && cd ~/.local/share/realesrgan
curl -sL -O "https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.5.0/realesrgan-ncnn-vulkan-20220424-ubuntu.zip"
unzip realesrgan-ncnn-vulkan-20220424-ubuntu.zip
rm realesrgan-ncnn-vulkan-20220424-ubuntu.zip
chmod +x realesrgan-ncnn-vulkan
ln -sf ~/.local/share/realesrgan/realesrgan-ncnn-vulkan ~/.local/bin/realesrgan-ncnn-vulkan
```

## Uso

El script acepta argumentos dinámicos de entrada y salida:

```bash
./upscale.sh <video_entrada> [video_salida] [factor_escala (2|3|4)]
```

### Ejemplos prácticos

1.  **Escalado rápido (automático a 2x con salida por defecto en la misma carpeta):**
    ```bash
    ./upscale.sh pelicula.mp4
    # Generará: pelicula_upscaled.mp4
    ```

2.  **Especificar nombre del archivo resultante y escala 4x:**
    ```bash
    ./upscale.sh mi_video.mp4 mi_video_4k.mp4 4
    ```

3.  **Ejecutar en segundo plano de forma segura (ideal para videos largos):**
    ```bash
    nohup ./upscale.sh input.mp4 output.mp4 2 > upscale.log 2>&1 &
    ```

## Características Técnicas
*   **Escritura Atómica:** Guarda los datos en archivos `.tmp` y solo realiza el renombrado al finalizar correctamente. Evita archivos corruptos si se interrumpe a la mitad.
*   **Consumo Eficiente de CPU/Disco:** Extrae frames temporales ligeros en formato `.jpg` en lugar de `.png` no comprimidos, ahorrando espacio en disco e hilos de CPU.
*   **Aceleración Dual GPU:** Realiza la decodificación (`NVDEC`) y la codificación final (`NVENC`) en hardware de video de NVIDIA, manteniendo baja la temperatura de tu procesador principal.
