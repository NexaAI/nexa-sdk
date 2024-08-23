import io
import os
import sys
import tempfile

import librosa
import streamlit as st
from st_audiorec import st_audiorec

from nexa.general import pull_model
from nexa.gguf.nexa_inference_voice import NexaVoiceInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    local_path, run_type = pull_model(model_path)
    nexa_model = NexaVoiceInference(model_path=model_path, local_path=local_path)
    return nexa_model


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


st.title("Nexa AI Voice Transcription")
st.caption("Powered by Nexa AI SDK🐙")

st.sidebar.header("Model Configuration")
model_path = st.sidebar.text_input("Model path", default_model)

if not model_path:
    st.warning("Please enter a valid S3 model filename to proceed.")
    st.stop()

# Initialize or update the model when the path changes
if (
    "current_model_path" not in st.session_state
    or st.session_state.current_model_path != model_path
):
    st.session_state.current_model_path = model_path
    st.session_state.nexa_model = load_model(model_path)
    if st.session_state.nexa_model is None:
        st.stop()

# Add sidebar options for new parameters
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

# Update model parameters
st.session_state.nexa_model.params.update(
    {
        "beam_size": beam_size,
        "task": task,
        "temperature": temperature,
    }
)

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
