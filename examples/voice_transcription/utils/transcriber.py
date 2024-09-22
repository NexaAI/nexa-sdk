import os
import wave
import streamlit as st
from utils.segmenter import Segmenter
from nexa.gguf import NexaVoiceInference
from nexa.gguf import NexaTextInference  # Import the text inference class
import logging

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class RealTimeTranscriber:
    def __init__(self, model_path, beam_size, task, temperature, output_directory="../transcriptions/audio/", verbose=False, language=None):
        self.segmenter = Segmenter()
        self.translation = task == "translate"
        self.output_directory = output_directory
        self.verbose = verbose

        # Initialize NexaVoiceInference with model and compute_type
        if self.verbose:
            logger.info(f"Loading model from {model_path}...")
        self.inference = NexaVoiceInference(
            model_path=model_path,
            beam_size=beam_size,
            language=language,
            task=task,
            temperature=temperature,
            compute_type="default"
        )
        if self.verbose:
            st.write(f"Model loaded: {self.inference.downloaded_path}")
            logger.info(f"Model loaded: {self.inference.downloaded_path}")

    def write_wav(self, filename, audio_data):
        try:
            with wave.open(filename, 'wb') as wf:
                wf.setnchannels(1)
                wf.setsampwidth(2)
                wf.setframerate(self.segmenter.sample_rate)
                wf.writeframes(audio_data)
            if self.verbose:
                st.write(f"Saved {filename}")
            return filename
        except Exception as e:
            st.error(f"Error writing WAV file: {e}")
            logger.error(f"Error writing WAV file: {e}")
            return None

    def transcribe_audio(self, audio_data):
        try:
            if self.verbose:
                if self.translation:
                    st.write(f"Translating audio to English...")
                else:
                    st.write(f"Transcribing audio...")

            # Transcribe directly from the audio data (in-memory)
            segments, _ = self.inference.model.transcribe(
                audio_data,
                beam_size=self.inference.params["beam_size"],
                language=self.inference.params["language"],
                task=self.inference.params["task"],
                temperature=self.inference.params["temperature"],
                vad_filter=True
            )
            transcription = "".join(segment.text for segment in segments)
            return transcription

        except Exception as e:
            logger.error(f"Transcription error: {e}")
            return None
    
    def process_chunks(self, transcription_container):
        self.segmenter.start_stream()
        audio_chunks = self.segmenter.vad_collector()

        if not os.path.exists(self.output_directory):
            os.makedirs(self.output_directory)

        for i, chunk in enumerate(audio_chunks):
            if not self.segmenter.running:
                break

            filename = f"{self.output_directory}/chunk_{i}.wav"
            saved_filename = self.write_wav(filename, chunk)
            self.last_transcription_set = False
            if saved_filename:
                transcription = self.transcribe_audio(saved_filename)
                logger.info(f"Adding transcription: {i} {transcription}")
                if transcription:
                    st.session_state["transcription"] += transcription
                    transcription_container.text_area("Transcription", value=st.session_state["transcription"], height=300)

        self.segmenter.stop_stream()

    def start_recording_foreground(self, transcription_container):
        """Start recording in the foreground."""
        if "transcription" not in st.session_state:
            st.session_state["transcription"] = ""
        self.segmenter.running = True
        st.session_state["recording_status"] = "Recording..."
        self.process_chunks(transcription_container)
        st.session_state["recording_status"] = "Recording completed"

    def stop_recording_foreground(self, transcription_container):
        """Stop recording in the foreground."""
        self.segmenter.running = False
        st.session_state["recording_status"] = "Recording stopped."

    def reset_transcription(self):
        """Reset the transcription."""
        st.session_state["transcription"] = ""
        st.session_state["recording_status"] = "Transcription reset."

class TextInference:
    def __init__(self, model_path="gemma", temperature=0.7, max_new_tokens=512, top_k=50, top_p=0.9):
        # Initialize NexaTextInference with the specified parameters
        try:
            self.inference = NexaTextInference(
                model_path=model_path,
                stop_words=[],
                temperature=temperature,
                max_new_tokens=max_new_tokens,
                top_k=top_k,
                top_p=top_p,
                profiling=False
            )
            logger.info(f"Text model loaded: {self.inference.downloaded_path}")
        except ValueError as e:
            logger.warning(str(e))
            raise e
        
    def generate_summary(self, prompt):
        for chunk in self.inference.create_completion(prompt, stream=True):
            yield chunk["choices"][0]["text"]