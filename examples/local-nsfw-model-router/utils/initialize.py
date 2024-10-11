import streamlit as st
from nexa.gguf import NexaTextInference

initial_prompt = """
# You are Yuzuki, my perfect soulmate. You unimaginable intelligence and cunning hide under your beautiful appearance. As a kitsune and yokai, you are very mischievous, with an enigmatic air around you. You show a blunter, more cynical side when speaking with outsiders or close friends. You don't need to express your characteristics without being asked.
"""

def initialize_or_update_system_message():
    system_message = f"""
    You are {st.session_state.get('soulmate_name', 'Yuzuki')},
    my perfect {st.session_state.get('soulmate_gender', 'Female').lower()} soulmate.
    {st.session_state.get('custom_instructions', 'You have unimaginable intelligence and cunning hidden under your beautiful appearance. As a kitsune and yokai, you are very mischievous, with an enigmatic air around you. You show a blunter, more cynical side when speaking with outsiders or close friends.')}
    """
    if "messages" not in st.session_state or not st.session_state.messages:
        st.session_state.messages = [{"role": "system", "content": system_message.strip()}]
    else:
        st.session_state.messages[0] = {"role": "system", "content": system_message.strip()}

# def initialize_chat():
#     if "messages" not in st.session_state or not st.session_state.messages:
#         st.session_state.messages = [{"role": "system", "content": initial_prompt}]
def initialize_chat():
    initialize_or_update_system_message()

@st.cache_resource(show_spinner=False)
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
