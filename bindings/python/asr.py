"""
NexaAI ASR Example - Speech to Text (non-streaming)

This example demonstrates how to use the NexaAI SDK to transcribe an audio file.
"""

import argparse
import os

from nexaai.asr import ASR, ASRConfig

def main():
    parser = argparse.ArgumentParser(description="NexaAI ASR Example")
    parser.add_argument("--model",
                       default="NexaAI/parakeet-npu",
                       help="Model id or path")
    parser.add_argument("--audio",
                       required=True,
                       help="Path to the input audio file")
    parser.add_argument("--language", default="en",
                       help="Language code (e.g., en, zh). Empty for auto-detect if supported")
    parser.add_argument("--beam-size", type=int, default=5,
                       help="Beam size for decoding")
    parser.add_argument("--timestamps", default="segment",
                       help="Timestamps granularity: none|segment|word (if supported)")
    parser.add_argument("--plugin-id", default="npu", help="Plugin ID to use")
    parser.add_argument("--device", default="npu", help="Device to run on (e.g., cpu, gpu, 0)")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    audio_path = os.path.expanduser(args.audio)

    if not os.path.exists(audio_path):
        raise FileNotFoundError(f"Audio file not found: {audio_path}")

    asr = ASR.from_(name_or_path=model_path, plugin_id=args.plugin_id, device_id=args.device)

    cfg = ASRConfig(timestamps=args.timestamps, beam_size=args.beam_size, stream=False)
    result = asr.transcribe(audio_path=audio_path, language=args.language, config=cfg)
    print(result.transcript)


if __name__ == "__main__":
    main()


