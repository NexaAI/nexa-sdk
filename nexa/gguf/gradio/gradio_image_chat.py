import sys
import os
import gradio as gr
from PIL import Image
from nexa.gguf.nexa_inference_image import NexaImageInference
from nexa.general import pull_model
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
import io

if len(sys.argv) < 4:
    print("Usage: python gradio_image_chat.py <model_path> <is_local_path> <hf>")
    sys.exit(1)

model_path = sys.argv[1]
is_local_path = (sys.argv[2] == "True")
hf = (sys.argv[3] == "True")

def load_model(model_path, is_local, huggingface):
    if is_local:
        local_path = os.path.abspath(model_path)
    elif huggingface:
        local_path, _ = pull_model(model_path, hf=True)
    else:
        local_path, _ = pull_model(model_path)

    nexa_model = NexaImageInference(
        model_path=model_path,
        local_path=local_path
    )
    return nexa_model

try:
    nexa_model = load_model(model_path, is_local_path, hf)
except Exception as e:
    print(f"Failed to load model: {e}")
    nexa_model = None


def process_txt2img(prompt, negative_prompt, steps, width, height, guidance, random_seed):
    if not nexa_model:
        return [None, "No model loaded. Please check logs or model path."]
    if not prompt:
        return [None, "Please provide a prompt."]

    nexa_model.params["num_inference_steps"] = steps
    nexa_model.params["width"] = width
    nexa_model.params["height"] = height
    nexa_model.params["guidance_scale"] = guidance
    nexa_model.params["random_seed"] = random_seed

    try:
        with suppress_stdout_stderr():
            images = nexa_model.txt2img(
                prompt=prompt,
                negative_prompt=negative_prompt,
                cfg_scale=guidance,
                width=width,
                height=height,
                sample_steps=steps,
                seed=random_seed
            )
        if not images:
            return [None, "No images generated."]
        img = images[0]
        return [img, "Image generated successfully!"]
    except Exception as e:
        return [None, f"Error during text-to-image generation: {e}"]

with gr.Blocks() as demo:
    gr.HTML(
        f"""
        <div style="display: flex; align-items: center; margin-bottom: 5px; padding-top: 10px;">
            <h1 style="font-family: Arial, sans-serif; font-size: 2.5em; font-weight: bold; margin: 0; padding-bottom: 5px;">
                Nexa AI Image Generation
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

    with gr.Row():
        prompt_text = gr.Textbox(
            label="Enter your prompt:",
            placeholder="e.g. A photograph of an astronaut riding a horse"
        )
        negative_prompt_text = gr.Textbox(
            label="Enter your negative prompt (optional):",
            placeholder="(optional) e.g. blur, low quality, ugly, mutated"
        )

    with gr.Accordion("Generation Parameters", open=False):
        steps_slider = gr.Slider(1, 8, value=4, step=1, label="Inference Steps")
        width_slider = gr.Slider(64, 1024, value=512, step=1, label="Width")
        height_slider = gr.Slider(64, 1024, value=512, step=1, label="Height")
        guidance_slider = gr.Slider(0.00, 2.00, value=1.00, step=0.01, label="Guidance Scale")
        random_seed_slider = gr.Slider(0, 10000, value=0, step=1, label="Random Seed")

    with gr.Row():
        generate_btn = gr.Button("Generate Image")

    output_image = gr.Image(label="Generated Image", type="pil", interactive=False)
    status_text = gr.Textbox(label="Status / Logs", interactive=False)

    generate_btn.click(
        fn=process_txt2img,
        inputs=[
            prompt_text, 
            negative_prompt_text,
            steps_slider,
            width_slider,
            height_slider,
            guidance_slider,
            random_seed_slider
        ],
        outputs=[output_image, status_text]
    )

if __name__ == "__main__":
    demo.launch(server_name="0.0.0.0", server_port=7860, share=True, inbrowser=True)
