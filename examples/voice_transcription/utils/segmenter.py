import collections
import webrtcvad
import pyaudio
import streamlit as st
import logging

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class Segmenter:
    def __init__(self, vad_mode=3, sample_rate=16000, chunk_duration_ms=20, padding_duration_ms=200, min_segment_duration_s=1, frames_per_buffer=None):
        self.vad = webrtcvad.Vad()
        self.vad.set_mode(vad_mode)
        self.sample_rate = sample_rate
        self.chunk_duration_ms = chunk_duration_ms
        self.padding_duration_ms = padding_duration_ms
        self.min_segment_duration_s = min_segment_duration_s
        self.frames_per_buffer = frames_per_buffer or int(self.sample_rate / 1000 * self.chunk_duration_ms)
        self.stream = None
        self.running = False

    def start_stream(self):
        try:
            p = pyaudio.PyAudio()
            self.stream = p.open(format=pyaudio.paInt16,
                                channels=1,
                                rate=self.sample_rate,
                                input=True,
                                frames_per_buffer=self.frames_per_buffer)
            self.running = True
            logger.info(f"Audio stream started with format: {pyaudio.paInt16}, channels: 1, rate: {self.sample_rate}")
        except Exception as e:
            st.error(f"Error initializing audio stream: {e}")
            self.running = False


    def stop_stream(self):
        try:
            if self.stream and self.stream.is_active():
                self.stream.stop_stream()
                self.stream.close()
            self.running = False
        except Exception as e:
            st.error(f"Error stopping audio stream: {e}")

    def read_audio(self):
        try:
            return self.stream.read(self.frames_per_buffer, exception_on_overflow=False)
        except IOError as e:
            st.error(f"Error reading audio: {e}")
            return None

    def vad_collector(self):
        num_padding_chunks = int(self.padding_duration_ms / self.chunk_duration_ms)
        ring_buffer = collections.deque(maxlen=num_padding_chunks)
        triggered = False
        voiced_frames = []
        segment_duration_ms = 0
        min_segment_duration_ms = self.min_segment_duration_s * 1000

        while self.running:
            frame = self.read_audio()
            if frame is None:
                continue

            is_speech = self.vad.is_speech(frame, self.sample_rate)
            segment_duration_ms += self.chunk_duration_ms

            if not triggered:
                ring_buffer.append((frame, is_speech))
                num_voiced = len([f for f, speech in ring_buffer if speech])
                if num_voiced > 0.9 * ring_buffer.maxlen:
                    triggered = True
                    voiced_frames.extend([f[0] for f in ring_buffer])
                    ring_buffer.clear()
            else:
                voiced_frames.append(frame)
                ring_buffer.append((frame, is_speech))
                num_unvoiced = len([f for f, speech in ring_buffer if not speech])
                
                if num_unvoiced > 0.9 * ring_buffer.maxlen and segment_duration_ms >= min_segment_duration_ms:
                    triggered = False
                    if len(voiced_frames) > 0:  # Only yield non-empty chunks
                        yield b''.join(voiced_frames)
                    ring_buffer.clear()
                    voiced_frames = []
                    segment_duration_ms = 0

        # Yield the final segment if it contains voiced frames
        if voiced_frames:
            yield b''.join(voiced_frames)
