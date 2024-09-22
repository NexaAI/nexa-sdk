import streamlit as st
import sounddevice as sd
from scipy.io.wavfile import write
from tempfile import NamedTemporaryFile
from nexa.gguf import NexaVoiceInference

voice_model = NexaVoiceInference(
    model_path="faster-whisper-base",
    local_path=None,
    beam_size=5,
    task="transcribe",
    temperature=0.0,
    compute_type="default",
)

def record_and_transcribe(duration=5, fs=16000):
    info_placeholder = st.empty()
    info_placeholder.info("Recording...")
    
    recording = sd.rec(int(duration * fs), samplerate=fs, channels=1)
    sd.wait()
    
    info_placeholder.empty()
    
    with NamedTemporaryFile(delete=False, suffix=".wav") as f:
        write(f.name, fs, recording)
        audio_path = f.name
    
    segments, _ = voice_model.model.transcribe(audio_path)
    transcription = "".join(segment.text for segment in segments)
    return transcription
