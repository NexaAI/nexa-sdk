
# Fix for ARM64 Windows matplotlib compatibility issue
import os
import sys

# Disable Gradio's matplotlib backend manager on ARM64 Windows
if sys.platform == "win32" and os.environ.get("PROCESSOR_ARCHITECTURE") == "ARM64":
    os.environ["MPLBACKEND"] = "Agg"
    os.environ["_GRADIO_SKIP_MATPLOTLIB_MANAGER"] = "1"

import platform
import subprocess
from typing import List, Tuple

import gradio as gr

from rag_nexa import (
    DEFAULT_MODEL, DEFAULT_ENDPOINT, DEFAULT_INDEX_JSON, DEFAULT_EMBED_MODEL,
    build_json_index, load_json_index, search_numpy, call_nexa_chat
)

DOCS_DIR_DEFAULT = "./docs"

# Helpers
def ensure_docs_dir(path: str) -> None:
    if not os.path.exists(path):
        os.makedirs(path, exist_ok=True)

def save_uploaded_files(files: List[gr.File], dest_dir: str) -> int:
    if not files:
        return 0
    n = 0
    for f in files:
        src = getattr(f, "name", None) or f
        if not src:
            continue
        filename = os.path.basename(src)
        out_path = os.path.join(dest_dir, filename)
        with open(src, "rb") as r, open(out_path, "wb") as w:
            w.write(r.read())
        n += 1
    return n

def open_folder(path: str) -> Tuple[bool, str | None]:
    try:
        system = platform.system()
        if system == "Windows":
            subprocess.Popen(["explorer", os.path.normpath(path)])
        elif system == "Darwin":
            subprocess.Popen(["open", path])
        else:
            subprocess.Popen(["xdg-open", path])
        return True, None
    except Exception as e:
        return False, str(e)


# Build / Rebuild pipeline (kept API shape; internally builds JSON and loads to memory)
def do_rebuild(docs_dir: str, k: int, chunk_size: int, chunk_overlap: int,
               _model: str, endpoint: str):
    """
    Return:
      index (dict) – in-memory JSON index (NumPy arrays inside)
      status (str)
    """
    ensure_docs_dir(docs_dir)
    # Index path resides inside docs folder to keep UX simple
    index_json_path = os.path.join(docs_dir, os.path.basename(DEFAULT_INDEX_JSON))
    # Build JSON index then load
    try:
        n_docs, n_chunks = build_json_index(
            docs_dir, index_json_path, DEFAULT_EMBED_MODEL, chunk_size, chunk_overlap, endpoint
        )
        index = load_json_index(index_json_path)
        status = f"Indexed text chunks: {n_chunks} (from {n_docs} files)."
    except Exception as e:
        index = None
        status = f"Failed to index: {e}"
    return index, status


# Chat (streaming) with NumPy search
def chat_stream(message: str,
                history: list,
                index,
                model: str,
                endpoint: str,
                k: int):
    if index is None:
        yield history + [[message, "Index is empty. Upload & Rebuild first."]]
        return

    # NumPy cosine search
    try:
        top_idx, top_sims = search_numpy(message, index, DEFAULT_EMBED_MODEL, endpoint, top_k=int(k))
    except Exception as e:
        yield history + [[message, f"(Search failed: {e})"]]
        return

    # Compose context
    context_text = "\n\n".join(index["texts"][i] for i in top_idx.tolist())

    messages = [
        {
            "role": "system",
            "content": (
                "You are a careful assistant. Use ONLY the provided context to answer.\n\n"
                f"<context>\n{context_text}\n</context>"
            ),
        },
        {"role": "user", "content": message},
    ]

    response = ""

    # create assistant turn in chat
    yield history + [[message, response]]

    try:
        for piece in call_nexa_chat(model, messages, endpoint, stream=True):
            response += piece or ""
            yield history + [[message, response]]
    except Exception as e:
        # non-stream fallback
        try:
            response = call_nexa_chat(model, messages, endpoint, stream=False) or ""
            yield history + [[message, response]]
        except Exception as e2:
            yield history + [[message, f"(Generation failed: {e2})"]]


# UI
with gr.Blocks(title="RAG System") as demo:
    gr.Markdown("## RAG System")

    with gr.Row():
        with gr.Column(scale=1):
            gr.Markdown("**Data & Settings**")
            docs_dir = gr.Textbox(label="Docs folder", value=DOCS_DIR_DEFAULT)
            btn_open = gr.Button("Open docs folder")

            uploader = gr.Files(
                label="Upload files (txt/pdf/docx)",
                file_types=[".txt", ".pdf", ".docx"],
                file_count="multiple",
            )

            model = gr.Textbox(label="Model", value=DEFAULT_MODEL)
            endpoint = gr.Textbox(label="Endpoint", value=DEFAULT_ENDPOINT)

            k = gr.Slider(1, 20, value=5, step=1, label="Top-k")
            chunk_size = gr.Slider(300, 2000, value=1000, step=50, label="Chunk size")
            chunk_overlap = gr.Slider(0, 400, value=150, step=10, label="Chunk overlap")

            with gr.Row():
                btn_rebuild = gr.Button("Build/Rebuild", variant="primary")
                btn_clear = gr.Button("Clear chat")

            status = gr.Textbox(label="Status", value="", interactive=False)

        with gr.Column(scale=2):
            chat = gr.Chatbot(height=480, show_copy_button=True)
            chat_input = gr.Textbox(placeholder="Ask something about your documents...", label="Your question")
            btn_send = gr.Button("Send", variant="primary")

    # State (stores in-memory NumPy index dict)
    index_state = gr.State(None)

    # Events
    def on_upload(files, folder):
        ensure_docs_dir(folder)
        n = save_uploaded_files(files, folder)
        return f"Saved {n} file(s) to {folder}"

    uploader.upload(fn=on_upload, inputs=[uploader, docs_dir], outputs=status)

    def on_rebuild(d, k_, cs, co, m, e):
        idx, msg = do_rebuild(d, k_, cs, co, m, e)
        return idx, msg

    btn_rebuild.click(
        fn=on_rebuild,
        inputs=[docs_dir, k, chunk_size, chunk_overlap, model, endpoint],
        outputs=[index_state, status],
    )

    def on_open(folder):
        ok, err = open_folder(folder)
        return "" if ok else f"Failed to open: {err}"

    btn_open.click(fn=on_open, inputs=docs_dir, outputs=status)

    def on_clear():
        return []
    btn_clear.click(fn=on_clear, outputs=chat)

    # Stream to chat
    btn_send.click(
        fn=chat_stream,
        inputs=[chat_input, chat, index_state, model, endpoint, k],
        outputs=chat,
    )
    chat_input.submit(
        fn=chat_stream,
        inputs=[chat_input, chat, index_state, model, endpoint, k],
        outputs=chat,
    )

if __name__ == "__main__":
    ensure_docs_dir(DOCS_DIR_DEFAULT)
    demo.launch()
