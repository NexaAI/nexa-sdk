import logging
import atexit
import threading
import numpy as np
from queue import Queue, Empty
from typing import Optional
from flask import Flask, jsonify, render_template, request
from flask_cors import CORS
from flask_socketio import SocketIO, emit
from nexaai import LLM, GenerationConfig, SamplerConfig, ASR, setup_logging, LlmChatMessage
from nexaai.asr import ASRStreamConfig

logging.basicConfig(level=logging.INFO, format='[%(asctime)s] %(levelname)s: %(message)s', force=True)
setup_logging(level=logging.DEBUG)
logger = logging.getLogger(__name__)

app = Flask(__name__, template_folder='.', static_folder='.')
CORS(app)

socketio = SocketIO(app, cors_allowed_origins='*', async_mode='threading', logger=False, engineio_logger=False)

asr_model: Optional[ASR] = None
llm_model: Optional[LLM] = None
stream_managers = {}


class TranslationStreamManager:
    def __init__(self, asr_model: ASR, llm_model: LLM, sid: str, target_language: str = 'zh'):
        self.asr = asr_model
        self.llm = llm_model
        self.sid = sid
        self.stream = None
        self.stream_context = None
        self.stream_active = False
        self.target_language = target_language
        self.audio_queue: Queue[bytes] = Queue(maxsize=50)
        self.audio_thread: Optional[threading.Thread] = None
        self.stop_event = threading.Event()
        self.translation_queue: Queue[Optional[str]] = Queue(maxsize=20)
        self.translation_thread: Optional[threading.Thread] = None
        self.translation_degraded = False
        self.last_committed_source = ''
        self.pending_text = ''
        self.stable_count = 0

    def start_stream(self, target_language: str = 'zh'):
        self.target_language = target_language
        self.stream_active = True
        logger.info('Starting ASR stream')
        config = ASRStreamConfig(
            sample_rate=16000,
            chunk_duration=4.0,
            overlap_duration=3.5,
            max_queue_size=10,
            buffer_size=1024,
            timestamps='segment',
            beam_size=4,
        )

        def on_transcription(text: str):
            try:
                if text and text.strip():
                    self.on_new_segment(text)
            except Exception as e:
                logger.error(f'Error in transcription callback: {e}', exc_info=True)
                socketio.emit('error', {'message': f'Transcription error: {e}'}, to=self.sid)

        try:
            self.stream_context = self.asr.stream(language='en', config=config)
            self.stream = self.stream_context.__enter__()
            self.stream.start(on_transcription=on_transcription)
            logger.info('ASR stream started successfully')
            self.stop_event.clear()
            self.audio_thread = threading.Thread(
                target=self._audio_worker, name=f'audio-worker-{self.sid}', daemon=True
            )
            self.audio_thread.start()
            self.translation_thread = threading.Thread(
                target=self._translation_worker, name=f'translation-worker-{self.sid}', daemon=True
            )
            self.translation_thread.start()
        except Exception as e:
            logger.error(f'Error starting stream: {e}', exc_info=True)
            if self.stream_context is not None:
                try:
                    self.stream_context.__exit__(None, None, None)
                except Exception:
                    pass
            self.stream = None
            self.stream_active = False
            socketio.emit('error', {'message': f'Failed to start stream: {e}'}, to=self.sid)

    def push_audio(self, audio_bytes: bytes):
        if not self.stream_active:
            return
        if len(audio_bytes) % 2 == 1:
            audio_bytes = audio_bytes[:-1]
        if not audio_bytes:
            return
        try:
            self.audio_queue.put_nowait(audio_bytes)
        except Exception as e:
            logger.error(f'Error enqueuing audio: {e}')
            socketio.emit('error', {'message': f'Audio queue error: {e}'}, to=self.sid)

    def _audio_worker(self):
        try:
            while not self.stop_event.is_set():
                try:
                    audio_bytes = self.audio_queue.get(timeout=0.5)
                except Empty:
                    continue

                try:
                    audio_array = np.frombuffer(audio_bytes, dtype=np.int16).astype(np.float32) / 32768.0
                    if self.stream and self.stream_active:
                        self.stream.push_audio(audio_array.tolist())
                except Exception as e:
                    logger.error(f'Error pushing audio in worker: {e}')
                    socketio.emit('error', {'message': f'Audio processing error: {e}'}, to=self.sid)
                finally:
                    self.audio_queue.task_done()
        except Exception as e:
            logger.error(f'Audio worker crashed: {e}', exc_info=True)
            socketio.emit('error', {'message': f'Audio worker error: {e}'}, to=self.sid)

    def on_new_segment(self, segment_text: str):
        if not segment_text.strip():
            return
        try:
            logger.info(f'New segment: {segment_text}')
            if segment_text == self.pending_text:
                self.stable_count += 1
            else:
                self.pending_text = segment_text
                self.stable_count = 1
            socketio.emit('transcription', {'original': segment_text}, to=self.sid)
            is_sentence_end = segment_text.rstrip().endswith(('.', '?', '!'))
            if self.stable_count < 3 and not is_sentence_end:
                logger.debug(f'Waiting for stabilization (count={self.stable_count}, sentence_end={is_sentence_end})')
                return
            if segment_text == self.last_committed_source:
                logger.debug('Skipping translation duplicate (unchanged source)')
                return
            self.last_committed_source = segment_text
            logger.info('Committing stabilized segment for translation')
            self._enqueue_translation(segment_text)
        except Exception as e:
            logger.error(f'Error processing segment: {e}', exc_info=True)
            socketio.emit('error', {'message': f'Segment processing error: {e}'}, to=self.sid)

    def _translate_text(self, text: str, target_lang: str) -> Optional[str]:
        try:
            prompt = f"""
Translate the following segment into {target_lang}, without additional explanation.\n\n{text}
"""
            prompt = self.llm.apply_chat_template([LlmChatMessage(role='user', content=prompt)], enable_thinking=False)
            logger.info(f'Translation prompt: {prompt}')
            result = self.llm.generate(
                prompt,
                GenerationConfig(max_tokens=256, sampler_config=SamplerConfig(temperature=0.3)),
            )
            self.llm.reset()
            logger.info(f'Translation result: {result.full_text}')
            return result.full_text
        except Exception as e:
            logger.error(f'Translation error: {e}', exc_info=True)
            return None

    def _enqueue_translation(self, text: str):
        if self.translation_degraded:
            logger.error(f'Translation degraded; dropping text: {text}')
            return
        try:
            self.translation_queue.put_nowait(text)
        except Exception as e:
            logger.error(f'Failed to enqueue translation: {e}')

    def _translation_worker(self):
        while not self.stop_event.is_set():
            try:
                text = self.translation_queue.get(timeout=0.5)
            except Empty:
                continue
            if text is None:
                self.translation_queue.task_done()
                break
            try:
                translated = self._translate_text(text, self.target_language)
                if translated:
                    logger.info(f'Translated to: {translated}')
                    socketio.emit(
                        'translation',
                        {'translated': translated, 'original': text, 'language': self.target_language},
                        to=self.sid,
                    )
                else:
                    socketio.emit('error', {'message': 'Translation failed'}, to=self.sid)
            except Exception as e:
                logger.error(f'Translation worker error: {e}', exc_info=True)
            finally:
                self.translation_queue.task_done()

    def stop_stream(self):
        self.stream_active = False

        self.stop_event.set()
        try:
            self.audio_queue.put_nowait(b'')
        except Exception:
            pass
        if self.audio_thread and self.audio_thread.is_alive():
            self.audio_thread.join(timeout=2.0)
        self.audio_thread = None

        try:
            self.translation_queue.put_nowait(None)
        except Exception:
            pass
        if self.translation_thread and self.translation_thread.is_alive():
            self.translation_thread.join(timeout=2.0)
        self.translation_thread = None

        if self.stream:
            try:
                self.stream.stop(graceful=True)
                self.stream = None
            except Exception as e:
                logger.error(f'Error stopping stream: {e}')

        if self.stream_context is not None:
            try:
                self.stream_context.__exit__(None, None, None)
            except Exception as e:
                logger.error(f'Error closing stream context: {e}')

        logger.info('ASR stream stopped')


def initialize_models():
    global asr_model, llm_model
    asr_model = ASR.from_(model='NexaAI/parakeet-tdt-0.6b-v3-npu')
    llm_model = LLM.from_(model='NexaAI/HY-MT1.5-1.8B-npu')
    logger.info('✓ LLM model loaded successfully')


# ============================================================================
# HTTP Routes
# ============================================================================


@app.route('/')
def index():
    return render_template('index.html')


@app.route('/api/health', methods=['GET'])
def health():
    return jsonify(
        {
            'status': 'ok',
            'asr_loaded': asr_model is not None,
            'llm_loaded': llm_model is not None,
        }
    )


@app.route('/api/translate-segment', methods=['POST'])
def translate_segment():
    try:
        data = request.get_json()
        text = data.get('text', '').strip()
        target_lang = data.get('target_lang', 'zh')
        if not text:
            return jsonify({'error': 'Empty text'}), 400
        logger.info(f"REST API: Translating '{text}' to {target_lang}")
        if asr_model is None or llm_model is None:
            return jsonify({'error': 'Models not loaded'}), 500
        manager = TranslationStreamManager(asr_model, llm_model, 'rest-api', target_lang)
        translated = manager._translate_text(text, target_lang)
        if translated is None:
            return jsonify({'error': 'Translation failed'}), 500
        return jsonify(
            {
                'original': text,
                'translated': translated,
                'target_lang': target_lang,
            }
        )

    except Exception as e:
        logger.error(f'Error in translate_segment: {e}', exc_info=True)
        return jsonify({'error': str(e)}), 500


# ============================================================================
# WebSocket Events
# ============================================================================


@socketio.on('connect')
def handle_connect():
    emit('connect', {'data': 'Connected to translation server'})


@socketio.on('disconnect')
def handle_disconnect():
    sid = request.sid
    logger.info(f'Client disconnected: {sid}')

    if sid in stream_managers:
        stream_managers[sid].stop_stream()
        del stream_managers[sid]


@socketio.on('start_stream')
def handle_start_stream(data):
    sid = request.sid
    if isinstance(data, dict):
        target_language = data.get('target_lang', 'zh')
    else:
        target_language = 'zh'
    try:
        logger.info(f'[{sid}] start_stream event received, target_language: {target_language}')
        logger.debug(f'[{sid}] asr_model loaded: {asr_model is not None}, llm_model loaded: {llm_model is not None}')
        if asr_model and llm_model:
            logger.info(f'[{sid}] Creating TranslationStreamManager...')
            manager = TranslationStreamManager(asr_model, llm_model, sid, target_language)
            logger.info(f'[{sid}] Starting ASR stream...')
            manager.start_stream(target_language)
            stream_managers[sid] = manager
            logger.info(f'[{sid}] Stream registered in stream_managers, total streams: {len(stream_managers)}')
            emit(
                'stream_started',
                {
                    'status': 'ok',
                    'target_language': target_language,
                },
            )
            logger.info(f'[{sid}] stream_started event emitted')
        else:
            logger.error(
                f'[{sid}] Models not loaded! asr_model={asr_model is not None}, llm_model={llm_model is not None}'
            )
            emit('error', {'message': 'Models not loaded'})
    except Exception as e:
        logger.error(f'[{sid}] Error starting stream: {e}', exc_info=True)
        emit('error', {'message': f'Failed to start stream: {e}'})


@socketio.on('audio_chunk')
def handle_audio_chunk(data):
    sid = request.sid

    if sid not in stream_managers:
        logger.warning(
            f'[{sid}] Received audio_chunk but stream not started. Active streams: {list(stream_managers.keys())}'
        )
        emit('error', {'message': 'Stream not started. Click "Start Recording" first.'})
        return

    try:
        if isinstance(data, (bytes, bytearray)):
            audio_bytes = bytes(data)
        else:
            raise ValueError(f'Unsupported audio payload type: {type(data)}')

        if not audio_bytes:
            raise ValueError('Empty audio payload')

        if len(audio_bytes) % 2 == 1:
            audio_bytes = audio_bytes[:-1]

        logger.debug(f'[{sid}] Received audio chunk: {len(audio_bytes)} bytes')
        stream_managers[sid].push_audio(audio_bytes)

    except Exception as e:
        logger.error(f'[{sid}] Error processing audio chunk: {e}', exc_info=True)
        emit('error', {'message': f'Audio processing error: {e}'})


@socketio.on('stop_stream')
def handle_stop_stream():
    sid = request.sid

    if sid and sid in stream_managers:
        stream_managers[sid].stop_stream()
        del stream_managers[sid]
        emit('stream_stopped', {'status': 'ok'})
        logger.info(f'[{sid}] Stream stopped')


def cleanup_models():
    global asr_model, llm_model
    logger.info('Cleaning up models...')

    if asr_model:
        try:
            del asr_model
            asr_model = None
        except Exception as e:
            logger.error(f'Error cleaning up ASR model: {e}')

    if llm_model:
        try:
            del llm_model
            llm_model = None
        except Exception as e:
            logger.error(f'Error cleaning up LLM model: {e}')

    logger.info('Models cleaned up')


if __name__ == '__main__':
    logger.info('=' * 80)
    logger.info('Starting NexaAI Live Translator...')
    logger.info('=' * 80)
    atexit.register(cleanup_models)
    initialize_models()

    logger.info('✓ Starting Flask+SocketIO server on http://127.0.0.1:5000')
    logger.info('=' * 80)
    socketio.run(
        app,
        host='127.0.0.1',
        port=5000,
    )
