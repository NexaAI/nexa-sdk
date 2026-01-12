# Copyright 2024-2026 Nexa AI, Inc.
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
IMG_TOPK_DEFAULT = 1


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
        # Empty evidence
        yield history + [[message, "Index is empty. Please upload files and click Rebuild."]], [], "### Retrieved text\n_(no text hits)_"
        return

    # Text retrieval
    ctx_docs = retriever.get_relevant_documents(message) if has_text else []
    context_text = "\n\n".join([d.page_content for d in ctx_docs]) if ctx_docs else ""

    # Build text evidence markdown (filename#chunkN + snippet)
    def _mk_snippet(text: str, n: int = 160) -> str:
        t = (text or "").replace("\n", " ").strip()
        return (t[:n] + "…") if len(t) > n else t

    if ctx_docs:
        lines = []
        for i, d in enumerate(ctx_docs, 1):
            src = os.path.basename(d.metadata.get("source", "") or "")
            idx = d.metadata.get("chunk_index", -1)
            snippet = _mk_snippet(d.page_content)
            lines.append(f"**{i}.** `{src}#chunk{idx}`\n\n> {snippet}")
        chunks_md = "### Retrieved text\n" + "\n\n".join(lines)
    else:
        chunks_md = "### Retrieved text\n_(no text hits)_"

    # Image retrieval
    topk_imgs = []
    if img_index is not None and clip_model is not None and img_paths_kept and img_topk > 0:
        topk_imgs = retrieve_topk_images(message, max(1, img_topk), img_index, img_paths_kept, clip_model)

    img_contents = [{"type": "image_url", "image_url": {"url": p}} for p in topk_imgs]
    img_gallery = topk_imgs  # Right panel gallery

    # Compose messages
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

    # Streaming: render evidence on first frame
    partial = ""
    yield history + [[message, partial]], img_gallery, chunks_md

    try:
        for piece in stream_nexa_chat_messages(model, messages, endpoint):
            partial += piece
            yield history + [[message, partial]], img_gallery, chunks_md
    except Exception as e:
        yield history + [[message, f"(Streaming failed: {e})"]], img_gallery, chunks_md


# UI
CUSTOM_CSS = """
/* Make cards cleaner and add subtle shadows */
.gradio-container { max-width: 1600px !important; }
.rounded-card { border-radius: 16px; box-shadow: 0 1px 8px rgba(0,0,0,.06); background: white; }
.pad { padding: 14px; }
.section-title { font-weight: 700; font-size: 14px; opacity: .8; margin-bottom: 8px; }
#info-panel .gallery { background: #101114; } /* darker bg for images */
"""

with gr.Blocks(title="RAG with Qwen3vl", css=CUSTOM_CSS) as demo:
    gr.Markdown("## RAG with Qwen3vl")

    with gr.Row():
        # Left column: settings / files
        with gr.Column(scale=1, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Data & Settings**")

            docs_dir = gr.Textbox(label="Docs folder", value=DOCS_DIR_DEFAULT)
            btn_open = gr.Button("Open docs folder")

            uploader = gr.Files(
                label="Upload files (txt/pdf/docx/png/jpg/jpeg/webp/bmp)",
                file_types=[".txt", ".pdf", ".docx", ".png", ".jpg", ".jpeg", ".webp", ".bmp"],
                file_count="multiple",
            )

            with gr.Accordion("Model & Retrieval settings", open=False):
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

        # Middle column: chat
        with gr.Column(scale=2, elem_classes=["rounded-card", "pad"]):
            chat = gr.Chatbot(height=600, show_copy_button=True)
            chat_input = gr.Textbox(placeholder="Ask something about your documents...", label="Your question")
            btn_send = gr.Button("Send", variant="primary", elem_id="btn-send")

        # Right column: information panel (evidence)
        with gr.Column(scale=2):
            with gr.Accordion("Retrieved evidence", open=True):
                evid_gallery = gr.Gallery(
                    label="Images",
                    columns=3,
                    height=260,
                    show_label=True
                )
                evid_text = gr.Markdown()

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
        outputs=[chat, evid_gallery, evid_text],
    )
    chat_input.submit(
        fn=chat_stream,
        inputs=[chat_input, chat, retriever_state, img_index_state, img_paths_state, clip_model_state, model, endpoint, img_topk],
        outputs=[chat, evid_gallery, evid_text],
    )

if __name__ == "__main__":
    ensure_docs_dir(DOCS_DIR_DEFAULT)
    demo.launch()
