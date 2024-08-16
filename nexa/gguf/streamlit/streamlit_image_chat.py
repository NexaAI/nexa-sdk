import os
import sys

from PIL import Image

import streamlit as st
from nexa.gguf.nexa_inference_image import NexaImageInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    nexa_model = NexaImageInference(model_path)
    return nexa_model


def generate_images(nexa_model, prompt, negative_prompt):
    output_dir = os.path.dirname(nexa_model.params["output_path"])
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    nexa_model._generate_images(prompt, negative_prompt)

    images = []
    for file_name in os.listdir(output_dir):
        if file_name.endswith(".png"):
            image_path = os.path.join(output_dir, file_name)
            images.append(Image.open(image_path))
    return images


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
st.title("Nexa AI Image Generation")
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
num_inference_steps = st.sidebar.slider(
    "Number of Inference Steps",
    1,
    100,
    st.session_state.nexa_model.params["num_inference_steps"],
)
num_images_per_prompt = st.sidebar.slider(
    "Number of Images per Prompt",
    1,
    10,
    st.session_state.nexa_model.params["num_images_per_prompt"],
)
height = st.sidebar.slider(
    "Height", 64, 1024, st.session_state.nexa_model.params["height"]
)
width = st.sidebar.slider(
    "Width", 64, 1024, st.session_state.nexa_model.params["width"]
)
guidance_scale = st.sidebar.slider(
    "Guidance Scale", 0.0, 20.0, st.session_state.nexa_model.params["guidance_scale"]
)
random_seed = st.sidebar.slider(
    "Random Seed", 0, 10000, st.session_state.nexa_model.params["random_seed"]
)

st.session_state.nexa_model.params.update(
    {
        "num_inference_steps": num_inference_steps,
        "num_images_per_prompt": num_images_per_prompt,
        "height": height,
        "width": width,
        "guidance_scale": guidance_scale,
        "random_seed": random_seed,
    }
)

prompt = st.text_input("Enter your prompt:")
negative_prompt = st.text_input("Enter your negative prompt (optional):")

if st.button("Generate Image"):
    if not prompt:
        st.warning("Please enter a prompt to proceed.")
    else:
        with st.spinner("Generating images..."):
            images = generate_images(
                st.session_state.nexa_model, prompt, negative_prompt
            )
            st.success("Images generated successfully!")
            for image in images:
                st.image(image, caption="Generated Image", use_column_width=True)
