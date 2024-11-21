import argparse
import os
from nexa import __version__
from nexa.constants import ModelType
import json


def _choose_files(local_path):
    """ Helper function for Multimodal inference only: select the model and projector ggufs from the local_path. """
    print(f"Files in {local_path}:")
    files = os.listdir(local_path)
    for i, file in enumerate(files):
        print(f"{i+1}. {file}")
    
    while True:
        try:
            model_choice = int(input(">>> Enter the index of the model gguf: ")) - 1
            if 0 <= model_choice < len(files):
                break
            else:
                print("Invalid selection. Please enter a valid number.")
        except ValueError:
            print("Invalid input. Please enter a number.")
    
    while True:
        try:
            projector_choice = int(input(">>> Enter the index of the projector gguf: ")) - 1
            if 0 <= projector_choice < len(files):
                break
            else:
                print("Invalid selection. Please enter a valid number.")
        except ValueError:
            print("Invalid input. Please enter a number.")
    
    return os.path.join(local_path, files[model_choice]), os.path.join(local_path, files[projector_choice])

def run_ggml_inference(args):
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    is_local_path = kwargs.pop("local_path", False)
    model_type = kwargs.pop("model_type", None)
    hf = kwargs.pop('huggingface', False)
    ms = kwargs.pop('modelscope', False)
    
    run_type = None
    if model_type:
        run_type = ModelType[model_type].value
    elif is_local_path or hf or ms:
        run_type = ModelType["NLP"].value

    local_path = None
    projector_local_path = None
    
    if is_local_path or hf or ms:
        if is_local_path:
            local_path = os.path.abspath(model_path)
            model_path = local_path
            if run_type == "Multimodal" or run_type == "AudioLM":
                if not os.path.isdir(local_path):
                    print("Error: For Multimodal models with --local_path, the provided path must be a directory containing both model and projector ggufs.")
                    return
                
                model_path, projector_local_path = _choose_files(local_path)
                
                if not model_path or not projector_local_path:
                    return
                
                local_path = model_path
            elif run_type == "Audio":
                if not os.path.isdir(local_path):
                    print("Error: For Audio models with --local_path, the provided path must be a directory containing all related files.")
                    return
        else:  # hf or ms case
            # TODO: remove this after adding support for Multimodal model in CLI
            if run_type == "Multimodal" or run_type == "Audio" or run_type == "TTS":
                print("Running multimodal model or audio model or TTS model from Hugging Face is currently not supported in CLI mode. Please use SDK to run Multimodal model or Audio model or TTS model.")
                return
            from nexa.general import pull_model
            local_path, _ = pull_model(model_path, hf=hf, ms=ms, run_type=run_type)
    else: # Model Hub
        from nexa.general import pull_model
        local_path, run_type = pull_model(model_path)

    stop_words = kwargs.pop("stop_words", None)

    try:
        if (is_local_path or hf or ms) and not model_type:
            print("No model type specified. Running with default model type: NLP")
            print("You can specify a different model type using the -mt flag")
            
        if run_type == "NLP":
            from nexa.gguf.nexa_inference_text import NexaTextInference
            inference = NexaTextInference(model_path=model_path, local_path=local_path, stop_words=stop_words, **kwargs)
        elif run_type == "Computer Vision":
            from nexa.gguf.nexa_inference_image import NexaImageInference
            inference = NexaImageInference(model_path=model_path, local_path=local_path, **kwargs)
            if hasattr(args, 'streamlit') and args.streamlit:
                inference.run_streamlit(model_path, is_local_path=is_local_path, hf=hf)
            elif args.img2img:
                inference.run_img2img()
            else:
                inference.run_txt2img()
            return
        elif run_type == "Multimodal":
            if is_local_path:
                if "omni" in local_path:
                    from nexa.gguf.nexa_inference_vlm_omni import NexaOmniVlmInference
                    inference = NexaOmniVlmInference(model_path=model_path, local_path=local_path, projector_local_path=projector_local_path, **kwargs)
                else:
                    from nexa.gguf.nexa_inference_vlm import NexaVLMInference
                    inference = NexaVLMInference(model_path=model_path, local_path=local_path, projector_local_path=projector_local_path, stop_words=stop_words, **kwargs)
            else:
                if "omni" in local_path:
                    from nexa.gguf.nexa_inference_vlm_omni import NexaOmniVlmInference
                    inference = NexaOmniVlmInference(model_path=model_path, local_path=local_path, **kwargs)
                else:
                    from nexa.gguf.nexa_inference_vlm import NexaVLMInference
                    inference = NexaVLMInference(model_path=model_path, local_path=local_path, stop_words=stop_words, **kwargs)
        elif run_type == "Audio":
            from nexa.gguf.nexa_inference_voice import NexaVoiceInference
            inference = NexaVoiceInference(model_path=model_path, local_path=local_path, **kwargs)
        elif run_type == "TTS":
            # # Temporarily disabled since version v0.0.9.3
            raise NotImplementedError("TTS model is not supported in CLI mode.")
            # from nexa.gguf.nexa_inference_tts import NexaTTSInference
            # inference = NexaTTSInference(model_path=model_path, local_path=local_path, **kwargs)
        elif run_type == "AudioLM":
            from nexa.gguf.nexa_inference_audio_lm import NexaAudioLMInference
            inference = NexaAudioLMInference(model_path=model_path, local_path=local_path, **kwargs)
        else:
            print(f"Unknown task: {run_type}. Skipping inference.")
            return
        
    except Exception as e:
        print(f"Error running ggml inference: {e}")
        print(f"Please refer to our docs to install nexaai package: https://docs.nexaai.com/getting-started/installation ")
        return

    if hasattr(args, 'streamlit') and args.streamlit:
        if run_type == "Multimodal" or run_type == "AudioLM":
            inference.run_streamlit(model_path, is_local_path=is_local_path, hf=hf, projector_local_path=projector_local_path)
        else:
            inference.run_streamlit(model_path, is_local_path=is_local_path, hf=hf)
    else:
        inference.run()

def run_ggml_server(args):
    from nexa.gguf.server.nexa_service import run_nexa_ai_service as NexaServer
    
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    is_local_path = kwargs.pop("local_path", False)
    model_type = kwargs.pop("model_type", None)
    hf = kwargs.pop('huggingface', False)
    ms = kwargs.pop('modelscope', False)
    
    run_type = None
    if model_type:
        run_type = ModelType[model_type].value
    elif is_local_path or hf or ms:
        run_type = ModelType["NLP"].value

    projector_local_path = None
    if run_type == "Multimodal" and is_local_path:
        local_path = os.path.abspath(model_path)
        if not os.path.isdir(local_path):
            print("Error: For Multimodal models with --local_path, the provided path must be a directory.")
            return
        
        model_path, projector_local_path = _choose_files(local_path)
        
        if not model_path or not projector_local_path:
            return
    elif run_type == "Audio" and is_local_path:
        local_path = os.path.abspath(model_path)
        if not os.path.isdir(local_path):
            print("Error: For Audio models with --local_path, the provided path must be a directory containing all related files.")
            return

    if (is_local_path or hf or ms) and not model_type:
        print("No model type specified. Running with default model type: NLP")
        print("You can specify a different model type using the -mt flag")

    NexaServer(
        model_path_arg=model_path,
        is_local_path_arg=is_local_path,
        model_type_arg=run_type,
        huggingface=hf,
        modelscope=ms,
        projector_local_path_arg=projector_local_path,
        **kwargs
    )

def run_onnx_inference(args):
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    is_local_path = kwargs.pop("local_path", False)
    model_type = kwargs.pop("model_type", None)
    
    run_type = None
    if model_type:
        run_type = ModelType[model_type].value
    elif is_local_path:
        run_type = ModelType["NLP"].value

    local_path = None
    if is_local_path:
        local_path = os.path.abspath(model_path)
        if not os.path.isdir(local_path):
            print("Error: For ONNX models, the provided path must be a directory.")
            return
        model_path = local_path
    else:
        from nexa.general import pull_model
        local_path, run_type = pull_model(model_path)

    try:
        if is_local_path and not model_type:
            print("No model type specified. Running with default model type: NLP")
            print("You can specify a different model type using the -mt flag")
            
        if run_type == "NLP":
            from nexa.onnx.nexa_inference_text import NexaTextInference as NexaTextOnnxInference
            inference = NexaTextOnnxInference(model_path=model_path, local_path=local_path, **kwargs)
        elif run_type == "Computer Vision":
            from nexa.onnx.nexa_inference_image import NexaImageInference as NexaImageOnnxInference
            inference = NexaImageOnnxInference(model_path=model_path, local_path=local_path, **kwargs)
        elif run_type == "Audio":
            from nexa.onnx.nexa_inference_voice import NexaVoiceInference as NexaVoiceOnnxInference
            inference = NexaVoiceOnnxInference(model_path=model_path, local_path=local_path, **kwargs)
        elif run_type == "TTS":
            from nexa.onnx.nexa_inference_tts import NexaTTSInference as NexaTTSOnnxInference
            inference = NexaTTSOnnxInference(model_path=model_path, local_path=local_path, **kwargs)
        else:
            print(f"Unknown task: {run_type}. Skipping inference.")
            return
    except Exception as e:
        print(f"Error running onnx inference: {e}")
        print(f"Please refer to our docs to install nexaai[onnx] package: https://docs.nexaai.com/getting-started/installation ")
        return

    if hasattr(args, 'streamlit') and args.streamlit:
        inference.run_streamlit(model_path, is_local_path=is_local_path)
    else:
        inference.run()

def run_eval_tasks(args):
    try:
        if args.tasks and 'do-not-answer' in args.tasks:
            if not os.getenv('OPENAI_API_KEY'):
                print("Warning: The 'do-not-answer' task requires an OpenAI API key.")
                print("Please set your API key in the terminal using the following command:")
                print("export OPENAI_API_KEY=your_openai_api_key_here")
                print("After setting the key, please try again")
                return

        kwargs = {k: v for k, v in vars(args).items() if v is not None}
        model_path = kwargs.pop("model_path")
        
        from nexa.eval.nexa_eval import NexaEval
        evaluator = NexaEval(model_path, args.tasks, args.limit, args.nctx, args.num_workers)
        if not args.tasks:
            evaluator.run_perf_eval(args.device, args.new_tokens)
        else:
            evaluator.run_evaluation()
    except Exception as e:
        print(f"Error running evaluation: {e}")
        print("Please run: pip install 'nexaai[eval]'")
        return

def run_embedding_generation(args):
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    prompt = kwargs.pop("prompt")
    is_local_path = kwargs.pop("local_path", False)
    hf = kwargs.pop('huggingface', False)
    ms = kwargs.pop('modelscope', False)
    normalize = kwargs.pop('normalize', False)
    no_truncate = kwargs.pop('no_truncate', False)

    local_path = None
    if is_local_path or hf or ms:
        if is_local_path:
            local_path = os.path.abspath(model_path)
            model_path = local_path
        else:  # hf or ms case
            from nexa.general import pull_model
            local_path, _ = pull_model(model_path, hf=hf, ms=ms, run_type="Text Embedding")
    else:  # Model Hub
        from nexa.general import pull_model
        local_path, _ = pull_model(model_path)

    try:
        from nexa.gguf.nexa_inference_text import NexaTextInference
        inference = NexaTextInference(model_path=model_path, local_path=local_path, embedding=True)
        embedding = inference.create_embedding(prompt, normalize=normalize, truncate=not no_truncate)
        print({"embedding": embedding})
    except Exception as e:
        print(f"Error generating embedding: {e}")
        print("Please refer to our docs to install nexaai package: https://docs.nexaai.com/getting-started/installation")

def _select_model_type():
    """Helper function to get model type selection from user."""
    print("\nSelect model type:")
    print("1. NLP (text generation)")
    print("2. COMPUTER_VISION (image generation)")
    
    while True:
        try:
            choice = int(input("\nSelect model type (enter number): "))
            if choice == 1:
                return "NLP"
            elif choice == 2:
                return "COMPUTER_VISION"
            print("Invalid selection. Please try again.")
        except ValueError:
            print("Please enter a valid number.")

def _select_quantization_type():
    """Helper function to get quantization type selection from user."""
    from nexa.gguf.converter.constants import LLAMA_QUANTIZATION_TYPES
    print("\nAvailable quantization types:")
    for i, qt in enumerate(LLAMA_QUANTIZATION_TYPES.keys(), 1):
        print(f"{i}. {qt}")
    
    while True:
        try:
            choice = int(input("\nSelect quantization type (enter number): ")) - 1
            if 0 <= choice < len(LLAMA_QUANTIZATION_TYPES):
                return list(LLAMA_QUANTIZATION_TYPES.keys())[choice]
            print("Invalid selection. Please try again.")
        except ValueError:
            print("Please enter a valid number.")

def _store_in_nexa_list(converted_path, model_type):
    """Helper function to store converted model in nexa list."""
    import shutil
    from nexa.general import add_model_to_list
    from nexa.gguf.converter.constants import NEXA_MODELS_HUB_CONVERTED_DIR
    
    # Create the converted directory if it doesn't exist
    os.makedirs(NEXA_MODELS_HUB_CONVERTED_DIR, exist_ok=True)
    
    # Copy the file to the converted directory
    nexa_list_path = os.path.join(NEXA_MODELS_HUB_CONVERTED_DIR, os.path.basename(converted_path))
    shutil.copy2(converted_path, nexa_list_path)
    
    # Add the new path to the model list
    add_model_to_list(os.path.basename(converted_path), nexa_list_path, "gguf", model_type)

def _run_converted_model(converted_path, model_type):
    """Helper function to run the converted model."""
    try:
        import subprocess
        command = f"nexa run {converted_path} -lp -mt {model_type}"
        print(f"Running command: {command}")
        subprocess.run(command.split(), check=True, text=True)
    except subprocess.CalledProcessError as e:
        print("Error running the converted model.")
        print("Change model type with -mt to run the model correctly. Or refer to our docs: https://docs.nexa.ai/sdk/cli-reference")

def run_convert(args):
    # Get model type and quantization type
    model_type = _select_model_type()
    ftype = args.ftype or _select_quantization_type()
    
    input_path = args.input_path
    
    # Check input path conditions
    if os.path.isdir(input_path):
        # Valid directory, proceed as is
        pass
    elif os.path.isfile(input_path) and input_path.endswith('.gguf'):
        # Valid GGUF file, proceed as is
        pass
    else:
        # Try downloading from HF or MS if path isn't a valid local directory/file
        if args.modelscope:
            from nexa.general import download_repo_from_ms
            success, local_path = download_repo_from_ms(input_path)
        else:
            from nexa.general import download_repo_from_hf
            success, local_path = download_repo_from_hf(input_path)
        
        if success:
            input_path = local_path
        else:
            print("Error: Input path must be either a directory, a .gguf file, or a valid Hugging Face model identifier.")
            return

    # Input_path here should be a valid directory
    kwargs = {k: v for k, v in vars(args).items() if v is not None and k not in ['input_path', 'ftype', 'output_file', 'convert_type']}
    
    try:
        from nexa.gguf.converter.nexa_convert import convert_hf_to_quantized_gguf
        converted_path = convert_hf_to_quantized_gguf(
            input_path,
            output_file=args.output_file,
            ftype=ftype,
            convert_type=args.convert_type,
            **kwargs
        )
        if converted_path:
            print(f"\nConversion completed successfully. Output file: {converted_path}")
            
            # Ask if user wants to store in nexa list
            store_choice = input("\nWould you like to store this model in nexa list so you can run it with `nexa run <model_name>` anywhere and anytime? (y/N): ").strip().lower()
            if store_choice == 'y':
                _store_in_nexa_list(converted_path, model_type)
            
            # Ask if user wants to run the model
            run_choice = input("\nWould you like to run the converted model? (y/N): ").strip().lower()
            if run_choice == 'y':
                _run_converted_model(converted_path, model_type)
            else:
                print("Exiting without running the model.")
            
            print(f"\nConverted model stored at {converted_path}")
            running_command = f"nexa run {converted_path.split('/')[-1]}"\
                if store_choice == 'y' else f"nexa run {converted_path} -lp -mt {model_type}"
            print(f"\nYou can run the converted model with command: {running_command}")
        else:
            print("Conversion failed.")
    except Exception as e:
        print(f"Error during conversion: {e}")

def main():
    parser = argparse.ArgumentParser(description="Nexa CLI tool for handling various model operations.")
    parser.add_argument("-V", "--version", action="version", version=__version__, help="Show the version of the Nexa SDK.")
    subparsers = parser.add_subparsers(dest="command", help="sub-command help")

    # Run command
    run_parser = subparsers.add_parser("run", help="Run inference for various tasks using GGUF models.")
    run_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    run_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")
    run_parser.add_argument("-pf", "--profiling", action="store_true", help="Enable profiling logs for the inference process")
    run_parser.add_argument("-lp", "--local_path", action="store_true", help="Indicate that the model path provided is the local path")
    run_parser.add_argument("-mt", "--model_type", type=str, choices=[e.name for e in ModelType], help="Indicate the model running type (default: NLP)")
    run_parser.add_argument("-hf", "--huggingface", action="store_true", help="Load model from Hugging Face Hub")
    run_parser.add_argument("-ms", "--modelscope", action="store_true", help="Load model from ModelScope Hub")

    # Text generation/vlm arguments
    text_group = run_parser.add_argument_group('Text generation/VLM options')
    text_group.add_argument("-t", "--temperature", type=float, help="Temperature for sampling")
    text_group.add_argument("-m", "--max_new_tokens", type=int, help="Maximum number of new tokens to generate")
    text_group.add_argument("-k", "--top_k", type=int, help="Top-k sampling parameter")
    text_group.add_argument("-p", "--top_p", type=float, help="Top-p sampling parameter")
    text_group.add_argument("-sw", "--stop_words", nargs="*", help="List of stop words for early stopping")
    text_group.add_argument("--lora_path", type=str, help="Path to a LoRA file to apply to the model.")
    text_group.add_argument("--nctx", type=int, default=2048, help="Maximum context length of the model you're using")

    # Image generation arguments
    image_group = run_parser.add_argument_group('Image generation options')
    image_group.add_argument("-i2i", "--img2img", action="store_true", help="Whether to run image-to-image generation")
    image_group.add_argument("-ns", "--num_inference_steps", type=int, help="Number of inference steps")
    image_group.add_argument("-H", "--height", type=int, help="Height of the output image")
    image_group.add_argument("-W", "--width", type=int, help="Width of the output image")
    image_group.add_argument("-g", "--guidance_scale", type=float, help="Guidance scale for diffusion")
    image_group.add_argument("-o", "--output", type=str, default="generated_images/image.png", help="Output path for the generated image")
    image_group.add_argument("-s", "--random_seed", type=int, help="Random seed for image generation")
    image_group.add_argument("--lora_dir", type=str, help="Path to directory containing LoRA files")
    image_group.add_argument("--wtype", type=str, help="Weight type (f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)")
    image_group.add_argument("--control_net_path", type=str, help="Path to control net model")
    image_group.add_argument("--control_image_path", type=str, help="Path to image condition for Control Net")
    image_group.add_argument("--control_strength", type=float, help="Strength to apply Control Net")

    # ASR arguments
    asr_group = run_parser.add_argument_group('Automatic Speech Recognition options')
    asr_group.add_argument("--beam_size", type=int, help="Beam size to use for transcription")
    asr_group.add_argument("--language", type=str, help="Language code for audio (e.g., 'en' or 'fr')")
    asr_group.add_argument("--task", type=str, help="Task to execute (transcribe or translate)")
    asr_group.add_argument("--compute_type", type=str, help="Type to use for computation")

    # TTS arguments
    tts_group = run_parser.add_argument_group('Text-to-Speech options')
    tts_group.add_argument("--output_dir", type=str, default="tts", help="Output directory for tts")
    tts_group.add_argument("--sampling_rate", type=int, default=24000, help="Sampling rate for audio processing")
    tts_group.add_argument("--n_threads", type=int, default=1, help="Number of threads to use for processing")
    tts_group.add_argument("--seed", type=int, default=0, help="Seed for random number generation")
    tts_group.add_argument("--verbosity", type=int, default=1, help="Verbosity level for the Bark model")

    # ONNX command
    onnx_parser = subparsers.add_parser("onnx", help="Run inference for various tasks using ONNX models.")
    onnx_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    onnx_parser.add_argument("-st", "--streamlit", action="store_true", help="Run the inference in Streamlit UI")
    onnx_parser.add_argument("-lp", "--local_path", action="store_true", help="Indicate that the model path provided is the local path")
    onnx_parser.add_argument("-mt", "--model_type", type=str, choices=[e.name for e in ModelType], help="Indicate the model running type")

    # ONNX Text generation arguments
    onnx_text_group = onnx_parser.add_argument_group('Text generation options')
    onnx_text_group.add_argument("-t", "--temperature", type=float, help="Temperature for sampling")
    onnx_text_group.add_argument("-m", "--max_new_tokens", type=int, help="Maximum number of new tokens to generate")
    onnx_text_group.add_argument("-k", "--top_k", type=int, help="Top-k sampling parameter")
    onnx_text_group.add_argument("-p", "--top_p", type=float, help="Top-p sampling parameter")
    onnx_text_group.add_argument("-sw", "--stop_words", nargs="*", help="List of stop words for early stopping")
    onnx_text_group.add_argument("-pf", "--profiling", action="store_true", help="Enable profiling logs for the inference process")

    # ONNX Image generation arguments
    onnx_image_group = onnx_parser.add_argument_group('Image generation options')
    onnx_image_group.add_argument("-ns", "--num_inference_steps", type=int, help="Number of inference steps")
    onnx_image_group.add_argument("-np", "--num_images_per_prompt", type=int, help="Number of images to generate per prompt")
    onnx_image_group.add_argument("-H", "--height", type=int, help="Height of the output image")
    onnx_image_group.add_argument("-W", "--width", type=int, help="Width of the output image")
    onnx_image_group.add_argument("-g", "--guidance_scale", type=float, help="Guidance scale for diffusion")
    onnx_image_group.add_argument("-O", "--output", type=str, help="Output path for the generated image")
    onnx_image_group.add_argument("-s", "--random_seed", type=int, help="Random seed for image generation")

    # ONNX Voice arguments
    onnx_voice_group = onnx_parser.add_argument_group('Voice generation options')
    onnx_voice_group.add_argument("-o", "--output_dir", type=str, default="voice_output", help="Output directory for audio processing")
    onnx_voice_group.add_argument("-r", "--sampling_rate", type=int, default=16000, help="Sampling rate for audio processing")
    
    # Embed command
    embed_parser = subparsers.add_parser("embed", help="Generate embeddings for a given prompt.")
    embed_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    embed_parser.add_argument("prompt", type=str, help="The prompt to generate an embedding for")
    embed_parser.add_argument("-lp", "--local_path", action="store_true", help="Indicate that the model path provided is the local path")
    embed_parser.add_argument("-hf", "--huggingface", action="store_true", help="Load model from Hugging Face Hub")
    embed_parser.add_argument("-ms", "--modelscope", action="store_true", help="Load model from ModelScope Hub")
    embed_parser.add_argument("-n", "--normalize", action="store_true", help="Normalize the embeddings")
    embed_parser.add_argument("-nt", "--no_truncate", action="store_true", help="Not truncate the embeddings")

    # Convert command
    convert_parser = subparsers.add_parser("convert", help="Convert and quantize a Hugging Face model to GGUF format.")
    convert_parser.add_argument("input_path", type=str, help="Path to the input Hugging Face model directory or GGUF file")
    convert_parser.add_argument("ftype", nargs='?', type=str, help="Quantization type")
    convert_parser.add_argument("output_file", nargs='?', type=str, help="Path to the output quantized GGUF file")    

    convert_hf_parser = convert_parser.add_argument_group('Convert from safetensors options')
    convert_hf_parser.add_argument("--convert_type", type=str, default="f16", help="Conversion type for safetensors to GGUF (default: f16)")
    convert_hf_parser.add_argument("--bigendian", action="store_true", help="Use big endian format")
    convert_hf_parser.add_argument("--use_temp_file", action="store_true", help="Use a temporary file during conversion")
    convert_hf_parser.add_argument("--no_lazy", action="store_true", help="Disable lazy loading")
    convert_hf_parser.add_argument("--metadata", type=json.loads, help="Additional metadata as JSON string")
    convert_hf_parser.add_argument("--split_max_tensors", type=int, default=0, help="Maximum number of tensors per split")
    convert_hf_parser.add_argument("--split_max_size", type=str, default="0", help="Maximum size per split")
    convert_hf_parser.add_argument("--no_tensor_first_split", action="store_true", help="Disable tensor-first splitting")
    convert_hf_parser.add_argument("--vocab_only", action="store_true", help="Only process vocabulary")
    convert_hf_parser.add_argument("--dry_run", action="store_true", help="Perform a dry run without actual conversion")
    convert_hf_parser.add_argument("-ms", "--modelscope", action="store_true", help="Download model from ModelScope Hub")

    quantization_parser = convert_parser.add_argument_group('Quantization options')
    quantization_parser.add_argument("--nthread", type=int, default=4, help="Number of threads to use (default: 4)")
    quantization_parser.add_argument("--output_tensor_type", type=str, help="Output tensor type")
    quantization_parser.add_argument("--token_embedding_type", type=str, help="Token embedding type")
    quantization_parser.add_argument("--allow_requantize", action="store_true", help="Allow quantizing non-f32/f16 tensors")
    quantization_parser.add_argument("--quantize_output_tensor", action="store_true", help="Quantize output.weight")
    quantization_parser.add_argument("--only_copy", action="store_true", help="Only copy tensors (ignores ftype, allow_requantize, and quantize_output_tensor)")
    quantization_parser.add_argument("--pure", action="store_true", help="Quantize all tensors to the default type")
    quantization_parser.add_argument("--keep_split", action="store_true", help="Quantize to the same number of shards")

    # GGML server parser
    server_parser = subparsers.add_parser("server", help="Run the Nexa AI Text Generation Service")
    server_parser.add_argument("model_path", type=str, nargs='?', help="Path or identifier for the model in Nexa Model Hub")
    server_parser.add_argument("-lp", "--local_path", action="store_true", help="Indicate that the model path provided is the local path")
    server_parser.add_argument("-mt", "--model_type", type=str, choices=[e.name for e in ModelType], help="Indicate the model running type, must be used with -lp, -hf or -ms")
    server_parser.add_argument("-hf", "--huggingface", action="store_true", help="Load model from Hugging Face Hub")
    server_parser.add_argument("-ms", "--modelscope", action="store_true", help="Load model from ModelScope Hub")
    server_parser.add_argument("--host", type=str, default="localhost", help="Host to bind the server to")
    server_parser.add_argument("--port", type=int, default=8000, help="Port to bind the server to")
    server_parser.add_argument("--reload", action="store_true", help="Enable automatic reloading on code changes")
    server_parser.add_argument("--nctx", type=int, default=2048, help="Maximum context length of the model you're using")

    # Other commands
    pull_parser = subparsers.add_parser("pull", help="Pull a model from official or hub.")
    pull_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")
    pull_parser.add_argument("-hf", "--huggingface", action="store_true", help="Pull model from Hugging Face Hub")
    pull_parser.add_argument("-ms", "--modelscope", action="store_true", help="Pull model from ModelScope Hub")
    pull_parser.add_argument("-o", "--output_path", type=str, help="Custom output path for the pulled model")

    remove_parser = subparsers.add_parser("remove", help="Remove a model from local machine.")
    remove_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")

    subparsers.add_parser("clean", help="Clean up all model files.")
    subparsers.add_parser("list", help="List all models in the local machine.")
    subparsers.add_parser("login", help="Login to Nexa API.")
    subparsers.add_parser("whoami", help="Show current user information.")
    subparsers.add_parser("logout", help="Logout from Nexa API.")

    # Benchmark Evaluation
    eval_parser = subparsers.add_parser("eval", help="Evaluate models on specified tasks.")
    eval_parser.add_argument("model_path", type=str, help="Path or identifier for the model in Nexa Model Hub")

    # General evaluation options
    general_eval_group = eval_parser.add_argument_group('General evaluation options')
    general_eval_group.add_argument("--tasks", type=str, help="Tasks to evaluate the model on, separated by commas.")
    general_eval_group.add_argument("--limit", type=float, help="Limit the number of examples per task. If <1, limit is a percentage of the total number of examples.", default=None)
    general_eval_group.add_argument("--num_workers", type=int, help="Number of workers to use for evaluation", default=1)
    general_eval_group.add_argument("--nctx", type=int, help="Length of context window", default=4096)

    # Performance evaluation options
    perf_eval_group = eval_parser.add_argument_group('Performance evaluation options')
    perf_eval_group.add_argument("--device", type=str, help="Device to run performance evaluation on, choose from 'cpu', 'cuda', 'mps'", default="cpu")
    perf_eval_group.add_argument("--new_tokens", type=int, help="Number of new tokens to evaluate", default=100)

    args = parser.parse_args()

    if args.command == "run":
        if args.local_path and args.huggingface:
            print("Error: --local_path and --huggingface flags cannot be used together")
            return
        if args.local_path and args.modelscope:
            print("Error: --local_path and --modelscope flags cannot be used together")
            return
        if args.huggingface and args.modelscope:
            print("Error: --huggingface and --modelscope flags cannot be used together")
            return
        run_ggml_inference(args)
    elif args.command == "server":
        if args.local_path and args.huggingface:
            print("Error: --local_path and --huggingface flags cannot be used together")
            return
        if args.local_path and args.modelscope:
            print("Error: --local_path and --modelscope flags cannot be used together")
            return
        if args.huggingface and args.modelscope:
            print("Error: --huggingface and --modelscope flags cannot be used together")
            return
        run_ggml_server(args)
    elif args.command == "onnx":
        run_onnx_inference(args)
    elif args.command == "eval":
        run_eval_tasks(args)
    elif args.command == "embed":
        run_embedding_generation(args)
    elif args.command == "pull":
        from nexa.general import pull_model
        import os

        hf = getattr(args, 'huggingface', False)
        ms = getattr(args, 'modelscope', False)
        local_download_path = None
        
        if args.output_path:
            if not os.path.exists(args.output_path):
                os.makedirs(args.output_path, exist_ok=True)
                print(f"Created output directory: {args.output_path}")
            local_download_path = os.path.abspath(args.output_path)
        pull_model(args.model_path, hf, ms, local_download_path=local_download_path)
    elif args.command == "convert":
        run_convert(args)
    elif args.command == "remove":
        from nexa.general import remove_model
        remove_model(args.model_path)
    elif args.command == "clean":
        from nexa.general import clean
        clean()
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
