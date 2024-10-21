import streamlit as st
from typing import Iterator
from nexa.gguf import NexaTextInference

def generate_chat_response(nexa_model: NexaTextInference) -> Iterator:
    messages = st.session_state.messages
    response = nexa_model.create_chat_completion(
        messages=messages,
        temperature=nexa_model.params["temperature"],
        max_tokens=nexa_model.params["max_new_tokens"],
        top_k=nexa_model.params["top_k"],
        top_p=nexa_model.params["top_p"],
        stream=True,
    )
    return response
