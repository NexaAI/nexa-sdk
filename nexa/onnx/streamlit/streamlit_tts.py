import os
import sys
import time

import soundfile as sf
import streamlit as st
from nexa.general import pull_model
from nexa.onnx.nexa_inference_tts import NexaTTSInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path: str):
    local_path, run_type = pull_model(model_path)  
    return NexaTTSInference(model_path=model_path, local_path=local_path)


def generate_audio(nexa_model: NexaTTSInference, text):
    if nexa_model.model is None or nexa_model.tokenizer is None:
        st.error("Model or tokenizer not loaded properly.")
        return None

    inputs = nexa_model.tokenizer(text)
    outputs = nexa_model.model.run(None, {"text": inputs})

    audio_file_path = os.path.join(
        nexa_model.params["output_path"], f"audio_{int(time.time())}.wav"
    )
    os.makedirs(nexa_model.params["output_path"], exist_ok=True)
    sf.write(audio_file_path, outputs[0], nexa_model.params["sampling_rate"])

    return audio_file_path


st.title("Nexa AI Text to Speech")
st.caption("Powered by Nexa AI SDK🐙")

st.sidebar.header("Model Configuration")

model_path = st.sidebar.text_input("Model path", default_model)

st.session_state.nexa_model = load_model(model_path)

if not st.session_state.nexa_model:
    st.warning("Please enter a valid S3 model filename to proceed.")
    st.stop()

st.sidebar.header("Audio Parameters")
sampling_rate = st.sidebar.slider(
    "Sampling Rate", 8000, 48000, st.session_state.nexa_model.params["sampling_rate"]
)

# Update model parameters
st.session_state.nexa_model.params.update({"sampling_rate": sampling_rate})

text_input = st.text_area("Enter text to convert to speech", "")
generate_button = st.button("Generate Speech")

if generate_button and text_input:
    with st.spinner("Generating audio..."):
        audio_path = generate_audio(st.session_state.nexa_model, text_input)
        if audio_path:
            st.success("Audio generated successfully!")
            audio_file = open(audio_path, "rb")
            audio_bytes = audio_file.read()
            st.audio(audio_bytes, format="audio/wav")
            audio_file.close()
        else:
            st.error("Failed to generate audio.")
