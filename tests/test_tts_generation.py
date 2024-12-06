# Temporarily disabled since version v0.0.9.3

# from nexa.gguf import NexaTTSInference

# def test_tts_generation():
#     tts = NexaTTSInference(
#         model_path="bark-small",
#         local_path=None,
#         n_threads=4,
#         seed=42,
#         sampling_rate=24000,
#         verbosity=2
#     )
    
#     # Generate audio from prompt
#     prompt = "Hello, this is a test of the Bark text to speech system."
#     audio_data = tts.audio_generation(prompt)
    
#     # Save the generated audio
#     tts._save_audio(audio_data, tts.sampling_rate, "tts_output")
#     print("TTS generation test completed successfully!")

# if __name__ == "__main__":
#     test_tts_generation()