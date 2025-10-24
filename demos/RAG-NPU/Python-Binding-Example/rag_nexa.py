
from __future__ import annotations

import os
import re
import json
import argparse
from typing import List, Dict, Any, Iterable, Tuple
from pathlib import Path
import requests

import numpy as np
import pdfplumber
import docx

import warnings
import sys
from io import StringIO
from contextlib import contextmanager

from nexaai.llm import LLM, GenerationConfig
from nexaai.common import ModelConfig
from nexaai.embedder import Embedder, EmbeddingConfig
from nexaai.rerank import Reranker, RerankConfig

# ============================================================================
# Configuration Constants
# ============================================================================

DEFAULT_MODEL = "NexaAI/Granite-4-Micro-NPU"
DEFAULT_EMBED_MODEL = "NexaAI/embeddinggemma-300m-npu"
DEFAULT_INDEX_JSON = "./vecdb.json"
DEFAULT_RERANK_MODEL = "NexaAI/jina-v2-rerank-npu"
DEFAULT_MODEL_FOLDER = "~/.cache/nexa.ai/nexa_sdk/models"

# ============================================================================
# File System Utilities
# ============================================================================
def get_default_data_folder() -> str:
    """
    Get the default data folder path in user's Downloads directory.
    Creates the folder if it doesn't exist.
    
    Returns:
        str: Path to the data folder
    """
    downloads_folder = Path.home() / "Downloads" / "nexa-rag-docs"
    downloads_folder.mkdir(parents=True, exist_ok=True)
    return str(downloads_folder)


# ============================================================================
# Nexa SDK Python binding API Calls
# ============================================================================
def call_nexa_chat(model: str, messages: List[Dict[str, Any]], model_folder: str, stream: bool = False):
    """
    Call python binding(llm).
    
    Args:
        model: Model name/identifier
        messages: List of chat messages with 'role' and 'content'
        model_folder: model folder path
        stream: If True, yields text pieces; if False, returns full text 
    Returns:
        str (if stream=False): Complete response text
        Generator[str] (if stream=True): Yields text pieces incrementally
    """
    
    m_cfg = ModelConfig()
    llm = LLM.from_(model, plugin_id="npu", device_id="npu", m_cfg=m_cfg)
    
    prompt = llm.apply_chat_template(messages)
    g_cfg=GenerationConfig(max_tokens=512)
    if not stream:
        # Non-streaming: return complete text
        return llm.generate(prompt, g_cfg=g_cfg)
    
    for token in llm.generate_stream(prompt, g_cfg=g_cfg):
        yield token

def call_nexa_embeddings(embed_model: str, inputs: List[str], model_folder: str) -> List[List[float]]:
    """
    Call Nexa-compatible /v1/embeddings endpoint to embed a batch of strings.
    
    Args:
        embed_model: Embedding model name
        inputs: List of text strings to embed
        model_folder: model folder path
        
    Returns:
        List[List[float]]: List of embedding vectors aligned to input order
    """
    if not inputs:
        return []
    
    out: List[List[float]] = []
     
    embedder = Embedder.from_(name_or_path=embed_model, plugin_id='npu')

    # Process in batches to avoid large payloads
    BATCH_SIZE = 64
    for i in range(0, len(inputs), BATCH_SIZE):
        batch = inputs[i:i+BATCH_SIZE]
        batch_size = len(batch)
        embeddings = embedder.generate(
        texts=batch, config=EmbeddingConfig(batch_size=batch_size))
        
        for embedding in embeddings:
            out.append(embedding.tolist())
           
    return out


def call_nexa_rerank(rerank_model: str, query: str, documents: List[str], model_folder: str, top_n: int = 3) -> List[int]:
    """
    Call Nexa-compatible /v1/reranking endpoint to rerank documents.
    
    Args:
        rerank_model: Reranking model name
        query: Search query string
        documents: List of document texts to rerank
        model_folder: model folder path
        top_n: Number of top documents to return
        
    Returns:
        List[int]: List of document indices in reranked order (best first)

    """
    if not documents:
        return []
    
    reranker = Reranker.from_(name_or_path=rerank_model, plugin_id="npu")
    
    batch_size = len(documents)
    scores = reranker.rerank(query=query, documents=documents, 
                           config=RerankConfig(batch_size=batch_size))

    # Sort by relevance score (descending) and return top_n indices
    sorted_results = sorted(scores,reverse=True)
    return [r["index"] for r in sorted_results[:top_n]]


# ============================================================================
# Document Loaders
# ============================================================================
def load_txt(path: str) -> str:
    """
    Load text file with multiple encoding fallbacks.
    
    Args:
        path: Path to text file
        
    Returns:
        str: File contents (empty string if all encodings fail)
    """
    for enc in ("utf-8", "utf-8-sig", "latin-1"):
        try:
            with open(path, "r", encoding=enc, errors="ignore") as f:
                return f.read()
        except Exception:
            continue
    return ""


@contextmanager
def suppress_stderr():
    """Context manager to temporarily suppress stderr output."""
    old_stderr = sys.stderr
    try:
        with open(os.devnull, 'w') as devnull:
            sys.stderr = devnull
            yield
    finally:
        sys.stderr = old_stderr


def fix_missing_spaces(text: str) -> str:
    """
    Fix common cases where spaces are missing in extracted PDF text.
    Applies heuristics to add spaces between:
    - lowercase followed by uppercase (camelCase issues)
    - letters and numbers
    - punctuation and following letters
    
    Args:
        text: Input text with potential spacing issues
        
    Returns:
        str: Text with improved spacing
    """
    # Add space between lowercase and uppercase
    text = re.sub(r'([a-z])([A-Z])', r'\1 \2', text)
    # Add space between letter and number
    text = re.sub(r'([a-zA-Z])(\d)', r'\1 \2', text)
    # Add space between number and letter
    text = re.sub(r'(\d)([a-zA-Z])', r'\1 \2', text)
    # Add space after punctuation if followed by letter
    text = re.sub(r'([.,;:!?])([A-Za-z])', r'\1 \2', text)
    # Add space after closing bracket/paren if followed by uppercase
    text = re.sub(r'([\)\]])([A-Z])', r'\1 \2', text)
    return text


def load_pdf(path: str) -> str:
    """
    Load PDF text using pdfplumber with proper spacing handling.
    
    Args:
        path: Path to PDF file
        
    Returns:
        str: Extracted text (empty string if extraction fails)
    """
    text_parts: List[str] = []
    try:
        with suppress_stderr():
            with pdfplumber.open(path) as pdf:
                for page in pdf.pages:
                    # Use layout=True for better spacing preservation
                    text = page.extract_text(layout=True)
                    if text:
                        # Apply space-fixing heuristics
                        text = fix_missing_spaces(text)
                        text_parts.append(text)
    except Exception as e:
        print(f"[warn] Error extracting from {path}: {e}")
        return ""
    
    return "\n".join(text_parts)


def load_docx(path: str) -> str:
    """
    Load text from Word document.
    
    Args:
        path: Path to .docx file
        
    Returns:
        str: Extracted text from all paragraphs
    """
    try:
        d = docx.Document(path)
        paras = [p.text for p in d.paragraphs]
        return "\n".join(paras)
    except Exception as e:
        print(f"[warn] Error reading docx {path}: {e}")
        return ""


def normalize_ws(s: str) -> str:
    """
    Normalize whitespace: collapse multiple spaces/tabs into single space.
    
    Args:
        s: Input string
        
    Returns:
        str: String with normalized whitespace
    """
    return re.sub(r"[ \t\u3000]+", " ", s).strip()


def yield_files(root: str, exts=(".txt", ".pdf", ".docx")) -> Iterable[str]:
    """
    Recursively yield file paths matching specified extensions.
    
    Args:
        root: Root directory to search
        exts: Tuple of file extensions to include
        
    Yields:
        str: File paths matching extensions
    """
    for base, _, files in os.walk(root):
        for name in files:
            if name.lower().endswith(exts):
                yield os.path.join(base, name)


# ============================================================================
# Text Chunking
# ============================================================================
def simple_chunk(text: str, chunk_size: int = 1000, overlap: int = 150) -> List[str]:
    """
    Simple character-level chunking with overlap.
    Creates chunks of approximately chunk_size characters with overlap
    to preserve context at chunk boundaries.
    
    Args:
        text: Input text to chunk
        chunk_size: Target size of each chunk in characters
        overlap: Number of overlapping characters between chunks
        
    Returns:
        List[str]: List of text chunks
    """
    text = text.replace("\r\n", "\n")
    n = len(text)
    
    if n == 0:
        return []
    
    chunks = []
    start = 0
    
    while start < n:
        end = min(start + chunk_size, n)
        chunks.append(text[start:end])
        
        if end == n:
            break
        
        # Move start forward with overlap
        start = end - overlap
        if start < 0:
            start = 0
    
    return chunks


# ============================================================================
# Index Building and Loading
# ============================================================================
def build_json_index(
    data_folder: str, 
    index_path: str, 
    embed_model: str, 
    chunk_size: int, 
    overlap: int,
    model_folder: str = DEFAULT_MODEL_FOLDER
) -> Tuple[int, int]:
    """
    Build JSON-based vector index from documents in a folder.
    
    Process:
    1. Read all supported documents from folder
    2. Chunk each document with overlap
    3. Embed chunks using Python binding API
    4. Save to JSON file with metadata
    
    Args:
        data_folder: Folder containing documents to index
        index_path: Output path for JSON index file
        embed_model: Embedding model name
        chunk_size: Size of text chunks in characters
        overlap: Overlap between chunks in characters
        model_folder: model folder path
        
    Returns:
        Tuple[int, int]: (number of documents processed, number of chunks created)
        
    Raises:
        RuntimeError: If no chunks were created
    """
    items = []
    num_docs = 0
    
    for path in yield_files(data_folder):
        num_docs += 1
        
        # Load document based on file type
        lower = path.lower()
        try:
            if lower.endswith(".txt"):
                raw = load_txt(path)
            elif lower.endswith(".pdf"):
                raw = load_pdf(path)
            elif lower.endswith(".docx"):
                raw = load_docx(path)
            else:
                continue
        except Exception as e:
            print(f"[warn] Failed to read {path}: {e}")
            continue

        # Normalize and validate
        raw = normalize_ws(raw)
        if not raw:
            print(f"[warn] Empty content from {path}, skipping")
            continue

        # Chunk the document
        chunks = simple_chunk(raw, chunk_size, overlap)
        if not chunks:
            continue
            
        # Embed chunks via API
        try:
            vectors = call_nexa_embeddings(embed_model, chunks, model_folder)
        except Exception as e:
            print(f"[warn] Failed to embed chunks from {path}: {e}")
            continue

        # Store chunk metadata
        for i, (txt, vec) in enumerate(zip(chunks, vectors)):
            items.append({
                "id": len(items),
                "text": txt,
                "source": os.path.abspath(path),
                "chunk_index": i,
                "vector": vec,
            })

    if not items:
        raise RuntimeError("No chunks found. Check your --data path and files.")

    # Build and save index
    dim = len(items[0]["vector"])
    payload = {
        "embed_model": embed_model,
        "dim": dim,
        "items": items,
    }
    
    with open(index_path, "w", encoding="utf-8") as f:
        json.dump(payload, f, ensure_ascii=False)

    return num_docs, len(items)


def load_json_index(index_path: str) -> Dict[str, Any]:
    """
    Load JSON vector index into memory as NumPy arrays.
    
    Args:
        index_path: Path to JSON index file
        
    Returns:
        dict: Index data containing:
            - embed_model: Model used for embeddings
            - dim: Embedding dimension
            - matrix: NumPy array of shape (N, D) with all vectors
            - texts: List of chunk texts
            - sources: List of source file paths
            - chunk_ids: List of chunk indices within documents
            
    Raises:
        FileNotFoundError: If index file doesn't exist
        json.JSONDecodeError: If index file is malformed
        ValueError: If index structure is invalid
    """
    if not os.path.exists(index_path):
        raise FileNotFoundError(f"Index file not found: {index_path}")
    
    try:
        with open(index_path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except json.JSONDecodeError as e:
        raise json.JSONDecodeError(f"Malformed index file: {index_path}", e.doc, e.pos)
    
    # Validate structure
    if "items" not in data:
        raise ValueError("Index file missing 'items' field")
    
    items = data["items"]
    if not items:
        raise ValueError("Index contains no items")
    
    # Extract fields
    texts = [it["text"] for it in items]
    sources = [it["source"] for it in items]
    chunk_ids = [it["chunk_index"] for it in items]
    
    # Build embedding matrix
    mat = np.array([it["vector"] for it in items], dtype=np.float32)
    
    return {
        "embed_model": data.get("embed_model", ""),
        "dim": data.get("dim", mat.shape[1]),
        "matrix": mat,         # (N, D)
        "texts": texts,        # list[str]
        "sources": sources,    # list[str]
        "chunk_ids": chunk_ids # list[int]
    }


# ============================================================================
# Vector Search
# ============================================================================
def embed_query_server(query: str, embed_model: str, model_folder: str) -> np.ndarray:
    """
    Embed a single query string via API.
    
    Args:
        query: Query text to embed
        embed_model: Embedding model name
        model_folder: model folder path
        
    Returns:
        np.ndarray: Query embedding vector
    """
    vecs = call_nexa_embeddings(embed_model, [query], model_folder)
    return np.array(vecs[0], dtype=np.float32)


def search_numpy(
    query: str, 
    index: dict, 
    embed_model: str, 
    model_folder: str,
    top_k: int = 5
) -> Tuple[np.ndarray, np.ndarray]:
    """
    Search vector index using cosine similarity.
    
    Args:
        query: Search query text
        index: Loaded index dictionary from load_json_index()
        embed_model: Embedding model name
        model_folder: model folder path
        top_k: Number of top results to return
        
    Returns:
        Tuple[np.ndarray, np.ndarray]: 
            - Array of top-k indices
            - Array of corresponding similarity scores
    """
    # Embed query
    q_vec = embed_query_server(query, embed_model, model_folder)  # (D,)
    db = index["matrix"]  # (N, D)

    # Normalize vectors to compute cosine similarity
    q_norm = q_vec / (np.linalg.norm(q_vec) + 1e-8)  # (D,)
    db_norm = db / (np.linalg.norm(db, axis=1, keepdims=True) + 1e-8)  # (N, D)

    # Compute cosine similarity = dot product of normalized vectors
    sims = db_norm @ q_norm  # (N,)
    
    # Get top-k indices (sorted by similarity descending)
    top_idx = np.argsort(-sims)[:top_k]
    return top_idx, sims[top_idx]


# ============================================================================
# Main CLI Application
# ============================================================================
def main():
    """Main CLI application for RAG system."""
    # Get default data folder in Downloads
    default_data_folder = get_default_data_folder()
    
    # Parse command-line arguments
    ap = argparse.ArgumentParser(
        description="Local-files RAG (text-only) using JSON index + NumPy search"
    )
    ap.add_argument(
        "--data", 
        default=default_data_folder, 
        help=f"Folder with txt/pdf/docx (default: {default_data_folder})"
    )
    ap.add_argument(
        "--index_json", 
        default=DEFAULT_INDEX_JSON, 
        help="Path to embeddings JSON index"
    )
    ap.add_argument(
        "--embed_model", 
        default=DEFAULT_EMBED_MODEL, 
        help="Embedding model name"
    )
    ap.add_argument(
        "--chunk_size", 
        type=int, 
        default=1000, 
        help="Chunk size in characters"
    )
    ap.add_argument(
        "--chunk_overlap", 
        type=int, 
        default=150, 
        help="Chunk overlap in characters"
    )
    ap.add_argument(
        "--k", 
        type=int, 
        default=5, 
        help="Top-k retrieval"
    )
    ap.add_argument(
        "--rerank_top_n", 
        type=int, 
        default=3, 
        help="Top-n after reranking"
    )
    ap.add_argument(
        "--rerank_model", 
        default=DEFAULT_RERANK_MODEL, 
        help="Rerank model name"
    )
    ap.add_argument(
        "--use_rerank", 
        action="store_true", 
        help="Enable reranking step"
    )
    ap.add_argument(
        "--model", 
        default=DEFAULT_MODEL, 
        help="LLM model for generation"
    )
    
    ap.add_argument(
        "--model_folder", 
        default=DEFAULT_MODEL_FOLDER, 
        help="Model folder path"
    )
    
    ap.add_argument(
        "--rebuild", 
        action="store_true", 
        help="Rebuild JSON index before starting chat"
    )
    args = ap.parse_args()

    # Setup directories
    os.makedirs(args.data, exist_ok=True)
    print(f"[info] Using data folder: {args.data}")
    
    os.makedirs(os.path.dirname(args.index_json) or ".", exist_ok=True)

    # Build or load index
    if args.rebuild or (not os.path.exists(args.index_json)):
        print(f"[build] Building JSON index via server embeddings → {args.index_json}")
        try:
            n_docs, n_chunks = build_json_index(
                args.data, 
                args.index_json, 
                args.embed_model, 
                args.chunk_size, 
                args.chunk_overlap,
                args.model_folder,
            )
            print(f"[build] Done. docs={n_docs}, chunks={n_chunks}")
        except Exception as e:
            print(f"[error] Failed to build index: {e}")
            return
    else:
        print(f"[info] Using existing index: {args.index_json}")

    # Load index into memory
    try:
        index = load_json_index(args.index_json)
        print(f"[info] Loaded index: dim={index['dim']}, rows={index['matrix'].shape[0]}, embed_model={index['embed_model']}")
    except Exception as e:
        print(f"[error] Failed to load index: {e}")
        return

    print(f"[info] Ready. model={args.model}")
    print("Type your question (Enter to quit). Commands: :reload (rebuild index)")

    # Interactive chat loop
    while True:
        try:
            q = input("[user] ").strip()
        except (EOFError, KeyboardInterrupt):
            break
            
        if not q:
            break
            
        # Handle special commands
        if q.lower() == ":reload":
            print("[build] Rebuilding JSON index ...")
            try:
                n_docs, n_chunks = build_json_index(
                    args.data, 
                    args.index_json, 
                    args.embed_model, 
                    args.chunk_size, 
                    args.chunk_overlap,
                    args.model_folder,
                )
                index = load_json_index(args.index_json)
                print(f"[build] Done. docs={n_docs}, chunks={n_chunks}")
            except Exception as e:
                print(f"[error] Failed to rebuild index: {e}")
            continue

        # Perform vector search
        try:
            top_idx, top_sims = search_numpy(
                q, 
                index, 
                args.embed_model, 
                args.model_folder,
                top_k=args.k
            )
        except Exception as e:
            print(f"[error] Search failed: {e}")
            continue

        # Optional reranking step
        if args.use_rerank and len(top_idx) > 0:
            try:
                # Get candidate documents from initial search
                candidate_docs = [index["texts"][i] for i in top_idx.tolist()]
                
                # Rerank and get top_n indices (relative to candidate_docs)
                reranked_local_idx = call_nexa_rerank(
                    args.rerank_model, 
                    q, 
                    candidate_docs,
                    args.model_folder,
                    top_n=args.rerank_top_n
                )
                
                # Map back to original index positions
                top_idx = top_idx[reranked_local_idx]
                print(f"[rerank] Reranked to top {len(reranked_local_idx)} documents")
            except Exception as e:
                print(f"[warn] Reranking failed, using original search results: {e}")

        # Build context from retrieved chunks
        context_text = "\n\n".join([index["texts"][i] for i in top_idx.tolist()])
        messages = [
            {
                "role": "system",
                "content": (
                    "You are a careful assistant. Use ONLY the provided context to answer.\n\n"
                    f"<context>\n{context_text}\n</context>"
                ),
            },
            {"role": "user", "content": q},
        ]

        # Generate response
        print("\n[assistant]", end="", flush=True)
        try:
            # Try streaming first
            for piece in call_nexa_chat(args.model, messages, model_folder=args.model_folder, stream=True):
                print(piece, end="", flush=True)
            print()
        except requests.HTTPError as e:
            # Fallback to non-streaming
            print(f"\n[warn] Streaming failed, fallback to non-stream. Reason: {e}")
            try:
                full = call_nexa_chat(args.model, messages, model_folder=args.model_folder, stream=False)
                print(full)
            except Exception as e2:
                print(f"[error] Non-stream request also failed: {e2}")

    print("[info] Bye.")


if __name__ == "__main__":
    main()
