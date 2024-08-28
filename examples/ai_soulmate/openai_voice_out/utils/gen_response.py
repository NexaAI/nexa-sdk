import streamlit as st
from typing import Iterator
from typing import Iterator
from nexa.gguf import NexaTextInference
from pathlib import Path
from openai import OpenAI

client = OpenAI(api_key="YOUR_OPENAI_API_KEY")


def generate_and_play_response(response_text: str):
    speech_file_path = Path(__file__).parent / "speech.mp3"
    response = client.audio.speech.create(
        model="tts-1",
        voice="shimmer",
        input=response_text
    )
    
    response.stream_to_file(speech_file_path)
    return str(speech_file_path)  # Return the path as a string

def generate_chat_response(nexa_model: NexaTextInference) -> Iterator:
    messages = st.session_state.messages
    response = nexa_model.create_chat_completion(
        messages=messages,
        temperature=nexa_model.params["temperature"],
        max_tokens=nexa_model.params["max_new_tokens"],
        top_k=nexa_model.params["top_k"],
        top_p=nexa_model.params["top_p"],
        stream=True
    )
    return response
