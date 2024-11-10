import sys
import tempfile
import subprocess
import re
from typing import List, Iterator
import streamlit as st
from PIL import Image
from nexa.general import pull_model
from nexa.gguf.nexa_inference_vlm import NexaVLMInference
from nexa.utils import (
    get_model_options,
    update_model_options,
)
from nexa.constants import NEXA_RUN_MODEL_MAP_VLM

specified_run_type = 'Multimodal'
model_map = NEXA_RUN_MODEL_MAP_VLM

# init from command line args:
try:
    default_model = sys.argv[1]
    is_local_path = sys.argv[2].lower() == "true"
    hf = sys.argv[3].lower() == "true"
    projector_local_path = sys.argv[4] if len(sys.argv) > 4 else None
except IndexError:
    st.error("Missing required command line arguments.")
    sys.exit(1)  # terminate with an error

@st.cache_resource(show_spinner=False)
def load_model(model_path, is_local=False, is_hf=False, projector_path=None):
    """Load model with model mapping logic."""
    try:
        if is_local:
            local_path = model_path
            nexa_model = NexaVLMInference(
                model_path=model_path,
                local_path=local_path,
                projector_local_path=projector_path
            )
        elif is_hf:
            local_path, _ = pull_model(model_path, hf=True)
            nexa_model = NexaVLMInference(model_path=model_path, local_path=local_path)
        else:
            # get the actual model name from the mapping if it exists:
            if model_path in NEXA_RUN_MODEL_MAP_VLM:
                real_model_path = NEXA_RUN_MODEL_MAP_VLM[model_path]
                local_path, run_type = pull_model(real_model_path)
            else:
                local_path, run_type = pull_model(model_path)
            nexa_model = NexaVLMInference(model_path=model_path, local_path=local_path)
        return nexa_model
    except Exception as e:
        st.error(f"Error loading model: {str(e)}")
        return None

def generate_response(
    nexa_model: NexaVLMInference, image_path: str, user_input: str
) -> Iterator:
    return nexa_model._chat(user_input, image_path)

# UI setup:
st.set_page_config(page_title="Nexa AI Multimodal Generation", layout="wide")
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
st.title("Nexa AI Multimodal Generation")
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
    st.session_state.current_projector_path = projector_local_path
    st.session_state.current_hub_model = None

    # init with default model:
    if is_local_path:
        try:
            with st.spinner(f"Loading local model: {default_model}"):
                st.session_state.nexa_model = load_model(
                    default_model,
                    is_local=True,
                    projector_path=projector_local_path
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
    elif st.session_state.current_local_path:
        current_index = st.session_state.model_options.index("Local Model 📁")
    elif st.session_state.current_hub_model:
        current_index = st.session_state.model_options.index("Use Model From Nexa Model Hub 🔍")

selected_option = st.sidebar.selectbox(
    "Select a Model",
    st.session_state.model_options,
    index=current_index
)

# handle model selection:
if selected_option == "Local Model 📁":
    model_path = st.sidebar.text_input(
        "Enter local model path",
        value=st.session_state.current_local_path if hasattr(st.session_state, 'current_local_path') else ""
    )

    projector_path = st.sidebar.text_input(
        "Enter local projector path",
        value=st.session_state.current_projector_path if hasattr(st.session_state, 'current_projector_path') else ""
    )

    if not model_path:
        st.warning("Please enter a valid local model path to proceed.")
        st.stop()

    if not projector_path:
        st.warning("Please enter a valid local projector path to proceed.")
        st.stop()

    if (not hasattr(st.session_state, 'current_local_path') or st.session_state.current_local_path != model_path or st.session_state.current_projector_path != projector_path):
        with st.spinner("Loading local model..."):
            st.session_state.nexa_model = load_model(
                model_path,
                is_local=True,
                projector_path=projector_path
            )
            if st.session_state.nexa_model:
                st.session_state.current_local_path = model_path
                st.session_state.current_projector_path = projector_path
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
        \n2. Find a multimodal model using the task filters
        \n3. Select your desired model and copy either:
        \n   - The full nexa run command, or
        \n   - Simply the model name
        \n4. Paste it into the field on the sidebar and press enter
        """)
        st.stop()

    if model_path.startswith("nexa run"):
        model_path = model_path.split("nexa run")[-1].strip()

    if (not hasattr(st.session_state, 'current_hub_model') or
        st.session_state.current_hub_model != model_path):
        with st.spinner("Loading model from hub..."):
            st.session_state.nexa_model = load_model(model_path)
            if st.session_state.nexa_model:
                st.session_state.current_hub_model = model_path
                st.session_state.current_model_path = model_path

else:
    model_path = selected_option
    if (not hasattr(st.session_state, 'current_model_path') or
        st.session_state.current_model_path != model_path):
        with st.spinner(f"Loading model: {model_path}"):
            st.session_state.nexa_model = load_model(model_path)
            if st.session_state.nexa_model:
                st.session_state.current_model_path = model_path

# only show generation params if model is loaded:
if hasattr(st.session_state, 'nexa_model') and st.session_state.nexa_model:
    # generation params:
    st.sidebar.header("Generation Parameters")
    temperature = st.sidebar.slider(
        "Temperature",
        0.0,
        1.0,
        st.session_state.nexa_model.params["temperature"]
    )
    max_new_tokens = st.sidebar.slider(
        "Max New Tokens",
        1,
        2048,
        st.session_state.nexa_model.params["max_new_tokens"]
    )
    top_k = st.sidebar.slider(
        "Top K",
        1,
        100,
        st.session_state.nexa_model.params["top_k"]
    )
    top_p = st.sidebar.slider(
        "Top P",
        0.0,
        1.0,
        st.session_state.nexa_model.params["top_p"]
    )

    st.session_state.nexa_model.params.update({
        "temperature": temperature,
        "max_new_tokens": max_new_tokens,
        "top_k": top_k,
        "top_p": top_p,
    })

    # generation interface:
    user_input = st.text_input("Enter your text input:")
    uploaded_file = st.file_uploader("Upload an image", type=["png", "jpg", "jpeg"])

    generate_button = st.button("Send")
    spinner_placeholder = st.empty()
    success_label = st.empty()
    response_placeholder = st.empty()
    image_placeholder = st.empty()

    if uploaded_file:
        image_placeholder.image(
            uploaded_file,
            caption="Uploaded Image",
            use_column_width=True
        )

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

                        full_response = ""
                        for chunk in generate_response(
                            st.session_state.nexa_model,
                            full_path,
                            user_input
                        ):
                            delta = chunk["choices"][0]["delta"]
                            if "role" in delta:
                                print(delta["role"], end=": ", flush=True)
                            elif "content" in delta:
                                print(delta["content"], end="", flush=True)
                                full_response += delta["content"]
                            response_placeholder.write(full_response)
                        success_label.success("Response generated successfully.")
else:
    st.warning("Please select or load a model to proceed.")
