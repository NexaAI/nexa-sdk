import sys
from typing import Iterator

import streamlit as st
from nexa.general import pull_model
from nexa.gguf.nexa_inference_text import NexaTextInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    st.session_state.messages = []
    local_path, run_type = pull_model(model_path)
    nexa_model = NexaTextInference(model_path=model_path, local_path=local_path)
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
    "Max New Tokens", 1, 500, st.session_state.nexa_model.params["max_new_tokens"]
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
