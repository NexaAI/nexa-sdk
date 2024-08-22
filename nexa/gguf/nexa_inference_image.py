import argparse
import logging
import multiprocessing
import os
import sys
import time
from pathlib import Path

from nexa.gguf.sd.stable_diffusion import StableDiffusion
from nexa.constants import (
    DEFAULT_IMG_GEN_PARAMS,
    EXIT_REMINDER,
    NEXA_RUN_MODEL_MAP,
    NEXA_RUN_MODEL_PRECISION_MAP,
    DEFAULT_IMG_GEN_PARAMS_LCM,
    DEFAULT_IMG_GEN_PARAMS_TURBO,
)
from nexa.utils import SpinningCursorAnimation, nexa_prompt, suppress_stdout_stderr
from streamlit.web import cli as stcli

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)


class NexaImageInference:
    """
    A class used for loading image models and running image generation.

    Methods:
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

    def __init__(self, model_path, local_path, **kwargs):
        self.model_path = model_path
        self.local_path = local_path

        if self.model_path == "lcm-dreamshaper":
            self.params = DEFAULT_IMG_GEN_PARAMS_LCM
        elif self.model_path == "sdxl-turbo":
            self.params = DEFAULT_IMG_GEN_PARAMS_TURBO
        else:
            self.params = DEFAULT_IMG_GEN_PARAMS

        self.params.update(kwargs)
        if not kwargs.get("streamlit", False):
            self._load_model()
            if self.model is None:
                logging.error("Failed to load the model or pipeline.")
                exit(1)

    @SpinningCursorAnimation()
    def _load_model(self):
        with suppress_stdout_stderr():
            self.model = StableDiffusion(
                model_path=self.local_path,
                lora_model_dir=self.params.get("lora_dir", ""),
                n_threads=self.params.get("n_threads", multiprocessing.cpu_count()),
                wtype=self.params.get(
                    "wtype", NEXA_RUN_MODEL_PRECISION_MAP.get(self.model_path, "default")
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
            logging.info(f"\nImage {i+1} saved to: {file_path}")

    def loop_txt2img(self):

        while True:
            try:
                prompt = nexa_prompt("Enter your prompt: ")
                negative_prompt = nexa_prompt(
                    "Enter your negative prompt (press Enter to skip): "
                )
                self._txt2img(prompt, negative_prompt)
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during generation: {e}", exc_info=True)

    def _txt2img(self, prompt: str, negative_prompt: str):
        """
        Generate images based on the given prompt, negative prompt, and parameters.
        """
        try:
            images = self.model.txt_to_img(
                prompt=prompt,
                negative_prompt=negative_prompt if negative_prompt else "",
                cfg_scale=self.params["guidance_scale"],
                width=self.params["width"],
                height=self.params["height"],
                sample_steps=self.params["num_inference_steps"],
                seed=self.params["random_seed"],
                control_cond=self.params.get("control_image_path", ""),
                control_strength=self.params.get("control_strength", 0.9),
            )
            self._save_images(images)
        except Exception as e:
            logging.error(f"Error during image generation: {e}")

    def loop_img2img(self):
        def _generate_images(image_path, prompt, negative_prompt):
            """
            Generate images based on the given prompt, negative prompt, and parameters.
            """
            try:
                images = self.model.img_to_img(
                    image=image_path,
                    prompt=prompt,
                    negative_prompt=negative_prompt if negative_prompt else "",
                    cfg_scale=self.params["guidance_scale"],
                    width=self.params["width"],
                    height=self.params["height"],
                    sample_steps=self.params["num_inference_steps"],
                    seed=self.params["random_seed"],
                    control_cond=self.params.get("control_image_path", ""),
                    control_strength=self.params.get("control_strength", 0.9),
                )
                self._save_images(images)
            except Exception as e:
                logging.error(f"Error during image generation: {e}")

        while True:
            try:
                image_path = nexa_prompt("Enter the path to your image: ")
                prompt = nexa_prompt("Enter your prompt: ")
                negative_prompt = nexa_prompt(
                    "Enter your negative prompt (press Enter to skip): "
                )
                _generate_images(image_path, prompt, negative_prompt)
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
            inference.loop_img2img()
        else:
            inference.loop_txt2img()
