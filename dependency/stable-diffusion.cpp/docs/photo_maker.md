## Using PhotoMaker to personalize image generation

You can use [PhotoMaker](https://github.com/TencentARC/PhotoMaker) to personalize generated images with your own ID.

**NOTE**, currently PhotoMaker **ONLY** works with **SDXL** (any SDXL model files will work).

Download PhotoMaker model file (in safetensor format) [here](https://huggingface.co/bssrdf/PhotoMaker). The official release of the model file (in .bin format) does not work with ```stablediffusion.cpp```.

- Specify the PhotoMaker model path using the `--stacked-id-embd-dir PATH` parameter.
- Specify the input images path using the `--input-id-images-dir PATH` parameter.
  - input images **must** have the same width and height for preprocessing (to be improved)

In prompt, make sure you have a class word followed by the trigger word ```"img"``` (hard-coded for now). The class word could be one of ```"man, woman, girl, boy"```. If input ID images contain asian faces, add ```Asian``` before the class
word.

Another PhotoMaker specific parameter:

- ```--style-ratio  (0-100)%```: default is 20 and 10-20 typically gets good results. Lower ratio means more faithfully following input ID (not necessarily better quality).

Other parameters recommended for running Photomaker:

- ```--cfg-scale 5.0```
- ```-H 1024```
- ```-W 1024```

If on low memory GPUs (<= 8GB), recommend running with ```--vae-on-cpu``` option to get artifact free images.

Example:

```bash
bin/sd -m ../models/sdxlUnstableDiffusers_v11.safetensors  --vae ../models/sdxl_vae.safetensors --stacked-id-embd-dir ../models/photomaker-v1.safetensors --input-id-images-dir ../assets/examples/scarletthead_woman -p "a girl img, retro futurism, retro game art style but extremely beautiful, intricate details, masterpiece, best quality, space-themed, cosmic, celestial, stars, galaxies, nebulas, planets, science fiction, highly detailed" -n "realistic, photo-realistic, worst quality, greyscale, bad anatomy, bad hands, error, text" --cfg-scale 5.0  --sampling-method euler -H 1024 -W 1024 --style-ratio 10 --vae-on-cpu -o output.png
```