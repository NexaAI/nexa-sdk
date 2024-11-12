import io
import os
import sys
import tempfile

import streamlit as st
from st_audiorec import st_audiorec

from nexa.general import pull_model
from nexa.gguf.nexa_inference_audio_lm import NexaAudioLMInference

# Initialize session state
if "has_result" not in st.session_state:
    st.session_state.has_result = False
    st.session_state.prompt = ""
    st.session_state.key = 0
    st.rerun()

default_model = sys.argv[1]
is_local_path = False if sys.argv[2] == "False" else True
hf = False if sys.argv[3] == "False" else True
projector_local_path = sys.argv[4] if len(sys.argv) > 4 else None

@st.cache_resource
def load_model(model_path):
    if is_local_path:
        local_path = model_path
    elif hf:
        local_path, _ = pull_model(model_path, hf=True)
    else:
        local_path, _ = pull_model(model_path)
        
    if is_local_path:
        nexa_model = NexaAudioLMInference(model_path=model_path, local_path=local_path, projector_local_path=projector_local_path)
    else:
        nexa_model = NexaAudioLMInference(model_path=model_path, local_path=local_path)
    return nexa_model

def process_audio(nexa_model, audio_file, prompt=""):
    # Save the uploaded audio data to a temporary file
    audio_data = audio_file.getvalue()
    with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as temp_audio:
        temp_audio.write(audio_data)
        temp_audio_path = temp_audio.name

    try:
        # Use the model's inference method directly
        response = nexa_model.inference(temp_audio_path, prompt)
        return response

    except Exception as e:
        st.error(f"Error during audio processing: {e}")
        return None
    finally:
        # Clean up the temporary audio file
        if os.path.exists(temp_audio_path):
            os.unlink(temp_audio_path)

def start_new_callback():
    st.session_state.has_result = False
    st.session_state.prompt = ""
    st.session_state.key += 1

st.markdown("# Nexa AI AudioLM Generation [![Nexa SDK](https://img.shields.io/badge/SDK-Nexa-blue)](https://github.com/NexaAI/nexa-sdk)")
st.caption("Powered by Nexa AI SDK🐙")

# Sidebar configuration
st.sidebar.header("Model Configuration")
model_path = st.sidebar.text_input("Model path", default_model)

if not model_path:
    st.warning("Please enter a valid model path to proceed.")
    st.stop()

# Initialize or update the model when the path changes
if ("current_model_path" not in st.session_state 
    or st.session_state.current_model_path != model_path):
    st.session_state.current_model_path = model_path
    st.session_state.nexa_model = load_model(model_path)
    if st.session_state.nexa_model is None:
        st.stop()

# Wrap input sections
input_container = st.container()

with input_container:
    # Text prompt input
    prompt = st.text_input("Enter optional prompt text:", value=st.session_state.prompt, key=f"prompt_{st.session_state.key}")

    
    # Option 1: Upload Audio File
    st.subheader("Option 1: Upload Audio File")
    uploaded_file = st.file_uploader("Choose an audio file", type=["wav", "mp3"], key=f"uploader_{st.session_state.key}")

    if uploaded_file is not None:
        st.audio(uploaded_file, format="audio/wav")
        
        if st.button("Process Audio", key="process_upload"):
            with st.spinner("Processing audio..."):
                response = process_audio(st.session_state.nexa_model, uploaded_file, prompt)
                if response:
                    st.session_state.has_result = True
            
            if response:
                st.subheader("Model Response:")
                st.write(response)
                if st.session_state.has_result:
                    if st.button("Start new", key="start_new_upload", on_click=start_new_callback):
                        pass
                    st.write("Start new conversation to try next prompt")
            else:
                st.error("Processing failed. Please try again with a different audio file.")
    
    # Option 2: Real-time Recording
    st.subheader("Option 2: Record Audio")
    wav_audio_data = st_audiorec()
    
    if wav_audio_data:
        if st.button("Process Audio", key="process_record"):
            with st.spinner("Processing audio..."):
                response = process_audio(st.session_state.nexa_model, io.BytesIO(wav_audio_data), prompt)
                if response:
                    st.session_state.has_result = True
            
            if response:
                st.subheader("Model Response:")
                st.write(response)
                if st.session_state.has_result:
                    if st.button("Start new", key="start_new_record", on_click=start_new_callback):
                        pass
                    st.write("Start new conversation to try next prompt")
            else:
                st.error("Processing failed. Please try recording again.")
    else:
        st.warning("No audio recorded. Please record some audio before processing.")