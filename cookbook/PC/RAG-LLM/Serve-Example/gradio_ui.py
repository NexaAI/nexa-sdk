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


# Fix for ARM64 Windows matplotlib compatibility issue
import os
import sys
from gradio import ChatMessage

# Disable Gradio's matplotlib backend manager on ARM64 Windows
if sys.platform == "win32" and os.environ.get("PROCESSOR_ARCHITECTURE") == "ARM64":
    os.environ["MPLBACKEND"] = "Agg"
    os.environ["_GRADIO_SKIP_MATPLOTLIB_MANAGER"] = "1"

import platform
import subprocess
from typing import List, Tuple, Optional, Dict, Any

import gradio as gr

from rag_nexa import (
    DEFAULT_MODEL, DEFAULT_ENDPOINT, DEFAULT_INDEX_JSON, DEFAULT_EMBED_MODEL,
    build_json_index, load_json_index, search_numpy, call_nexa_chat
)

DOCS_DIR_DEFAULT = "../docs"


# ============================================================================
# Helper Functions
# ============================================================================

def ensure_docs_dir(path: str) -> None:
    """
    Create directory if it doesn't exist.
    
    Args:
        path: Directory path to create
    """
    try:
        if not os.path.exists(path):
            os.makedirs(path, exist_ok=True)
    except OSError as e:
        raise OSError(f"Failed to create directory {path}: {e}")


def save_uploaded_files(files: List[gr.File], dest_dir: str) -> int:
    """
    Save uploaded files to destination directory.
    
    Args:
        files: List of uploaded Gradio File objects
        dest_dir: Destination directory path
        
    Returns:
        Number of files successfully saved
    """
    if not files:
        return 0
    
    saved_count = 0
    for file in files:
        try:
            # Extract file path from Gradio File object
            src = getattr(file, "name", None) or file
            if not src:
                continue
            
            # Copy file to destination
            filename = os.path.basename(src)
            out_path = os.path.join(dest_dir, filename)
            
            with open(src, "rb") as r, open(out_path, "wb") as w:
                w.write(r.read())
            
            saved_count += 1
        except Exception as e:
            # Log error but continue with other files
            print(f"Warning: Failed to save {src}: {e}")
            continue
    
    return saved_count


def open_folder(path: str) -> Tuple[bool, Optional[str]]:
    """
    Open folder in system file explorer.
    
    Args:
        path: Folder path to open
        
    Returns:
        Tuple of (success: bool, error_message: str or None)
    """
    if not os.path.exists(path):
        return False, f"Path does not exist: {path}"
    
    try:
        system = platform.system()
        
        # Platform-specific folder opening
        if system == "Windows":
            subprocess.Popen(["explorer", os.path.normpath(path)])
        elif system == "Darwin":  # macOS
            subprocess.Popen(["open", path])
        else:  # Linux and others
            subprocess.Popen(["xdg-open", path])
        
        return True, None
    except Exception as e:
        return False, str(e)


# ============================================================================
# Core RAG Functions
# ============================================================================

def do_rebuild(docs_dir: str, k: int, chunk_size: int, chunk_overlap: int,
               _model: str, endpoint: str) -> Tuple[Optional[Dict[str, Any]], str]:
    """
    Build or rebuild the document index from files in docs_dir.
    
    Args:
        docs_dir: Directory containing documents to index
        k: Number of top results (unused in this function)
        chunk_size: Size of text chunks for splitting
        chunk_overlap: Overlap between consecutive chunks
        _model: Model name (unused, using DEFAULT_EMBED_MODEL)
        endpoint: API endpoint for embedding service
        
    Returns:
        Tuple of (index_dict or None, status_message)
    """
    try:
        # Validate inputs
        if chunk_overlap >= chunk_size:
            return None, "Error: Chunk overlap must be less than chunk size"
        
        # Ensure docs directory exists
        ensure_docs_dir(docs_dir)
        
        # Index path resides inside docs folder to keep UX simple
        index_json_path = os.path.join(docs_dir, os.path.basename(DEFAULT_INDEX_JSON))
        
        # Build JSON index then load into memory
        n_docs, n_chunks = build_json_index(
            docs_dir, index_json_path, DEFAULT_EMBED_MODEL, 
            chunk_size, chunk_overlap, endpoint
        )
        
        # Load the built index into memory
        index = load_json_index(index_json_path)
        status = f"✓ Indexed {n_chunks} text chunks from {n_docs} document(s)"
        
        return index, status
        
    except FileNotFoundError as e:
        return None, f"Error: Directory not found - {e}"
    except ValueError as e:
        return None, f"Error: Invalid parameter value - {e}"
    except Exception as e:
        return None, f"Error: Failed to build index - {e}"


def chat_stream(message: str, history: list, index: Optional[Dict[str, Any]],
                model: str, endpoint: str, k: int):
    """
    Stream chat responses using RAG (Retrieval-Augmented Generation).
    
    Args:
        message: User's question
        history: Chat history as list of [user_msg, assistant_msg] pairs
        index: In-memory index with document chunks and embeddings
        model: LLM model name
        endpoint: API endpoint for chat service
        k: Number of top-k chunks to retrieve
        
    Yields:
        Updated chat history with streaming response
    """
    if history is None:
        history = []
        
    # Validate message
    if not message or not message.strip():
        history.append(ChatMessage(role="assistant", content="⚠️ Please enter a question."))
        yield history, ""
        return
    
    history.append(ChatMessage(role="user", content=message))
    yield history, ""
    
    # Validate index exists
    if index is None:
        history.append(ChatMessage(role="assistant", content="⚠️ Index is empty. Please upload documents and click 'Build/Rebuild' first."))
        yield history, ""
        return
    
    
    # Retrieve relevant document chunks using NumPy cosine similarity search
    try:
        top_idx, top_sims = search_numpy(
            message, index, DEFAULT_EMBED_MODEL, endpoint, top_k=int(k)
        )
    except Exception as e:
        history.append(ChatMessage(role="assistant", content="⚠️ Search failed: {e}"))
        yield history, ""
        return
    
    # No results found
    if len(top_idx) == 0:
        history.append(ChatMessage(role="assistant", content="⚠️ No relevant documents found."))
        yield history, ""
        return
    
    # Compose context from retrieved chunks
    context_text = "\n\n".join(index["texts"][i] for i in top_idx.tolist())
    
    # Build messages for LLM with system prompt containing context
    messages = [
        {
            "role": "user", "content": (
                "You are a careful assistant. Use ONLY the provided context to answer. "
                f"Context:\n{context_text}\n"
                f"Question:\n {message}\n"
        )},
    ]
    
    
    # Initialize assistant turn in chat history
    history.append(ChatMessage(role="assistant", content=""))
    yield history, ""
    
    # Stream response from LLM
    try:
        for piece in call_nexa_chat(model, messages, endpoint, stream=True):
            history[-1].content += piece or ""
            yield history, ""
            
    except Exception as e:
        # Fallback to non-streaming mode if streaming fails
        try:
            response = call_nexa_chat(model, messages, endpoint, stream=False) or ""
            history[-1].content = response
            yield history, ""
        except Exception as e2:
            history[-1].content = f"⚠️ Generation failed: {e2}"
            yield history, ""


# ============================================================================
# Gradio UI
# ============================================================================

with gr.Blocks(title="RAG System") as demo:
    gr.Markdown("## RAG System - Retrieval-Augmented Generation")
    
    with gr.Row():
        # Left column: Data upload and settings
        with gr.Column(scale=1):
            gr.Markdown("**Data & Settings**")
            
            # Document folder management
            docs_dir = gr.Textbox(label="Docs folder", value=DOCS_DIR_DEFAULT)
            btn_open = gr.Button("📁 Open docs folder")
            
            # File uploader for documents
            uploader = gr.Files(
                label="Upload files (txt/pdf/docx)",
                file_types=[".txt", ".pdf", ".docx"],
                file_count="multiple",
            )
            
            # Model configuration
            model = gr.Textbox(label="Model", value=DEFAULT_MODEL)
            endpoint = gr.Textbox(label="Endpoint", value=DEFAULT_ENDPOINT)
            
            # Retrieval and chunking parameters
            k = gr.Slider(1, 20, value=5, step=1, label="Top-k (number of chunks to retrieve)")
            chunk_size = gr.Slider(300, 2000, value=1000, step=50, label="Chunk size (characters)")
            chunk_overlap = gr.Slider(0, 400, value=150, step=10, label="Chunk overlap (characters)")
            
            # Action buttons
            with gr.Row():
                btn_rebuild = gr.Button("🔄 Build/Rebuild Index", variant="primary")
                btn_clear = gr.Button("🗑️ Clear chat")
            
            # Status display
            status = gr.Textbox(label="Status", value="Ready", interactive=False)
        
        # Right column: Chat interface
        with gr.Column(scale=2):
            chat = gr.Chatbot(type="messages", height=480, show_copy_button=True)
            chat_input = gr.Textbox(
                placeholder="Ask something about your documents...", 
                label="Your question"
            )
            btn_send = gr.Button("📤 Send", variant="primary")
    
    # State: stores in-memory NumPy index dictionary
    index_state = gr.State(None)
    
    # ============================================================================
    # Event Handlers
    # ============================================================================
    
    def on_upload(files, folder):
        """Handle file upload event."""
        try:
            ensure_docs_dir(folder)
            n = save_uploaded_files(files, folder)
            return f"✓ Saved {n} file(s) to {folder}"
        except Exception as e:
            return f"⚠️ Upload failed: {e}"
    
    uploader.upload(fn=on_upload, inputs=[uploader, docs_dir], outputs=status)
    
    def on_rebuild(d, k_, cs, co, m, e):
        """Handle index rebuild event."""
        idx, msg = do_rebuild(d, k_, cs, co, m, e)
        return idx, msg
    
    btn_rebuild.click(
        fn=on_rebuild,
        inputs=[docs_dir, k, chunk_size, chunk_overlap, model, endpoint],
        outputs=[index_state, status],
    )
    
    def on_open(folder):
        """Handle open folder event."""
        ok, err = open_folder(folder)
        return "" if ok else f"⚠️ Failed to open: {err}"
    
    btn_open.click(fn=on_open, inputs=docs_dir, outputs=status)
    
    def on_clear():
        """Handle clear chat event."""
        return []
    
    btn_clear.click(fn=on_clear, outputs=chat)
    
    # Stream chat responses (both button click and enter key)
    btn_send.click(
        fn=chat_stream,
        inputs=[chat_input, chat, index_state, model, endpoint, k],
        outputs=[chat, chat_input],
    )
    chat_input.submit(
        fn=chat_stream,
        inputs=[chat_input, chat, index_state, model, endpoint, k],
        outputs=[chat, chat_input]
    )


if __name__ == "__main__":
    # Ensure default docs directory exists on startup
    ensure_docs_dir(DOCS_DIR_DEFAULT)
    demo.launch()
