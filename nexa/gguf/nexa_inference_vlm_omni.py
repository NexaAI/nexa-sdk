
import ctypes
import logging
import os
from pathlib import Path
from nexa.utils import nexa_prompt, SpinningCursorAnimation
from nexa.constants import (
    DEFAULT_TEXT_GEN_PARAMS,
    NEXA_RUN_OMNI_VLM_PROJECTOR_MAP,
    NEXA_RUN_OMNI_VLM_MAP
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama import omni_vlm_cpp
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model

class NexaOmniVlmInference:
    """
    A class used for vision language model inference.
    """

    def __init__(
        self,
        model_path=None,
        local_path=None,
        projector_local_path=None,
        device="auto",
        **kwargs,
    ):
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")

        self.params = DEFAULT_TEXT_GEN_PARAMS.copy()
        self.params.update(kwargs)
        self.model = None
        self.projector = None
        self.projector_path = NEXA_RUN_OMNI_VLM_PROJECTOR_MAP.get(model_path, None)
        self.downloaded_path = local_path
        self.projector_downloaded_path = projector_local_path
        self.device = device
        self.context = None
        if self.device == "auto" or self.device == "gpu":
            self.n_gpu_layers = -1 if is_gpu_available() else 0
        else:
            self.n_gpu_layers = 0

        if (
            self.downloaded_path is not None
            and self.projector_downloaded_path is not None
        ):
            # when running from local, both path should be provided
            pass
        elif self.downloaded_path is not None:
            if model_path in NEXA_RUN_OMNI_VLM_MAP:
                self.projector_path = NEXA_RUN_OMNI_VLM_PROJECTOR_MAP[model_path]
                self.projector_downloaded_path, _ = pull_model(
                    self.projector_path, **kwargs
                )
        elif model_path in NEXA_RUN_OMNI_VLM_MAP:
            self.model_path = NEXA_RUN_OMNI_VLM_MAP[model_path]
            self.projector_path = NEXA_RUN_OMNI_VLM_PROJECTOR_MAP[model_path]
            self.downloaded_path, _ = pull_model(self.model_path, **kwargs)
            self.projector_downloaded_path, _ = pull_model(
                self.projector_path, **kwargs
            )
        elif Path(model_path).parent.exists():
            local_dir = Path(model_path).parent
            model_name = Path(model_path).name
            tag_and_ext = model_name.split(":")[-1]
            self.downloaded_path = local_dir / f"model-{tag_and_ext}"
            self.projector_downloaded_path = local_dir / f"projector-{tag_and_ext}"
            if not (
                self.downloaded_path.exists()
                and self.projector_downloaded_path.exists()
            ):
                logging.error(
                    f"Model or projector not found in {local_dir}. "
                    "Make sure to name them as 'model-<tag>.gguf' and 'projector-<tag>.gguf'."
                )
                exit(1)
        else:
            logging.error("VLM user model from hub is not supported yet.")
            exit(1)
        with suppress_stdout_stderr():
            self._load_model()

    @SpinningCursorAnimation()
    def _load_model(self):
        try:
            self.ctx_params_model = ctypes.c_char_p(
                self.downloaded_path.encode("utf-8")
            )
            self.ctx_params_mmproj = ctypes.c_char_p(
                self.projector_downloaded_path.encode("utf-8")
            )
            omni_vlm_cpp.omnivlm_init(self.ctx_params_model, self.ctx_params_mmproj)
        except Exception as e:
            logging.error(f"Error loading model: {e}")
            raise
        
    def run(self):
        while True:
            try:
                image_path = nexa_prompt("Image Path (required): ")
                if not os.path.exists(image_path):
                    print(f"Image path: {image_path} not found, running omni VLM without image input.")

                user_input = nexa_prompt()
                image_path = ctypes.c_char_p(image_path.encode("utf-8"))
                user_input = ctypes.c_char_p(user_input.encode("utf-8"))
                omni_vlm_cpp.omnivlm_inference(user_input, image_path)

            except KeyboardInterrupt:
                print("\nExiting...")
                break
            except Exception as e:
                logging.error(f"\nError during audio generation: {e}", exc_info=True)
            print("\n")

    def __del__(self):
        omni_vlm_cpp.omnivlm_free()


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="Run audio-in text-out generation with nexa-omni-audio model"
    )
    parser.add_argument(
        "model_path",
        type=str,
        help="Path or identifier for the model in Nexa Model Hub",
    )
    parser.add_argument(
        "-d",
        "--device",
        type=str,
        choices=["auto", "cpu", "gpu"],
        default="auto",
        help="Device to use for inference (auto, cpu, or gpu)",
    )
    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    device = kwargs.pop("device", "auto")

    inference = NexaOmniVlmInference(model_path, device=device, **kwargs)
    inference.run()