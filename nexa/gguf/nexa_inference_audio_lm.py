import ctypes
import logging
import os
from pathlib import Path
from nexa.utils import SpinningCursorAnimation, nexa_prompt
from nexa.constants import (
    DEFAULT_TEXT_GEN_PARAMS,
    NEXA_RUN_MODEL_MAP_AUDIO_LM,
    NEXA_RUN_AUDIO_LM_PROJECTOR_MAP,
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama import audio_lm_cpp
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model


def is_qwen(model_name):
    if "qwen" in model_name.lower():  # TEMPORARY SOLUTION : this hardcode can be risky
        return True
    return False


assert set(NEXA_RUN_MODEL_MAP_AUDIO_LM.keys()) == set(
    NEXA_RUN_AUDIO_LM_PROJECTOR_MAP.keys()
), "Model, projector, and handler should have the same keys"


class NexaAudioLMInference:
    """
    A class used for loading Bark text-to-speech models and running text-to-speech generation.

    Methods:
        run: Run the audio LM generation loop.

    Args:
        model_path (str): Path to the model file.
        mmproj_path (str): Path to the audio projector file.
        n_gpu_layers(int): Number of gpu layers to use for processing. Defaults to -1.
        output_dir (str): Output directory for tts. Defaults to "tts".
        verbosity (int): Verbosity level for the Bark model. Defaults to 0.
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
        self.projector_path = NEXA_RUN_AUDIO_LM_PROJECTOR_MAP.get(model_path, None)
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
            if model_path in NEXA_RUN_MODEL_MAP_AUDIO_LM:
                self.projector_path = NEXA_RUN_AUDIO_LM_PROJECTOR_MAP[model_path]
                self.projector_downloaded_path, _ = pull_model(
                    self.projector_path, **kwargs
                )
        elif model_path in NEXA_RUN_MODEL_MAP_AUDIO_LM:
            self.model_path = NEXA_RUN_MODEL_MAP_AUDIO_LM[model_path]
            self.projector_path = NEXA_RUN_AUDIO_LM_PROJECTOR_MAP[model_path]
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

        if self.downloaded_path is None or self.projector_downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)
        self.is_qwen = is_qwen(self.downloaded_path) # TEMPORARY SOLUTION : this hardcode can be risky
        self.ctx_params = audio_lm_cpp.context_default_params(self.is_qwen)
        with suppress_stdout_stderr():
            self._load_model()

    @SpinningCursorAnimation()
    def _load_model(self):
        try:
            self.ctx_params.model = ctypes.c_char_p(
                self.downloaded_path.encode("utf-8")
            )
            self.ctx_params.mmproj = ctypes.c_char_p(
                self.projector_downloaded_path.encode("utf-8")
            )
            self.ctx_params.n_gpu_layers = (
                0x7FFFFFFF if self.n_gpu_layers == -1 else self.n_gpu_layers
            )  # 0x7FFFFFFF is INT32 max, will be auto set to all layers

            self.context = audio_lm_cpp.init_context(
                ctypes.byref(self.ctx_params), is_qwen=self.is_qwen
            )
            if not self.context:
                raise RuntimeError("Failed to load audio language model")
            logging.debug("Model loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model: {e}")
            raise

    def run(self):
        while True:
            try:
                while True:
                    audio_path = nexa_prompt("Enter the path to your audio file (required): ")
                    if os.path.exists(audio_path):
                        break
                    print(f"'{audio_path}' is not a valid audio path. Please try again.")

                user_input = nexa_prompt("Enter text (leave empty if no prompt): ")

                self.ctx_params.file = ctypes.c_char_p(audio_path.encode("utf-8"))
                self.ctx_params.prompt = ctypes.c_char_p(user_input.encode("utf-8"))

                audio_lm_cpp.process_full(
                    self.context, ctypes.byref(self.ctx_params), is_qwen=self.is_qwen
                )

            except KeyboardInterrupt:
                print("\nExiting...")
                break

            except Exception as e:
                logging.error(f"\nError during audio generation: {e}", exc_info=True)

    def __del__(self):
        """
        Destructor to free the Bark context when the instance is deleted.
        """
        if self.context:
            audio_lm_cpp.free(self.context, is_qwen=self.is_qwen)


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

    inference = NexaAudioLMInference(model_path, device=device, **kwargs)
    inference.run()
