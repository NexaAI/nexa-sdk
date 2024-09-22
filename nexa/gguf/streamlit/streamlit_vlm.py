import sys
import tempfile
from typing import Iterator

import streamlit as st
from PIL import Image
from nexa.general import pull_model
from nexa.gguf.nexa_inference_vlm import NexaVLMInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    local_path, run_type = pull_model(model_path)
    nexa_model = NexaVLMInference(model_path=model_path, local_path=local_path)
    return nexa_model


def generate_response(
    nexa_model: NexaVLMInference, image_path: str, user_input: str
) -> Iterator:
    return nexa_model._chat(user_input, image_path)


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
st.title("Nexa AI VLM Generation")
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

st.sidebar.header("Generation Parameters")
temperature = st.sidebar.slider(
    "Temperature", 0.0, 1.0, st.session_state.nexa_model.params["temperature"]
)
max_new_tokens = st.sidebar.slider(
    "Max New Tokens", 1, 2048, st.session_state.nexa_model.params["max_new_tokens"]
)
top_k = st.sidebar.slider("Top K", 1, 100, st.session_state.nexa_model.params["top_k"])
top_p = st.sidebar.slider(
    "Top P", 0.0, 1.0, st.session_state.nexa_model.params["top_p"]
)

st.session_state.nexa_model.params.update(
    {
        "temperature": temperature,
        "max_new_tokens": max_new_tokens,
        "top_k": top_k,
        "top_p": top_p,
    }
)

user_input = st.text_input("Enter your text input:")
uploaded_file = st.file_uploader("Upload an image", type=["png", "jpg", "jpeg"])

generate_button = st.button("Send")
spinner_placeholder = st.empty()
success_label = st.empty()
response_placeholder = st.empty()
image_placeholder = st.empty()

if uploaded_file:
    image_placeholder.image(
        uploaded_file, caption="Uploaded Image", use_column_width=True
    )

if generate_button:
    if not user_input and not uploaded_file:
        st.warning("Please enter text input and upload an image to proceed.")
    else:
        with spinner_placeholder:
            with st.spinner("Generating description..."):
                with tempfile.NamedTemporaryFile() as image_path:
                    if uploaded_file:
                        ext = uploaded_file.name.split(".")[-1]
                        full_path = f"{image_path.name}.{ext}"
                        with Image.open(uploaded_file) as img:
                            img.save(full_path)

                    full_response = ""
                    for chunk in generate_response(
                        st.session_state.nexa_model, full_path, user_input
                    ):
                        delta = chunk["choices"][0]["delta"]
                        if "role" in delta:
                            print(delta["role"], end=": ", flush=True)
                        elif "content" in delta:
                            print(delta["content"], end="", flush=True)
                            full_response += delta["content"]
                        response_placeholder.write(full_response)
                    success_label.success("Response generated successfully.")
