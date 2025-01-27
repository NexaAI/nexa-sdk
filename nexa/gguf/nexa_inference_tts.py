import os
import ctypes
import logging
import numpy as np
import time
import threading
import platform
import functools
from nexa.general import pull_model
from nexa.gguf.lib_utils import is_gpu_available

from nexa.gguf.bark import bark_cpp
import nexa.gguf.outetts as outetts

class NexaTTSInference:
    """
    A class used for loading either Bark or OuteTTS text-to-speech models and running text-to-speech generation.

    Methods:
        run: Run the text-to-speech generation loop.
        audio_generation: Generate audio from the user input.

    Args:
        model_path (str): Path to the model file.
        local_path (str): If provided, use this local path instead of downloading.
        tts_engine (str): "bark" or "outetts" to choose which TTS engine to use.
        n_threads (int): Number of threads to use for processing. Defaults to 1.
        seed (int): Seed for random number generation. Defaults to 0.
        sampling_rate (int): Sampling rate for audio processing. Defaults to 24000.
        verbosity (int): Verbosity level for the Bark model. Defaults to 0.
        device (str): "auto"/"gpu"/"cpu" for Bark GPU acceleration.
        language (str): (For OuteTTS) Language code, e.g. "en".
        speaker_name (str): (For OuteTTS) Default speaker name to load.
    """
    
    def __init__(self, model_path=None, local_path=None, tts_engine="outetts", 
                 n_threads=1, seed=0, sampling_rate=24000, verbosity=0,
                 win_stack_size=16*1024*1024, device="auto", language="en", speaker_name="male_1", **kwargs):
        
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")
            
        self.model_path = model_path
        self.downloaded_path = local_path
        self.tts_engine = tts_engine.lower()
        self.n_threads = n_threads
        self.seed = seed
        self.sampling_rate = sampling_rate
        self.verbosity = verbosity
        self.win_stack_size = win_stack_size
        self.device = device
        self.language = language
        self.speaker_name = speaker_name
        self.params = {
            "output_path": os.path.join(os.getcwd(), "tts"),
        }
        self.params.update(kwargs)
        self.context = None
        self.interface = None
        self.speaker = None

        if self.downloaded_path is None:
            self.downloaded_path, _ = pull_model(self.model_path, **kwargs)

        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        self._load_model()


    def _windows_operation(func):
        """
        Method decorator to ensure proper stack size for Windows operations.
        Only affects Windows systems; on other platforms, calls the function directly.
        Uses the instance's win_stack_size parameter.
        """
        @functools.wraps(func)
        def wrapper(self, *args, **kwargs):
            if platform.system() != 'Windows':
                return func(self, *args, **kwargs)
                
            def threaded_func():
                self._thread_result = func(self, *args, **kwargs)
                
            original_stack_size = threading.stack_size(self.win_stack_size)
            thread = threading.Thread(target=threaded_func, name=f"BarkThread_{func.__name__}")
            thread.start()
            thread.join()
            threading.stack_size(original_stack_size)
            
            if hasattr(self, '_thread_result'):
                result = self._thread_result
                delattr(self, '_thread_result')
                return result
            return None
            
        return wrapper


    @_windows_operation
    def _load_model(self):
        if self.tts_engine == "bark":
            # Bark loading
            logging.debug(f"Loading Bark model from {self.downloaded_path}")
            try:
                params = bark_cpp.bark_context_default_params()
                params.sample_rate = self.sampling_rate
                params.verbosity = self.verbosity

                # Use configured n_gpu_layers when device is auto/gpu and GPU is available
                if self.device == "auto" or self.device == "gpu":
                    if is_gpu_available():
                        params.n_gpu_layers = 4 
                    else:
                        params.n_gpu_layers = 0
                        logging.info("GPU not available, falling back to CPU")
                else:
                    params.n_gpu_layers = 0
                    logging.info("Using CPU mode")

                c_model_path = ctypes.c_char_p(self.downloaded_path.encode('utf-8'))
                c_seed = ctypes.c_uint32(self.seed)
                
                try:
                    self.context = bark_cpp.bark_load_model(c_model_path, params, c_seed)
                except Exception as e:
                    logging.error(f"Failed to load model with GPU. Falling back to CPU: {e}")
                    params.n_gpu_layers = 0
                    self.context = bark_cpp.bark_load_model(c_model_path, params, c_seed)

                if not self.context:
                    raise RuntimeError("Failed to load Bark model")
                logging.debug("Bark model loaded successfully")
            except Exception as e:
                logging.error(f"Error loading Bark model: {e}")
                raise

        elif self.tts_engine == "outetts":
            # OuteTTS loading
            logging.debug(f"Loading OuteTTS model from {self.downloaded_path}")
            # For OuteTTS we assume model_path is a GGUF model.
            # Example from run_outetts.py:
            model_config = outetts.GGUFModelConfig_v1(
                model_path=self.downloaded_path,
                language=self.language,
                n_gpu_layers=-1 if is_gpu_available() else 0
            )
            if "OuteTTS-0.2-500M" in self.model_path:
                self.interface = outetts.InterfaceGGUF(model_version="0.2", cfg=model_config)
            elif "OuteTTS-0.1-350M" in self.model_path:
                self.interface = outetts.InterfaceGGUF(model_version="0.1", cfg=model_config)
            else:
                raise ValueError(f"Unknown model path: {self.model_path}")
                
            # Load a default speaker (if desired)
            self.speaker = self.interface.load_default_speaker(name=self.speaker_name)
            logging.debug("OuteTTS model loaded successfully")
        else:
            raise ValueError(f"Unknown TTS engine: {self.tts_engine}")


    def run(self):
        # For Bark, we can still use the spinner; for OuteTTS it's optional
        from nexa.gguf.llama._utils_spinner import start_spinner, stop_spinner

        while True:
            try:
                user_input = input("Enter text to generate audio: ")
                stop_event, spinner_thread = start_spinner(
                    style="default",
                    message=""
                )

                audio_data = self.audio_generation(user_input)
               
                self._save_audio(audio_data, self.sampling_rate, self.params["output_path"])
                logging.info(f"Audio saved to {self.params['output_path']}")                
            
                stop_spinner(stop_event, spinner_thread)
            except KeyboardInterrupt:
                print("Exiting...")
                break
            except Exception as e:
                logging.error(f"Error during audio generation: {e}", exc_info=True)


    @_windows_operation
    def audio_generation(self, user_input):
        """
        Generate audio from the user input using the selected TTS engine.
        """
        if self.tts_engine == "bark":
            # Bark Generation
            c_text = ctypes.c_char_p(user_input.encode('utf-8'))
            success = bark_cpp.bark_generate_audio(self.context, c_text, self.n_threads)
            
            if not success:
                raise RuntimeError("Failed to generate audio with Bark")

            audio_size = bark_cpp.bark_get_audio_data_size(self.context)
            audio_data = bark_cpp.bark_get_audio_data(self.context)
            return np.ctypeslib.as_array(audio_data, shape=(audio_size,))

        elif self.tts_engine == "outetts":
            # OuteTTS Generation
            output = self.interface.generate(
                text=user_input,
                temperature=0.1,
                repetition_penalty=1.1,
                max_length=4096,
                speaker=self.speaker
            )
            return output
        
        else:
            raise ValueError(f"Unknown TTS engine: {self.tts_engine}")


    def _save_audio(self, audio_data, sampling_rate, output_path):
        import soundfile as sf
        os.makedirs(output_path, exist_ok=True)
        file_name = f"audio_{int(time.time())}.wav"
        file_path = os.path.join(output_path, file_name)

        if self.tts_engine == "bark":
            sf.write(file_path, audio_data, sampling_rate)
        elif self.tts_engine == "outetts":
            audio_data.save(file_path)
        return file_path


    def get_load_time(self):
        """
        Get the time taken to load the model for Bark. For OuteTTS, not implemented.
        """
        if self.tts_engine == "bark":
            return bark_cpp.bark_get_load_time(self.context)
        else:
            return None


    def get_eval_time(self):
        """
        Get the time taken to evaluate (generate audio) for Bark. For OuteTTS, not implemented.
        """
        if self.tts_engine == "bark":
            return bark_cpp.bark_get_eval_time(self.context)
        else:
            return None


    def reset_statistics(self):
        """
        Reset the internal statistics of the Bark context. For OuteTTS, not implemented.
        """
        if self.tts_engine == "bark" and self.context:
            bark_cpp.bark_reset_statistics(self.context)


    def __del__(self):
        """
        Destructor to free the Bark context when the instance is deleted.
        For OuteTTS, not currently needed.
        """
        if self.tts_engine == "bark":
            if self.context:
                bark_cpp.bark_free(self.context)

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Run text to speech generation with Bark or OuteTTS model")
    parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    parser.add_argument("-e", "--tts_engine", type=str, default="bark", help="TTS engine to use: 'bark' or 'outetts'")
    parser.add_argument("-o", "--output_dir", type=str, default="tts", help="Output directory for tts")
    parser.add_argument("-r", "--sampling_rate", type=int, default=24000, help="Sampling rate for audio processing")
    parser.add_argument("-t", "--n_threads", type=int, default=1, help="Number of threads to use for processing")
    parser.add_argument("-s", "--seed", type=int, default=0, help="Seed for random number generation")
    parser.add_argument("-v", "--verbosity", type=int, default=1, help="Verbosity level for the Bark model")
    parser.add_argument("-d", "--device", type=str, default="auto", help="Device to use for Bark: 'auto', 'gpu', or 'cpu'")
    parser.add_argument("-l", "--language", type=str, default="en", help="Language for OuteTTS model")
    parser.add_argument("-sp", "--speaker_name", type=str, default="male_1", help="Default speaker name for OuteTTS")

    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if k not in ["model_path", "n_threads", "seed", "sampling_rate", "verbosity", "tts_engine", "device", "n_gpu_layers", "language", "speaker_name"]}

    inference = NexaTTSInference(
        model_path=args.model_path,
        tts_engine=args.tts_engine,
        n_threads=args.n_threads,
        seed=args.seed, 
        sampling_rate=args.sampling_rate, 
        verbosity=args.verbosity,
        device=args.device,
        language=args.language,
        speaker_name=args.speaker_name,
        **kwargs
    )
    inference.run()