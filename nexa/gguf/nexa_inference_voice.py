import argparse
import logging
import os
import sys
import time
from pathlib import Path

from nexa.constants import (
    DEFAULT_VOICE_GEN_PARAMS,
    EXIT_REMINDER,
    NEXA_RUN_MODEL_MAP_VOICE,
)
from nexa.general import pull_model
from nexa.utils import nexa_prompt, SpinningCursorAnimation
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr


logging.basicConfig(level=logging.INFO)


class NexaVoiceInference:
    """
    A class used for loading voice models and running voice transcription.

    Methods:
      run: Run the voice transcription loop.
      run_streamlit: Run the Streamlit UI.
      transcribe: Transcribe the audio file.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str): Local path of the model.
    output_dir (str): Output directory for transcriptions.
    beam_size (int): Beam size to use for transcription.
    language (str): The language spoken in the audio.
    task (str): Task to execute (transcribe or translate).
    temperature (float): Temperature for sampling.
    compute_type (str): Type to use for computation (e.g., float16, int8, int8_float16).
    output_dir (str): Output directory for transcriptions.

    """
    def __init__(self, model_path, local_path=None, **kwargs):
        self.model_path = model_path
        self.downloaded_path = local_path
        self.params = DEFAULT_VOICE_GEN_PARAMS

        if self.downloaded_path is None:
            self.downloaded_path, run_type = pull_model(self.model_path)

        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        self.params.update(kwargs)
        self.model = None

        if not kwargs.get("streamlit", False):
            self._load_model()
            if self.model is None:
                logging.error(
                    "Failed to load model, Exiting.", exc_info=True
                )
                exit(1)


    @SpinningCursorAnimation()
    def _load_model(self):
        from faster_whisper import WhisperModel

        logging.debug(f"Loading model from: {self.downloaded_path}")
        with suppress_stdout_stderr():
            os.environ["KMP_DUPLICATE_LIB_OK"] = "TRUE"
            self.model = WhisperModel(
                self.downloaded_path,
                device="cpu",
                compute_type=self.params["compute_type"],
            )
        logging.debug("Model loaded successfully")

    def run(self):
        while True:
            try:
                audio_path = nexa_prompt("Enter the path to your audio file: ")
                self._transcribe_audio(audio_path)
            except KeyboardInterrupt:
                print(EXIT_REMINDER)
            except Exception as e:
                logging.error(f"Error during text generation: {e}", exc_info=True)

    def transcribe(self, audio, **kwargs):
        """
        Transcribe the audio file.

        Arguments:
          audio: Path to the input file (or a file-like object), or the audio waveform.
          language: The language spoken in the audio. It should be a language code such
            as "en" or "fr". If not set, the language will be detected in the first 30 seconds
            of audio.
          task: Task to execute (transcribe or translate).
          beam_size: Beam size to use for decoding.
          best_of: Number of candidates when sampling with non-zero temperature.
          patience: Beam search patience factor.
          length_penalty: Exponential length penalty constant.
          repetition_penalty: Penalty applied to the score of previously generated tokens
            (set > 1 to penalize).
          no_repeat_ngram_size: Prevent repetitions of ngrams with this size (set 0 to disable).
          temperature: Temperature for sampling. It can be a tuple of temperatures,
            which will be successively used upon failures according to either
            `compression_ratio_threshold` or `log_prob_threshold`.
          compression_ratio_threshold: If the gzip compression ratio is above this value,
            treat as failed.
          log_prob_threshold: If the average log probability over sampled tokens is
            below this value, treat as failed.
          no_speech_threshold: If the no_speech probability is higher than this value AND
            the average log probability over sampled tokens is below `log_prob_threshold`,
            consider the segment as silent.
          condition_on_previous_text: If True, the previous output of the model is provided
            as a prompt for the next window; disabling may make the text inconsistent across
            windows, but the model becomes less prone to getting stuck in a failure loop,
            such as repetition looping or timestamps going out of sync.
          prompt_reset_on_temperature: Resets prompt if temperature is above this value.
            Arg has effect only if condition_on_previous_text is True.
          initial_prompt: Optional text string or iterable of token ids to provide as a
            prompt for the first window.
          prefix: Optional text to provide as a prefix for the first window.
          suppress_blank: Suppress blank outputs at the beginning of the sampling.
          suppress_tokens: List of token IDs to suppress. -1 will suppress a default set
            of symbols as defined in the model config.json file.
          without_timestamps: Only sample text tokens.
          max_initial_timestamp: The initial timestamp cannot be later than this.
          word_timestamps: Extract word-level timestamps using the cross-attention pattern
            and dynamic time warping, and include the timestamps for each word in each segment.
          prepend_punctuations: If word_timestamps is True, merge these punctuation symbols
            with the next word
          append_punctuations: If word_timestamps is True, merge these punctuation symbols
            with the previous word
          vad_filter: Enable the voice activity detection (VAD) to filter out parts of the audio
            without speech. This step is using the Silero VAD model
            https://github.com/snakers4/silero-vad.
          vad_parameters: Dictionary of Silero VAD parameters or VadOptions class (see available
            parameters and default values in the class `VadOptions`).
          max_new_tokens: Maximum number of new tokens to generate per-chunk. If not set,
            the maximum will be set by the default max_length.
          chunk_length: The length of audio segments. If it is not None, it will overwrite the
            default chunk_length of the FeatureExtractor.
          clip_timestamps:
            Comma-separated list start,end,start,end,... timestamps (in seconds) of clips to
             process. The last end timestamp defaults to the end of the file.
             vad_filter will be ignored if clip_timestamps is used.
          hallucination_silence_threshold:
            When word_timestamps is True, skip silent periods longer than this threshold
             (in seconds) when a possible hallucination is detected
          hotwords:
            Hotwords/hint phrases to provide the model with. Has no effect if prefix is not None.
          language_detection_threshold: If the maximum probability of the language tokens is higher
           than this value, the language is detected.
          language_detection_segments: Number of segments to consider for the language detection.

        Returns:
          A tuple with:

            - a generator over transcribed segments
            - an instance of TranscriptionInfo
        """
        return self.model.transcribe(
            audio,
            **kwargs,
        )


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
            print(f"Transcription: {transcription}")
        except Exception as e:
            logging.error(f"Error during transcription: {e}", exc_info=True)

    def _save_transcription(self, transcription):
        os.makedirs(self.params["output_dir"], exist_ok=True)

        # Generate a filename with timestamp
        filename = f"transcription_{int(time.time())}.txt"
        output_path = os.path.join(self.params["output_dir"], filename)
        try:
            with open(output_path, "w", encoding="utf-8") as f:
                f.write(transcription)
        except UnicodeEncodeError:
            # Fallback to writing with 'ignore' error handler
            with open(output_path, "w", encoding="utf-8", errors="ignore") as f:
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
