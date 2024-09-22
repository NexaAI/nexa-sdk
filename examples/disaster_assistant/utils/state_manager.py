from typing import List, Dict, Any
import streamlit as st
import json 

with open('victim_json_template_flat.json', 'r') as f:
    victim_info = json.load(f)

class StateManager:
    def __init__(self):
        # Initialize session state variables
        if "messages" not in st.session_state:
            st.session_state.messages = []
        if "victim_info" not in st.session_state:
            st.session_state.victim_info = victim_info
        if "json_template" not in st.session_state:
            st.session_state['json_template'] = victim_info

    def add_message(self, role: str, content: str) -> None:
        """Add a new message to the chat history."""
        st.session_state.messages.append({"role": role, "content": content})

    def display_messages(self) -> None:
        """Display all messages in the chat history."""
        for message in st.session_state.messages:
            with st.chat_message(message["role"]):
                st.markdown(message["content"])

    def get_last_message(self) -> Dict[str, str]:
        """Get the last message from the chat history."""
        return st.session_state.messages[-1] if st.session_state.messages else {"role": "", "content": ""}

    def clear_messages(self) -> None:
        """Clear all messages from the chat history."""
        st.session_state.messages.clear()

    def get_conversation_history(self) -> str:
        """Get the entire conversation history as a single string."""
        return "\n".join(f"{m['role']}: {m['content']}" for m in st.session_state.messages)

    def update_victim_info(self, new_info: Dict[str, Any]) -> None:
        """Update the victim information in the session state."""
        st.session_state.victim_info.update(new_info)