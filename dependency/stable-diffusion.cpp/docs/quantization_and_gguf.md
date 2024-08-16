## Quantization

You can specify the model weight type using the `--type` parameter. The weights are automatically converted when loading the model.

- `f16` for 16-bit floating-point
- `f32` for 32-bit floating-point
- `q8_0` for 8-bit integer quantization
- `q5_0` or `q5_1` for 5-bit integer quantization
- `q4_0` or `q4_1` for 4-bit integer quantization


### Memory Requirements of Stable Diffusion 1.x

| precision | f32  | f16  |q8_0  |q5_0  |q5_1  |q4_0  |q4_1  |
| ----         | ----  |----  |----  |----  |----  |----  |----  |
|  **Memory** (txt2img - 512 x 512) | ~2.8G | ~2.3G | ~2.1G | ~2.0G | ~2.0G | ~2.0G | ~2.0G |
|  **Memory** (txt2img - 512 x 512) *with Flash Attention* | ~2.4G | ~1.9G | ~1.6G | ~1.5G | ~1.5G | ~1.5G | ~1.5G |

## Convert to GGUF

You can also convert weights in the formats `ckpt/safetensors/diffusers` to gguf and perform quantization in advance, avoiding the need for quantization every time you load them.

For example:

```sh
./bin/sd -M convert -m ../models/v1-5-pruned-emaonly.safetensors -o  ../models/v1-5-pruned-emaonly.q8_0.gguf -v --type q8_0
```