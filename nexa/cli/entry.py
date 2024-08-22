import argparse
import uvicorn

def run_onnx_inference(args):
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    if args.onnx_command == "gen-text":
        from nexa.onnx.nexa_inference_text import \
            NexaTextInference as NexaTextOnnxInference
        inference = NexaTextOnnxInference(model_path, **kwargs)
    elif args.onnx_command == "gen-image":
        from nexa.onnx.nexa_inference_image import \
            NexaImageInference as NexaImageOnnxInference
        inference = NexaImageOnnxInference(model_path, **kwargs)
    elif args.onnx_command == "asr":
        from nexa.onnx.nexa_inference_voice import \
            NexaVoiceInference as NexaVoiceOnnxInference
        inference = NexaVoiceOnnxInference(model_path, **kwargs)
    elif args.onnx_command == "tts":
        from nexa.onnx.nexa_inference_tts import \
            NexaTTSInference as NexaTTSOnnxInference
        inference = NexaTTSOnnxInference(model_path, **kwargs)
    elif args.onnx_command == "server":
        from nexa.onnx.server.nexa_service import app as NexaOnnxServer
        uvicorn.run(NexaOnnxServer, host=args.host, port=args.port, reload=args.reload)
    else:
        raise ValueError(f"Unknown ONNX command: {args.onnx_command}")

    if hasattr(args, 'streamlit') and args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()

def run_ggml_inference(args):
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")

    if args.command == "server":
        from nexa.gguf.server.nexa_service import run_nexa_ai_service as NexaServer
        NexaServer(model_path, **kwargs)
        return

    stop_words = kwargs.pop("stop_words", [])

    if args.command == "gen-text":
        from nexa.gguf.nexa_inference_text import NexaTextInference
        inference = NexaTextInference(model_path, stop_words=stop_words, **kwargs)
    elif args.command == "gen-image":
        from nexa.gguf.nexa_inference_image import NexaImageInference
        inference = NexaImageInference(model_path, **kwargs)
        if hasattr(args, 'streamlit') and args.streamlit:
            inference.run_streamlit(model_path)
        elif args.img2img:
            inference.run_img2img()
        else:
            inference.run_txt2img()
        return
    elif args.command == "vlm":
        from nexa.gguf.nexa_inference_vlm import NexaVLMInference
        inference = NexaVLMInference(model_path, stop_words=stop_words, **kwargs)
    elif args.command == "asr":
        from nexa.gguf.nexa_inference_voice import NexaVoiceInference
        inference = NexaVoiceInference(model_path, **kwargs)
    else:
        raise ValueError(f"Unknown command: {args.command}")

    if hasattr(args, 'streamlit') and args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()

def main():
    parser = argparse.ArgumentParser(
        description="Nexa CLI tool for handling various model operations."
    )
    subparsers = parser.add_subparsers(dest="command", help="sub-command help")

    # ONNX subparsers
    onnx_parser = subparsers.add_parser("onnx", help="Run ONNX models for inference.")
    onnx_subparsers = onnx_parser.add_subparsers(dest="onnx_command", help="ONNX sub-command help")

    # ONNX Text Generation
    onnx_text_parser = onnx_subparsers.add_parser("gen-text", help="Run ONNX model for text generation.")
    onnx_text_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    onnx_text_parser.add_argument("-t", "--temperature", type=float, default=0.8, help="Temperature for sampling")
    onnx_text_parser.add_argument("-m", "--max_new_tokens", type=int, default=512, help="Maximum number of new tokens to generate")
    onnx_text_parser.add_argument("-k", "--top_k", type=int, default=50, help="Top-k sampling parameter")
    onnx_text_parser.add_argument("-p", "--top_p", type=float, default=1.0, help="Top-p sampling parameter")
    onnx_text_parser.add_argument("-sw", "--stop_words", nargs="*", default=[], help="List of stop words for early stopping")
    onnx_text_parser.add_argument("-pf", "--profiling", action="store_true", help="Enable profiling logs for the inference process")
    onnx_text_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")

    # ONNX Image Generation
    onnx_image_parser = onnx_subparsers.add_parser("gen-image", help="Run ONNX model for image generation.")
    onnx_image_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    onnx_image_parser.add_argument("-ns", "--num_inference_steps", type=int, default=20, help="Number of inference steps")
    onnx_image_parser.add_argument("-np", "--num_images_per_prompt", type=int, default=1, help="Number of images to generate per prompt")
    onnx_image_parser.add_argument("-H", "--height", type=int, default=512, help="Height of the output image")
    onnx_image_parser.add_argument("-W", "--width", type=int, default=512, help="Width of the output image")
    onnx_image_parser.add_argument("-g", "--guidance_scale", type=float, default=7.5, help="Guidance scale for diffusion")
    onnx_image_parser.add_argument("-o", "--output", type=str, default="generated_images/image.png", help="Output path for the generated image")
    onnx_image_parser.add_argument("-s", "--random_seed", type=int, default=41, help="Random seed for image generation")
    onnx_image_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")

    # ONNX ASR parser
    onnx_asr_parser = onnx_subparsers.add_parser('asr', help='Run ONNX model for auto-speech-recognition.')
    onnx_asr_parser.add_argument("model_path", type=str, help="Path or identifier for the model in S3")
    onnx_asr_parser.add_argument("-o", "--output_dir", type=str, default="transcriptions", help="Output directory for transcriptions")
    onnx_asr_parser.add_argument("-r", "--sampling_rate", type=int, default=16000, help="Sampling rate for audio processing")
    onnx_asr_parser.add_argument("-st", "--streamlit", action='store_true', help="Run the inference in Streamlit UI")

    # ONNX voice-generation parser
    onnx_tts_parser = onnx_subparsers.add_parser('tts', help='Run ONNX model for text-to-speech generation.')
    onnx_tts_parser.add_argument("model_path", type=str, help="Path or identifier for the model in S3")
    onnx_tts_parser.add_argument("-o", "--output_dir", type=str, default="tts", help="Output directory for tts")
    onnx_tts_parser.add_argument("-r", "--sampling_rate", type=int, default=16000, help="Sampling rate for audio processing")
    onnx_tts_parser.add_argument("-st", "--streamlit", action='store_true', help="Run the inference in Streamlit UI")

    # ONNX server parser
    onnx_server_parser = onnx_subparsers.add_parser("server", help="Run the Nexa AI Text Generation Service")
    onnx_server_parser.add_argument("model_path", type=str, help="Path or identifier for the model in S3")
    onnx_server_parser.add_argument("--host", type=str, default="0.0.0.0", help="Host to bind the server to")
    onnx_server_parser.add_argument("--port", type=int, default=8000, help="Port to bind the server to")
    onnx_server_parser.add_argument("--reload", action="store_true", help="Enable automatic reloading on code changes")

    # GGML Text Generation
    gen_text_parser = subparsers.add_parser("gen-text", help="Run a GGUF model locally for text generation.")
    gen_text_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    gen_text_parser.add_argument("-t", "--temperature", type=float, default=0.8, help="Temperature for sampling")
    gen_text_parser.add_argument("-m", "--max_new_tokens", type=int, default=512, help="Maximum number of new tokens to generate")
    gen_text_parser.add_argument("-k", "--top_k", type=int, default=50, help="Top-k sampling parameter")
    gen_text_parser.add_argument("-p", "--top_p", type=float, default=1.0, help="Top-p sampling parameter")
    gen_text_parser.add_argument("-sw", "--stop_words", nargs="*", default=[], help="List of stop words for early stopping")
    gen_text_parser.add_argument("-pf", "--profiling", action="store_true", help="Enable profiling logs for the inference process")
    gen_text_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")

    # GGML Image Generation
    gen_image_parser = subparsers.add_parser("gen-image", help="Run a GGUF model locally for image generation.")
    gen_image_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    gen_image_parser.add_argument("-i2i","--img2img",action="store_true",help="Whether to run image-to-image generation")
    gen_image_parser.add_argument("-ns", "--num_inference_steps", type=int, help="Number of inference steps")
    gen_image_parser.add_argument("-np", "--num_images_per_prompt", type=int, default=1, help="Number of images to generate per prompt")
    gen_image_parser.add_argument("-H", "--height", type=int, help="Height of the output image")
    gen_image_parser.add_argument("-W", "--width", type=int, help="Width of the output image")
    gen_image_parser.add_argument("-g", "--guidance_scale", type=float, help="Guidance scale for diffusion")
    gen_image_parser.add_argument("-o", "--output", type=str, default="generated_images/image.png", help="Output path for the generated image")
    gen_image_parser.add_argument("-s", "--random_seed", type=int, help="Random seed for image generation")
    gen_image_parser.add_argument("--lora_dir", type=str, help="Path to directory containing LoRA files")
    gen_image_parser.add_argument("--wtype", type=str, help="weight type (f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0). If not specified, the default is the type of the weight file.")
    gen_image_parser.add_argument("--control_net_path", type=str, help="Path to control net model")
    gen_image_parser.add_argument("--control_image_path", type=str, help="Path to image condition for Control Net")
    gen_image_parser.add_argument("--control_strength", type=str, help="Strength to apply Control Net")
    gen_image_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")

    # GGML VLM Inference
    vlm_parser = subparsers.add_parser("vlm", help="Run a GGUF model locally for VLM inference.")
    vlm_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    vlm_parser.add_argument("-t", "--temperature", type=float, default=0.8, help="Temperature for sampling")
    vlm_parser.add_argument("-m", "--max_new_tokens", type=int, default=2048, help="Maximum number of new tokens to generate")
    vlm_parser.add_argument("-k", "--top_k", type=int, default=50, help="Top-k sampling parameter")
    vlm_parser.add_argument("-p", "--top_p", type=float, default=1.0, help="Top-p sampling parameter")
    vlm_parser.add_argument("-sw", "--stop_words", nargs="*", default=[], help="List of stop words for early stopping")
    vlm_parser.add_argument("-pf", "--profiling", action="store_true", help="Enable profiling logs for the inference process")
    vlm_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")

    # Voice Inference
    asr_parser = subparsers.add_parser("asr", help="Run a GGUF model locally for voice inference.")
    asr_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    asr_parser.add_argument("-o", "--output_dir", type=str, default="transcriptions", help="Output directory for transcriptions")
    asr_parser.add_argument("-b", "--beam_size", type=int, default=5, help="Beam size to use for transcription")
    asr_parser.add_argument("-l", "--language", type=str, default=None, help="The language spoken in the audio. It should be a language code such as 'en' or 'fr'.")
    asr_parser.add_argument("--task", type=str, default="transcribe", help="Task to execute (transcribe or translate)")
    asr_parser.add_argument("-t", "--temperature", type=float, default=0.0, help="Temperature for sampling")
    asr_parser.add_argument("-c", "--compute_type", type=str, default="default", help="Type to use for computation (e.g., float16, int8, int8_float16)")
    asr_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")

    # GGML server parser
    server_parser = subparsers.add_parser("server", help="Run the Nexa AI Text Generation Service")
    server_parser.add_argument("model_path", type=str, help="Path or identifier for the model in S3")
    server_parser.add_argument("--host", type=str, default="0.0.0.0", help="Host to bind the server to")
    server_parser.add_argument("--port", type=int, default=8000, help="Port to bind the server to")
    server_parser.add_argument("--reload", action="store_true", help="Enable automatic reloading on code changes")

    # GGML general
    subparsers.add_parser("pull", help="Pull a model from official or hub.").add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    subparsers.add_parser("remove", help="Remove a model from local machine.").add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    subparsers.add_parser("list", help="List all models in the local machine.")
    subparsers.add_parser("login", help="Login to Nexa API.")
    subparsers.add_parser("whoami", help="Show current user information.")
    subparsers.add_parser("logout", help="Logout from Nexa API.")

    args = parser.parse_args()

    if args.command == "onnx":
        run_onnx_inference(args)
    elif args.command in ["gen-text", "gen-image", "vlm", "asr", "server"]:
        run_ggml_inference(args)
    elif args.command == "pull":
        from nexa.general import pull_model
        pull_model(args.model_path)
    elif args.command == "remove":
        from nexa.general import remove_model
        remove_model(args.model_path)
    elif args.command == "list":
        from nexa.general import list_models
        list_models()
    elif args.command == "login":
        from nexa.general import login
        login()
    elif args.command == "logout":
        from nexa.general import logout
        logout()
    elif args.command == "whoami":
        from nexa.general import whoami
        whoami()
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
