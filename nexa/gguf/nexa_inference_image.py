import argparse
import logging
import multiprocessing
import os
import sys
import time
from pathlib import Path
from nexa.constants import (
    DEFAULT_IMG_GEN_PARAMS,
    EXIT_REMINDER,
    NEXA_RUN_MODEL_PRECISION_MAP,
    DEFAULT_IMG_GEN_PARAMS_LCM,
    DEFAULT_IMG_GEN_PARAMS_TURBO,
    NEXA_RUN_MODEL_MAP_FLUX,
    NEXA_RUN_T5XXL_MAP,
)
from nexa.utils import SpinningCursorAnimation, nexa_prompt
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr

from streamlit.web import cli as stcli
from nexa.general import pull_model

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)

# image generation retry attempts
RETRY_ATTEMPTS = (
    3  # a temporary fix for the issue of segmentation fault for stable-diffusion-cpp
)

# FLUX vae and clip model paths
FLUX_VAE_PATH = "FLUX.1-schnell:ae-fp16"
FLUX_CLIP_L_PATH = "FLUX.1-schnell:clip_l-fp16"

class NexaImageInference:
    """
    A class used for loading image models and running image generation.

    Methods:
        txt2img(prompt): Generate images from text.
        img2img(image_path, prompt): Generate images from an image.
        run_txt2img: Run the text-to-image generation loop.
        run_img2img: Run the image-to-image generation loop.
        run_streamlit: Run the Streamlit UI.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str): Local path of the model.
    num_inference_steps (int): Number of inference steps.
    width (int): Width of the output image.
    height (int): Height of the output image.
    guidance_scale (float): Guidance scale for diffusion.
    output_path (str): Output path for the generated image.
    random_seed (int): Random seed for image generation.
    lora_dir (str): Path to directory containing LoRA files.
    lora_path (str): Path to a LoRA file to apply to the model.
    wtype (str): Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0).
    control_net_path (str): Path to control net model.
    control_image_path (str): Path to image condition for Control Net.
    control_strength (float): Strength to apply Control Net.
    streamlit (bool): Run the inference in Streamlit UI.
    profiling (bool): Enable profiling logs for the inference process.
    """

    def __init__(self, model_path=None, local_path=None, **kwargs):
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")
        
        self.model_path = model_path
        self.downloaded_path = local_path

        # FLUX model components
        self.t5xxl_path = None
        self.ae_path = None
        self.clip_l_path = None
        self.t5xxl_downloaded_path = None
        self.ae_downloaded_path = None
        self.clip_l_downloaded_path = None

        # Download base model if not provided
        if self.downloaded_path is None:
            self.downloaded_path, _ = pull_model(self.model_path, **kwargs)
            if self.downloaded_path is None:
                logging.error(
                    f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                    exc_info=True,
                )
                exit(1)

        # Check if the model is a FLUX model and download additional components
        if self.model_path in NEXA_RUN_MODEL_MAP_FLUX:
            self.t5xxl_path = NEXA_RUN_T5XXL_MAP.get(model_path)
            self.ae_path = FLUX_VAE_PATH
            self.clip_l_path = FLUX_CLIP_L_PATH

            if self.t5xxl_path:
                self.t5xxl_downloaded_path, _ = pull_model(self.t5xxl_path, **kwargs)
            if self.ae_path:
                self.ae_downloaded_path, _ = pull_model(self.ae_path, **kwargs)
            if self.clip_l_path:
                self.clip_l_downloaded_path, _ = pull_model(self.clip_l_path, **kwargs)
        if "lcm-dreamshaper" in self.model_path or "flux" in self.model_path:
            self.params = DEFAULT_IMG_GEN_PARAMS_LCM.copy() # both lcm-dreamshaper and flux use the same params
        elif "sdxl-turbo" in self.model_path:
            self.params = DEFAULT_IMG_GEN_PARAMS_TURBO.copy()
        else:
            self.params = DEFAULT_IMG_GEN_PARAMS.copy()

        self.profiling = kwargs.get("profiling", False)
        self.params.update({k: v for k, v in kwargs.items() if v is not None})
        if not kwargs.get("streamlit", False):
            self._load_model(model_path)
            if self.model is None:
                logging.error("Failed to load the model or pipeline.")
                exit(1)

    @SpinningCursorAnimation()
    def _load_model(self, model_path: str):
        with suppress_stdout_stderr():
            from nexa.gguf.sd.stable_diffusion import StableDiffusion
            if self.t5xxl_downloaded_path and self.ae_downloaded_path and self.clip_l_downloaded_path:
                self.model = StableDiffusion(
                    diffusion_model_path=self.downloaded_path,
                    clip_l_path=self.clip_l_downloaded_path,
                    t5xxl_path=self.t5xxl_downloaded_path,
                    vae_path=self.ae_downloaded_path,
                    n_threads=self.params.get("n_threads", multiprocessing.cpu_count()),
                    wtype=self.params.get(
                        "wtype", NEXA_RUN_MODEL_PRECISION_MAP.get(model_path, "default")
                    ),  # Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
                    verbose=self.profiling,
                )
            else:
                self.model = StableDiffusion(
                    model_path=self.downloaded_path,
                    lora_model_dir=self.params.get("lora_dir", ""),
                    n_threads=self.params.get("n_threads", multiprocessing.cpu_count()),
                    wtype=self.params.get(
                        "wtype", NEXA_RUN_MODEL_PRECISION_MAP.get(model_path, "default")
                    ),  # Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
                    control_net_path=self.params.get("control_net_path", ""),
                    verbose=self.profiling,
                )

    def _save_images(self, images):
        """
        Save the generated images to the specified output path.
        """
        output_dir = os.path.dirname(self.params["output_path"])
        os.makedirs(output_dir, exist_ok=True)

        for i, image in enumerate(images):
            file_name = f"image_{i+1}_{int(time.time())}.png"
            file_path = os.path.join(output_dir, file_name)
            image.save(file_path)
            print(f"\nImage {i+1} saved to: {os.path.abspath(file_path)}")

    def _retry(self, func, *args, **kwargs):
        for attempt in range(RETRY_ATTEMPTS):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                logging.error(f"Attempt {attempt + 1} failed with error: {e}")
                time.sleep(1)
        print("All retry attempts failed becase of Out of Memory error, Try to use smaller models...")
        return None

    def txt2img(
        self,
        prompt,
        negative_prompt="",
        cfg_scale=7.5,
        width=512,
        height=512,
        sample_steps=20,
        seed=0,
        control_cond="",
        control_strength=0.9,
    ):
        """
        Used for SDK. Generate images from text.

        Args:
            prompt (str): Prompt for the image generation.
            negative_prompt (str): Negative prompt for the image generation.

        Returns:
            list: List of generated images.
        """
        images = self._retry(
            self.model.txt_to_img,
            prompt=prompt,
            negative_prompt=negative_prompt,
            cfg_scale=cfg_scale,
            width=width,
            height=height,
            sample_steps=sample_steps,
            seed=seed,
            control_cond=control_cond,
            control_strength=control_strength,
        )
        return images

    def run_txt2img(self):
        while True:
            try:
                prompt = nexa_prompt("Enter your prompt: ")
                negative_prompt = nexa_prompt(
                    "Enter your negative prompt (press Enter to skip): "
                )
                try:
                    images = self.txt2img(
                        prompt,
                        negative_prompt,
                        cfg_scale=self.params["guidance_scale"],
                        width=self.params["width"],
                        height=self.params["height"],
                        sample_steps=self.params["num_inference_steps"],
                        seed=self.params["random_seed"],
                        control_cond=self.params.get("control_image_path", ""),
                        control_strength=self.params.get("control_strength", 0.9),
                    )
                    if images:
                        self._save_images(images)
                except Exception as e:
                    logging.error(f"Error during text to image generation: {e}")
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during generation: {e}", exc_info=True)

    def img2img(
        self,
        image_path,
        prompt,
        negative_prompt="",
        cfg_scale=7.5,
        width=512,
        height=512,
        sample_steps=20,
        seed=0,
        control_cond="",
        control_strength=0.9,
    ):
        """
        Used for SDK. Generate images from an image.

        Args:
            image_path (str): Path to the input image.
            prompt (str): Prompt for the image generation.
            negative_prompt (str): Negative prompt for the image generation.

        Returns:
            list: List of generated images.
        """
        images = self._retry(
            self.model.img_to_img,
            image=image_path,
            prompt=prompt,
            negative_prompt=negative_prompt,
            cfg_scale=cfg_scale,
            width=width,
            height=height,
            sample_steps=sample_steps,
            seed=seed,
            control_cond=control_cond,
            control_strength=control_strength,
        )
        return images

    def run_img2img(self):
        while True:
            try:
                image_path = nexa_prompt("Enter the path to your image: ")
                prompt = nexa_prompt("Enter your prompt: ")
                negative_prompt = nexa_prompt(
                    "Enter your negative prompt (press Enter to skip): "
                )
                images = self.img2img(
                    image_path,
                    prompt,
                    negative_prompt,
                    cfg_scale=self.params["guidance_scale"],
                    width=self.params["width"],
                    height=self.params["height"],
                    sample_steps=self.params["num_inference_steps"],
                    seed=self.params["random_seed"],
                    control_cond=self.params.get("control_image_path", ""),
                    control_strength=self.params.get("control_strength", 0.9),
                )

                if images:
                    self._save_images(images)
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during generation: {e}", exc_info=True)

    def run_streamlit(self, model_path: str, is_local_path = False, hf = False):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")

        streamlit_script_path = (
            Path(os.path.abspath(__file__)).parent
            / "streamlit"
            / "streamlit_image_chat.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path, str(is_local_path), str(hf)]
        sys.exit(stcli.main())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run image generation with a specified model"
    )
    parser.add_argument(
        "model_path",
        type=str,
        help="Path or identifier for the model in Nexa Model Hub",
    )
    parser.add_argument(
        "-i2i",
        "--img2img",
        action="store_true",
        help="Whether to run image-to-image generation",
    )
    parser.add_argument(
        "-ns",
        "--num_inference_steps",
        type=int,
        default=20,
        help="Number of inference steps",
    )
    parser.add_argument(
        "-H", "--height", type=int, default=512, help="Height of the output image"
    )
    parser.add_argument(
        "-w", "--width", type=int, default=512, help="Width of the output image"
    )
    parser.add_argument(
        "-g",
        "--guidance_scale",
        type=float,
        default=7.5,
        help="Guidance scale for diffusion",
    )
    parser.add_argument(
        "-o",
        "--output",
        type=str,
        default="generated_images/image.png",
        help="Output path for the generated image",
    )
    parser.add_argument(
        "-s",
        "--random_seed",
        type=int,
        default=0,
        help="Random seed for image generation",
    )
    parser.add_argument(
        "--lora_dir",
        type=str,
        help="Path to directory containing LoRA files.",
    )
    parser.add_argument(
        "--wtype",
        type=str,
        help="Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)",
    )
    parser.add_argument(
        "--control_net_path",
        type=str,
        help="Path to control net model.",
    )
    parser.add_argument(
        "--control_image_path",
        type=str,
        help="Path to image condition for Control Net.",
    )
    parser.add_argument(
        "--control_strength",
        type=float,
        help="Strength to apply Control Net.",
    )
    parser.add_argument(
        "-st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
    )
    parser.add_argument(
        "-pf",
        "--profiling",
        action="store_true",
        help="Enable profiling logs for the inference process",
    )
    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    inference = NexaImageInference(model_path, **kwargs)
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        if args.img2img:
            inference.run_img2img()
        else:
            inference.run_txt2img()
