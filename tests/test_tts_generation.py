from nexa.gguf import NexaTTSInference


def test_tts_generation_barkcpp():
    tts = NexaTTSInference(
        model_path="bark-small",
        tts_engine='bark',
        local_path=None,
        n_threads=4,
        seed=42,
        sampling_rate=24000,
        verbosity=2
    )

    # Generate audio from prompt
    prompt = "Hello, this is a test of the Bark text to speech system."
    audio_data = tts.audio_generation(prompt)

    # Save the generated audio
    tts._save_audio(audio_data, tts.sampling_rate, "tts_output/barkcpp")


def test_tts_generation_outetts():
    tts = NexaTTSInference(
        model_path="OuteTTS-0.2-500M:q4_K_M",
        local_path=None,
        n_threads=4,
        seed=42,
        sampling_rate=24000,
        verbosity=2
    )

    # Generate audio from prompt
    prompt = "Hello, this is a test of the OuteTTS text to speech system."
    audio_data = tts.audio_generation(prompt)

    # Save the generated audio
    tts._save_audio(audio_data, tts.sampling_rate, "tts_output/outetts")


if __name__ == "__main__":
    test_tts_generation_barkcpp()
    print('Bark TTS test completed')
    test_tts_generation_outetts()
    print("TTS generation test completed successfully!")
