# Copyright 2024-2026 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
NexaAI Live Translator - Advanced Backend with ASR Streaming

This module provides the WebSocket server implementation for real-time ASR
and LLM translation. It handles audio streaming from the frontend and returns
transcriptions and translations in real-time.
"""

import asyncio
import json
import logging
import numpy as np
from typing import Optional

from flask import Flask
from flask_socketio import SocketIO, emit, disconnect
from nexaai.asr import ASR, ASRStreamConfig
from nexaai.llm import LLM, GenerationConfig, LlmChatMessage

logger = logging.getLogger(__name__)


class TranslationStreamManager:
    """Manages audio streaming, ASR transcription, and LLM translation."""

    def __init__(self, asr_model: ASR, llm_model: LLM):
        """
        Initialize the stream manager.

        Args:
            asr_model: ASR model instance
            llm_model: LLM model instance
        """
        self.asr = asr_model
        self.llm = llm_model
        self.stream = None
        self.stream_active = False
        self.source_language = 'en'
        self.target_language = 'zh'
        self.transcription_buffer = []
        self.last_segment = ''

    def start_stream(self, language: str, emit_callback):
        """
        Start ASR streaming session.

        Args:
            language: Source language ('en' or 'zh')
            emit_callback: Callback function to emit events to frontend
        """
        self.source_language = language
        self.target_language = 'zh' if language == 'en' else 'en'
        self.transcription_buffer = []
        self.emit_callback = emit_callback
        self.stream_active = True

        # Configure ASR stream with 3-second chunks and segment-based output
        config = ASRStreamConfig(
            sample_rate=16000,
            chunk_duration=3.0,
            overlap_duration=1.0,
            max_queue_size=10,
            buffer_size=4096,
            timestamps='segment',
            beam_size=5,
        )

        # Create transcription callback
        def on_transcription(text: str):
            if text and text.strip():
                self.on_new_segment(text)

        # Open streaming context
        self.stream = self.asr.stream(language=self.source_language, config=config)
        self.stream.start(on_transcription=on_transcription)

        logger.info(f'ASR stream started for language: {self.source_language}')

    def push_audio(self, audio_bytes: bytes):
        """
        Push audio data to ASR stream.

        Args:
            audio_bytes: Raw audio bytes (16-bit PCM at 16kHz)
        """
        if not self.stream_active or not self.stream:
            return

        try:
            # Convert bytes to float32 array (-1.0 to 1.0 range)
            audio_array = np.frombuffer(audio_bytes, dtype=np.int16).astype(np.float32) / 32768.0
            self.stream.push_audio(audio_array.tolist())
        except Exception as e:
            logger.error(f'Error pushing audio: {e}')
            self.emit_callback('error', {'message': f'Audio processing error: {e}'})

    def on_new_segment(self, segment_text: str):
        """
        Handle new transcription segment from ASR.

        Args:
            segment_text: Transcribed text segment
        """
        if not segment_text.strip():
            return

        try:
            logger.info(f'New segment: {segment_text}')

            # Emit transcription to frontend
            self.emit_callback(
                'transcription',
                {
                    'original': segment_text,
                    'language': self.source_language,
                },
            )

            # Translate segment
            translated = self._translate_text(
                segment_text,
                self.source_language,
                self.target_language,
            )

            if translated:
                logger.info(f'Translated to: {translated}')
                self.emit_callback(
                    'translation',
                    {
                        'translated': translated,
                        'original': segment_text,
                        'language': self.target_language,
                    },
                )
            else:
                self.emit_callback('error', {'message': 'Translation failed for segment'})

        except Exception as e:
            logger.error(f'Error processing segment: {e}')
            self.emit_callback('error', {'message': f'Segment processing error: {e}'})

    def _translate_text(self, text: str, source_lang: str, target_lang: str) -> Optional[str]:
        """
        Translate text using LLM.

        Args:
            text: Text to translate
            source_lang: Source language
            target_lang: Target language

        Returns:
            Translated text or None if translation fails
        """
        if not text.strip():
            return ''

        if source_lang == target_lang:
            return text

        try:
            if target_lang == 'zh':
                system_prompt = (
                    'You are a professional translator. Translate the following English text to Chinese. '
                    'Output ONLY the Chinese translation, no explanations.'
                )
                user_prompt = f'Translate to Chinese: {text}'
            else:  # target_lang == 'en'
                system_prompt = (
                    'You are a professional translator. Translate the following Chinese text to English. '
                    'Output ONLY the English translation, no explanations.'
                )
                user_prompt = f'Translate to English: {text}'

            messages = [
                LlmChatMessage(role='system', content=system_prompt),
                LlmChatMessage(role='user', content=user_prompt),
            ]

            prompt = self.llm.apply_chat_template(messages)
            result = self.llm.generate(
                prompt,
                GenerationConfig(max_tokens=256, temperature=0.3),
            )

            return result.strip() if result else None

        except Exception as e:
            logger.error(f'Translation error: {e}')
            return None

    def stop_stream(self):
        """Stop ASR streaming and cleanup."""
        if self.stream:
            try:
                self.stream.stop(graceful=True)
                self.stream = None
            except Exception as e:
                logger.error(f'Error stopping stream: {e}')

        self.stream_active = False
        self.transcription_buffer = []
        logger.info('ASR stream stopped')


def create_socketio_app(app: Flask, asr_model: ASR, llm_model: LLM):
    """
    Create Flask-SocketIO app with translation streaming.

    Args:
        app: Flask application instance
        asr_model: ASR model instance
        llm_model: LLM model instance

    Returns:
        SocketIO instance
    """
    socketio = SocketIO(app, cors_allowed_origins='*')

    # Store stream managers per session
    stream_managers = {}

    @socketio.on('connect')
    def handle_connect():
        """Handle client connection."""
        sid = request.sid if hasattr(request, 'sid') else None
        logger.info(f'Client connected: {sid}')
        emit('connect', {'data': 'Connected to translation server'})

    @socketio.on('disconnect')
    def handle_disconnect():
        """Handle client disconnection."""
        sid = request.sid if hasattr(request, 'sid') else None
        logger.info(f'Client disconnected: {sid}')

        # Cleanup stream
        if sid and sid in stream_managers:
            stream_managers[sid].stop_stream()
            del stream_managers[sid]

    @socketio.on('start_stream')
    def handle_start_stream(data):
        """
        Start ASR streaming.

        Expects data:
        {
            "language": "en" or "zh"
        }
        """
        sid = request.sid if hasattr(request, 'sid') else None
        language = data.get('language', 'en')

        try:
            # Create new stream manager
            def emit_callback(event_type: str, event_data: dict):
                emit(event_type, event_data)

            manager = TranslationStreamManager(asr_model, llm_model)
            manager.start_stream(language, emit_callback)
            stream_managers[sid] = manager

            emit(
                'stream_started',
                {
                    'status': 'ok',
                    'source_language': language,
                    'target_language': 'zh' if language == 'en' else 'en',
                },
            )

            logger.info(f'Stream started for session {sid}, language: {language}')

        except Exception as e:
            logger.error(f'Error starting stream: {e}')
            emit('error', {'message': f'Failed to start stream: {e}'})

    @socketio.on('audio_chunk')
    def handle_audio_chunk(data):
        """
        Receive audio chunk from frontend.

        Expects data:
        {
            "audio": base64-encoded audio bytes
        }
        """
        sid = request.sid if hasattr(request, 'sid') else None

        if sid not in stream_managers:
            emit('error', {'message': 'Stream not started'})
            return

        try:
            # Decode base64 audio
            import base64

            audio_bytes = base64.b64decode(data.get('audio', ''))
            stream_managers[sid].push_audio(audio_bytes)

        except Exception as e:
            logger.error(f'Error processing audio chunk: {e}')
            emit('error', {'message': f'Audio processing error: {e}'})

    @socketio.on('stop_stream')
    def handle_stop_stream():
        """Stop ASR streaming."""
        sid = request.sid if hasattr(request, 'sid') else None

        if sid and sid in stream_managers:
            stream_managers[sid].stop_stream()
            del stream_managers[sid]
            emit('stream_stopped', {'status': 'ok'})
            logger.info(f'Stream stopped for session {sid}')

    return socketio
