import sys
import subprocess
import re
from typing import Iterator, List
import streamlit as st
from nexa.general import pull_model
from nexa.gguf.nexa_inference_text import NexaTextInference

# init:
DEFAULT_PARAMS = {
    "temperature": 0.9,
    "max_new_tokens": 256,
    "top_k": 50,
    "top_p": 1.0,
    "nctx": 2048
}

def parse_nexa_list_output() -> List[str]:
    """Parse the output of nexa list command to get local NLP models."""
    try:
        result = subprocess.run(['nexa', 'list'], capture_output=True, text=True)
        pattern = r'\|\s*(.*?)\s*\|\s*(.*?)\s*\|\s*(.*?)\s*\|\s*(.*?)\s*\|'
        matches = re.findall(pattern, result.stdout)

        # filter for Run Type == NLP models and get their names:
        nlp_models = set()  # to avoid duplicates
        for match in matches:
            model_name = match[0].strip()
            run_type = match[2].strip()
            if run_type == 'NLP' and not model_name.startswith('Model Name'):
                nlp_models.add(model_name)

        return sorted(list(nlp_models))  # convert to sorted list for consistent ordering
    except Exception as e:
        st.error(f"Error parsing nexa list output: {str(e)}")
        return []

def get_model_options() -> List[str]:
    """Get list of model options from nexa list."""
    models = parse_nexa_list_output()
    # add special options at the end of the dropdown menu:
    models.extend(["Use Model From Nexa Model Hub 🔍", "Local Model 📁"])
    return models

def update_model_options():
    """Update the model options in session state and force a refresh."""
    # refresh list of models:
    fresh_options = get_model_options()

    # update session state with new options:
    st.session_state.model_options = fresh_options

    # if we have a current model path, ensure it's in the list:
    if hasattr(st.session_state, 'current_model_path') and st.session_state.current_model_path:
        if st.session_state.current_model_path in fresh_options:
            st.session_state.current_model_index = fresh_options.index(st.session_state.current_model_path)
        else:
            # if current model not in list, reset to Model Hub option:
            hub_index = fresh_options.index("Use Model From Nexa Model Hub 🔍")
            st.session_state.current_model_index = hub_index

@st.cache_resource(show_spinner=False)
def load_model(model_path: str, is_local: bool = False, is_hf: bool = False):
    """Load model with proper error handling and state management."""
    try:
        st.session_state.messages = []

        if is_local:
            local_path = model_path
        elif is_hf:
            try:
                local_path, _ = pull_model(model_path, hf=True)
                update_model_options()  # update options after successful pull
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
                update_model_options()  # update options after successful pull
            except ValueError as e:
                st.error(f"Error pulling model from Nexa Model Hub: {str(e)}")
                return None
            except Exception as e:
                st.error(f"Unexpected error while pulling model: {str(e)}")
                return None

        # create the model:
        try:
            nexa_model = NexaTextInference(
                model_path=model_path,
                local_path=local_path,
                **DEFAULT_PARAMS
            )

            # force refresh of model options after successful load:
            update_model_options()

            # reset the model index to include the new model:
            if model_path in st.session_state.model_options:
                st.session_state.current_model_index = st.session_state.model_options.index(model_path)

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
        st.session_state.messages = []
        nexa_model = NexaTextInference(
            model_path="local_model",
            local_path=local_path,
            **DEFAULT_PARAMS
        )
        update_model_options()  # update options after successful local model load
        return nexa_model
    except Exception as e:
        st.error(f"Error loading local model: {str(e)}")
        return None

def generate_response(nexa_model: NexaTextInference) -> Iterator:
    """Generate response from the model."""
    user_input = st.session_state.messages[-1]["content"]
    if hasattr(nexa_model, "chat_format") and nexa_model.chat_format:
        return nexa_model._chat(user_input)
    else:
        return nexa_model._complete(user_input)

# main execution:
try:
    # get command line arguments with proper error handling:
    if len(sys.argv) < 4:
        st.error("Missing required command line arguments.")
        sys.exit(1)

    default_model = sys.argv[1]
    is_local_path = sys.argv[2].lower() == "true"
    hf = sys.argv[3].lower() == "true"

    # UI setup:
    st.set_page_config(page_title="Nexa AI Text Generation", layout="wide")
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
    st.title("Nexa AI Text Generation")
    st.caption("Powered by Nexa AI SDK🐙")

    # force refresh model options on every page load:
    if 'model_options' not in st.session_state:
        st.session_state.model_options = get_model_options()
    else:
        update_model_options()

    # init session state variables:
    if 'initialized' not in st.session_state:
        st.session_state.messages = []
        st.session_state.current_model_path = None
        st.session_state.current_local_path = None
        st.session_state.current_hub_model = None

        if not is_local_path and not hf:
            try:
                # try to load the model first:
                with st.spinner(f"Loading model: {default_model}"):
                    st.session_state.nexa_model = load_model(default_model)
                    if st.session_state.nexa_model:
                        st.session_state.current_hub_model = default_model
            except Exception as e:
                st.error(f"Error loading default model: {str(e)}")

            # set to model hub option if not found in list:
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

    # Update the selectbox index based on the currently loaded model
    if 'nexa_model' in st.session_state:
        if st.session_state.current_hub_model:
            # If we have a hub model loaded, select the hub option
            current_index = st.session_state.model_options.index("Use Model From Nexa Model Hub 🔍")
        elif st.session_state.current_local_path:
            # If we have a local model loaded, select the local option
            current_index = st.session_state.model_options.index("Local Model 📁")
        elif st.session_state.current_model_path:
            # If we have a listed model loaded, find its index
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

    # Update current model index when selection changes:
    current_index = st.session_state.model_options.index(model_path)
    if current_index != st.session_state.current_model_index:
        st.session_state.current_model_index = current_index
        # Clear the current model to force reload:
        if 'nexa_model' in st.session_state:
            del st.session_state.nexa_model
            st.session_state.messages = []
            # Also clear other model path variables
            st.session_state.current_model_path = None
            st.session_state.current_local_path = None
            st.session_state.current_hub_model = None

    # handle model loading based on selection:
    if model_path == "Local Model 📁":
        local_model_path = st.sidebar.text_input("Enter local model path")
        if not local_model_path:
            st.warning("Please enter a valid local model path to proceed.")
            st.stop()

        local_model_path = local_model_path.strip()  # remove spaces
        if 'nexa_model' not in st.session_state or st.session_state.current_local_path != local_model_path:
            with st.spinner("Loading local model..."):
                st.session_state.nexa_model = load_local_model(local_model_path)
                st.session_state.current_local_path = local_model_path

    elif model_path == "Use Model From Nexa Model Hub 🔍":
        # pre-fill with default_model if it's a Nexa Hub model:
        initial_value = default_model if not is_local_path and not hf else ""
        hub_model_name = st.sidebar.text_input("Enter model name from Nexa Model Hub",
                                             value=initial_value)

        # empty string check:
        if not hub_model_name:
            st.warning(f"""
            How to add a model from Nexa Model Hub:
            \n1. Visit [Nexa Model Hub](https://nexaai.com/models)
            \n2. Find a model using the task filters (chat, uncensored, etc.)
            \n3. Select your desired model and copy either:
            \n   - The full nexa run command (e.g., nexa run Sao10K/MN-BackyardAI-Party-12B-v1:gguf-q4_K_M), or
            \n   - Simply the model name (e.g., Sao10K/MN-BackyardAI-Party-12B-v1:gguf-q4_K_M)
            \n4. Paste it into the "Enter model name from Nexa Model Hub" field on the sidebar and press enter
            """)
            st.stop()

        # process the input after checking it's not empty:
        if hub_model_name.startswith("nexa run"):
            hub_model_name = hub_model_name.split("nexa run")[-1].strip()
        else:
            hub_model_name = hub_model_name.strip()

        if 'nexa_model' not in st.session_state or st.session_state.current_hub_model != hub_model_name:
            with st.spinner("Loading model from hub..."):
                st.session_state.nexa_model = load_model(hub_model_name)
                if st.session_state.nexa_model:  # only update if load was successful
                    st.session_state.current_hub_model = hub_model_name

    else:
        # load selected model if it's not already loaded:
        if ('nexa_model' not in st.session_state or
            getattr(st.session_state, 'current_model_path', None) != model_path):
            with st.spinner(f"Loading model: {model_path}"):
                st.session_state.nexa_model = load_model(model_path)
                if st.session_state.nexa_model:  # only update if load was successful
                    st.session_state.current_model_path = model_path

    # generation params:
    if 'nexa_model' in st.session_state and st.session_state.nexa_model:
        st.sidebar.header("Generation Parameters")
        model_params = st.session_state.nexa_model.params

        temperature = st.sidebar.slider(
            "Temperature", 0.0, 1.0, model_params.get("temperature", DEFAULT_PARAMS["temperature"])
        )
        max_new_tokens = st.sidebar.slider(
            "Max New Tokens", 1, 500, model_params.get("max_new_tokens", DEFAULT_PARAMS["max_new_tokens"])
        )
        top_k = st.sidebar.slider(
            "Top K", 1, 100, model_params.get("top_k", DEFAULT_PARAMS["top_k"])
        )
        top_p = st.sidebar.slider(
            "Top P", 0.0, 1.0, model_params.get("top_p", DEFAULT_PARAMS["top_p"])
        )
        nctx = st.sidebar.slider(
            "Context length", 1000, 9999, model_params.get("nctx", DEFAULT_PARAMS["nctx"])
        )

        st.session_state.nexa_model.params.update({
            "temperature": temperature,
            "max_new_tokens": max_new_tokens,
            "top_k": top_k,
            "top_p": top_p,
            "nctx": nctx,
        })

    # chat interface:
    for message in st.session_state.messages:
        with st.chat_message(message["role"]):
            st.markdown(message["content"])

    if prompt := st.chat_input("Say something..."):
        if 'nexa_model' not in st.session_state or not st.session_state.nexa_model:
            st.error("Please wait for the model to load or select a valid model.")
        else:
            st.session_state.messages.append({"role": "user", "content": prompt})
            with st.chat_message("user"):
                st.markdown(prompt)

            with st.chat_message("assistant"):
                response_placeholder = st.empty()
                full_response = ""
                for chunk in generate_response(st.session_state.nexa_model):
                    choice = chunk["choices"][0]
                    if "delta" in choice:
                        delta = choice["delta"]
                        content = delta.get("content", "")
                    elif "text" in choice:
                        delta = choice["text"]
                        content = delta

                    full_response += content
                    response_placeholder.markdown(full_response, unsafe_allow_html=True)
                response_placeholder.markdown(full_response)

            st.session_state.messages.append({"role": "assistant", "content": full_response})

except Exception as e:
    st.error(f"An unexpected error occurred: {str(e)}")
    import traceback
    st.error(f"Traceback: {traceback.format_exc()}")
