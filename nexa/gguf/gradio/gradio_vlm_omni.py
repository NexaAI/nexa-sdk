import sys
import os
import gradio as gr
from nexa.gguf.nexa_inference_vlm_omni import NexaOmniVlmInference
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model

if len(sys.argv) < 4:
    print("Usage: python gradio_vlm_omni.py <model_path> <is_local_path> <hf> [<projector_local_path>]")
    sys.exit(1)

model_path = sys.argv[1]
is_local_path = (sys.argv[2] == "True")
hf = (sys.argv[3] == "True")
projector_local_path = sys.argv[4] if len(sys.argv) > 4 else None

def load_model(model_path, is_local, huggingface, projector_path):
    if is_local:
        local_path = os.path.abspath(model_path)
    elif huggingface:
        local_path, _ = pull_model(model_path, hf=True)
    else:
        local_path, _ = pull_model(model_path)

    if is_local and projector_path:
        nexa_model = NexaOmniVlmInference(
            model_path=model_path,
            local_path=local_path,
            projector_local_path=projector_path
        )
    else:
        nexa_model = NexaOmniVlmInference(
            model_path=model_path,
            local_path=local_path
        )
    return nexa_model

try:
    nexa_model = load_model(model_path, is_local_path, hf, projector_local_path)
except Exception as e:
    print(f"Failed to load model: {e}")
    nexa_model = None

def process_image_fn(image_file, prompt=""):
    if not nexa_model:
        return "No model loaded. Please load a model first."
    if not image_file:
        return "Please provide an image input."

    try:
        with suppress_stdout_stderr():
            response = nexa_model.inference(prompt, image_file)
        return response
    except Exception as e:
        return f"Error during image processing: {e}"

with gr.Blocks() as demo: 
    gr.HTML(
        f"""
        <div style="display: flex; align-items: center; margin-bottom: 5px; padding-top: 10px;">
            <h1 style="font-family: Arial, sans-serif; font-size: 2.5em; font-weight: bold; margin: 0; padding-bottom: 5px;">
                Nexa AI Omni VLM Generation
            </h1>
            <a href='https://github.com/NexaAI/nexa-sdk' style='text-decoration: none; margin-left: 10px;'>
                <img src='https://img.shields.io/badge/SDK-Nexa-blue' alt='Nexa SDK' style='vertical-align: middle;'>
            </a>
        </div>
        <div style="font-family: Arial, sans-serif; font-size: 1em; color: #444; margin-bottom: 0.5em;">
            <b>Powered by Nexa AI SDK🐙</b> <br>
            <b>Model path: {model_path}</b>
        </div>
        """
    )

    gr.HTML("<h3 style='font-family: Arial, sans-serif; font-size: 1.1em; font-weight: bold;'>Enter your text input:</h3>")
    prompt_textbox = gr.Textbox(
        placeholder="e.g., Describe the image content or ask a question about it.", 
        lines=1,
        label="Prompt Box"
    )

    gr.HTML("<h2 style='font-family: Arial, sans-serif; font-size: 1.5em; font-weight: bold;'>Upload an image</h2>")
    with gr.Row():
        uploaded_image = gr.Image(type="filepath", label="Upload an image (png, jpg, jpeg) / Webcam / Paste")
        upload_response = gr.Textbox(label="Model Response", interactive=False)

    process_upload_button = gr.Button("Send")
    process_upload_button.click(
        process_image_fn, 
        inputs=[uploaded_image, prompt_textbox], 
        outputs=upload_response
    )

if __name__ == "__main__":
    demo.launch(server_name="0.0.0.0", server_port=7860, share=True, inbrowser=True)
