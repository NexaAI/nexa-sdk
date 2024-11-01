import ctypes
import logging
import os

from nexa.constants import (
    DEFAULT_TEXT_GEN_PARAMS,
    NEXA_RUN_MODEL_MAP_VLM,
    NEXA_RUN_PROJECTOR_MAP,
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama import nexa_audio_lm_cpp
from nexa.general import pull_model

def is_qwen(model_name):
    if "qwen2" in model_name: # TEMPORARY SOLUTION : this hardcode can be risky
        return True
    return False

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

    def __init__(self, model_path: str, mmproj_path: str, verbosity=0, device="auto", **kwargs):
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")
        self.params = DEFAULT_TEXT_GEN_PARAMS.copy()
        self.params.update(kwargs)
        self.model = None
        self.device = device
        
        self.model_path = model_path
        self.mmproj_path = mmproj_path
        if self.device == "auto" or self.device == "gpu":
            self.n_gpu_layers = -1 if is_gpu_available() else 0
        else:
            self.n_gpu_layers = 0
        self.is_qwen = is_qwen(model_path)
        self.ctx_params = nexa_audio_lm_cpp.context_default_params(self.is_qwen)
        self.context = None
        self.verbosity = verbosity
        self.params = {
            "output_path": os.path.join(os.getcwd(), "audio-lm"),
        }
        self.params.update(kwargs)
        self.downloaded_path, _ = pull_model(self.model_path, **kwargs)
        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)
        self._load_model()

    def _load_model(self):
        logging.debug(f"Loading model from {self.model_path} and {self.mmproj_path}")
        try:
            self.ctx_params.model = ctypes.c_char_p(self.model_path.encode("utf-8"))
            self.ctx_params.mmproj = ctypes.c_char_p(self.mmproj_path.encode("utf-8"))
            self.ctx_params.n_gpu_layers = (
                0x7FFFFFFF if self.n_gpu_layers == -1 else self.n_gpu_layers
            )  # 0x7FFFFFFF is INT32 max, will be auto set to all layers

            self.context = nexa_audio_lm_cpp.init_context(
                ctypes.byref(self.ctx_params),
                self.is_qwen
            )
            if not self.context:
                raise RuntimeError("Failed to load Bark model")
            logging.debug("Model loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model: {e}")
            raise

    def run(self):
        while True:
            try:
                audio_path = input("Audio Path (leave empty if no audio): ")
                if audio_path and not os.path.exists(audio_path):
                    print(f"'{audio_path}' is not a path to audio. Will ignore.")

                user_input = input("Enter text: ")

                self.ctx_params.file = ctypes.c_char_p(audio_path.encode("utf-8"))
                self.ctx_params.prompt = ctypes.c_char_p(user_input.encode("utf-8"))

                nexa_audio_lm_cpp.process_full(
                    self.context, ctypes.byref(self.ctx_params),
                    self.is_qwen
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
            nexa_audio_lm_cpp.free_context(self.context)


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

    inference = NexaAudioLMInference(
        model_path,
        device=device, 
        **kwargs
    )
    inference.run()