# !/usr/bin/env python3

import gradio as gr
from gradio import ChatMessage

from typing import Dict, Any
from vlm_viewmodel import (VLMViewModel)

vlm_vm = VLMViewModel()

# Call streaming response
def stream_response(
    history: list,
    message: Dict[str, Any],
    system_prompt: str,
    repo_id: str,
    plugin_id: str,
    device_id: str,
    max_tokens: int, 
    temperature: float, 
    top_p: float
):
    """Handle send button click event."""
    
    images = []
    prompt = ""
    for file in message["files"]:
        history.append(ChatMessage(content=gr.Image(file), role="user"))
        images.append(file)
    if message["text"] is not None:
        history.append(ChatMessage(content=message["text"], role="user"))
        prompt = message["text"]
        
    yield history, gr.MultimodalTextbox(value=None, interactive=False)
    
    history.append(ChatMessage(content="Thinking...", role="assistant"))
    
    yield history, None
    
    try:
        vlm_vm.create(
            repo_id=repo_id,
            plugin_id=plugin_id,
            device=device_id,
            system_prompt=system_prompt
        )
    
        history[-1].content = "```json\n"
        for token in vlm_vm.stream_response(
            prompt=prompt, 
            images=images, 
            max_tokens=max_tokens, 
            temperature=temperature, 
            top_p=top_p
        ):
            history[-1].content += token
            yield history, None
    
    except Exception as e:
        history[-1].content += f"\n\n[Error occurred: {str(e)}]"
        yield history, None

# Build Gradio UI
with gr.Blocks(title="VLM Example with Nexa Python Binding", fill_height=True) as demo:
    gr.Markdown("## VLM Example with Nexa Python Binding")

    with gr.Row(equal_height=True):
        with gr.Column(scale=1, elem_classes=["rounded-card", "pad"]):
            system_prompt = gr.Textbox(
                label="System Prompt", 
                placeholder="Enter system prompt here...",
                value="", 
                lines=20, 
                interactive=True,
            )
            
        with gr.Column(scale=4):
            chatbot = gr.Chatbot(
                type='messages', 
                show_copy_button=True,
                height=500
            )
            
            # Input area with chat input 
            chat_input = gr.MultimodalTextbox(
                interactive=True,
                placeholder="Enter message...",
                file_types=["image"],
                file_count="multiple",
                show_label=False,
            )
        with gr.Column(scale=1):
            with gr.Row():
                default_model = vlm_vm.models[0]
                all_models_repids = [model.repo_id for model in vlm_vm.models]
                repo_id = gr.Dropdown(all_models_repids, label="Model repo Id", value=default_model.repo_id)
                plugin_id = gr.Dropdown(default_model.plugin_ids, label="Plugin Id", value=default_model.plugin_ids[0], interactive=True)
                device = gr.Dropdown(default_model.devices, label="Device", value=default_model.devices[0], interactive=True)
                
                temperature = gr.Slider(0.0, 1.0, value=0.8, step=0.05, label="Temperature", interactive=True)
                max_tokens = gr.Slider(200, 65535, value=200, step=10, label="Max Tokens", interactive=True)
                top_p = gr.Slider(0.0, 1.0, value=0.95, step=0.05, label="Top-p", interactive=True)
    
        def on_model_repo_change(repo_id):
            vlm_vm.reset()
            selected_model = next((model for model in vlm_vm.models if model.repo_id == repo_id), None)
            if selected_model:
                return [
                    gr.update(
                        choices=selected_model.plugin_ids,
                        value=selected_model.plugin_ids[0]
                    ), 
                    gr.update(
                        choices=selected_model.devices,
                        value=selected_model.devices[0]
                    ), 
                    []
                ]
            
            return [
                gr.update(
                    choices=["nexaml"],
                    value="nexaml"
                ), 
                gr.update(
                    choices=["gpu"],
                    value="gpu"
                ),
                []
            ]
           
        def on_system_prompt_change(new_prompt):
            vlm_vm.update_system_prompt(new_prompt)
         
        def on_id_change():
            vlm_vm.reset
            
        system_prompt.change(
            fn=on_system_prompt_change,
            inputs=system_prompt,
            outputs=[]
        )
        
        repo_id.change(
            fn=on_model_repo_change,
            inputs=repo_id,
            outputs=[plugin_id, device, chatbot]
        )
        
        plugin_id.change(
            fn=on_id_change
        )
        
        device.change(
            fn=on_id_change
        )
        
        stream = chat_input.submit(
            fn=stream_response, 
            inputs=[chatbot, chat_input, system_prompt, repo_id, plugin_id, device, max_tokens, temperature, top_p],
            outputs=[chatbot, chat_input]
        )
        
        stream.then(lambda: gr.MultimodalTextbox(interactive=True), None, [chat_input])


if __name__ == "__main__":
	demo.launch()

