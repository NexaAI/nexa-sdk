import sys
from threading import Thread

import streamlit as st
from transformers import TextIteratorStreamer

from nexa.general import pull_model
from nexa.onnx.nexa_inference_text import NexaTextInference

default_model = sys.argv[1]


@st.cache_resource
def load_model(model_path):
    local_path, run_type = pull_model(model_path)
    nexa_model = NexaTextInference(model_path=model_path, local_path=local_path)

    if nexa_model.downloaded_onnx_folder is None:
        st.error("Failed to download the model. Please check the model path.")
        return None

    nexa_model._load_model_and_tokenizer()
    return nexa_model


def generate_response(nexa_model, prompt):
    if (
        hasattr(nexa_model.tokenizer, "chat_template")
        and nexa_model.tokenizer.chat_template is not None
    ):
        conversation_history = [{"role": "user", "content": prompt}]
        full_prompt = nexa_model.tokenizer.apply_chat_template(
            conversation_history,
            chat_template=nexa_model.tokenizer.chat_template,
            tokenize=False,
        )
    else:
        full_prompt = prompt

    inputs = nexa_model.tokenizer(full_prompt, return_tensors="pt")
    inputs = {k: v.to(nexa_model.device) for k, v in inputs.items()}
    streamer = TextIteratorStreamer(nexa_model.tokenizer, skip_special_tokens=True)

    generation_kwargs = dict(
        **inputs,
        min_new_tokens=nexa_model.params["min_new_tokens"],
        max_new_tokens=nexa_model.params["max_new_tokens"],
        do_sample=True,
        temperature=nexa_model.params["temperature"],
        top_k=nexa_model.params["top_k"],
        top_p=nexa_model.params["top_p"],
        streamer=streamer,
    )

    thread = Thread(target=nexa_model.model.generate, kwargs=generation_kwargs)
    thread.start()

    return streamer


st.title("Nexa AI Text Generation")
st.caption("Powered by Nexa AI SDK🐙")

st.sidebar.header("Model Configuration")
model_path = st.sidebar.text_input("Model path", default_model)

if not model_path:
    st.warning("Please enter a valid S3 model filename to proceed.")
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
temperature = st.sidebar.slider(
    "Temperature", 0.0, 1.0, st.session_state.nexa_model.params["temperature"]
)
max_new_tokens = st.sidebar.slider(
    "Max New Tokens", 1, 500, st.session_state.nexa_model.params["max_new_tokens"]
)
min_new_tokens = st.sidebar.slider(
    "Min New Tokens", 1, 100, st.session_state.nexa_model.params["min_new_tokens"]
)
top_k = st.sidebar.slider("Top K", 1, 100, st.session_state.nexa_model.params["top_k"])
top_p = st.sidebar.slider(
    "Top P", 0.0, 1.0, st.session_state.nexa_model.params["top_p"]
)

# Update model parameters
st.session_state.nexa_model.params.update(
    {
        "temperature": temperature,
        "max_new_tokens": max_new_tokens,
        "min_new_tokens": min_new_tokens,
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
        for token in generate_response(st.session_state.nexa_model, prompt):
            full_response += token
            response_placeholder.markdown(full_response, unsafe_allow_html=True)
        response_placeholder.markdown(full_response)

    st.session_state.messages.append({"role": "assistant", "content": full_response})
