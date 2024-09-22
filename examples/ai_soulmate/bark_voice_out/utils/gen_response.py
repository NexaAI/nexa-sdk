from typing import List, Iterator
import numpy as np
from nexa.gguf import NexaTextInference
from bark import SAMPLE_RATE, generate_audio, preload_models
from bark.api import semantic_to_waveform, generate_text_semantic
import streamlit as st
import sounddevice as sd

def split_text(text: str, max_length: int = 200) -> List[str]:
    words = text.split()
    chunks = []
    chunk = []
    
    for word in words:
        if len(" ".join(chunk + [word])) > max_length:
            chunks.append(" ".join(chunk))
            chunk = [word]
        else:
            chunk.append(word)
            
    if chunk:
        chunks.append(" ".join(chunk))
        
    return chunks

def generate_and_play_response(response_text: str):
    text_chunks = split_text(response_text)
    
    silence = np.zeros(int(0.25 * SAMPLE_RATE))
    GEN_TEMP = 0.6
    SPEAKER = "v2/en_speaker_9"
    
    pieces = []
    for sentence in text_chunks:
        semantic_tokens = generate_text_semantic(
            sentence,
            history_prompt=SPEAKER,
            temp=GEN_TEMP,
            min_eos_p=0.05
        )
        audio_array = semantic_to_waveform(semantic_tokens, history_prompt=SPEAKER)
        pieces.append(audio_array)
        pieces.append(silence.copy())
        
    combined_audio = np.concatenate(pieces)
    play_audio(SAMPLE_RATE, combined_audio)

def play_audio(sample_rate, audio_array):
    sd.play(audio_array, sample_rate)
    sd.wait()

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
