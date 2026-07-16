# Memory: bananascaler

Persistent notes for agents working on **bananascaler**. Update as decisions are made.

## Known Constraints

- Real-ESRGAN model hardcoded to `realesr-animevideov3-x2`. Scale factor is a CLI parameter; model selection is not yet exposed.
- Temp dirs created in `/tmp/`. Systems with small `/tmp` partitions may fail on very long videos.
- Audio remux assumes single audio stream (`-map 1:a`). Multi-audio files may need explicit stream selection.

## Past Decisions

- Chose JPEG for intermediate frames over PNG: ~60-70% disk reduction at negligible quality cost for super-resolution input.
- Used `ncnn-vulkan` backend instead of CUDA-only: broader GPU vendor support via Vulkan.
- Atomic rename pattern adopted from day 1: non-negotiable.
