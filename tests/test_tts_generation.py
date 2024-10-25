import os
import tempfile
import numpy as np
from pathlib import Path
from nexa.gguf import NexaTTSInference
from unittest.mock import patch

def test_tts_generation():
    # Initialize the NexaTTSInference with a temporary output directory
    temp_output_dir = tempfile.mkdtemp()
    tts_inference = NexaTTSInference(
        model_path="suno/bark-small",
        local_path=None,
        output_dir=temp_output_dir,
        n_threads=1,
        seed=0,
        sampling_rate=24000,
        verbosity=0
    )

    # Mock test input text
    test_text = "This is a test audio generation."

    # Create mock audio data
    mock_audio_data = np.zeros(24000)  # 1 second of silence at 24kHz

    # Mock the bark_cpp functions
    with patch('nexa.gguf.bark.bark_cpp.bark_generate_audio', return_value=True), \
         patch('nexa.gguf.bark.bark_cpp.bark_get_audio_data_size', return_value=len(mock_audio_data)), \
         patch('nexa.gguf.bark.bark_cpp.bark_get_audio_data', return_value=mock_audio_data):

        # Test audio generation
        generated_audio = tts_inference.audio_generation(test_text)

        # Assertions
        assert isinstance(generated_audio, np.ndarray), "Generated audio should be a numpy array"
        assert len(generated_audio) == 24000, "Generated audio length should match mock data"
        assert generated_audio.dtype == np.float32, "Generated audio should be float32"

        # Test audio saving
        tts_inference._save_audio(generated_audio, tts_inference.sampling_rate, temp_output_dir)

        # Check if audio file was created
        audio_files = list(Path(temp_output_dir).glob("audio_*.wav"))
        assert len(audio_files) == 1, "Audio file was not created"

        # Check if the audio file has the correct size
        assert os.path.getsize(audio_files[0]) > 0, "Audio file is empty"

    # Test timing methods
    with patch('nexa.gguf.bark.bark_cpp.bark_get_load_time', return_value=1000):
        load_time = tts_inference.get_load_time()
        assert load_time == 1000, "Load time should match mock value"

    with patch('nexa.gguf.bark.bark_cpp.bark_get_eval_time', return_value=2000):
        eval_time = tts_inference.get_eval_time()
        assert eval_time == 2000, "Eval time should match mock value"

    # Test error handling
    with patch('nexa.gguf.bark.bark_cpp.bark_generate_audio', return_value=False):
        try:
            tts_inference.audio_generation(test_text)
            assert False, "Should have raised RuntimeError"
        except RuntimeError:
            pass

    print("TTS generation test passed successfully!")

    # Clean up
    for file in audio_files:
        file.unlink()
    os.rmdir(temp_output_dir)

if __name__ == "__main__":
    test_tts_generation()