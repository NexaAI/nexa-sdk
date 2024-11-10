import os
import sys
from PIL import Image
import streamlit as st
from nexa.general import pull_model
from nexa.gguf.nexa_inference_image import NexaImageInference
from nexa.utils import (
    get_model_options,
    update_model_options,
)
from nexa.constants import (
    DEFAULT_IMG_GEN_PARAMS_LCM,
    DEFAULT_IMG_GEN_PARAMS_TURBO,
    DEFAULT_IMG_GEN_PARAMS,
    NEXA_RUN_MODEL_MAP_IMAGE,
    NEXA_RUN_MODEL_MAP_FLUX,
)
import io

specified_run_type = 'Computer Vision'
model_map = NEXA_RUN_MODEL_MAP_IMAGE | NEXA_RUN_MODEL_MAP_FLUX

def get_default_params(model_path: str) -> dict:
    """Get default parameters based on model type."""
    if "lcm-dreamshaper" in model_path or "flux" in model_path:
        return DEFAULT_IMG_GEN_PARAMS_LCM.copy()  # fast LCM models: 4 steps @ 1.0 guidance
    elif "sdxl-turbo" in model_path:
        return DEFAULT_IMG_GEN_PARAMS_TURBO.copy()  # sdxl-turbo: 5 steps @ 5.0 guidance
    else:
        return DEFAULT_IMG_GEN_PARAMS.copy()  # standard SD models: 20 steps @ 7.5 guidance

@st.cache_resource(show_spinner=False)
def load_model(model_path: str, is_local: bool = False, is_hf: bool = False):
    """Load model with proper error handling."""
    try:
        if is_local:
            local_path = model_path
        elif is_hf:
            try:
                local_path, _ = pull_model(model_path, hf=True)
                update_model_options(specified_run_type, model_map)
            except Exception as e:
                st.error(f"Error pulling HuggingFace model: {str(e)}")
                return None
        else:
            try:
                # model hub case:
                local_path, run_type = pull_model(model_path)
                if not local_path or not run_type:
                    st.error(f"Failed to pull model {model_path} from Nexa Model Hub")
                    return None
                update_model_options(specified_run_type, model_map)  # update options after successful pull
            except ValueError as e:
                st.error(f"Error pulling model from Nexa Model Hub: {str(e)}")
                return None
            except Exception as e:
                st.error(f"Unexpected error while pulling model: {str(e)}")
                return None

        try:
            nexa_model = NexaImageInference(
                model_path=model_path,
                local_path=local_path
            )

            # force refresh of model options after successful load:
            update_model_options(specified_run_type, model_map)

            return nexa_model

        except Exception as e:
            st.error(f"Error initializing model: {str(e)}")
            return None

    except Exception as e:
        st.error(f"Error in load_model: {str(e)}")
        return None

@st.cache_resource(show_spinner=False)
def load_local_model(local_path: str):
    """Load local model with default parameters."""
    try:
        nexa_model = NexaImageInference(
            model_path="local_model",
            local_path=local_path
        )
        update_model_options(specified_run_type, model_map)  # update options after successful local model load
        return nexa_model
    except Exception as e:
        st.error(f"Error loading local model: {str(e)}")
        return None

def generate_images(nexa_model: NexaImageInference, prompt: str, negative_prompt: str):
    """Generate images using the model."""
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

# main execution:
try:
    # get command line arguments with proper error handling:
    if len(sys.argv) < 4:
        st.error("Missing required command line arguments.")
        sys.exit(1)  # program terminated with an error

    default_model = sys.argv[1]
    is_local_path = sys.argv[2].lower() == "true"
    hf = sys.argv[3].lower() == "true"

    # UI setup:
    st.set_page_config(page_title="Nexa AI Image Generation", layout="wide")
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

    # force refresh model options on every page load:
    if 'model_options' not in st.session_state:
        st.session_state.model_options = get_model_options(specified_run_type, model_map)
    else:
        update_model_options(specified_run_type, model_map)

    # init session state variables:
    if 'initialized' not in st.session_state:
        st.session_state.current_model_path = None
        st.session_state.current_local_path = None
        st.session_state.current_hub_model = None

        if not is_local_path and not hf:
            try:
                with st.spinner(f"Loading model: {default_model}"):
                    st.session_state.nexa_model = load_model(default_model)
                    if st.session_state.nexa_model:
                        st.session_state.current_hub_model = default_model
            except Exception as e:
                st.error(f"Error loading default model: {str(e)}")

            if default_model not in st.session_state.model_options:
                st.session_state.current_model_index = st.session_state.model_options.index("Use Model From Nexa Model Hub 🔍")
        else:
            try:
                st.session_state.current_model_index = st.session_state.model_options.index(default_model)
            except ValueError:
                st.session_state.current_model_index = 0

        st.session_state.initialized = True

    # model selection sidebar:
    st.sidebar.header("Model Configuration")

    # update selectbox index based on current model
    if 'nexa_model' in st.session_state:
        if st.session_state.current_hub_model:
            current_index = st.session_state.model_options.index("Use Model From Nexa Model Hub 🔍")
        elif st.session_state.current_local_path:
            current_index = st.session_state.model_options.index("Local Model 📁")
        elif st.session_state.current_model_path:
            current_index = st.session_state.model_options.index(st.session_state.current_model_path)
        else:
            current_index = st.session_state.current_model_index
    else:
        current_index = st.session_state.current_model_index

    model_path = st.sidebar.selectbox(
        "Select a Model",
        st.session_state.model_options,
        index=current_index,
        key='model_selectbox'
    )

    # handle model path input:
    if model_path == "Local Model 📁":
        local_model_path = st.sidebar.text_input("Enter local model path")
        if not local_model_path:
            st.warning("Please enter a valid local model path to proceed.")
            st.stop()
        local_model_path = local_model_path.strip()  # remove spaces

        # handle local model path changes:
        if 'nexa_model' not in st.session_state or st.session_state.current_local_path != local_model_path:
            with st.spinner("Loading local model..."):
                st.session_state.nexa_model = load_local_model(local_model_path)
                st.session_state.current_local_path = local_model_path

    elif model_path == "Use Model From Nexa Model Hub 🔍":
        initial_value = default_model if not is_local_path and not hf else ""
        hub_model_name = st.sidebar.text_input(
            "Enter model name from Nexa Model Hub",
            value=initial_value
        )

        # empty string check:
        if not hub_model_name:
            st.warning("""
            How to add a model from Nexa Model Hub:
            \n1. Visit [Nexa Model Hub](https://nexaai.com/models)
            \n2. Find a vision model using the task filters
            \n3. Select your desired model and copy either:
            \n   - The full nexa run command, or (e.g., nexa run stable-diffusion-v1-4:q4_0)
            \n   - Simply the model name (e.g., stable-diffusion-v1-4:q4_0)
            \n4. Paste it into the field on the sidebar and press enter
            """)
            st.stop()

        # process the input after checking it's not empty:
        if hub_model_name.startswith("nexa run"):
            hub_model_name = hub_model_name.split("nexa run")[-1].strip()
        else:
            hub_model_name = hub_model_name.strip()

        # handle hub model name changes:
        if 'nexa_model' not in st.session_state or st.session_state.current_hub_model != hub_model_name:
            with st.spinner("Loading model from hub..."):
                st.session_state.nexa_model = load_model(hub_model_name)
                if st.session_state.nexa_model:  # only update if load was successful
                    st.session_state.current_hub_model = hub_model_name

    else:
        # load selected model if it's not already loaded:
        if ('nexa_model' not in st.session_state or getattr(st.session_state, 'current_model_path', None) != model_path):
            with st.spinner(f"Loading model: {model_path}"):
                st.session_state.nexa_model = load_model(model_path)
                if st.session_state.nexa_model:  # only update if load was successful
                    st.session_state.current_model_path = model_path

    # generation params:
    if 'nexa_model' in st.session_state and st.session_state.nexa_model:
        st.sidebar.header("Generation Parameters")

        model_to_check = (st.session_state.current_hub_model if st.session_state.current_hub_model else st.session_state.current_local_path if st.session_state.current_local_path else st.session_state.current_model_path)

        # get model specific defaults:
        default_params = get_default_params(model_to_check)

        # adjust step range based on model type:
        max_steps = 100
        if "lcm-dreamshaper" in model_to_check or "flux" in model_to_check:
            max_steps = 8  # 4-8 steps
        elif "sdxl-turbo" in model_to_check:
            max_steps = 10  # 5-10 steps

        # adjust guidance scale range based on model type:
        max_guidance = 20.0
        if "lcm-dreamshaper" in model_to_check or "flux" in model_to_check:
            max_guidance = 2.0  # 1.0-2.0
        elif "sdxl-turbo" in model_to_check:
            max_guidance = 10.0  # 5.0-10.0

        num_inference_steps = st.sidebar.slider(
            "Number of Inference Steps",
            1,
            max_steps,
            default_params["num_inference_steps"]
        )
        height = st.sidebar.slider(
            "Height",
            64,
            1024,
            default_params["height"]
        )
        width = st.sidebar.slider(
            "Width",
            64,
            1024,
            default_params["width"]
        )
        guidance_scale = st.sidebar.slider(
            "Guidance Scale",
            0.0,
            max_guidance,
            default_params["guidance_scale"]
        )
        random_seed = st.sidebar.slider(
            "Random Seed",
            0,
            10000,
            default_params["random_seed"]
        )

        st.session_state.nexa_model.params.update({
            "num_inference_steps": num_inference_steps,
            "height": height,
            "width": width,
            "guidance_scale": guidance_scale,
            "random_seed": random_seed,
        })

    # image generation interface:
    prompt = st.text_input("Enter your prompt:")
    negative_prompt = st.text_input("Enter your negative prompt (optional):")

    if st.button("Generate Image"):
        if not prompt:
            st.warning("Please enter a prompt to proceed.")
        else:
            with st.spinner("Generating images..."):
                images = generate_images(
                    st.session_state.nexa_model,
                    prompt,
                    negative_prompt
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

except Exception as e:
    st.error(f"An unexpected error occurred: {str(e)}")
    import traceback
    st.error(f"Traceback: {traceback.format_exc()}")
