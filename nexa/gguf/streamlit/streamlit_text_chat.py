import sys
from typing import Iterator
import streamlit as st
from nexa.general import pull_model
from nexa.gguf.nexa_inference_text import NexaTextInference

default_model = sys.argv[1]
is_local_path = False if sys.argv[2] == "False" else True
hf = False if sys.argv[3] == "False" else True

model_mapping = {
    "llama3-uncensored": "llama3-uncensored",
    "Llama-3SOME-8B-v2": "TheDrummer/Llama-3SOME-8B-v2:gguf-q4_K_M",
    "Rocinante-12B-v1.1": "TheDrummer/Rocinante-12B-v1.1:gguf-q4_K_M",
}

model_options = list(model_mapping.keys()) + ["Use Model From Nexa Model Hub 🔍","Local Model 📁"]

@st.cache_resource
def load_model(model_path):
    st.session_state.messages = []
    if is_local_path:
        local_path = model_path
    elif hf:
        local_path, _ = pull_model(model_path, hf=True)
    else:
        local_path, run_type = pull_model(model_path)
    nexa_model = NexaTextInference(model_path=model_path, local_path=local_path)
    return nexa_model

@st.cache_resource(show_spinner=False)
def load_local_model(local_path):
    st.session_state.messages = []
    nexa_model = NexaTextInference(
        model_path="llama3-uncensored",
        local_path=local_path,
        temperature=0.9,
        max_new_tokens=256,
        top_k=50,
        top_p=1.0,
    )
    return nexa_model

def generate_response(nexa_model: NexaTextInference) -> Iterator:
    user_input = st.session_state.messages[-1]["content"]
    if hasattr(nexa_model, "chat_format") and nexa_model.chat_format:
        return nexa_model._chat(user_input)
    else:
        return nexa_model._complete(user_input)

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

st.sidebar.header("Model Configuration")
model_path = st.sidebar.selectbox("Select a Model", model_options, index=model_options.index(default_model))

# if not model_path:
#     st.warning(
#         "Please enter a valid path or identifier for the model in Nexa Model Hub to proceed."
#     )
#     st.stop()

if model_path == "Local Model 📁":
    local_model_path = st.sidebar.text_input("Enter local model path")
    if not local_model_path:
        st.warning("Please enter a valid local model path to proceed.")
        st.stop()
    hub_model_name = None
elif model_path == "Use Model From Nexa Model Hub 🔍":
    hub_model_name = st.sidebar.text_input("Enter model name from Nexa Model Hub")
    if hub_model_name:
        if hub_model_name.startswith("nexa run"):
            hub_model_name = hub_model_name.split("nexa run")[-1].strip()
        else:
            hub_model_name = hub_model_name.strip()
    if not hub_model_name:
        st.warning(f"""
            How to Explore and Run Models on Nexa Model Hub:
            \n1. Visit [Nexa Model Hub](https://nexaai.com/models).
            \n2. Use the task filter to find models suitable for your use case (e.g., "uncensored").
            \n3. Select your desired model (e.g., Sao10K/MN-BackyardAI-Party-12B-v1).
            \n4. Copy the Nexa run command (e.g., nexa run Sao10K/MN-BackyardAI-Party-12B-v1:gguf-q4_K_M).
            \n5. Paste the command into the "Enter model name from Nexa Model Hub" field on the sidebar on the left.
            \n6. Wait a few moments for the new model to download and load into the chat.
            \nNote: Different quantization options (like q4_K_M in the example) affect the model's performance and resource usage. Choose the option that best suits your hardware capabilities and performance requirements.
        """)
        st.stop()
    local_model_path = None
else:
    local_model_path = None
    hub_model_name = None

if (
    "current_model_path" not in st.session_state
    or st.session_state.current_model_path != model_path
    or (
        model_path == "Local Model 📁"
        and local_model_path != st.session_state.current_local_model_path
    )
    or (
        model_path == "Use Model From Nexa Model Hub 🔍"
        and hub_model_name != st.session_state.current_hub_model_name
    )
):
    st.session_state.current_model_path = model_path
    st.session_state.current_local_model_path = local_model_path
    st.session_state.current_hub_model_name = hub_model_name
    with st.spinner(
        "Hang tight! Loading model, you can check the progress in the terminal. I'll be right back with you : )"
    ):
        if model_path == "Local Model 📁" and local_model_path:
            st.session_state.nexa_model = load_local_model(local_model_path)
        elif model_path == "Use Model From Nexa Model Hub 🔍" and hub_model_name:
            st.session_state.nexa_model = load_model(hub_model_name)
        else:
            st.session_state.nexa_model = load_model(model_mapping[model_path])
    st.session_state.messages = []

    if "intro_sent" in st.session_state:
        del st.session_state["intro_sent"]

if not model_path:
    st.warning(
        "Please enter a valid path or identifier for the model in Nexa Model Hub to proceed."
    )
    st.stop()

st.sidebar.header("Generation Parameters")
temperature = st.sidebar.slider(
    "Temperature", 0.0, 1.0, st.session_state.nexa_model.params["temperature"]
)
max_new_tokens = st.sidebar.slider(
    "Max New Tokens", 1, 500, st.session_state.nexa_model.params["max_new_tokens"]
)
top_k = st.sidebar.slider("Top K", 1, 100, st.session_state.nexa_model.params["top_k"])
top_p = st.sidebar.slider(
    "Top P", 0.0, 1.0, st.session_state.nexa_model.params["top_p"]
)
nctx = st.sidebar.slider(
    "Context length", 1000, 9999, st.session_state.nexa_model.params["nctx"]
)

st.session_state.nexa_model.params.update(
    {
        "temperature": temperature,
        "max_new_tokens": max_new_tokens,
        "top_k": top_k,
        "top_p": top_p,
        "Context length": nctx,
    }
)

if "messages" not in st.session_state:
    st.session_state.messages = []

for message in st.session_state.messages:
    with st.chat_message(message["role"]):
        st.markdown(message["content"])

if prompt := st.chat_input("Say something..."):
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
