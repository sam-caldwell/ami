GPU Demo (S-8)

This example demonstrates authoring inline gpu(...) blocks for Metal, OpenCL, and CUDA and wiring them up as a Transform worker. The build produces a workers shared library with dynamic GPU dispatch.

Quick start:
- macOS (Metal): AMI_GPU_FORCE_METAL=1
- Linux (OpenCL): AMI_GPU_FORCE_OPENCL=1 (install libOpenCL)
- Linux (CUDA): AMI_GPU_FORCE_CUDA=1 (NVIDIA driver provides libcuda)

Build:
- cd examples/gpu_demo
- ami build

Notes:
- The demo kernel writes out[i] = i*3 for n elements.
- Attribute list forms are accepted: grid=[gx,gy,gz], tpg=[tx,ty,tz].
- Optional args spec supports "1buf1u32" (default) and "1buf" for OpenCL/CUDA; Metal path currently supports only 1buf1u32.

