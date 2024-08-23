import os
import tempfile
from pathlib import Path
from nexa.gguf import NexaVoiceInference
from unittest.mock import patch, MagicMock

def test_voice_generation():
    # Initialize the NexaVoiceInference
    voice_inference = NexaVoiceInference(
        model_path="faster-whisper-tiny",
        local_path=None,
        output_dir=tempfile.mkdtemp(),  # Use a temporary directory for output
        beam_size=5,
        language="en",
        task="transcribe",
        temperature=0.0,
        compute_type="float32"
    )

    # Create a mock audio file
    mock_audio_path = Path(tempfile.mkdtemp()) / "test_audio.wav"
    mock_audio_path.touch()  # Create an empty file

    # Mock the WhisperModel's transcribe method
    mock_segments = [
        MagicMock(text="This is a test transcription."),
        MagicMock(text=" It works perfectly.")
    ]
    mock_transcribe = MagicMock(return_value=(mock_segments, None))

    # Test _transcribe_audio method
    with patch.object(voice_inference.model, 'transcribe', mock_transcribe):
        voice_inference._transcribe_audio(str(mock_audio_path))

    # Assertions
    mock_transcribe.assert_called_once_with(
        str(mock_audio_path),
        beam_size=5,
        language="en",
        task="transcribe",
        temperature=0.0,
        vad_filter=True
    )

    # Check if the transcription was saved
    transcription_files = list(Path(voice_inference.params["output_dir"]).glob("transcription_*.txt"))
    assert len(transcription_files) == 1, "Transcription file was not created"

    # Check the content of the transcription file
    with open(transcription_files[0], 'r') as f:
        content = f.read()
    assert content == "This is a test transcription. It works perfectly.", "Transcription content is incorrect"

    print("Voice generation test passed successfully!")

if __name__ == "__main__":
    test_voice_generation()