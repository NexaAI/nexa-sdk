import ctypes
import logging
import os
import sys
import librosa
import tempfile
import soundfile as sf
from pathlib import Path
from streamlit.web import cli as stcli
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
        self.temp_file = None

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

            # self.context = audio_lm_cpp.init_context(
            #     ctypes.byref(self.ctx_params), is_qwen=self.is_qwen
            # )
            # if not self.context:
            #     raise RuntimeError("Failed to load audio language model")
            # logging.debug("Model loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model: {e}")
            raise

    def run(self):
        """
        Run the audio language model inference loop.
        """
        from nexa.gguf.llama._utils_spinner import start_spinner, stop_spinner

        try:
            while True:
                audio_path = self._get_valid_audio_path()
                user_input = nexa_prompt("Enter text (leave empty if no prompt): ")
            
                stop_event, spinner_thread = start_spinner(
                style="default", 
                message=""  
                )
            
                try:
                    with suppress_stdout_stderr():
                        response = self.inference(audio_path, user_input)
                finally:
                    stop_spinner(stop_event, spinner_thread)
            
                print(f"{response}")
                self.cleanup()

        except KeyboardInterrupt:
            print("\nExiting...")
        except Exception as e:
            logging.error(f"\nError during audio generation: {e}", exc_info=True)
        # finally:
        #     self.cleanup()

    def _get_valid_audio_path(self) -> str:
        """
        Helper method to get a valid audio file path from user
        """
        while True:
            audio_path = nexa_prompt("Enter the path to your audio file (required): ")
            if os.path.exists(audio_path):
                # Check if it's a supported audio format
                if any(audio_path.lower().endswith(ext) for ext in ['.wav', '.mp3', '.m4a', '.flac', '.ogg']):
                    return audio_path
                print(f"Unsupported audio format. Please use WAV, MP3, M4A, FLAC, or OGG files.")
            else:
                print(f"'{audio_path}' is not a valid audio path. Please try again.")

    def inference(self, audio_path: str, prompt: str = "") -> str:
        """
        Perform a single inference with the audio language model.
        """
        if not os.path.exists(audio_path):
            raise FileNotFoundError(f"Audio file not found: {audio_path}")

        try:
            # Ensure audio is at 16kHz before processing
            audio_path = self._ensure_16khz(audio_path)

            self.ctx_params.file = ctypes.c_char_p(audio_path.encode("utf-8"))
            self.ctx_params.prompt = ctypes.c_char_p(prompt.encode("utf-8"))

            self.context = audio_lm_cpp.init_context(
                ctypes.byref(self.ctx_params), is_qwen=self.is_qwen
            )
            if not self.context:
                raise RuntimeError("Failed to load audio language model")
            logging.debug("Model loaded successfully")

            response = audio_lm_cpp.process_full(
                self.context, ctypes.byref(self.ctx_params), is_qwen=self.is_qwen
            )
            return response.decode("utf-8") if isinstance(response, bytes) else response
        except Exception as e:
            raise RuntimeError(f"Error during inference: {str(e)}")
        finally:
            if self.temp_file:
                try:
                    self.temp_file.close()
                    if os.path.exists(self.temp_file.name):
                        os.unlink(self.temp_file.name)
                except:
                    pass
                self.temp_file = None

    def cleanup(self):
        """
        Explicitly cleanup resources
        """
        if self.context:
            audio_lm_cpp.free(self.context, is_qwen=self.is_qwen)
            self.context = None
        
        if self.temp_file:
            try:
                self.temp_file.close()
                if os.path.exists(self.temp_file.name):
                    os.unlink(self.temp_file.name)
            except:
                pass
            self.temp_file = None

    # def __del__(self):
    #     """
    #     Destructor to free the Bark context when the instance is deleted.
    #     """
    #     if self.context:
    #         audio_lm_cpp.free(self.context, is_qwen=self.is_qwen)

    def _ensure_16khz(self, audio_path: str) -> str:
        """
        Check if audio is 16kHz, resample if necessary.
        Supports various audio formats (mp3, wav, m4a, etc.)
        """
        try:
            y, sr = librosa.load(audio_path, sr=None)
    
            if sr == 16000:
                return audio_path
            
            # Resample to 16kHz
            print(f"Resampling audio from {sr} to 16000")
            y_resampled = librosa.resample(y=y, orig_sr=sr, target_sr=16000)
            self.temp_file = tempfile.NamedTemporaryFile(
                suffix='.wav',
                delete=False
            )
            sf.write(
                self.temp_file.name, 
                y_resampled, 
                16000,
                subtype='PCM_16'
            )
            return self.temp_file.name

        except Exception as e:
            raise RuntimeError(f"Error processing audio file: {str(e)}")

    def run_streamlit(self, model_path: str, is_local_path = False, hf = False, projector_local_path = None):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")

        streamlit_script_path = (
            Path(os.path.abspath(__file__)).parent / "streamlit" / "streamlit_audio_lm.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path, str(is_local_path), str(hf), str(projector_local_path)]
        sys.exit(stcli.main())

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
    parser.add_argument(
        "-st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
    )

    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    device = kwargs.pop("device", "auto")

    inference = NexaAudioLMInference(model_path, device=device, **kwargs)
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()
