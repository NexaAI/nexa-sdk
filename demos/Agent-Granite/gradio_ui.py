import os
import platform
import subprocess
from typing import List, Tuple

import gradio as gr

from agent_nexa import (
    nexa_start_search
)

# Chat
def chat_search(query: str, history: list):
    # Ensure history is a list
    if history is None:
        history = []

    # Call the agent/search function
    result = nexa_start_search(query)
    history.append([query, result])

    # Return both updated history (for the Chatbot) and an empty string to clear the input
    return history, ""

# UI
CUSTOM_CSS = """
/* Make cards cleaner and add subtle shadows */
.gradio-container { max-width: 1600px !important; }
.rounded-card { border-radius: 16px; box-shadow: 0 1px 8px rgba(0,0,0,.06); background: white; }
.pad { padding: 14px; }
.section-title { font-weight: 700; font-size: 14px; opacity: .8; margin-bottom: 8px; }
#info-panel .gallery { background: #101114; } /* darker bg for images */
"""

with gr.Blocks(title="Web search with NexaSDK", css=CUSTOM_CSS) as demo:
    gr.Markdown("## Web search with NexaSDK")

    with gr.Column(scale=2, elem_classes=["rounded-card", "pad"]):
        chat = gr.Chatbot(height=600, show_copy_button=True)
        chat_input = gr.Textbox(placeholder="Ask something about your documents...", label="Your question")
        with gr.Row():
            btn_send = gr.Button("Send", variant="primary", elem_id="btn-send")
            btn_clear = gr.Button("Clear chat")

    # Events
    def on_clear():
        # Clear chat history and input box
        return [], ""

    btn_clear.click(fn=on_clear, outputs=[chat, chat_input])

    # Chat send (streaming)
    btn_send.click(
        fn=chat_search,
        inputs=[chat_input, chat],
        outputs=[chat, chat_input],
    )
    chat_input.submit(
        fn=chat_search,
        inputs=[chat_input, chat],
        outputs=[chat, chat_input],
    )

if __name__ == "__main__":
    demo.launch()
