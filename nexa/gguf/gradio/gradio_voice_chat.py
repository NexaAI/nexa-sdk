import sys
import os
import gradio as gr
from nexa.gguf.nexa_inference_voice import NexaVoiceInference
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model

if len(sys.argv) < 4:
    print("Usage: python gradio_voice_chat.py <model_path> <is_local_path> <hf>")
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
    
    nexa_model = NexaVoiceInference(
        model_path=model_path,
        local_path=local_path,
    )
    return nexa_model

try:
    nexa_model = load_model(model_path, is_local_path, hf)
except Exception as e:
    print(f"Failed to load model: {e}")
    nexa_model = None

def process_audio_fn(audio_file, beam_size, task, temperature):
    """
    Args:
        audio_file (str): Path to audio
        beam_size (int): Beam size 
        task (str): 'transcribe' or 'translate'
        temperature (float): sampling temperature
    """
    if not nexa_model:
        return "No model loaded. Please ensure the model is correctly initialized."
    if not audio_file:
        return "Please provide an audio input."

    try:
        with suppress_stdout_stderr():
            segments, _ = nexa_model.model.transcribe(
                audio_file,
                beam_size=beam_size,
                task=task,
                temperature=temperature,
                vad_filter=True
            )
        transcription = "".join(segment.text for segment in segments)
        return transcription
    except Exception as e:
        return f"Error during audio transcription: {e}"

with gr.Blocks() as demo:
    gr.HTML(
        f"""
        <div style="display: flex; align-items: center; margin-bottom: 5px; padding-top: 10px;">
            <h1 style="font-family: Arial, sans-serif; font-size: 2.5em; font-weight: bold; margin: 0; padding-bottom: 5px;">
                Nexa AI Voice Transcription
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

    gr.HTML("<h3 style='font-family: Arial, sans-serif; font-size: 1.5em; font-weight: bold;'>Transcription Parameters</h3>")
    with gr.Row():
        beam_size_slider = gr.Slider(
            minimum=1, maximum=10, step=1, value=5, label="Beam Size"
        )
        task_dropdown = gr.Dropdown(
            choices=["transcribe", "translate"],
            value="transcribe",
            label="Task"
        )
        temperature_slider = gr.Slider(
            minimum=0.00, maximum=1.00, step=0.01, value=0.0, label="Temperature"
        )

    gr.HTML("<h2 style='font-family: Arial, sans-serif; font-size: 1.5em; font-weight: bold;'>Upload / Record Audio</h2>")
    with gr.Row():
        uploaded_audio = gr.Audio(type="filepath", label="Chooese an audio file under 200MB (wav, mp3) / Record Audio from Microphone")
        upload_response = gr.Textbox(label="Model Response", interactive=False)
    process_upload_button = gr.Button("Process Audio")

    process_upload_button.click(
        process_audio_fn,
        inputs=[uploaded_audio, beam_size_slider, task_dropdown, temperature_slider],
        outputs=upload_response
    )

if __name__ == "__main__":
    demo.launch(server_name="0.0.0.0", server_port=7860, share=True, inbrowser=True)
