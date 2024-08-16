from typing import List, Optional, Union, Callable
import random
import ctypes
import multiprocessing
from PIL import Image


import nexa.gguf.sd.stable_diffusion_cpp as sd_cpp
from nexa.gguf.sd.stable_diffusion_cpp import GGMLType


from nexa.gguf.sd._internals_diffusion import _StableDiffusionModel, _UpscalerModel
# from nexa._logger_diffusion import set_verbose


class StableDiffusion:
    """High-level Python wrapper for a stable-diffusion.cpp model."""

    def __init__(
        self,
        model_path: str = "",
        vae_path: str = "",
        taesd_path: str = "",
        control_net_path: str = "",
        upscaler_path: str = "",
        lora_model_dir: str = "",
        embed_dir: str = "",
        stacked_id_embed_dir: str = "",
        vae_decode_only: bool = False,
        vae_tiling: bool = False,
        free_params_immediately: bool = False,
        n_threads: int = -1,
        wtype: str = "default",
        rng_type: int = sd_cpp.RNGType.STD_DEFAULT_RNG,
        schedule: int = sd_cpp.Schedule.DISCRETE,
        keep_clip_on_cpu: bool = False,
        keep_control_net_cpu: bool = False,
        keep_vae_on_cpu: bool = False,
        verbose: bool = True,
    ):
        """Load a stable-diffusion.cpp model from `model_path`.

        Examples:
            Basic usage

            >>> import stable_diffusion_cpp
            >>> model = stable_diffusion_cpp.StableDiffusion(
            ...     model_path="path/to/model",
            ... )
            >>> images = stable_diffusion.txt_to_img(prompt="a lovely cat")
            >>> images[0].save("output.png")

        Args:
            model_path: Path to the model.
            vae_path: Path to the vae.
            taesd_path: Path to the taesd.
            control_net_path: Path to the control net.
            upscaler_path: Path to esrgan model (Upscale images after generation).
            lora_model_dir: Lora model directory.
            embed_dir: Path to embeddings.
            stacked_id_embed_dir: Path to PHOTOMAKER stacked id embeddings.
            vae_decode_only: Process vae in decode only mode.
            vae_tiling: Process vae in tiles to reduce memory usage.
            free_params_immediately: Free parameters immediately after use.
            n_threads: Number of threads to use for generation.
            wtype: The weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0) (default: the weight type of the model file).
            rng_type: RNG (default: cuda).
            schedule: Denoiser sigma schedule (default: discrete).
            keep_clip_on_cpu: Keep clip in CPU (for low vram).
            keep_control_net_cpu: Keep controlnet in CPU (for low vram).
            keep_vae_on_cpu: Keep vae in CPU (for low vram).
            verbose: Print verbose output to stderr.

        Raises:
            ValueError: If a model path does not exist.

        Returns:
            A Stable Diffusion instance.
        """
        # Params
        self.model_path = model_path
        self.vae_path = vae_path
        self.taesd_path = taesd_path
        self.control_net_path = control_net_path
        self.lora_model_dir = lora_model_dir
        self.embed_dir = embed_dir
        self.stacked_id_embed_dir = stacked_id_embed_dir
        self.vae_decode_only = vae_decode_only
        self.vae_tiling = vae_tiling
        self.free_params_immediately = free_params_immediately
        self.n_threads = n_threads or max(multiprocessing.cpu_count() // 2, 1)
        self.wtype = wtype
        self.rng_type = rng_type
        self.schedule = schedule
        self.keep_clip_on_cpu = keep_clip_on_cpu
        self.keep_control_net_cpu = keep_control_net_cpu
        self.keep_vae_on_cpu = keep_vae_on_cpu

        # =========== Logging ===========

        self.verbose = verbose
        # set_verbose(verbose)

        # =========== SD Model loading ===========

        # Set the correspondoing weight type for type
        if self.wtype == "default":
            self.wtype = GGMLType.SD_TYPE_COUNT
        elif self.wtype == "f32":
            self.wtype = GGMLType.SD_TYPE_F32
        elif self.wtype == "f16":
            self.wtype = GGMLType.SD_TYPE_F16
        elif self.wtype == "q4_0":
            self.wtype = GGMLType.SD_TYPE_Q4_0
        elif self.wtype == "q4_1":
            self.wtype = GGMLType.SD_TYPE_Q4_1
        elif self.wtype == "q5_0":
            self.wtype = GGMLType.SD_TYPE_Q5_0
        elif self.wtype == "q5_1":
            self.wtype = GGMLType.SD_TYPE_Q5_1
        elif self.wtype == "q8_0":
            self.wtype = GGMLType.SD_TYPE_Q8_0
        else:
            raise ValueError(
                f"error: invalid weight format {self.wtype}, must be one of [default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0]"
            )

        # Load the Stable Diffusion model
        self._model = _StableDiffusionModel(
            self.model_path,
            self.vae_path,
            self.taesd_path,
            self.control_net_path,
            self.lora_model_dir,
            self.embed_dir,
            self.stacked_id_embed_dir,
            self.vae_decode_only,
            self.vae_tiling,
            self.free_params_immediately,
            self.n_threads,
            self.wtype,
            self.rng_type,
            self.schedule,
            self.keep_clip_on_cpu,
            self.keep_control_net_cpu,
            self.keep_vae_on_cpu,
            self.verbose,
        )

        # =========== Upscaling Model loading ===========

        self._upscaler = _UpscalerModel(
            upscaler_path, self.n_threads, self.wtype, self.verbose
        )

    @property
    def model(self) -> sd_cpp.sd_ctx_t_p:
        assert self._model.model is not None
        return self._model.model

    @property
    def upscaler(self) -> sd_cpp.upscaler_ctx_t_p:
        assert self._upscaler.upscaler is not None
        return self._upscaler.upscaler

    # ============================================
    # Text to Image
    # ============================================

    def txt_to_img(
        self,
        prompt: str,
        negative_prompt: str = "",
        clip_skip: int = -1,
        cfg_scale: float = 7.0,
        width: int = 512,
        height: int = 512,
        sample_method: int = sd_cpp.SampleMethod.EULER_A,
        sample_steps: int = 20,
        seed: int = 42,
        batch_count: int = 1,
        control_cond: Optional[Union[Image.Image, str]] = None,
        control_strength: float = 0.9,
        style_strength: float = 20.0,
        normalize_input: bool = False,
        input_id_images_path: str = "",
        canny: bool = False,
        upscale_factor: int = 1,
        progress_callback: Optional[Callable] = None,
    ) -> List[Image.Image]:
        """Generate images from a text prompt.

        Args:
            prompt: The prompt to render.
            negative_prompt: The negative prompt.
            clip_skip: Ignore last layers of CLIP network; 1 ignores none, 2 ignores one layer (default: -1).
            cfg_scale: Unconditional guidance scale: (default: 7.0).
            width: Image height, in pixel space (default: 512).
            height: Image width, in pixel space (default: 512).
            sample_method: Sampling method (default: "euler_a").
            sample_steps: Number of sample steps (default: 20).
            seed: RNG seed (default: 42, use random seed for < 0).
            batch_count: Number of images to generate.
            control_cond: A control condition image path or Pillow Image. (default: None).
            control_strength: Strength to apply Control Net (default: 0.9).
            style_strength: Strength for keeping input identity (default: 20%).
            normalize_input: Normalize PHOTOMAKER input id images.
            input_id_images_path: Path to PHOTOMAKER input id images dir.
            canny: Apply canny edge detection preprocessor to the control_cond image.
            upscale_factor: The image upscaling factor (default: 1).
            progress_callback: Callback function to call on each step end.

        Returns:
            A list of Pillow Images."""

        if self.model is None:
            raise Exception(
                "Stable diffusion model not loaded. Make sure you have set the `model_path`."
            )

        # Set a random seed if seed is negative
        if seed < 0:
            seed = random.randint(0, 10000)

        # ==================== Set the callback function ====================

        if progress_callback is not None:

            @sd_cpp.sd_progress_callback
            def sd_progress_callback(
                step: int,
                steps: int,
                time: float,
                data: ctypes.c_void_p,
            ):
                progress_callback(step, steps, time)

            sd_cpp.sd_set_progress_callback(sd_progress_callback, ctypes.c_void_p(0))

        # ==================== Convert the control condition to a C sd_image_t ====================

        control_cond = self._format_control_cond(
            control_cond, canny, self.control_net_path
        )

        # Run the txt2img to generate images
        c_images = sd_cpp.txt2img(
            self.model,
            prompt.encode("utf-8"),
            negative_prompt.encode("utf-8"),
            clip_skip,
            cfg_scale,
            width,
            height,
            sample_method,
            sample_steps,
            seed,
            batch_count,
            control_cond,
            control_strength,
            style_strength,
            normalize_input,
            input_id_images_path.encode("utf-8"),
        )

        # Convert the C array of images to a Python list of images
        return self._sd_image_t_p_to_images(c_images, batch_count, upscale_factor)

    # ============================================
    # Image to Image
    # ============================================

    def img_to_img(
        self,
        image: Union[Image.Image, str],
        prompt: str,
        negative_prompt: str = "",
        clip_skip: int = -1,
        cfg_scale: float = 7.0,
        width: int = 512,
        height: int = 512,
        sample_method: int = sd_cpp.SampleMethod.EULER_A,
        sample_steps: int = 20,
        strength: float = 0.75,
        seed: int = 42,
        batch_count: int = 1,
        control_cond: Optional[Union[Image.Image, str]] = None,
        control_strength: float = 0.9,
        style_strength: float = 20.0,
        normalize_input: bool = False,
        input_id_images_path: str = "",
        canny: bool = False,
        upscale_factor: int = 1,
        progress_callback: Optional[Callable] = None,
    ) -> List[Image.Image]:
        """Generate images from an image input and text prompt.

        Args:
            image: The input image path or Pillow Image to direct the generation.
            prompt: The prompt to render.
            negative_prompt: The negative prompt.
            clip_skip: Ignore last layers of CLIP network; 1 ignores none, 2 ignores one layer (default: -1).
            cfg_scale: Unconditional guidance scale: (default: 7.0).
            width: Image height, in pixel space (default: 512).
            height: Image width, in pixel space (default: 512).
            sample_method: Sampling method (default: "euler_a").
            sample_steps: Number of sample steps (default: 20).
            strength: Strength for noising/unnoising (default: 0.75).
            seed: RNG seed (default: 42, use random seed for < 0).
            batch_count: Number of images to generate.
            control_cond: A control condition image path or Pillow Image. (default: None).
            control_strength: Strength to apply Control Net (default: 0.9).
            style_strength: Strength for keeping input identity (default: 20%).
            normalize_input: Normalize PHOTOMAKER input id images.
            input_id_images_path: Path to PHOTOMAKER input id images dir.
            canny: Apply canny edge detection preprocessor to the control_cond image.
            upscale_factor: The image upscaling factor (default: 1).
            progress_callback: Callback function to call on each step end.

        Returns:
            A list of Pillow Images."""

        if self.model is None:
            raise Exception(
                "Stable diffusion model not loaded. Make sure you have set the `model_path`"
            )

        # Set a random seed if seed is negative
        if seed < 0:
            seed = random.randint(0, 10000)

        # ==================== Set the callback function ====================

        if progress_callback is not None:

            @sd_cpp.sd_progress_callback
            def sd_progress_callback(
                step: int,
                steps: int,
                time: float,
                data: ctypes.c_void_p,
            ):
                progress_callback(step, steps, time)

            sd_cpp.sd_set_progress_callback(sd_progress_callback, ctypes.c_void_p(0))

        # ==================== Convert the control condition to a C sd_image_t ====================

        control_cond = self._format_control_cond(
            control_cond, canny, self.control_net_path
        )

        # ==================== Resize the input image ====================

        image = self._resize_image(
            image, width, height
        )  # Input image and generated image must have the same size

        # ==================== Convert the image to a byte array ====================

        image_pointer = self._image_to_sd_image_t_p(image)

        c_images = sd_cpp.img2img(
            self.model,
            image_pointer,
            prompt.encode("utf-8"),
            negative_prompt.encode("utf-8"),
            clip_skip,
            cfg_scale,
            width,
            height,
            sample_method,
            sample_steps,
            strength,
            seed,
            batch_count,
            control_cond,
            control_strength,
            style_strength,
            normalize_input,
            input_id_images_path.encode("utf-8"),
        )
        return self._sd_image_t_p_to_images(c_images, batch_count, upscale_factor)

    # ============================================
    # Image to Video
    # ============================================

    def img_to_vid(
        self,
        image: Union[Image.Image, str],
        width: int = 512,
        height: int = 512,
        video_frames: int = 6,
        motion_bucket_id: int = 127,
        fps: int = 6,
        augmentation_level: float = 0.0,
        min_cfg: float = 1.0,
        cfg_scale: float = 7.0,
        sample_method: int = sd_cpp.SampleMethod.EULER_A,
        sample_steps: int = 20,
        strength: float = 0.75,
        seed: int = 42,
        progress_callback: Optional[Callable] = None,
    ) -> List[Image.Image]:
        """Generate a video from an image input.

        Args:
            image: The input image path or Pillow Image to direct the generation.
            width: Video height, in pixel space (default: 512).
            height: Video width, in pixel space (default: 512).
            video_frames: Number of frames in the video.
            motion_bucket_id: Motion bucket id.
            fps: Frames per second.
            augmentation_level: The augmentation level.
            min_cfg: The minimum cfg.
            cfg_scale: Unconditional guidance scale: (default: 7.0).
            sample_method: Sampling method (default: "euler_a").
            sample_steps: Number of sample steps (default: 20).
            strength: Strength for noising/unnoising (default: 0.75).
            seed: RNG seed (default: 42, use random seed for < 0).
            progress_callback: Callback function to call on each step end.

        Returns:
            A list of Pillow Images."""

        if self.model is None:
            raise Exception(
                "Stable diffusion model not loaded. Make sure you have set the `model_path`"
            )

        # # Set a random seed if seed is negative
        # if seed < 0:
        #     seed = random.randint(0, 10000)

        # # ==================== Set the callback function ====================

        # if progress_callback is not None:

        #     @sd_cpp.sd_progress_callback
        #     def sd_progress_callback(
        #         step: int,
        #         steps: int,
        #         time: float,
        #         data: ctypes.c_void_p,
        #     ):
        #         progress_callback(step, steps, time)

        #     sd_cpp.sd_set_progress_callback(sd_progress_callback, ctypes.c_void_p(0))

        # # ==================== Resize the input image ====================

        # image = self._resize_image(
        #     image, width, height
        # )  # Input image and generated image must have the same size

        # # ==================== Convert the image to a byte array ====================

        # image_pointer = self._image_to_sd_image_t_p(image)

        # c_video = sd_cpp.img2vid(
        #     self.model,
        #     image_pointer,
        #     width,
        #     height,
        #     video_frames,
        #     motion_bucket_id,
        #     fps,
        #     augmentation_level,
        #     min_cfg,
        #     cfg_scale,
        #     sample_method,
        #     sample_steps,
        #     strength,
        #     seed,
        # )

        # return self._sd_image_t_p_to_images(c_video, video_frames, 1)
        raise NotImplementedError("Not yet implemented.")

    # ============================================
    # Preprocess Canny
    # ============================================

    def preprocess_canny(
        self,
        image: Union[Image.Image, str],
        width: int = 512,
        height: int = 512,
        high_threshold: float = 0.08,
        low_threshold: float = 0.08,
        weak: float = 0.8,
        strong: float = 1.0,
        inverse: bool = False,
        output_as_c_uint8: bool = False,
    ) -> Image.Image:
        """Apply canny edge detection to an input image.

        Args:
            image: The input image path or Pillow Image.
            width: Output image height, in pixel space (default: 512).
            height: Output image width, in pixel space (default: 512).
            high_threshold: High edge detection threshold.
            low_threshold: Low edge detection threshold.
            weak: Weak edge thickness.
            strong: Strong edge thickness.
            inverse: Invert the edge detection.
            output_as_c_uint8: Return the output as a c_types uint8 pointer.

        Returns:
            A Pillow Image."""

        # Convert the image to a C uint8 pointer
        data, width, height = self._cast_image(image)

        # Run the preprocess canny
        c_image = sd_cpp.preprocess_canny(
            data,
            width,
            height,
            high_threshold,
            low_threshold,
            weak,
            strong,
            inverse,
        )

        # Return the c_image if output_as_c_uint8 (for running inside txt2img/img2img pipeline)
        if output_as_c_uint8:
            return c_image

        # Calculate the size of the data buffer (channels * width * height)
        buffer_size = 3 * width * height

        # Convert c_image to a Pillow Image
        image = self._c_array_to_bytes(c_image, buffer_size)
        image = self._bytes_to_image(image, width, height)
        return image

    # ============================================
    # Image Upscaling
    # ============================================

    def upscale(
        self,
        images: Union[List[Union[Image.Image, str]], Union[Image.Image, str]],
        upscale_factor: int = 4,
        progress_callback: Optional[Callable] = None,
    ) -> List[Image.Image]:
        """Upscale a list of images using the upscaler model.

        Args:
            images: A list of image paths or Pillow Images to upscale.
            upscale_factor: The image upscaling factor (default: 4).

        Returns:
            A list of Pillow Images."""

        if self.upscaler is None:
            raise Exception(
                "Upscaling model not loaded. Make sure you have set the `upscaler_path`"
            )

        # ==================== Set the callback function ====================

        if progress_callback is not None:

            @sd_cpp.sd_progress_callback
            def sd_progress_callback(
                step: int,
                steps: int,
                time: float,
                data: ctypes.c_void_p,
            ):
                progress_callback(step, steps, time)

            sd_cpp.sd_set_progress_callback(sd_progress_callback, ctypes.c_void_p(0))

        if not isinstance(images, list):
            images = [images]  # Wrap single image in a list

        # ==================== Upscale images ====================

        upscaled_images = []

        for image in images:

            # Convert the image to a byte array
            image_bytes = self._image_to_sd_image_t_p(image)

            # Upscale the image
            image = sd_cpp.upscale(
                self.upscaler,
                image_bytes,
                upscale_factor,
            )

            # Load the image from the C sd_image_t and convert it to a PIL Image
            image = self._dereference_sd_image_t_p(image)
            image = self._bytes_to_image(image["data"], image["width"], image["height"])
            upscaled_images.append(image)

        return upscaled_images

    # ============================================
    # Utility functions
    # ============================================

    def _resize_image(
        self, image: Union[Image.Image, str], width: int, height: int
    ) -> Image.Image:
        """Resize an image to a new width and height."""
        image, _, _ = self._format_image(image)

        # Resize the image if the width and height are different
        if image.width != width or image.height != height:
            image = image.resize((width, height), Image.Resampling.BILINEAR)
        return image

    def _format_image(
        self,
        image: Union[Image.Image, str],
    ) -> Image.Image:
        """Convert an image path or Pillow Image to a Pillow Image of RGBA format."""
        # Convert image path to image if str
        if isinstance(image, str):
            image = Image.open(image)

        # Convert any non RGBA to RGBA
        if image.format != "PNG":
            image = image.convert("RGBA")

        # Ensure the image is in RGB mode
        if image.mode != "RGB":
            image = image.convert("RGB")

        return image, image.width, image.height

    def _format_control_cond(
        self,
        control_cond: Optional[Union[Image.Image, str]],
        canny: bool,
        control_net_path: str,
    ) -> Optional[Image.Image]:
        """Convert an image path or Pillow Image to an C sd_image_t image."""

        if not control_cond:
            return None

        if not control_net_path:
            # log_event(1, "'control_net_path' not set. Skipping control condition.")
            return None

        if canny:
            # Convert Pillow Image to canny edge detection image then format into C sd_image_t
            image, width, height = self._format_image(control_cond)
            image = self.preprocess_canny(image, output_as_c_uint8=True)
            image = self._c_uint8_to_sd_image_t_p(image, width, height)
        else:
            # Convert Pillow Image to C sd_image_t
            image = self._image_to_sd_image_t_p(control_cond)
        return image

    # ============= Image to C uint8 pointer =============

    def _cast_image(self, image: Union[Image.Image, str]):
        """Cast a PIL Image to a C uint8 pointer."""

        image, width, height = self._format_image(image)

        # Convert the PIL Image to a byte array
        image_bytes = image.tobytes()

        data = ctypes.cast(
            (ctypes.c_byte * len(image_bytes))(*image_bytes),
            ctypes.POINTER(ctypes.c_uint8),
        )
        return data, width, height

    # ============= Image to C sd_image_t =============

    def _c_uint8_to_sd_image_t_p(
        self, image: ctypes.c_uint8, width, height, channel: int = 3
    ):
        # Create a new C sd_image_t
        c_image = sd_cpp.sd_image_t(
            width=width,
            height=height,
            channel=channel,
            data=image,
        )
        return c_image

    def _image_to_sd_image_t_p(self, image: Union[Image.Image, str]):
        """Convert a PIL Image or image path to a C sd_image_t."""

        data, width, height = self._cast_image(image)

        # Create a new C sd_image_t
        c_image = self._c_uint8_to_sd_image_t_p(data, width, height)
        return c_image

    # ============= C sd_image_t to Image =============

    def _c_array_to_bytes(self, c_array, buffer_size: int):
        return bytearray(
            ctypes.cast(c_array, ctypes.POINTER(ctypes.c_byte * buffer_size)).contents
        )

    def _dereference_sd_image_t_p(self, c_image: sd_cpp.sd_image_t):
        """Dereference a C sd_image_t pointer to a Python dictionary with height, width, channel and data (bytes)."""

        # Calculate the size of the data buffer
        buffer_size = c_image.channel * c_image.width * c_image.height

        image = {
            "width": c_image.width,
            "height": c_image.height,
            "channel": c_image.channel,
            "data": self._c_array_to_bytes(c_image.data, buffer_size),
        }
        return image

    def _image_slice(
        self, c_images: sd_cpp.sd_image_t_p, count: int, upscale_factor: int
    ):
        """Slice a C array of images."""
        image_array = ctypes.cast(
            c_images, ctypes.POINTER(sd_cpp.sd_image_t * count)
        ).contents

        images = []

        for i in range(count):
            c_image = image_array[i]

            # Upscale the image
            if upscale_factor > 1:
                if self.upscaler is None:
                    raise Exception(
                        "Upscaling model not loaded. Make sure you have set the `upscaler_path`"
                    )
                else:
                    c_image = sd_cpp.upscale(
                        self.upscaler,
                        c_image,
                        upscale_factor,
                    )

            image = self._dereference_sd_image_t_p(c_image)
            images.append(image)

        # Return the list of images
        return images

    def _sd_image_t_p_to_images(
        self, c_images: sd_cpp.sd_image_t_p, count: int, upscale_factor: int
    ):
        """Convert C sd_image_t_p images to a Python list of images."""

        # Convert C array to Python list of images
        images = self._image_slice(c_images, count, upscale_factor)

        # Convert each image to PIL Image
        for i in range(len(images)):
            image = images[i]
            images[i] = self._bytes_to_image(
                image["data"], image["width"], image["height"]
            )

        return images

    # ============= Bytes to Image =============

    def _bytes_to_image(self, byte_data: bytes, width: int, height: int):
        """Convert a byte array to a PIL Image."""
        image = Image.new("RGBA", (width, height))

        for y in range(height):
            for x in range(width):
                idx = (y * width + x) * 3
                image.putpixel(
                    (x, y),
                    (byte_data[idx], byte_data[idx + 1], byte_data[idx + 2], 255),
                )
        return image