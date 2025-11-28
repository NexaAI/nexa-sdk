
import gradio as gr
import json
from serve import LLMService
from agent import AgentRunner
from gradio import ChatMessage

agent = AgentRunner()

def run_task(history, audio):
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
        task = LLMService.speech_to_text(audio)
    except Exception as e:
        history.append(ChatMessage(
                        role="assistant",
                        content=f"(Error: {e})",
                        metadata={"title": f"**Error occurred**"},
                    ))
        yield history, None
        return
    
    print(task)
    # task = """
    # give me the time right now, and tell me the weather for New York then send email to Mengsheng
    # """
    
    for raw in agent.run(task):
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
    with gr.Column(scale=2):
        chatbox = gr.Chatbot(height=500)
        audio_input = gr.Audio(
            sources=["microphone"], 
            type="filepath",
            format='wav',
            show_label=False
        )
    audio_input.stop_recording(fn=run_task, inputs=[chatbox, audio_input], outputs=[chatbox, audio_input])

if __name__ == "__main__":
    demo.launch()