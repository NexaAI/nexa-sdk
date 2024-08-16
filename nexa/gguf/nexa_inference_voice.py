import argparse
import logging
import os
import sys
import time
from pathlib import Path

from nexa.constants import EXIT_REMINDER, NEXA_RUN_MODEL_MAP_VOICE, DEFAULT_VOICE_GEN_PARAMS
from nexa.general import pull_model
from nexa.utils import nexa_prompt
from faster_whisper import WhisperModel

logging.basicConfig(level=logging.INFO)


class NexaVoiceInference:
    """
    A class used for loading voice models and running voice transcription.

    Methods:
    run: Run the voice transcription loop.
    run_streamlit: Run the Streamlit UI.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    output_dir (str): Output directory for transcriptions.
    beam_size (int): Beam size to use for transcription.
    language (str): The language spoken in the audio.
    task (str): Task to execute (transcribe or translate).
    temperature (float): Temperature for sampling.
    compute_type (str): Type to use for computation (e.g., float16, int8, int8_float16).
    output_dir (str): Output directory for transcriptions.

    """
    def __init__(self, model_path, **kwargs):
        self.model_path = None
        self.downloaded_path = None
        self.params = DEFAULT_VOICE_GEN_PARAMS        
        if model_path in NEXA_RUN_MODEL_MAP_VOICE:
            logging.debug(f"Found model {model_path} in public hub")
            self.model_path = NEXA_RUN_MODEL_MAP_VOICE.get(model_path)
            self.downloaded_path = pull_model(self.model_path)
        elif os.path.exists(model_path):
            logging.debug(f"Using local model at {model_path}")
            self.downloaded_path = model_path
        else:
            logging.error("Using voice model from hub is not supported yet.")
            exit(1)

        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)
        
        self.params.update(kwargs)
        self.model = None

    def run(self):
        self._load_model(self.downloaded_path)
        self._dialogue_mode()

    def _load_model(self, model_path):
        logging.debug(f"Loading model from: {model_path}")
        try:
            self.model = WhisperModel(model_path, device="auto", compute_type=self.params["compute_type"])
            logging.debug("Model loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model: {e}")

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
        logging.debug(f"Transcribing audio from: {audio_path}")
        try:
            segments, _ = self.model.transcribe(
                audio_path, 
                beam_size=self.params["beam_size"], 
                language=self.params["language"], 
                task=self.params["task"], 
                temperature=self.params["temperature"], 
                vad_filter=True,
            )
            transcription = "".join(segment.text for segment in segments)
            self._save_transcription(transcription)
            logging.info(f"Transcription: {transcription}")
        except Exception as e:
            logging.error(f"Error during transcription: {e}", exc_info=True)


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
        "model_path", type=str, help="Path or identifier for the model"
    )
    parser.add_argument(
        "-o",
        "--output_dir",
        type=str,
        default="transcriptions",
        help="Output directory for transcriptions",
    )
    parser.add_argument(
        "-b",
        "--beam_size",
        type=int,
        default=5,
        help="Beam size to use for transcription",
    )
    parser.add_argument(
        "-l",
        "--language",
        type=str,
        default=None,
        help="The language spoken in the audio. It should be a language code such as 'en' or 'fr'.",
    )
    parser.add_argument(
        "--task",
        type=str,
        default="transcribe",
        help="Task to execute (transcribe or translate)",
    )
    parser.add_argument(
        "-t",
        "--temperature",
        type=float,
        default=0.0,
        help="Temperature for sampling",
    )
    parser.add_argument(
        "-c",
        "--compute_type",
        type=str,
        default="default",
        help="Type to use for computation (e.g., float16, int8, int8_float16)",
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
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()
