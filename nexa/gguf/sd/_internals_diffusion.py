import os
from contextlib import ExitStack

from nexa.gguf.sd._utils_diffusion import suppress_stdout_stderr

import nexa.gguf.sd.stable_diffusion_cpp as sd_cpp


# ============================================
# Stable Diffusion Model
# ============================================


class _StableDiffusionModel:
    """Intermediate Python wrapper for a stable-diffusion.cpp stable_diffusion_model."""

    _free_sd_ctx = None
    # NOTE: this must be "saved" here to avoid exceptions when calling __del__

    def __init__(
        self,
        model_path: str,
        clip_l_path: str,
        t5xxl_path: str,
        diffusion_model_path: str,
        vae_path: str,
        taesd_path: str,
        control_net_path: str,
        lora_model_dir: str,
        embed_dir: str,
        stacked_id_embed_dir: str,
        vae_decode_only: bool,
        vae_tiling: bool,
        free_params_immediately: bool,
        n_threads: int,
        wtype: int,
        rng_type: int,
        schedule: int,
        keep_clip_on_cpu: bool,
        keep_control_net_cpu: bool,
        keep_vae_on_cpu: bool,
        verbose: bool,
    ):
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
        self.n_threads = n_threads
        self.wtype = wtype
        self.rng_type = rng_type
        self.schedule = schedule
        self.keep_clip_on_cpu = keep_clip_on_cpu
        self.keep_control_net_cpu = keep_control_net_cpu
        self.keep_vae_on_cpu = keep_vae_on_cpu
        self.verbose = verbose
        self._exit_stack = ExitStack()

        self.model = None

        # Load the free_sd_ctx function
        self._free_sd_ctx = sd_cpp._lib.free_sd_ctx

        # Load the model from the file if the path is provided
        if model_path:
            if not os.path.exists(model_path):
                raise ValueError(f"Model path does not exist: {model_path}")

        if diffusion_model_path:
            if not os.path.exists(diffusion_model_path):
                raise ValueError(f"Diffusion model path does not exist: {diffusion_model_path}")

        if model_path or diffusion_model_path:
            with suppress_stdout_stderr(disable=verbose):
                # Load the Stable Diffusion model ctx
                self.model = sd_cpp.new_sd_ctx(
                    self.model_path.encode("utf-8"),
                    self.clip_l_path.encode("utf-8"),
                    self.t5xxl_path.encode("utf-8"),
                    self.diffusion_model_path.encode("utf-8"),
                    self.vae_path.encode("utf-8"),
                    self.taesd_path.encode("utf-8"),
                    self.control_net_path.encode("utf-8"),
                    self.lora_model_dir.encode("utf-8"),
                    self.embed_dir.encode("utf-8"),
                    self.stacked_id_embed_dir.encode("utf-8"),
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
                )

            # Check if the model was loaded successfully
            if self.model is None:
                raise ValueError(f"Failed to load model from file: {model_path}")

        def free_ctx():
            """Free the model from memory."""
            if self.model is not None and self._free_sd_ctx is not None:
                self._free_sd_ctx(self.model)
                self.model = None

        self._exit_stack.callback(free_ctx)

    def close(self):
        """Closes the exit stack, ensuring all context managers are exited."""
        self._exit_stack.close()

    def __del__(self):
        """Free memory when the object is deleted."""
        self.close()


# ============================================
# Upscaler Model
# ============================================


class _UpscalerModel:
    """Intermediate Python wrapper for an Esrgan image upscaling model."""

    _free_upscaler_ctx = None
    # NOTE: this must be "saved" here to avoid exceptions when calling __del__

    def __init__(
        self,
        upscaler_path: str,
        n_threads: int,
        wtype: int,
        verbose: bool,
    ):
        self.upscaler_path = upscaler_path
        self.n_threads = n_threads
        self.wtype = wtype
        self.verbose = verbose
        self._exit_stack = ExitStack()

        self.upscaler = None

        # Load the model from the file if the path is provided
        if upscaler_path:

            # Load the free_upscaler_ctx function
            self._free_upscaler_ctx = sd_cpp._lib.free_upscaler_ctx

            if not os.path.exists(upscaler_path):
                raise ValueError(f"Upscaler model path does not exist: {upscaler_path}")

            # Load the image upscaling model ctx
            self.upscaler = sd_cpp.new_upscaler_ctx(upscaler_path.encode("utf-8"), self.n_threads, self.wtype)

            # Check if the model was loaded successfully
            if self.upscaler is None:
                raise ValueError(f"Failed to load upscaler model from file: {upscaler_path}")

        def free_ctx():
            """Free the model from memory."""
            if self.upscaler is not None and self._free_upscaler_ctx is not None:
                self._free_upscaler_ctx(self.upscaler)
                self.upscaler = None

        self._exit_stack.callback(free_ctx)

    def close(self):
        """Closes the exit stack, ensuring all context managers are exited."""
        self._exit_stack.close()

    def __del__(self):
        """Free memory when the object is deleted."""
        self.close()
