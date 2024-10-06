import os
import sys
from PIL import Image
from nexa.general import pull_model
import streamlit as st
from nexa.gguf.nexa_inference_image import NexaImageInference
import io

default_model = sys.argv[1]
is_local_path = False if sys.argv[2] == "False" else True
hf = False if sys.argv[3] == "False" else True


@st.cache_resource
def load_model(model_path):
    if is_local_path:
        local_path = model_path
    elif hf:
        local_path, _ = pull_model(model_path, hf=True)
    else:
        local_path, run_type = pull_model(model_path)
    nexa_model = NexaImageInference(model_path=model_path, local_path=local_path)
    return nexa_model


def generate_images(nexa_model: NexaImageInference, prompt: str, negative_prompt: str):
    output_dir = os.path.dirname(nexa_model.params["output_path"])
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    images = nexa_model.txt2img(
        prompt=prompt,
        negative_prompt=negative_prompt,
        cfg_scale=nexa_model.params["guidance_scale"],
        width=nexa_model.params["width"],
        height=nexa_model.params["height"],
        sample_steps=nexa_model.params["num_inference_steps"],
        seed=nexa_model.params["random_seed"]
    )
    
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
            for i, image in enumerate(images):
                st.image(image, caption=f"Generated Image", use_column_width=True)
                
                img_byte_arr = io.BytesIO()
                image.save(img_byte_arr, format='PNG')
                img_byte_arr = img_byte_arr.getvalue()

                st.download_button(
                    label=f"Download Image",
                    data=img_byte_arr,
                    file_name=f"generated_image.png",
                    mime="image/png"
                )
