import streamlit as st
import base64
from utils.transcribe import record_and_transcribe
from utils.gen_response import generate_chat_response, generate_and_play_response
import datetime
from utils.state_manager import StateManager
from utils.initialize import initialize_chat, load_model, display_victim_info
from utils.validate_json import process_and_upload_victim_info, update_victim_json
from utils.json_cleaner import upload_victim_info, process_json_response
import logging
import json

# Function to process JSON response
def process_json_response(response: str) -> None:
    try:
        process_and_upload_victim_info(update_victim_json(response), schema=st.session_state['json_template'])
    except Exception as e:
        logger.error(f"Error processing JSON response: {e}")
        st.warning("Error processing the response.")

logger = logging.getLogger(__name__)


state_manager = StateManager()


ai_avatar = "utils/avatar.jpg"
#default_model = "gemma-1.1-2b-instruct:q4_0"
#default_model = "Octopus-v2:fp16"
default_model = "llama3-uncensored"
def main():

 # create 3 columns
    left, middle, right = st.columns([.5, .1, .4])
    # Chat input and display
    with left:
        st.title("Natural Disaster Rescue Assistant")
    
        st.caption("NexaAI x SafeGuardianAI")

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
            "Max New Tokens", 1, 1000, st.session_state.nexa_model.params["max_new_tokens"]
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
                "stop": ["<end_of_turn>"]
            }
        )

        initialize_chat()
        for message in st.session_state.messages:
            if message["role"] != "system":
                if message["role"] == "user":
                    with st.chat_message(message["role"]):
                        st.markdown(message["content"])
                else:
                    with st.chat_message(message["role"], avatar=ai_avatar):
                        st.markdown(message["content"])

        if st.button("🎙️ Start Voice Chat"):
            transcribed_text = record_and_transcribe()
            if transcribed_text:
                # First AI run for JSON extraction
                json_prompt = transcribed_text + 'You are a disaster assistant. Your task is to return a JSON of the provided information. Only return JSON based on provided elements, don\'t return unknown key-value pairs. Return only the JSON, no additional comments or text:'
                st.session_state.messages.append({"role": "user", "content": json_prompt})
                
                json_response = ""
                history = []
                for chunk in generate_chat_response(st.session_state.nexa_model):
                    choice = chunk["choices"][0]
                    content = choice.get("delta", {}).get("content") or choice.get("text", "")
                    json_response += content
                    # Check if the end token is in the content

                st.session_state['json_response'] = json_response

                st.session_state.messages.append({"role": "assistant", "content": json_response})

                # Process the JSON response
                try:
                    process_json_response(json_response)
                    # history.append(json_response)
                except Exception as e:
                    st.warning(f"Error processing JSON: {e}")

                history.append(json_response)
                history_str = json.dumps(history)

                # Second AI run for the actual response
                response_prompt = f"Given the history {history_str} and the most recent response: {json_response}\n\nAssist the user's with the adequate response as a disaster rescue assistant: {transcribed_text}"
                st.session_state.messages.append({"role": "user", "content": response_prompt})

                with st.chat_message("user"):
                    st.markdown(transcribed_text)

                with st.chat_message("assistant", avatar=ai_avatar):
                    response_placeholder = st.empty()
                    full_response = ""
                    for chunk in generate_chat_response(st.session_state.nexa_model):
                        choice = chunk["choices"][0]
                        content = choice.get("delta", {}).get("content") or choice.get("text", "")
                        full_response += content
                        response_placeholder.markdown(full_response, unsafe_allow_html=True)
                    response_placeholder.markdown(full_response)

                st.session_state.messages.append({"role": "assistant", "content": full_response})

                # Generate and play audio response
                audio_path = generate_and_play_response(full_response)
                
                with open(audio_path, "rb") as audio_file:
                    audio_bytes = audio_file.read()
                    audio_base64 = base64.b64encode(audio_bytes).decode("utf-8")
                st.markdown(f"""
                    <audio autoplay>
                        <source src="data:audio/mp3;base64,{audio_base64}" type="audio/mp3">
                    </audio>
                """, unsafe_allow_html=True)

                st.session_state.messages.append({"role": "assistant", "content": full_response})

        if prompt := st.chat_input("Say something..."):
            st.session_state.messages.append({"role": "user", "content": prompt})
            with st.chat_message("user"):
                st.markdown(prompt)

            with st.chat_message("assistant", avatar=ai_avatar):
                response_placeholder = st.empty()
                full_response = ""
                for chunk in generate_chat_response(st.session_state.nexa_model):
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

            try:
                audio_path = generate_and_play_response(full_response)
            except Exception as e:
                st.warning(f"Error generating audio: {e}")
            with open(audio_path, "rb") as audio_file:
                audio_bytes = audio_file.read()
                audio_base64 = base64.b64encode(audio_bytes).decode("utf-8")
            st.markdown(f"""
                <audio autoplay>
                    <source src="data:audio/mp3;base64,{audio_base64}" type="audio/mp3">
                </audio>
            """, unsafe_allow_html=True)
            st.session_state.messages.append({"role": "assistant", "content": full_response})

            try:
                process_json_response(full_response)
            except Exception as e:
                st.warning(f"Error processing JSON: {e}")

    with right:
        # with st.container(height=300):
        #     st.subheader("Parsed JSON")
        #     st.markdown(st.session_state['json_response'])

        with st.container(height=800):
            st.subheader("Full JSON")
            display_victim_info()

          



if __name__ == "__main__":
    main()