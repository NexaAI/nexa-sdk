from typing import List, Dict, Optional, Union, Callable
import random
import ctypes
import multiprocessing
import contextlib
from PIL import Image
import nexa.gguf.sd.stable_diffusion_cpp as sd_cpp
from nexa.gguf.sd.stable_diffusion_cpp import GGMLType, RNGType, Schedule, SampleMethod
from nexa.gguf.sd._internals_diffusion import _StableDiffusionModel, _UpscalerModel
from nexa.gguf.sd._utils_diffusion import suppress_stdout_stderr


class StableDiffusion:
    """High-level Python wrapper for a stable-diffusion.cpp model."""

    def __init__(
        self,
        model_path: str = "",
        clip_l_path: str = "",
        t5xxl_path: str = "",
        diffusion_model_path: str = "",
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
        wtype: Union[str, GGMLType, int, float, None] = "default",
        rng_type: Union[str, RNGType, int, float, None] = "default",
        schedule: Union[str, Schedule, int, float, None] = "default",
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
            clip_l_path: Path to the clip_l.
            t5xxl_path: Path to the t5xxl.
            diffusion_model_path: Path to the diffusion model.
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
            wtype: The weight type (default: automatically determines the weight type of the model file).
            rng_type: Random number generator.
            schedule: Denoiser sigma schedule.
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
        self.clip_l_path = clip_l_path
        self.t5xxl_path = t5xxl_path
        self.diffusion_model_path = diffusion_model_path
        self.vae_path = vae_path
        self.taesd_path = taesd_path
        self.control_net_path = control_net_path
        self.lora_model_dir = lora_model_dir
        self.embed_dir = embed_dir
        self.stacked_id_embed_dir = stacked_id_embed_dir
        self.vae_decode_only = vae_decode_only
        self.vae_tiling = vae_tiling
        self.free_params_immediately = free_params_immediately
        self.n_threads = n_threads or max(multiprocessing.cpu_count() // 2, 1) # Default to half the number of CPUs
        self.wtype = wtype
        self.rng_type = rng_type
        self.schedule = schedule
        self.keep_clip_on_cpu = keep_clip_on_cpu
        self.keep_control_net_cpu = keep_control_net_cpu
        self.keep_vae_on_cpu = keep_vae_on_cpu
        self._stack = contextlib.ExitStack()

        # =========== Logging ===========

        self.verbose = verbose
        # set_verbose(verbose)

        # =========== Validate string and int inputs ===========

        self.wtype = validate_and_set_input(self.wtype, GGML_TYPE_MAP, "wtype")
        self.rng_type = validate_and_set_input(self.rng_type, RNG_TYPE_MAP, "rng_type")
        self.schedule = validate_and_set_input(self.schedule, SCHEDULE_MAP, "schedule")

        # =========== SD Model loading ===========

        self._model = self._stack.enter_context(
            contextlib.closing(
                _StableDiffusionModel(
                    self.model_path,
                    self.clip_l_path,
                    self.t5xxl_path,
                    self.diffusion_model_path,
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
            )
        )

        # =========== Upscaling Model loading ===========

        self._upscaler = self._stack.enter_context(
            contextlib.closing(
                _UpscalerModel(
                    upscaler_path,
                    self.n_threads,
                    self.wtype,
                    self.verbose,
                )
            )
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
        guidance: float = 3.5,
        width: int = 512,
        height: int = 512,
        sample_method: Union[str, SampleMethod, int, float, None] = "euler_a",
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
            clip_skip: Ignore last layers of CLIP network; 1 ignores none, 2 ignores one layer.
            cfg_scale: Unconditional guidance scale.
            guidance: Guidance scale.
            width: Image height, in pixel space.
            height: Image width, in pixel space.
            sample_method: Sampling method.
            sample_steps: Number of sample steps.
            seed: RNG seed (default: 42, use random seed for < 0).
            batch_count: Number of images to generate.
            control_cond: A control condition image path or Pillow Image.
            control_strength: Strength to apply Control Net.
            style_strength: Strength for keeping input identity (default: 20%).
            normalize_input: Normalize PHOTOMAKER input id images.
            input_id_images_path: Path to PHOTOMAKER input id images dir.
            canny: Apply canny edge detection preprocessor to the control_cond image.
            upscale_factor: The image upscaling factor.
            progress_callback: Callback function to call on each step end.

        Returns:
            A list of Pillow Images."""

        if self.model is None:
            raise Exception("Stable diffusion model not loaded.")

        # =========== Validate string and int inputs ===========

        sample_method = validate_and_set_input(sample_method, SAMPLE_METHOD_MAP, "sample_method")

        # Ensure dimensions are multiples of 64
        width = validate_dimensions(width, "width")
        height = validate_dimensions(height, "height")

        # =========== Set seed ===========

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

        control_cond = self._format_control_cond(control_cond, canny, self.control_net_path)

        with suppress_stdout_stderr(disable=True):
            # Generate images
            c_images = sd_cpp.txt2img(
                self.model,
                prompt.encode("utf-8"),
                negative_prompt.encode("utf-8"),
                clip_skip,
                cfg_scale,
                guidance,
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
        guidance: float = 3.5,
        width: int = 512,
        height: int = 512,
        sample_method: Union[str, SampleMethod, int, float, None] = "euler_a",
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
            clip_skip: Ignore last layers of CLIP network; 1 ignores none, 2 ignores one layer.
            cfg_scale: Unconditional guidance scale.
            guidance: Guidance scale.
            width: Image height, in pixel space.
            height: Image width, in pixel space.
            sample_method: Sampling method.
            sample_steps: Number of sample steps.
            strength: Strength for noising/unnoising.
            seed: RNG seed (default: 42, use random seed for < 0).
            batch_count: Number of images to generate.
            control_cond: A control condition image path or Pillow Image.
            control_strength: Strength to apply Control Net.
            style_strength: Strength for keeping input identity (default: 20%).
            normalize_input: Normalize PHOTOMAKER input id images.
            input_id_images_path: Path to PHOTOMAKER input id images dir.
            canny: Apply canny edge detection preprocessor to the control_cond image.
            upscale_factor: The image upscaling factor.
            progress_callback: Callback function to call on each step end.

        Returns:
            A list of Pillow Images."""

        if self.model is None:
            raise Exception("Stable diffusion model not loaded.")

        # =========== Validate string and int inputs ===========

        sample_method = validate_and_set_input(sample_method, SAMPLE_METHOD_MAP, "sample_method")

        # Ensure dimensions are multiples of 64
        width = validate_dimensions(width, "width")
        height = validate_dimensions(height, "height")

        # =========== Set seed ===========

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

        control_cond = self._format_control_cond(control_cond, canny, self.control_net_path)

        # ==================== Resize the input image ====================

        image = self._resize_image(image, width, height) # Input image and generated image must have the same size

        # ==================== Convert the image to a byte array ====================

        image_pointer = self._image_to_sd_image_t_p(image)

        with suppress_stdout_stderr(disable=True):
            # Generate images
            c_images = sd_cpp.img2img(
                self.model,
                image_pointer,
                prompt.encode("utf-8"),
                negative_prompt.encode("utf-8"),
                clip_skip,
                cfg_scale,
                guidance,
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
        sample_method: Union[str, SampleMethod, int, float, None] = "euler_a",
        sample_steps: int = 20,
        strength: float = 0.75,
        seed: int = 42,
        progress_callback: Optional[Callable] = None,
    ) -> List[Image.Image]:
        """Generate a video from an image input.

        Args:
            image: The input image path or Pillow Image to direct the generation.
            width: Video height, in pixel space.
            height: Video width, in pixel space.
            video_frames: Number of frames in the video.
            motion_bucket_id: Motion bucket id.
            fps: Frames per second.
            augmentation_level: The augmentation level.
            min_cfg: The minimum cfg.
            cfg_scale: Unconditional guidance scale.
            sample_method: Sampling method.
            sample_steps: Number of sample steps.
            strength: Strength for noising/unnoising.
            seed: RNG seed (default: 42, use random seed for < 0).
            progress_callback: Callback function to call on each step end.

        Returns:
            A list of Pillow Images."""

        raise NotImplementedError("Not yet implemented.")

        # if self.model is None:
        #     raise Exception("Stable diffusion model not loaded.")

        # # =========== Validate string and int inputs ===========

        # sample_method = validate_and_set_input(sample_method, SAMPLE_METHOD_MAP, "sample_method")

        # # Ensure dimensions are multiples of 64
        # width = validate_dimensions(width, "width")
        # height = validate_dimensions(height, "height")

        # # =========== Set seed ===========

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

        # with suppress_stdout_stderr(disable=self.verbose):
        #     # Generate the video
        #     c_video = sd_cpp.img2vid(
        #         self.model,
        #         image_pointer,
        #         width,
        #         height,
        #         video_frames,
        #         motion_bucket_id,
        #         fps,
        #         augmentation_level,
        #         min_cfg,
        #         cfg_scale,
        #         sample_method,
        #         sample_steps,
        #         strength,
        #         seed,
        #     )

        # return self._sd_image_t_p_to_images(c_video, video_frames, 1)

    # ============================================
    # Preprocess Canny
    # ============================================

    def preprocess_canny(
        self,
        image: Union[Image.Image, str],
        high_threshold: float = 0.08,
        low_threshold: float = 0.08,
        weak: float = 0.8,
        strong: float = 1.0,
        inverse: bool = False,
        output_as_c_uint8: bool = False,
    ) -> Image.Image:
        """Apply canny edge detection to an input image. Width and height determined automatically.

        Args:
            image: The input image path or Pillow Image.
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

        with suppress_stdout_stderr(disable=self.verbose):
            # Run the preprocess canny
            c_image = sd_cpp.preprocess_canny(
                data,
                int(width),
                int(height),
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
            upscale_factor: The image upscaling factor.

        Returns:
            A list of Pillow Images."""

        if self.upscaler is None:
            raise Exception("Upscaling model not loaded.")

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

            with suppress_stdout_stderr(disable=self.verbose):
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

    def _resize_image(self, image: Union[Image.Image, str], width: int, height: int) -> Image.Image:
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

        # if not control_net_path:
        #     log_event(1, "'control_net_path' not set. Skipping control condition.")
        #     return None

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

    def _c_uint8_to_sd_image_t_p(self, image: ctypes.c_uint8, width, height, channel: int = 3):
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
        return bytearray(ctypes.cast(c_array, ctypes.POINTER(ctypes.c_byte * buffer_size)).contents)

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

    def _image_slice(self, c_images: sd_cpp.sd_image_t, count: int, upscale_factor: int):
        """Slice a C array of images."""
        image_array = ctypes.cast(c_images, ctypes.POINTER(sd_cpp.sd_image_t * count)).contents

        images = []

        for i in range(count):
            c_image = image_array[i]

            # Upscale the image
            if upscale_factor > 1:
                if self.upscaler is None:
                    raise Exception("Upscaling model not loaded.")
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

    def _sd_image_t_p_to_images(self, c_images: sd_cpp.sd_image_t, count: int, upscale_factor: int):
        """Convert C sd_image_t_p images to a Python list of images."""

        # Convert C array to Python list of images
        images = self._image_slice(c_images, count, upscale_factor)

        # Convert each image to PIL Image
        for i in range(len(images)):
            image = images[i]
            images[i] = self._bytes_to_image(image["data"], image["width"], image["height"])

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

    def __setstate__(self, state):
        self.__init__(**state)

    def close(self) -> None:
        """Explicitly free the model from memory."""
        self._stack.close()

    def __del__(self) -> None:
        self.close()


# ============================================
# Validate dimension parameters
# ============================================


def validate_dimensions(dimension: int | float, attribute_name: str) -> int:
    """Dimensions must be a multiple of 64 otherwise a GGML_ASSERT error is encountered."""
    dimension = int(dimension)
    if dimension <= 0 or dimension % 64 != 0:
        raise ValueError(f"The '{attribute_name}' must be a multiple of 64.")
    return dimension


# ============================================
# Mapping from strings to constants
# ============================================


def validate_and_set_input(user_input: Union[str, int, float], type_map: Dict, attribute_name: str):
    """Validate an input strinbg or int from a map of strings to integers."""
    if isinstance(user_input, float):
        user_input = int(user_input)  # Convert float to int

    # Handle string input
    if isinstance(user_input, str):
        user_input = user_input.strip().lower()
        if user_input in type_map:
            return int(type_map[user_input])
        else:
            raise ValueError(f"Invalid {attribute_name} type '{user_input}'. Must be one of {list(type_map.keys())}.")
    elif isinstance(user_input, int) and user_input in type_map.values():
        return int(user_input)
    else:
        raise ValueError(f"{attribute_name} must be a string or an integer and must be a valid type.")


RNG_TYPE_MAP = {
    "default": RNGType.STD_DEFAULT_RNG,
    "cuda": RNGType.CUDA_RNG,
}

SAMPLE_METHOD_MAP = {
    "euler_a": SampleMethod.EULER_A,
    "euler": SampleMethod.EULER,
    "heun": SampleMethod.HEUN,
    "dpm2": SampleMethod.DPM2,
    "dpmpp2s_a": SampleMethod.DPMPP2S_A,
    "dpmpp2m": SampleMethod.DPMPP2M,
    "dpmpp2mv2": SampleMethod.DPMPP2Mv2,
    "ipndm": SampleMethod.IPNDM,
    "ipndm_v": SampleMethod.IPNDM_V,
    "lcm": SampleMethod.LCM,
}

SCHEDULE_MAP = {
    "default": Schedule.DEFAULT,
    "discrete": Schedule.DISCRETE,
    "karras": Schedule.KARRAS,
    "exponential": Schedule.EXPONENTIAL,
    "ays": Schedule.AYS,
    "gits": Schedule.GITS,
}

GGML_TYPE_MAP = {
    "f32": GGMLType.SD_TYPE_F32,
    "f16": GGMLType.SD_TYPE_F16,
    "q4_0": GGMLType.SD_TYPE_Q4_0,
    "q4_1": GGMLType.SD_TYPE_Q4_1,
    "q5_0": GGMLType.SD_TYPE_Q5_0,
    "q5_1": GGMLType.SD_TYPE_Q5_1,
    "q8_0": GGMLType.SD_TYPE_Q8_0,
    "q8_1": GGMLType.SD_TYPE_Q8_1,
    # k-quantizations
    "q2_k": GGMLType.SD_TYPE_Q2_K,
    "q3_k": GGMLType.SD_TYPE_Q3_K,
    "q4_k": GGMLType.SD_TYPE_Q4_K,
    "q5_k": GGMLType.SD_TYPE_Q5_K,
    "q6_k": GGMLType.SD_TYPE_Q6_K,
    "q8_k": GGMLType.SD_TYPE_Q8_K,
    "iq2_xxs": GGMLType.SD_TYPE_IQ2_XXS,
    "iq2_xs": GGMLType.SD_TYPE_IQ2_XS,
    "iq3_xxs": GGMLType.SD_TYPE_IQ3_XXS,
    "iq1_s": GGMLType.SD_TYPE_IQ1_S,
    "iq4_nl": GGMLType.SD_TYPE_IQ4_NL,
    "iq3_s": GGMLType.SD_TYPE_IQ3_S,
    "iq2_s": GGMLType.SD_TYPE_IQ2_S,
    "iq4_xs": GGMLType.SD_TYPE_IQ4_XS,
    "i8": GGMLType.SD_TYPE_I8,
    "i16": GGMLType.SD_TYPE_I16,
    "i32": GGMLType.SD_TYPE_I32,
    "i64": GGMLType.SD_TYPE_I64,
    "f64": GGMLType.SD_TYPE_F64,
    "iq1_m": GGMLType.SD_TYPE_IQ1_M,
    "bf16": GGMLType.SD_TYPE_BF16,
    "q4_0_4_4": GGMLType.SD_TYPE_Q4_0_4_4,
    "q4_0_4_8": GGMLType.SD_TYPE_Q4_0_4_8,
    "q4_0_8_8": GGMLType.SD_TYPE_Q4_0_8_8,
    # Default
    "default": GGMLType.SD_TYPE_COUNT,
}
