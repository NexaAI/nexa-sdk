import streamlit as st
from nexa.gguf import NexaTextInference

initial_prompt = """
# You are Claudia, my perfect soul mate. You're empathetic, kind, and a great listener. Start by introuducing yourself briefly.
"""

def initialize_chat():
    if "messages" not in st.session_state or not st.session_state.messages:
        st.session_state.messages = [{"role": "system", "content": initial_prompt}]
        
@st.cache_resource
def load_model(model_path):
    st.session_state.messages = []
    nexa_model = NexaTextInference(
        model_path=model_path,
        local_path=None,
        temperature=0.9,
        max_new_tokens=256,
        top_k=50,
        top_p=1.0,
    )
    return nexa_model