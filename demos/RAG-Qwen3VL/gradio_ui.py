import os
import platform
import subprocess
from typing import List, Tuple

import gradio as gr

from rag_nexa import (
    DEFAULT_MODEL, DEFAULT_ENDPOINT, DEFAULT_EMBED_MODEL,
    build_chunks_from_folder, build_retriever,
    stream_nexa_chat_messages,
    yield_images, build_image_index, retrieve_topk_images,
)

DOCS_DIR_DEFAULT = "./docs"
IMG_TOPK_DEFAULT = 3


# Helpers
def ensure_docs_dir(path: str) -> None:
    if not os.path.exists(path):
        os.makedirs(path, exist_ok=True)

def save_uploaded_files(files: List[gr.File], dest_dir: str) -> int:
    """Save uploaded files to dest_dir (txt/pdf/docx and common images)."""
    if not files:
        return 0
    n = 0
    for f in files:
        src = getattr(f, "name", None) or f  # local temp path
        if not src:
            continue
        filename = os.path.basename(src)
        out_path = os.path.join(dest_dir, filename)
        with open(src, "rb") as r, open(out_path, "wb") as w:
            w.write(r.read())
        n += 1
    return n

def open_folder(path: str) -> Tuple[bool, str | None]:
    """Open folder in OS file explorer (success returns no UI message)."""
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


# Build / Rebuild pipeline
def do_rebuild(docs_dir: str, k: int, chunk_size: int, chunk_overlap: int,
                _model: str, _endpoint: str):
    """
    Return:
        retriever (text),
        img_index, img_paths_kept, clip_model (images),
        status text
    """
    ensure_docs_dir(docs_dir)
    docs = build_chunks_from_folder(docs_dir, chunk_size=chunk_size, chunk_overlap=chunk_overlap)
    retriever = build_retriever(docs, k=k, endpoint=_endpoint, embed_model=DEFAULT_EMBED_MODEL) if docs else None

    # Build image index from docs_dir
    all_imgs = yield_images(docs_dir)
    img_index, img_paths_kept, clip_model = build_image_index(all_imgs)

    if (not docs) and (not img_paths_kept):
        status = f"No files found in {docs_dir}. Please upload txt/pdf/docx or images."
    else:
        status = f"Indexed text chunks: {len(docs) if docs else 0}, images: {len(img_paths_kept)}."

    return retriever, img_index, img_paths_kept, clip_model, status


# Chat
def chat_stream(message: str,
                history: list,
                retriever,
                img_index,
                img_paths_kept,
                clip_model,
                model: str,
                endpoint: str,
                img_topk: int):
    """
    Generator for Gradio Chatbot: retrieval + streaming generation.
    - Text retrieval from retriever
    - Image retrieval via CLIP FAISS (retrieve_topk_images)
    - Generation via /v1/chat/completions, stream=True
    """
    has_text = retriever is not None
    has_images = (img_index is not None) and (clip_model is not None) and bool(img_paths_kept)

    if not has_text and not has_images:
        yield history + [[message, "Index is empty. Please upload files and click Rebuild."]]
        return

    # Text retrieval
    ctx_docs = retriever.get_relevant_documents(message) if has_text else []
    context_text = "\n\n".join([d.page_content for d in ctx_docs]) if ctx_docs else ""

    # Image retrieval
    topk_imgs = []
    if img_index is not None and clip_model is not None and img_paths_kept:
        topk_imgs = retrieve_topk_images(message, max(1, img_topk), img_index, img_paths_kept, clip_model)

    img_contents = [{"type": "image_url", "image_url": {"url": p}} for p in topk_imgs]

    # Messages for Nexa chat completions
    messages = [
        {
            "role": "system",
            "content": (
                "You are a careful assistant. Use ONLY the provided context to answer.\n\n"
                f"<context>\n{context_text}\n</context>"
            ),
        },
        {
            "role": "user",
            "content": [{"type": "text", "text": message}] + img_contents
        },
    ]

    # Streaming
    partial = ""
    yield history + [[message, partial]]  # create assistant turn
    try:
        for piece in stream_nexa_chat_messages(model, messages, endpoint):
            partial += piece
            yield history + [[message, partial]]
    except Exception as e:
        yield history + [[message, f"(Streaming failed: {e})"]]


# UI
with gr.Blocks(title="RAG with Qwen3vl") as demo:
    gr.Markdown("## RAG with Qwen3vl")

    with gr.Row():
        with gr.Column(scale=1):
            gr.Markdown("**Data & Settings**")

            docs_dir = gr.Textbox(label="Docs folder", value=DOCS_DIR_DEFAULT)
            btn_open = gr.Button("Open docs folder")

            uploader = gr.Files(
                label="Upload files (txt/pdf/docx/png/jpg/jpeg/webp/bmp)",
                file_types=[".txt", ".pdf", ".docx", ".png", ".jpg", ".jpeg", ".webp", ".bmp"],
                file_count="multiple",
            )

            model = gr.Textbox(label="Model", value=DEFAULT_MODEL)
            endpoint = gr.Textbox(label="Endpoint", value=DEFAULT_ENDPOINT)

            k = gr.Slider(1, 20, value=5, step=1, label="Top-k (text)")
            img_topk = gr.Slider(0, 6, value=IMG_TOPK_DEFAULT, step=1, label="Top-k images (0=disable)")
            chunk_size = gr.Slider(300, 2000, value=1000, step=50, label="Chunk size")
            chunk_overlap = gr.Slider(0, 400, value=150, step=10, label="Chunk overlap")

            with gr.Row():
                btn_rebuild = gr.Button("Build/Rebuild", variant="primary")
                btn_clear = gr.Button("Clear chat")

            status = gr.Textbox(label="Status", value="", interactive=False)

        with gr.Column(scale=2):
            chat = gr.Chatbot(height=520, show_copy_button=True)
            chat_input = gr.Textbox(placeholder="Ask something about your documents...", label="Your question")
            btn_send = gr.Button("Send", variant="primary")

    # States
    retriever_state = gr.State(None)
    img_index_state = gr.State(None)
    img_paths_state = gr.State([])
    clip_model_state = gr.State(None)

    # Events
    def on_upload(files, folder):
        ensure_docs_dir(folder)
        n = save_uploaded_files(files, folder)
        return f"Saved {n} file(s) to {folder}"

    uploader.upload(fn=on_upload, inputs=[uploader, docs_dir], outputs=status)

    def on_rebuild(d, k_, cs, co, m, e):
        r, img_idx, img_paths, clip_m, msg = do_rebuild(d, k_, cs, co, m, e)
        return r, img_idx, img_paths, clip_m, msg

    btn_rebuild.click(
        fn=on_rebuild,
        inputs=[docs_dir, k, chunk_size, chunk_overlap, model, endpoint],
        outputs=[retriever_state, img_index_state, img_paths_state, clip_model_state, status],
    )

    def on_open(folder):
        ok, err = open_folder(folder)
        return "" if ok else f"Failed to open: {err}"

    btn_open.click(fn=on_open, inputs=docs_dir, outputs=status)

    def on_clear():
        return []
    btn_clear.click(fn=on_clear, outputs=chat)

    # Chat send (streaming)
    btn_send.click(
        fn=chat_stream,
        inputs=[chat_input, chat, retriever_state, img_index_state, img_paths_state, clip_model_state, model, endpoint, img_topk],
        outputs=chat,
    )
    chat_input.submit(
        fn=chat_stream,
        inputs=[chat_input, chat, retriever_state, img_index_state, img_paths_state, clip_model_state, model, endpoint, img_topk],
        outputs=chat,
    )

if __name__ == "__main__":
    ensure_docs_dir(DOCS_DIR_DEFAULT)
    demo.launch()
