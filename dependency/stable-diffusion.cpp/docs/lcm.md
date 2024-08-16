## LCM/LCM-LoRA

- Download LCM-LoRA form https://huggingface.co/latent-consistency/lcm-lora-sdv1-5
- Specify LCM-LoRA by adding `<lora:lcm-lora-sdv1-5:1>` to prompt
- It's advisable to set `--cfg-scale` to `1.0` instead of the default `7.0`. For `--steps`, a range of `2-8` steps is recommended. For `--sampling-method`, `lcm`/`euler_a` is recommended.

Here's a simple example:

```
./bin/sd -m ../models/v1-5-pruned-emaonly.safetensors -p "a lovely cat<lora:lcm-lora-sdv1-5:1>" --steps 4 --lora-model-dir ../models -v --cfg-scale 1
```

| without LCM-LoRA (--cfg-scale 7)  | with LCM-LoRA (--cfg-scale 1)  |
| ----  |----    |
| ![](../assets/without_lcm.png) |![](../assets/with_lcm.png)  |