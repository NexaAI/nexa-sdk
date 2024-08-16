## Using ESRGAN to upscale results

You can use ESRGAN to upscale the generated images. At the moment, only the [RealESRGAN_x4plus_anime_6B.pth](https://github.com/xinntao/Real-ESRGAN/releases/download/v0.2.2.4/RealESRGAN_x4plus_anime_6B.pth) model is supported. Support for more models of this architecture will be added soon.

- Specify the model path using the `--upscale-model PATH` parameter. example:

```bash
sd -m ../models/v1-5-pruned-emaonly.safetensors -p "a lovely cat" --upscale-model ../models/RealESRGAN_x4plus_anime_6B.pth
```
