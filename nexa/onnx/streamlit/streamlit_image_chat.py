import io
import sys

import numpy as np
import streamlit as st

from nexa.general import pull_model
from nexa.onnx.nexa_inference_image import NexaImageInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    local_path, run_type = pull_model(model_path)
    nexa_model = NexaImageInference(model_path=model_path, local_path=local_path)

    if nexa_model.downloaded_onnx_folder is None:
        st.error("Failed to download the model. Please check the model path.")
        return None

    nexa_model._load_model(nexa_model.downloaded_onnx_folder)
    return nexa_model


def generate_images(nexa_model: NexaImageInference, prompt, negative_prompt):
    if nexa_model.pipeline is None:
        st.error("Model not loaded. Please check the model path and try again.")
        return None

    generator = np.random.RandomState(nexa_model.params["random_seed"])

    images = nexa_model.pipeline(
        prompt=prompt,
        negative_prompt=negative_prompt if negative_prompt else None,
        num_inference_steps=nexa_model.params["num_inference_steps"],
        num_images_per_prompt=nexa_model.params["num_images_per_prompt"],
        height=nexa_model.params["height"],
        width=nexa_model.params["width"],
        generator=generator,
        guidance_scale=nexa_model.params["guidance_scale"],
    ).images

    return images


st.title("Nexa AI Image Generation")
st.caption("Powered by Nexa AI SDK🐙")

st.sidebar.header("Model Configuration")
model_path = st.sidebar.text_input("Model path", default_model)

if not model_path:
    st.warning("Please enter a valid path or identifier for the model in Nexa Model Hub to proceed.")
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

st.sidebar.header("Generation Parameters")
num_inference_steps = st.sidebar.slider(
    "Inference Steps", 1, 100, st.session_state.nexa_model.params["num_inference_steps"]
)
height = st.sidebar.slider(
    "Image Height", 128, 1024, st.session_state.nexa_model.params["height"]
)
width = st.sidebar.slider(
    "Image Width", 128, 1024, st.session_state.nexa_model.params["width"]
)
guidance_scale = st.sidebar.slider(
    "Guidance Scale", 1.0, 20.0, st.session_state.nexa_model.params["guidance_scale"]
)
num_images_per_prompt = st.sidebar.slider(
    "Number of Images",
    1,
    4,
    st.session_state.nexa_model.params["num_images_per_prompt"],
)
random_seed = st.sidebar.number_input(
    "Random Seed", value=st.session_state.nexa_model.params["random_seed"]
)

# Update model parameters
st.session_state.nexa_model.params.update(
    {
        "num_inference_steps": num_inference_steps,
        "height": height,
        "width": width,
        "guidance_scale": guidance_scale,
        "random_seed": random_seed,
        "num_images_per_prompt": num_images_per_prompt,
    }
)

prompt = st.text_input("Enter your prompt:")
negative_prompt = st.text_input("Enter your negative prompt (optional):")

if st.button("Generate Images"):
    if prompt:
        with st.spinner("Generating images..."):
            images = generate_images(
                st.session_state.nexa_model, prompt, negative_prompt
            )
        if images:
            cols = st.columns(num_images_per_prompt)
            for idx, image in enumerate(images):
                with cols[idx]:
                    st.image(
                        image, caption=f"Generated Image {idx+1}", use_column_width=True
                    )

                    # Convert PIL Image to bytes
                    img_byte_arr = io.BytesIO()
                    image.save(img_byte_arr, format="PNG")
                    img_byte_arr = img_byte_arr.getvalue()

                    # Provide a download button for each image
                    st.download_button(
                        label=f"Download Image {idx+1}",
                        data=img_byte_arr,
                        file_name=f"generated_image_{idx+1}.png",
                        mime="image/png",
                    )
    else:
        st.warning("Please enter a prompt to generate images.")
