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
    NEXA_RUN_MODEL_MAP,
    NEXA_RUN_MODEL_PRECISION_MAP,
    DEFAULT_IMG_GEN_PARAMS_LCM,
    DEFAULT_IMG_GEN_PARAMS_TURBO,
)
from nexa.utils import SpinningCursorAnimation, nexa_prompt
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr

from streamlit.web import cli as stcli
from nexa.general import pull_model

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)
RETRY_ATTEMPTS = (
    3  # a temporary fix for the issue of segmentation fault for stable-diffusion-cpp
)


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
    streamlit (bool): Run the inference in Streamlit UI.

    """

    def __init__(self, model_path, local_path=None, **kwargs):
        self.model_path = model_path
        self.downloaded_path = local_path

        if self.downloaded_path is None:
            self.downloaded_path, run_type = pull_model(self.model_path)

        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        if "lcm-dreamshaper" in self.model_path:
            self.params = DEFAULT_IMG_GEN_PARAMS_LCM
        elif "sdxl-turbo" in self.model_path:
            self.params = DEFAULT_IMG_GEN_PARAMS_TURBO
        else:
            self.params = DEFAULT_IMG_GEN_PARAMS

        self.params.update(kwargs)
        if not kwargs.get("streamlit", False):
            self._load_model(model_path)
            if self.model is None:
                logging.error("Failed to load the model or pipeline.")
                exit(1)

    @SpinningCursorAnimation()
    def _load_model(self, model_path: str):
        with suppress_stdout_stderr():
            from nexa.gguf.sd.stable_diffusion import StableDiffusion

            self.model = StableDiffusion(
                model_path=self.downloaded_path,
                lora_model_dir=self.params.get("lora_dir", ""),
                n_threads=self.params.get("n_threads", multiprocessing.cpu_count()),
                wtype=self.params.get(
                    "wtype", NEXA_RUN_MODEL_PRECISION_MAP.get(model_path, "f32")
                ),  # Weight type (options: default, f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
                control_net_path=self.params.get("control_net_path", ""),
                verbose=False,
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
        logging.error("All retry attempts failed.")
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

    def run_streamlit(self, model_path: str):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")

        streamlit_script_path = (
            Path(os.path.abspath(__file__)).parent
            / "streamlit"
            / "streamlit_image_chat.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path]
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
    # parser.add_argument("--device", type=str, default='cuda' if torch.cuda.is_available() else 'cpu', help="Device to run the model on (default: cuda if available, else cpu)")
    parser.add_argument(
        "-st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
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
