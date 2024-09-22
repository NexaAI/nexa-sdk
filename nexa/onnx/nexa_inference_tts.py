import argparse
import logging
import os
import sys
import time
from pathlib import Path

import onnxruntime
import soundfile as sf
import yaml
from ttstokenizer import TTSTokenizer

from nexa.constants import EXIT_REMINDER, NEXA_RUN_MODEL_MAP_ONNX
from nexa.utils import nexa_prompt
from nexa.general import pull_model
logging.basicConfig(level=logging.INFO)


# TODO: https://huggingface.co/NeuML/ljspeech-jets-onnx is the only onnx on huggingface, need to add it to ort pipeline
class NexaTTSInference:
    """
    A class used for loading text-to-speech models and running text-to-speech generation.

    Methods:
        run: Run the text-to-speech generation loop.
        run_streamlit: Run the Streamlit UI.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str): Local path of the model.
    output_dir (str): Output directory for tts.
    sampling_rate (int): Sampling rate for audio processing.
    streamlit (bool): Run the inference in Streamlit UI.
    """
    
    def __init__(self, model_path, local_path=None, **kwargs):
        self.model_path = NEXA_RUN_MODEL_MAP_ONNX.get(model_path, model_path)
        self.yaml_file_name = None
        self.params = {
            "output_path": os.path.join(os.getcwd(), "tts"),
            "sampling_rate": 23000,
        }
        self.params.update(kwargs)
        self.model = None
        self.processor = None
        self.config = None
        self.downloaded_onnx_folder = local_path

        if self.downloaded_onnx_folder is None:
            self.downloaded_onnx_folder, run_type = pull_model(self.model_path)
        
        if self.downloaded_onnx_folder is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        self.yaml_file_name = os.path.join(self.downloaded_onnx_folder, "config.yaml")
        with open(self.yaml_file_name, "r", encoding="utf-8") as f:
            self.config = yaml.safe_load(f)

        self._load_model()

    def _load_model(self):
        logging.debug(f"Loading model from {self.downloaded_onnx_folder}")
        try:
            self.tokenizer = TTSTokenizer(self.config["token"]["list"])
            print(self.tokenizer)
            self.model = onnxruntime.InferenceSession(
                os.path.join(self.downloaded_onnx_folder, "model.onnx"),
                providers=["CPUExecutionProvider"],
            )
            print(self.model)
            logging.debug("Model and tokenizer loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model or tokenizer: {e}")

    def run(self):
        while True:
            try:
                user_input = nexa_prompt("Enter text to generate audio: ")
                outputs = self.audio_generation(user_input)
                self._save_audio(
                    outputs[0], self.params["sampling_rate"], self.params["output_path"]
                )
                logging.info(f"Audio saved to {self.params['output_path']}")                
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during text generation: {e}", exc_info=True)

    def audio_generation(self, user_input):
        """
        Used for SDK. Generate audio from the user input.

        Args:
            user_input (str): User input for audio generation.

        Returns:
            np.array: Audio data.
        """
        inputs = self.tokenizer(user_input)
        outputs = self.model.run(None, {"text": inputs})
        return outputs


    def _save_audio(self, audio_data, sampling_rate, output_path):
        os.makedirs(output_path, exist_ok=True)
        file_name = f"audio_{int(time.time())}.wav"
        file_path = os.path.join(output_path, file_name)
        sf.write(file_path, audio_data, sampling_rate)

    def run_streamlit(self, model_path: str):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")
        from streamlit.web import cli as stcli

        streamlit_script_path = (
            Path(__file__).resolve().parent / "streamlit" / "streamlit_tts.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path]
        sys.exit(stcli.main())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run text to speech generation with a specified model"
    )
    parser.add_argument(
        "model_path", type=str, help="Path or identifier for the model in S3"
    )
    parser.add_argument(
        "-o", "--output_dir", type=str, default="tts", help="Output directory for tts"
    )
    parser.add_argument(
        "-r",
        "--sampling_rate",
        type=int,
        default=23000,
        help="Sampling rate for audio processing",
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
    inference = NexaTTSInference(model_path, **kwargs)
    inference.run()
