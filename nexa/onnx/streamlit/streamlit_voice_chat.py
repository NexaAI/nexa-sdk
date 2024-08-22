import io
import os
import sys
import tempfile

import librosa
import streamlit as st
from st_audiorec import st_audiorec

from nexa.general import pull_model
from nexa.onnx.nexa_inference_voice import NexaVoiceInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    local_path, run_type = pull_model(model_path)    
    nexa_model = NexaVoiceInference(model_path=model_path, local_path=local_path)

    if nexa_model.downloaded_onnx_folder is None:
        st.error("Failed to download the model. Please check the model path.")
        return None

    nexa_model._load_model(nexa_model.downloaded_onnx_folder)
    return nexa_model


def transcribe_audio(nexa_model, audio_file):
    with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as temp_audio:
        temp_audio.write(audio_file.getvalue())
        temp_audio_path = temp_audio.name

    try:
        audio, sr = librosa.load(temp_audio_path, sr=nexa_model.params["sampling_rate"])
        inputs = nexa_model.processor(
            audio, return_tensors="pt", sampling_rate=nexa_model.params["sampling_rate"]
        )

        input_features = inputs.input_features
        attention_mask = (
            inputs.attention_mask if hasattr(inputs, "attention_mask") else None
        )

        gen_tokens = nexa_model.model.generate(
            input_features=input_features,
            attention_mask=attention_mask,
        )

        transcription = nexa_model.processor.batch_decode(
            gen_tokens, skip_special_tokens=True
        )[0]

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
            audio, _ = librosa.load(
                io.BytesIO(wav_audio_data),
                sr=st.session_state.nexa_model.params["sampling_rate"],
            )
            transcription = transcribe_audio(
                st.session_state.nexa_model, io.BytesIO(wav_audio_data)
            )

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
