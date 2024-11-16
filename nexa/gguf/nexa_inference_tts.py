import os
import ctypes
import logging
import numpy as np
import time
import threading
import platform
import functools
from .bark import bark_cpp

from nexa.general import pull_model
from nexa.gguf.lib_utils import is_gpu_available

class NexaTTSInference:
    """
    A class used for loading Bark text-to-speech models and running text-to-speech generation.

    Methods:
        run: Run the text-to-speech generation loop.
        audio_generation: Generate audio from the user input.

    Args:
        model_path (str): Path to the Bark model file.
        n_threads (int): Number of threads to use for processing. Defaults to 1.
        seed (int): Seed for random number generation. Defaults to 0.
        output_dir (str): Output directory for tts. Defaults to "tts".
        sampling_rate (int): Sampling rate for audio processing. Defaults to 24000.
        verbosity (int): Verbosity level for the Bark model. Defaults to 0.
    """
    
    def __init__(self, model_path=None, local_path=None, n_threads=1, seed=0, 
                 sampling_rate=24000, verbosity=0, win_stack_size=16*1024*1024,
                 device="auto", n_gpu_layers=4, **kwargs):
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")
            
        self.model_path = model_path
        self.downloaded_path = local_path
        self.n_threads = n_threads
        self.seed = seed
        self.sampling_rate = sampling_rate
        self.verbosity = verbosity
        self.win_stack_size = win_stack_size
        self.device = device
        self.n_gpu_layers = n_gpu_layers
        self.params = {
            "output_path": os.path.join(os.getcwd(), "tts"),
        }
        self.params.update(kwargs)
        self.context = None

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
        logging.debug(f"Loading model from {self.downloaded_path}")
        try:
            params = bark_cpp.bark_context_default_params()
            params.sample_rate = self.sampling_rate
            params.verbosity = self.verbosity

            # Use configured n_gpu_layers when device is auto/gpu and GPU is available
            if self.device == "auto" or self.device == "gpu":
                if is_gpu_available():
                    params.n_gpu_layers = self.n_gpu_layers
                    logging.info(f"Using GPU acceleration with {self.n_gpu_layers} layers")
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
            logging.debug("Model loaded successfully")
        except Exception as e:
            logging.error(f"Error loading model: {e}")
            raise


    def run(self):
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
        Generate audio from the user input.

        Args:
            user_input (str): User input for audio generation.

        Returns:
            np.array: Audio data.
        """
        c_text = ctypes.c_char_p(user_input.encode('utf-8'))
        success = bark_cpp.bark_generate_audio(self.context, c_text, self.n_threads)
        
        if not success:
            raise RuntimeError("Failed to generate audio")

        audio_size = bark_cpp.bark_get_audio_data_size(self.context)
        audio_data = bark_cpp.bark_get_audio_data(self.context)
        
        return np.ctypeslib.as_array(audio_data, shape=(audio_size,))


    def _save_audio(self, audio_data, sampling_rate, output_path):
        os.makedirs(output_path, exist_ok=True)
        file_name = f"audio_{int(time.time())}.wav"
        file_path = os.path.join(output_path, file_name)
        import soundfile as sf
        sf.write(file_path, audio_data, sampling_rate)


    def get_load_time(self):
        """
        Get the time taken to load the model.

        Returns:
            int: The load time in microseconds.
        """
        return bark_cpp.bark_get_load_time(self.context)


    def get_eval_time(self):
        """
        Get the time taken to evaluate (generate audio).

        Returns:
            int: The evaluation time in microseconds.
        """
        return bark_cpp.bark_get_eval_time(self.context)


    def reset_statistics(self):
        """
        Reset the internal statistics of the Bark context.
        """
        bark_cpp.bark_reset_statistics(self.context)


    def __del__(self):
        """
        Destructor to free the Bark context when the instance is deleted.
        """
        if self.context:
            bark_cpp.bark_free(self.context)

if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Run text to speech generation with Bark model")
    parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    parser.add_argument("-o", "--output_dir", type=str, default="tts", help="Output directory for tts")
    parser.add_argument("-r", "--sampling_rate", type=int, default=24000, help="Sampling rate for audio processing")
    parser.add_argument("-t", "--n_threads", type=int, default=1, help="Number of threads to use for processing")
    parser.add_argument("-s", "--seed", type=int, default=0, help="Seed for random number generation")
    parser.add_argument("-v", "--verbosity", type=int, default=1, help="Verbosity level for the Bark model")

    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if k not in ["model_path", "n_threads", "seed", "sampling_rate", "verbosity"]}
    
    inference = NexaTTSInference(args.model_path, n_threads=args.n_threads, seed=args.seed, 
                                sampling_rate=args.sampling_rate, verbosity=args.verbosity, **kwargs)
    inference.run()