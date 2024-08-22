import argparse
import logging
import os
import sys
import time
from pathlib import Path

import librosa
from optimum.onnxruntime.modeling_seq2seq import ORTModelForSpeechSeq2Seq
from transformers import AutoProcessor

from nexa.constants import EXIT_REMINDER, NEXA_RUN_MODEL_MAP_ONNX
from nexa.utils import nexa_prompt
from nexa.general import pull_model

logging.basicConfig(level=logging.INFO)


class NexaVoiceInference:
    """
    A class used for loading voice models and running voice transcription.

    Methods:
    run: Run the voice transcription loop.
    run_streamlit: Run the Streamlit UI.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str): Local path of the model.
    output_dir (str): Output directory for transcriptions.
    sampling_rate (int): Sampling rate for audio processing.
    streamlit (bool): Run the inference in Streamlit UI.
    """

    def __init__(self, model_path, local_path=None, **kwargs):
        self.model_path = NEXA_RUN_MODEL_MAP_ONNX.get(model_path, model_path)
        self.downloaded_onnx_folder = local_path
        self.params = {"output_dir": "transcriptions", "sampling_rate": 16000}
        self.params.update(kwargs)
        self.model = None
        self.processor = None

    def run(self):
        if self.downloaded_onnx_folder is None:
            self.downloaded_onnx_folder, run_type = pull_model(self.model_path)

        if self.downloaded_onnx_folder is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        self._load_model(self.downloaded_onnx_folder)
        self._dialogue_mode()

    def _load_model(self, model_path):
        logging.debug(f"Loading model from {model_path}")
        try:
            self.processor = AutoProcessor.from_pretrained(model_path)
            self.model = ORTModelForSpeechSeq2Seq.from_pretrained(model_path)
            logging.debug("Model and processor loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model or processor: {e}")

    def _dialogue_mode(self):
        while True:
            try:
                audio_path = nexa_prompt("Enter the path to your audio file: ")
                self._transcribe_audio(audio_path)
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during text generation: {e}", exc_info=True)

    def _transcribe_audio(self, audio_path):
        if self.model is None or self.processor is None:
            logging.error("Model or processor not loaded. Exiting.")
            return

        if not os.path.exists(audio_path):
            logging.error(f"Audio file not found: {audio_path}")
            return

        try:
            audio, sr = librosa.load(audio_path, sr=self.params["sampling_rate"])

            inputs = self.processor(
                audio, return_tensors="pt", sampling_rate=self.params["sampling_rate"]
            )

            input_features = inputs.input_features
            attention_mask = (
                inputs.attention_mask if hasattr(inputs, "attention_mask") else None
            )

            logging.info("Generating transcription...")
            gen_tokens = self.model.generate(
                input_features=input_features,
                attention_mask=attention_mask,
            )

            transcription = self.processor.batch_decode(
                gen_tokens, skip_special_tokens=True
            )[0]

            self._save_transcription(transcription)
            print(f"Transcription: {transcription}")

        except Exception as e:
            logging.error(f"Error during audio transcription: {e}")

    def _save_transcription(self, transcription):
        os.makedirs(self.params["output_dir"], exist_ok=True)

        # Generate a filename with timestamp
        filename = f"transcription_{int(time.time())}.txt"
        output_path = os.path.join(self.params["output_dir"], filename)
        with open(output_path, "w") as f:
            f.write(transcription)

        logging.info(f"Transcription saved to: {output_path}")
        return output_path

    def run_streamlit(self, model_path: str):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")
        from streamlit.web import cli as stcli

        streamlit_script_path = (
            Path(__file__).resolve().parent / "streamlit" / "streamlit_voice_chat.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path]
        sys.exit(stcli.main())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run voice transcription with a specified model"
    )
    parser.add_argument(
        "model_path", type=str, help="Path or identifier for the model in S3"
    )
    parser.add_argument(
        "-o",
        "--output_dir",
        type=str,
        default="transcriptions",
        help="Output directory for transcriptions",
    )
    parser.add_argument(
        "-r",
        "--sampling_rate",
        type=int,
        default=16000,
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
    inference = NexaVoiceInference(model_path, **kwargs)
    inference.run()
