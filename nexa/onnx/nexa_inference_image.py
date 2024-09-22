import argparse
import json
import logging
import os
import sys
import time
from pathlib import Path

import numpy as np
from optimum.onnxruntime import (
    ORTLatentConsistencyModelPipeline,
    ORTStableDiffusionPipeline,
    ORTStableDiffusionXLPipeline,
)
from nexa.general import pull_model
from nexa.constants import EXIT_REMINDER, NEXA_RUN_MODEL_MAP_ONNX
from nexa.utils import nexa_prompt, SpinningCursorAnimation

logging.basicConfig(level=logging.INFO)

ORT_PIPELINES_MAPPING = {
    "ORTStableDiffusionPipeline": ORTStableDiffusionPipeline,
    "ORTLatentConsistencyModelPipeline": ORTLatentConsistencyModelPipeline,
    "ORTStableDiffusionXLPipeline": ORTStableDiffusionXLPipeline,
}


class NexaImageInference:
    """
    A class used for loading image models and running image generation.

    Methods:
        run: Run the image generation loop.
        run_streamlit: Run the Streamlit UI.
        generate_images: Generate images based on the given prompt, negative prompt, and parameters.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str): Local path of the model.
    num_inference_steps (int): Number of inference steps.
    num_images_per_prompt (int): Number of images to generate per prompt.
    width (int): Width of the output image.
    height (int): Height of the output image.
    guidance_scale (float): Guidance scale for diffusion.
    output_path (str): Output path for the generated image. exapmle: generated_images/image.png
    random_seed (int): Random seed for image generation.
    streamlit (bool): Run the inference in Streamlit UI.
    """
    def __init__(self, model_path, local_path=None, **kwargs):
        self.model_path = NEXA_RUN_MODEL_MAP_ONNX.get(model_path, model_path)
        self.download_onnx_folder = local_path
        self.params = {
            "num_inference_steps": 20,
            "num_images_per_prompt": 1,
            "height": 512,
            "width": 512,
            "guidance_scale": 7.5,
            "output_path": "generated_images/image.png",
            "random_seed": 0,
        }
        self.params.update(kwargs)
        self.pipeline = None

    def run(self):

        if self.download_onnx_folder is None:
            self.download_onnx_folder, run_type = pull_model(self.model_path)

        if self.download_onnx_folder is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        self._load_model(self.download_onnx_folder)
        self._dialogue_mode()

    @SpinningCursorAnimation()
    def _load_model(self, model_path):
        """
        Load the model from the given model path using the appropriate pipeline.
        """
        logging.debug(f"Loading model from {model_path}")
        try:
            model_index_path = os.path.join(model_path, "model_index.json")
            with open(model_index_path, "r") as index_file:
                model_index = json.load(index_file)

            pipeline_class_name = model_index.get(
                "_class_name", "ORTStableDiffusionPipeline"
            )
            PipelineClass = ORT_PIPELINES_MAPPING.get(
                pipeline_class_name, ORTStableDiffusionPipeline
            )
            self.pipeline = PipelineClass.from_pretrained(model_path)
            logging.debug(f"Model loaded successfully using {pipeline_class_name}")
        except Exception as e:
            logging.error(f"Error loading model: {e}")

    def _dialogue_mode(self):
        """
        Enter a dialogue mode where the user can input prompts and negative prompts repeatedly.
        """
        while True:
            try:
                prompt = nexa_prompt("Enter your prompt: ")
                negative_prompt = nexa_prompt(
                    "Enter your negative prompt (press Enter to skip): "
                )
                images = self.generate_images(prompt, negative_prompt)
                self._save_images(images)
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during text generation: {e}", exc_info=True)

    def generate_images(self, prompt, negative_prompt):
        """
        Used for SDK. Generate images based on the given prompt, negative prompt, and parameters.

        Arg:
            prompt (str): Prompt for the image generation.
            negative_prompt (str): Negative prompt for the image generation.

        Returns:
            list: List of generated images.
        """
        if self.pipeline is None:
            logging.error("Model not loaded. Exiting.")
            return

        generator = np.random.RandomState(self.params["random_seed"])

        is_lcm_pipeline = isinstance(
            self.pipeline, ORTLatentConsistencyModelPipeline
        )

        pipeline_kwargs = {
            "prompt": prompt,
            "num_inference_steps": self.params["num_inference_steps"],
            "num_images_per_prompt": self.params["num_images_per_prompt"],
            "height": self.params["height"],
            "width": self.params["width"],
            "generator": generator,
            "guidance_scale": self.params["guidance_scale"],
        }
        if not is_lcm_pipeline and negative_prompt:
            pipeline_kwargs["negative_prompt"] = negative_prompt

        images = self.pipeline(**pipeline_kwargs).images
        return images



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
            print(f"Image {i+1} saved to: {file_path}")

    def run_streamlit(self, model_path: str):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")
        from streamlit.web import cli as stcli

        streamlit_script_path = (
            Path(__file__).resolve().parent / "streamlit" / "streamlit_image_chat.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path]
        sys.exit(stcli.main())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run image generation with a specified model"
    )
    parser.add_argument(
        "model_path", type=str, help="Path or identifier for the model in S3"
    )
    parser.add_argument(
        "-ns",
        "--num_inference_steps",
        type=int,
        default=20,
        help="Number of inference steps",
    )
    parser.add_argument(
        "-np",
        "--num_images_per_prompt",
        type=int,
        default=1,
        help="Number of images to generate per prompt",
    )
    parser.add_argument(
        "-H", "--height", type=int, default=512, help="Height of the output image"
    )
    parser.add_argument(
        "-W", "--width", type=int, default=512, help="Width of the output image"
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
        "-st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
    )
    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    inference = NexaImageInference(model_path, **kwargs)
    inference.run()
