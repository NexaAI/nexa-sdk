import os
import platform
import subprocess
from typing import List, Tuple

import gradio as gr
from gradio import ChatMessage
import json

from agent_nexa import (
    nexa_start_search_stream
)

# Chat streaming with Nexa
def chat_stream_search(query: str, history: list):
    """
    Parse JSON lines emitted by backend stream and update chat + input.

    Backend lines are JSON strings with format:
      {"status": "loading", "message": "start query..."}
      {"status": "function", "message": result}
      {"status": "stream", "message": piece}

    We write status messages into the input box so the user sees progress.
    """
    if history is None:
        history = []

    # Append the user's message to history
    history.append(ChatMessage(role="user", content=query))

    yield history, ""

    assistant_appended = False

    # Keep last_message if needed by backend (not used here)
    last_message = history[-2]['content'] if len(history) > 1 else ""

    try:
        for raw in nexa_start_search_stream(query, last_message):
            # raw is expected to be a JSON string
            parsed = None
            if isinstance(raw, str):
                try:
                    parsed = json.loads(raw)
                except Exception:
                    # Not JSON: treat as raw stream chunk
                    parsed = None

            if parsed and isinstance(parsed, dict) and "status" in parsed:
                st = parsed.get("status")
                msg = parsed.get("message", "")

                if st == "function_call_error":
                    history.append(ChatMessage(
                        role="assistant",
                        content=f"{msg}",
                        metadata={"title": f"🛠️ **Function call error**"},
                    ))
                    history.append(
                        ChatMessage(
                            role="assistant",
                            content="I will try again, please wait...\n",
                        )
                    )
                    yield history, ""
                    continue
                    
                if st == "error":
                    history.append(ChatMessage(
                        role="assistant",
                        content=f"(Error: {msg})",
                        metadata={"title": f"❌ **Error occurred**"},
                    ))
                    yield history, ""
                    continue
                
                if st == "proccess":
                    history.append(ChatMessage(
                        role="assistant",
                        content="",
                        metadata={"title": f"🛠️ **{msg}**"}
                    ))
                    yield history, ""
                    continue
                    
                if st == "function":
                    try: 
                        func_desc = json.loads(msg)
                        func_name = func_desc.get("name")
                        content = f"""```json
                        {json.dumps(func_desc)}
                        ```"""
                        history.append(ChatMessage(
                            role="assistant",
                            content=content,
                            metadata={"title": f"🛠️ **Used tool '{func_name}'**"},
                        ))
                    except Exception:
                        history.append(ChatMessage(
                            role="assistant",
                            content=f"{msg}",
                        ))
                    yield history, ""
                    continue

                if st == "stream":
                    piece = msg
                    # append or create assistant message
                    if not assistant_appended:
                        history.append(ChatMessage(role="assistant", content=piece))
                        assistant_appended = True
                    else:
                        # extend last assistant content
                        history[-1].content += piece

                    # clear input while streaming
                    yield history, ""
                    continue

                # unknown status -> show raw message in input
                history.append(ChatMessage(role="assistant", content=str(msg)))
                yield history, ""
                continue

            # fallback: raw chunk string
            if isinstance(raw, str):
                piece = raw
                if not assistant_appended:
                    history.append(ChatMessage(role="assistant", content=piece))
                    assistant_appended = True
                else:
                    history[-1].content += piece
                yield history, ""

    except Exception as e:
        # append error to assistant message
        if assistant_appended:
            history[-1].content += f"\n(Streaming failed: {e})"
        else:
            history.append(ChatMessage(role="assistant", content=f"(Streaming failed: {e})"))
        yield history, ""

# UI
CUSTOM_CSS = """
/* Make cards cleaner and add subtle shadows */
.gradio-container { max-width: 1600px !important; }
.rounded-card { border-radius: 16px; box-shadow: 0 1px 8px rgba(0,0,0,.06); background: white; }
.pad { padding: 14px; }
.section-title { font-weight: 700; font-size: 16px; opacity: .8; margin-bottom: 8px; }
#info-panel .gallery { background: #101114; } /* darker bg for images */
"""

with gr.Blocks(title="Function call with NexaSDK", css=CUSTOM_CSS) as demo:
    gr.Markdown("## Function call with NexaSDK")

    with gr.Column(scale=2, elem_classes=["rounded-card", "pad"]):
        chat = gr.Chatbot(type="messages", height=500, show_copy_button=True)
        chat_input = gr.Textbox(placeholder="Input your query...", label="Your question")
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
        fn=chat_stream_search,
        inputs=[chat_input, chat],
        outputs=[chat, chat_input],
    )
    chat_input.submit(
        fn=chat_stream_search,
        inputs=[chat_input, chat],
        outputs=[chat, chat_input],
    )

if __name__ == "__main__":
    demo.launch()
