import io
import os
import sys
import tempfile
import subprocess
import re
from typing import List
import streamlit as st
from st_audiorec import st_audiorec
from nexa.general import pull_model
from nexa.gguf.nexa_inference_voice import NexaVoiceInference
from nexa.utils import (
    get_model_options,
    update_model_options,
)
from nexa.constants import NEXA_RUN_MODEL_MAP_VOICE

specified_run_type = 'Audio'
model_map = NEXA_RUN_MODEL_MAP_VOICE

# init from command line args:
try:
    default_model = sys.argv[1]
    is_local_path = sys.argv[2].lower() == "true"
    hf = sys.argv[3].lower() == "true"
except IndexError:
    st.error("Missing required command line arguments.")
    sys.exit(1)  # terminate with an error

@st.cache_resource(show_spinner=False)
def load_model(model_path, is_local=False, is_hf=False):
    """Load model with model mapping logic."""
    try:
        if is_local:
            # for local paths, use the path directly:
            nexa_model = NexaVoiceInference(model_path=model_path, local_path=model_path)
        else:
            # for non-local paths:
            if is_hf:
                local_path, _ = pull_model(model_path, hf=True)
                nexa_model = NexaVoiceInference(model_path=model_path, local_path=local_path)
            else:
                # handle Model Hub models:
                if model_path in NEXA_RUN_MODEL_MAP_VOICE:
                    real_model_path = NEXA_RUN_MODEL_MAP_VOICE[model_path]
                    local_path, _ = pull_model(real_model_path)
                    nexa_model = NexaVoiceInference(model_path=real_model_path, local_path=local_path)
                else:
                    local_path, _ = pull_model(model_path)
                    nexa_model = NexaVoiceInference(model_path=model_path, local_path=local_path)
        return nexa_model
    except Exception as e:
        st.error(f"Error loading model: {str(e)}")
        return None

def transcribe_audio(nexa_model, audio_file):
    with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as temp_audio:
        temp_audio.write(audio_file.getvalue())
        temp_audio_path = temp_audio.name

    try:
        segments, _ = nexa_model.model.transcribe(
            temp_audio_path,
            beam_size=nexa_model.params["beam_size"],
            language=nexa_model.params["language"],
            task=nexa_model.params["task"],
            temperature=nexa_model.params["temperature"],
            vad_filter=True
        )
        transcription = "".join(segment.text for segment in segments)
        return transcription

    except Exception as e:
        st.error(f"Error during audio transcription: {e}")
        return None
    finally:
        os.unlink(temp_audio_path)

# UI setup:
st.set_page_config(page_title="Nexa AI Voice Transcription", layout="wide")
st.title("Nexa AI Voice Transcription")
st.caption("Powered by Nexa AI SDK🐙")

# force refresh model options on every page load:
if 'model_options' not in st.session_state:
    st.session_state.model_options = get_model_options(specified_run_type, model_map)
else:
    update_model_options(specified_run_type, model_map)

# init session state variables:
if 'initialized' not in st.session_state:
    st.session_state.model_options = get_model_options(specified_run_type, model_map)
    st.session_state.current_model_path = default_model
    st.session_state.current_local_path = None
    st.session_state.current_hub_model = None

    # init with default model:
    if is_local_path:
        try:
            with st.spinner(f"Loading local model: {default_model}"):
                st.session_state.nexa_model = load_model(
                    default_model,
                    is_local=True,
                    is_hf=hf
                )
                if st.session_state.nexa_model:
                    st.session_state.current_local_path = default_model
                    st.session_state.current_model_path = default_model
        except Exception as e:
            st.error(f"Error loading local model: {str(e)}")
    elif hf:
        try:
            with st.spinner(f"Loading HuggingFace model: {default_model}"):
                st.session_state.nexa_model = load_model(default_model, is_hf=True)
                if st.session_state.nexa_model:
                    st.session_state.current_hub_model = default_model
                    st.session_state.current_model_path = default_model
        except Exception as e:
            st.error(f"Error loading HuggingFace model: {str(e)}")

    else:
        try:
            with st.spinner(f"Loading model: {default_model}"):
                st.session_state.nexa_model = load_model(default_model)
                if st.session_state.nexa_model:
                    st.session_state.current_model_path = default_model
                    st.session_state.current_hub_model = default_model
        except Exception as e:
            st.error(f"Error loading model: {str(e)}")

    st.session_state.initialized = True

# model selection UI:
st.sidebar.header("Model Configuration")

# update selectbox index based on current model:
current_index = st.session_state.model_options.index("Use Model From Nexa Model Hub 🔍")
if 'nexa_model' in st.session_state:
    if st.session_state.current_model_path in st.session_state.model_options:
        current_index = st.session_state.model_options.index(st.session_state.current_model_path)
    elif st.session_state.current_hub_model:
        current_index = st.session_state.model_options.index("Use Model From Nexa Model Hub 🔍")
    elif st.session_state.current_local_path:
        current_index = st.session_state.model_options.index("Local Model 📁")

selected_option = st.sidebar.selectbox(
    "Select a Model",
    st.session_state.model_options,
    index=current_index
)

# handle model selection:
if selected_option == "Local Model 📁":
    model_path = st.sidebar.text_input(
        "Enter local model path",
        value=st.session_state.current_local_path if hasattr(st.session_state, 'current_local_path') else "",
        help="Enter the full path to your local model directory (e.g., /home/user/.cache/nexa/hub/official/model-name)"
    )

    if not model_path:
        st.warning("Please enter a valid local model path to proceed.")
        st.stop()

    if (not hasattr(st.session_state, 'current_local_path') or
        st.session_state.current_local_path != model_path):
        with st.spinner("Loading local model..."):
            st.session_state.nexa_model = load_model(
                    model_path,  # use the user input path
                    is_local=True,
                    is_hf=hf
                )
            if st.session_state.nexa_model:
                st.session_state.current_local_path = model_path
                st.session_state.current_model_path = model_path

elif selected_option == "Use Model From Nexa Model Hub 🔍":
    model_path = st.sidebar.text_input(
        "Enter model name from Nexa Model Hub",
        value=st.session_state.current_hub_model if hasattr(st.session_state, 'current_hub_model') else default_model
    )
    if not model_path:
        st.warning("""
        How to add a model from Nexa Model Hub:
        \n1. Visit [Nexa Model Hub](https://nexaai.com/models)
        \n2. Find an audio model using the task filters
        \n3. Select your desired model and copy either:
        \n   - The full nexa run command (e.g., nexa run faster-whisper-tiny:bin-cpu-fp16), or
        \n   - Simply the model name (e.g., faster-whisper-tiny:bin-cpu-fp16)
        \n4. Paste it into the field on the sidebar and press enter
        """)
        st.stop()

    # process the input after checking it's not empty:
    if model_path.startswith("nexa run"):
        model_path = model_path.split("nexa run")[-1].strip()

    if (not hasattr(st.session_state, 'current_hub_model') or st.session_state.current_hub_model != model_path):
        with st.spinner("Loading model from hub..."):
            st.session_state.nexa_model = load_model(model_path, is_local=False, is_hf=False)
            if st.session_state.nexa_model:
                st.session_state.current_hub_model = model_path
                st.session_state.current_model_path = model_path
                st.session_state.current_local_path = None  # clear local path state when switching to hub model

else:
    model_path = selected_option
    if (not hasattr(st.session_state, 'current_model_path') or
        st.session_state.current_model_path != model_path):
        with st.spinner(f"Loading model: {model_path}"):
            st.session_state.nexa_model = load_model(model_path, is_local=False, is_hf=False)
            if st.session_state.nexa_model:
                st.session_state.current_model_path = model_path
                st.session_state.current_local_path = None
                st.session_state.current_hub_model = None

# only show transcription parameters if model is loaded:
if hasattr(st.session_state, 'nexa_model') and st.session_state.nexa_model:
    # transcription parameters:
    st.sidebar.header("Transcription Parameters")
    beam_size = st.sidebar.slider(
        "Beam Size",
        1, 10,
        st.session_state.nexa_model.params["beam_size"]
    )
    task = st.sidebar.selectbox(
        "Task",
        ["transcribe", "translate"],
        index=0 if st.session_state.nexa_model.params["task"] == "transcribe" else 1
    )
    temperature = st.sidebar.slider(
        "Temperature",
        0.0, 1.0,
        st.session_state.nexa_model.params["temperature"],
        step=0.1
    )

    # update model parameters:
    st.session_state.nexa_model.params.update({
        "beam_size": beam_size,
        "task": task,
        "temperature": temperature,
    })

    # Option 1: Upload Audio File
    st.header("Option 1: Upload Audio File")
    uploaded_file = st.file_uploader("Choose an audio file", type=["wav", "mp3"])

    if uploaded_file is not None:
        st.audio(uploaded_file, format="audio/wav")

        if st.button("Transcribe Uploaded Audio"):
            with st.spinner("Transcribing audio..."):
                transcription = transcribe_audio(st.session_state.nexa_model, uploaded_file)

            if transcription:
                st.subheader("Transcription:")
                st.write(transcription)

                # Provide a download button for the transcription
                transcription_bytes = transcription.encode()
                st.download_button(
                    label="Download Transcription",
                    data=transcription_bytes,
                    file_name="transcription.txt",
                    mime="text/plain",
                )
            else:
                st.error(
                    "Transcription failed. Please try again with a different audio file."
                )

    # Option 2: Real-time Recording
    st.header("Option 2: Record Audio")
    wav_audio_data = st_audiorec()

    if wav_audio_data:
        if st.button("Transcribe Recorded Audio"):
            with st.spinner("Transcribing audio..."):
                transcription = transcribe_audio(st.session_state.nexa_model, io.BytesIO(wav_audio_data))

            if transcription:
                st.subheader("Transcription:")
                st.write(transcription)

                # Provide a download button for the transcription
                transcription_bytes = transcription.encode()
                st.download_button(
                    label="Download Transcription",
                    data=transcription_bytes,
                    file_name="transcription.txt",
                    mime="text/plain",
                )
            else:
                st.error("Transcription failed. Please try recording again.")
    else:
        st.warning("No audio recorded. Please record some audio before transcribing.")
else:
    st.warning("Please select or load a model to proceed.")
