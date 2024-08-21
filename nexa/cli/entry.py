import argparse
from nexa.general import pull_model
import uvicorn

def run_inference(args):
    local_path, run_type = pull_model(args.model_path)
    kwargs = {k: v for k, v in vars(args).items() if v is not None}

    if run_type == "gen-text":
        from nexa.gguf.nexa_inference_text import NexaTextInference
        inference = NexaTextInference(local_path, **kwargs)
    elif run_type == "gen-image":
        from nexa.gguf.nexa_inference_image import NexaImageInference
        inference = NexaImageInference(local_path, **kwargs)
    elif run_type == "vlm":
        from nexa.gguf.nexa_inference_vlm import NexaVLMInference
        inference = NexaVLMInference(local_path, **kwargs)
    elif run_type == "asr":
        from nexa.gguf.nexa_inference_voice import NexaVoiceInference
        inference = NexaVoiceInference(local_path, **kwargs)
    else:
        raise ValueError(f"Unknown task: {run_type}")

    if args.streamlit:
        inference.run_streamlit(local_path)
    else:
        inference.run()

def run_onnx_inference(args):
    local_path, run_type = pull_model(args.model_path)
    kwargs = {k: v for k, v in vars(args).items() if v is not None}

    if run_type == "gen-text":
        from nexa.onnx.nexa_inference_text import \
            NexaTextInference as NexaTextOnnxInference
        inference = NexaTextOnnxInference(local_path, **kwargs)
    elif run_type == "gen-image":
        from nexa.onnx.nexa_inference_image import \
            NexaImageInference as NexaImageOnnxInference
        inference = NexaImageOnnxInference(local_path, **kwargs)
    elif run_type == "asr":
        from nexa.onnx.nexa_inference_voice import \
            NexaVoiceInference as NexaVoiceOnnxInference
        inference = NexaVoiceOnnxInference(local_path, **kwargs)
    elif run_type == "tts":
        from nexa.onnx.nexa_inference_tts import \
            NexaTTSInference as NexaTTSOnnxInference
        inference = NexaTTSOnnxInference(local_path, **kwargs)
    else:
        raise ValueError(f"Unknown ONNX command: {args.onnx_command}")

    if hasattr(args, 'streamlit') and args.streamlit:
        inference.run_streamlit(local_path)
    else:
        inference.run()

def main():
    parser = argparse.ArgumentParser(
        description="Nexa CLI tool for handling various model operations."
    )
    subparsers = parser.add_subparsers(dest="command", help="sub-command help")

    # Run command
    run_parser = subparsers.add_parser("run", help="Run inference for various tasks.")
    run_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    
    # Common arguments
    common_group = run_parser.add_argument_group('Common options')
    common_group.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")
    common_group.add_argument("-pf", "--profiling", action="store_true", help="Enable profiling logs for the inference process")

    # Text generation/vlm arguments
    text_group = run_parser.add_argument_group('Text generation/VLM options')
    text_group.add_argument("-t", "--temperature", type=float, help="Temperature for sampling")
    text_group.add_argument("-m", "--max_new_tokens", type=int, help="Maximum number of new tokens to generate")
    text_group.add_argument("-k", "--top_k", type=int, help="Top-k sampling parameter")
    text_group.add_argument("-p", "--top_p", type=float, help="Top-p sampling parameter")
    text_group.add_argument("-sw", "--stop_words", nargs="*", help="List of stop words for early stopping")

    # Image generation arguments
    image_group = run_parser.add_argument_group('Image generation options')
    image_group.add_argument("-i2i", "--img2img", action="store_true", help="Whether to run image-to-image generation")
    image_group.add_argument("-ns", "--num_inference_steps", type=int, help="Number of inference steps")
    image_group.add_argument("-np", "--num_images_per_prompt", type=int, help="Number of images to generate per prompt")
    image_group.add_argument("-H", "--height", type=int, help="Height of the output image")
    image_group.add_argument("-W", "--width", type=int, help="Width of the output image")
    image_group.add_argument("-g", "--guidance_scale", type=float, help="Guidance scale for diffusion")
    image_group.add_argument("-o", "--output", type=str, help="Output path for the generated image")
    image_group.add_argument("-s", "--random_seed", type=int, help="Random seed for image generation")
    image_group.add_argument("--lora_dir", type=str, help="Path to directory containing LoRA files")
    image_group.add_argument("--wtype", type=str, help="Weight type (f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)")
    image_group.add_argument("--control_net_path", type=str, help="Path to control net model")
    image_group.add_argument("--control_image_path", type=str, help="Path to image condition for Control Net")
    image_group.add_argument("--control_strength", type=str, help="Strength to apply Control Net")

    # ASR arguments
    asr_group = run_parser.add_argument_group('Automatic Speech Recognition options')
    asr_group.add_argument("-b", "--beam_size", type=int, help="Beam size to use for transcription")
    asr_group.add_argument("-l", "--language", type=str, help="Language code for audio (e.g., 'en' or 'fr')")
    asr_group.add_argument("--task", type=str, help="Task to execute (transcribe or translate)")
    asr_group.add_argument("-c", "--compute_type", type=str, help="Type to use for computation")

    # Other commands (unchanged)
    subparsers.add_parser("pull", help="Pull a model from official or hub.").add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    subparsers.add_parser("remove", help="Remove a model from local machine.").add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    subparsers.add_parser("list", help="List all models in the local machine.")
    subparsers.add_parser("login", help="Login to Nexa API.")
    subparsers.add_parser("whoami", help="Show current user information.")
    subparsers.add_parser("logout", help="Logout from Nexa API.")

    # GGML server parser
    server_parser = subparsers.add_parser("server", help="Run the Nexa AI Text Generation Service")
    server_parser.add_argument("model_path", type=str, help="Path or identifier for the model in S3")
    server_parser.add_argument("--host", type=str, default="0.0.0.0", help="Host to bind the server to")
    server_parser.add_argument("--port", type=int, default=8000, help="Port to bind the server to")
    server_parser.add_argument("--reload", action="store_true", help="Enable automatic reloading on code changes")

    args = parser.parse_args()

    if args.command == "run":
        run_inference(args)
    elif args.command == "onnx":
        run_onnx_inference(args)
    elif args.command == "serve":
        from nexa.gguf.server.nexa_service import run_nexa_ai_service as NexaServer
        kwargs = {k: v for k, v in vars(args).items() if v is not None}
        NexaServer(args.model_path, **kwargs)
        return
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