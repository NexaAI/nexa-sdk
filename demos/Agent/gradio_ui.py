# Copyright 2024-2025 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import gradio as gr
import json
from serve import (LLMService, ALL_ASR_MODELS, ALL_INFER_MODELS, BASE_URL)
from agent import AgentRunner
from gradio import ChatMessage

agent = AgentRunner()

def run_task(history, audio, base_url, asr_model, llm_model):
    if history is None:
        history = []

    history.append(
        ChatMessage(
            role="assistant",
            content="",
            metadata={"title": f"**Process audio...**"}
        ))
    yield history, None
    
    try:
        task = LLMService.speech_to_text(base_url=base_url, audio=audio, model=asr_model)
    except Exception as e:
        history.append(ChatMessage(
                        role="assistant",
                        content=f"(Error: {e})",
                        metadata={"title": f"**Error occurred**"},
                    ))
        yield history, None
        return
    
    # task = """
    # give me the time right now, and tell me the weather for New York then send email
    # """
    
    for raw in agent.run(base_url=base_url, task=task, model=llm_model):
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

                if st == "error":
                    history.append(ChatMessage(
                        role="assistant",
                        content=f"(Error: {msg})",
                        metadata={"title": f"**Error occurred**"},
                    ))
                    yield history, None
                    continue
                if st == "function":
                    history.append(ChatMessage(
                        role="assistant",
                        content=f"""
                        ```json
                        {msg}
                        ```
                        """,
                        metadata={"title": f"**Call Tool**"},
                    ))
                    yield history, None
                    continue
                
                if st == "proccess" or st == "task":
                    history.append(ChatMessage(
                        role="assistant",
                        content="",
                        metadata={"title": f"**{msg}**"}
                    ))
                    yield history, None
                    continue
                
                if st == "finished":
                    history.append(ChatMessage(
                        role="assistant",
                        content="",
                        metadata={"title": f"**{msg}**"}
                    ))
                    yield history, None
                    continue

with gr.Blocks() as demo:
    gr.Markdown("## Agent with Nexa serve")
    with gr.Row():
        with gr.Column(scale=2):
            chatbox = gr.Chatbot(height=500)
            audio_input = gr.Audio(
                sources=["microphone"], 
                type="filepath",
                format='wav',
                show_label=False
            )
            
        with gr.Column(scale=1):
            base_url=gr.Textbox(BASE_URL, label="Base URL")
            asr_repo_id = gr.Dropdown(ALL_ASR_MODELS, label="Asr model repo Id", value=ALL_ASR_MODELS[0])
            llm_repo_id = gr.Dropdown(ALL_INFER_MODELS, label="LLM model repo Id", value=ALL_INFER_MODELS[0])
        
    audio_input.stop_recording(fn=run_task, inputs=[chatbox, audio_input, base_url, asr_repo_id, llm_repo_id], outputs=[chatbox, audio_input])

if __name__ == "__main__":
    demo.launch()
