import sys
import tempfile

import streamlit as st
from PIL import Image
from nexa.general import pull_model
from nexa.gguf.nexa_inference_vlm_omni import NexaOmniVlmInference
import ctypes

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
        nexa_model = NexaOmniVlmInference(model_path=model_path, local_path=local_path, projector_local_path=projector_local_path)
    else:
        nexa_model = NexaOmniVlmInference(model_path=model_path, local_path=local_path)
    return nexa_model

def generate_response(nexa_model: NexaOmniVlmInference, image_path: str, user_input: str) -> str:
    return nexa_model.inference(user_input, image_path)

st.markdown(
    r"""
    <style>
    .stDeployButton {
            visibility: hidden;
        }
    </style>
    """,
    unsafe_allow_html=True,
)
st.title("Nexa AI Omni VLM Generation")
st.caption("Powered by Nexa AI SDK🐙")

st.sidebar.header("Model Configuration")
model_path = st.sidebar.text_input("Model path", default_model)

if not model_path:
    st.warning(
        "Please enter a valid path or identifier for the model in Nexa Model Hub to proceed."
    )
    st.stop()

if (
    "current_model_path" not in st.session_state
    or st.session_state.current_model_path != model_path
):
    st.session_state.current_model_path = model_path
    st.session_state.nexa_model = load_model(model_path)
    if st.session_state.nexa_model is None:
        st.stop()

user_input = st.text_input("Enter your text input:")
uploaded_file = st.file_uploader("Upload an image", type=["png", "jpg", "jpeg"])

generate_button = st.button("Send")
spinner_placeholder = st.empty()
success_label = st.empty()
response_placeholder = st.empty()
image_placeholder = st.empty()

if uploaded_file:
    image_placeholder.image(uploaded_file, caption="Uploaded Image")

if generate_button:
    if not user_input and not uploaded_file:
        st.warning("Please enter text input and upload an image to proceed.")
    else:
        with spinner_placeholder:
            with st.spinner("Generating description..."):
                with tempfile.NamedTemporaryFile() as image_path:
                    full_path = None
                    if uploaded_file:
                        ext = uploaded_file.name.split(".")[-1]
                        full_path = f"{image_path.name}.{ext}"
                        with Image.open(uploaded_file) as img:
                            img.save(full_path)

                    response = generate_response(
                        st.session_state.nexa_model, full_path, user_input
                    )

                    response_placeholder.write(response)
                    success_label.success("Response generated successfully.")