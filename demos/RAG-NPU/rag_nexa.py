
from __future__ import annotations

import os
import re
import json
import argparse
from typing import List, Dict, Any, Iterable, Tuple
import requests

import numpy as np
import fitz
import docx

from langchain_huggingface import HuggingFaceEmbeddings
from langchain.schema import Document


# Config
DEFAULT_MODEL = "NexaAI/Qwen3-4B-GGUF"
DEFAULT_ENDPOINT = "http://127.0.0.1:18181"
DEFAULT_EMBED_MODEL = "djuna/jina-embeddings-v2-small-en-Q5_K_M-GGUF"
DEFAULT_INDEX_JSON = "./vecdb.json"


# HTTP helper
def _post_json(url: str, payload: dict, timeout: int = 300) -> dict:
    headers = {"Content-Type": "application/json"}
    resp = requests.post(url, headers=headers, data=json.dumps(payload), timeout=timeout)
    if resp.status_code >= 400:
        raise requests.HTTPError(f"{resp.status_code} {url}\n{resp.text}", response=resp)
    return resp.json()

def call_nexa_chat(model: str, messages: List[Dict[str, Any]], base: str, stream: bool = False):
    """
    Call Nexa-compatible /v1/chat/completions endpoint.
    - If stream=False: return full text.
    - If stream=True: yield incremental text pieces.
    """
    url = base.rstrip("/") + "/v1/chat/completions"
    headers = {"Content-Type": "application/json"}
    payload = {"model": model, "messages": messages, "stream": stream, "max_tokens": 512}

    if not stream:
        data = _post_json(url, payload)
        try:
            return data["choices"][0]["message"]["content"]
        except Exception:
            return data.get("text", "") or data.get("response", "")

    # stream=True (SSE)
    with requests.post(url, headers=headers, data=json.dumps(payload), stream=True, timeout=300) as resp:
        resp.raise_for_status()
        for raw in resp.iter_lines(decode_unicode=True):
            if not raw:
                continue
            if raw.startswith("data:"):
                chunk = raw[len("data:"):].strip()
                if chunk == "[DONE]":
                    break
                try:
                    obj = json.loads(chunk)
                except Exception:
                    continue
                choices = obj.get("choices", [])
                if choices:
                    delta = choices[0].get("delta") or {}
                    piece = delta.get("content", "")
                    if piece:
                        yield piece


def call_nexa_embeddings(embed_model: str, inputs: List[str], base: str) -> List[List[float]]:
    """
    Call Nexa-compatible /v1/embeddings endpoint to embed a batch of strings.
    Returns a list of vectors aligned to 'inputs' order.
    """
    url = base.rstrip("/") + "/v1/embeddings"
    # Split into small batches to avoid large payloads
    out: List[List[float]] = []
    B = 64
    for i in range(0, len(inputs), B):
        batch = inputs[i:i+B]
        payload = {
            "model": embed_model,
            "input": batch,
            "encoding_format": "float"
        }
        data = _post_json(url, payload)
        # Expected shape: {"data":[{"embedding":[...],"index":0}, ...]}
        # Sort by index to be safe
        vecs = [None] * len(batch)
        for item in data.get("data", []):
            idx = item.get("index", 0)
            vec = item.get("embedding", [])
            if 0 <= idx < len(batch):
                vecs[idx] = vec
        # Fallback if no 'index' present — assume same order
        if any(v is None for v in vecs):
            vecs = [item.get("embedding", []) for item in data.get("data", [])]
        out.extend(vecs)
    return out


# File loaders (txt/pdf/docx)
def load_txt(path: str) -> str:
    for enc in ("utf-8", "utf-8-sig", "latin-1"):
        try:
            with open(path, "r", encoding=enc, errors="ignore") as f:
                return f.read()
        except Exception:
            continue
    return ""

def load_pdf(path: str) -> str:
    text_parts: List[str] = []
    with fitz.open(path) as doc:
        for page in doc:
            text_parts.append(page.get_text("text"))
    return "\n".join(text_parts)

def load_docx(path: str) -> str:
    d = docx.Document(path)
    paras = [p.text for p in d.paragraphs]
    return "\n".join(paras)

def normalize_ws(s: str) -> str:
    return re.sub(r"[ \t\u3000]+", " ", s).strip()

def yield_files(root: str, exts=(".txt", ".pdf", ".docx")) -> Iterable[str]:
    for base, _, files in os.walk(root):
        for name in files:
            if name.lower().endswith(exts):
                yield os.path.join(base, name)


# Chunking
def simple_chunk(text: str, chunk_size: int = 1000, overlap: int = 150) -> List[str]:
    """
    Simple character-level chunking with overlap.
    This keeps dependencies light for a JSON-based index.
    """
    text = text.replace("\r\n", "\n")
    n = len(text)
    chunks = []
    start = 0
    while start < n:
        end = min(start + chunk_size, n)
        chunks.append(text[start:end])
        if end == n:
            break
        start = end - overlap
        if start < 0:
            start = 0
    return chunks


# Build JSON index
def build_json_index(data_folder: str, index_path: str, embed_model: str, chunk_size: int, overlap: int, endpoint_base: str) -> Tuple[int, int]:
    """
    Read documents → chunk → embed → write a single JSON file with:
    {
      "embed_model": "...",
      "dim": 384,
      "items": [
        {"id": 0, "text": "...", "source": "...", "chunk_index": 0, "vector": [floats...]},
        ...
      ]
    }
    Returns: (num_docs, num_chunks)
    """
    items = []
    num_docs = 0
    for path in yield_files(data_folder):
        num_docs += 1
        # Load
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

        raw = normalize_ws(raw)
        if not raw:
            continue

        # Chunk
        chunks = simple_chunk(raw, chunk_size, overlap)
        # Embed through Nexa Serve
        vectors = call_nexa_embeddings(embed_model, chunks, endpoint_base)

        # Collect
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

    dim = len(items[0]["vector"])
    payload = {
        "embed_model": embed_model,
        "dim": dim,
        "items": items,
    }
    with open(index_path, "w", encoding="utf-8") as f:
        json.dump(payload, f, ensure_ascii=False)

    return num_docs, len(items)


# Load JSON index → NumPy arrays
def load_json_index(index_path: str):
    with open(index_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    items = data["items"]
    texts = [it["text"] for it in items]
    sources = [it["source"] for it in items]
    chunk_ids = [it["chunk_index"] for it in items]
    # Build matrix [N, D]
    mat = np.array([it["vector"] for it in items], dtype=np.float32)
    return {
        "embed_model": data.get("embed_model", ""),
        "dim": data.get("dim", mat.shape[1]),
        "matrix": mat,         # (N, D)
        "texts": texts,        # list[str]
        "sources": sources,    # list[str]
        "chunk_ids": chunk_ids # list[int]
    }


# NumPy search 
def embed_query_server(query: str, embed_model: str, endpoint_base: str) -> np.ndarray:
    vecs = call_nexa_embeddings(embed_model, [query], endpoint_base)
    return np.array(vecs[0], dtype=np.float32)

def search_numpy(query: str, index: dict, embed_model: str, endpoint_base: str, top_k: int = 5):
    q_vec = embed_query_server(query, embed_model, endpoint_base)  # (D,)
    db = index["matrix"]  # (N, D)

    # Normalize (avoid division by zero)
    q_norm = q_vec / (np.linalg.norm(q_vec) + 1e-8)  # (D,)
    db_norm = db / (np.linalg.norm(db, axis=1, keepdims=True) + 1e-8)  # (N, D)

    # Cosine similarity = dot(q_norm, db_norm.T)
    sims = db_norm @ q_norm  # (N,)
    top_idx = np.argsort(-sims)[:top_k]
    return top_idx, sims[top_idx]


# CLI
def main():
    ap = argparse.ArgumentParser(description="Local-files RAG (text-only) using JSON index + NumPy search")
    ap.add_argument("--data", default="./docs", help="Folder with txt/pdf/docx")
    ap.add_argument("--index_json", default=DEFAULT_INDEX_JSON, help="Path to embeddings JSON index")
    ap.add_argument("--embed_model", default=DEFAULT_EMBED_MODEL, help="Embedding model name")
    ap.add_argument("--chunk_size", type=int, default=1000, help="Chunk size")
    ap.add_argument("--chunk_overlap", type=int, default=150, help="Chunk overlap")
    ap.add_argument("--k", type=int, default=5, help="Top-k retrieval")
    ap.add_argument("--model", default=DEFAULT_MODEL, help="LLM model for generation")
    ap.add_argument("--endpoint", default=DEFAULT_ENDPOINT, help="Nexa endpoint base, e.g. http://127.0.0.1:18181")
    ap.add_argument("--rebuild", action="store_true", help="Rebuild JSON index before starting chat")
    args = ap.parse_args()

    os.makedirs(os.path.dirname(args.index_json) or ".", exist_ok=True)

    # Rebuild on start
    if args.rebuild or (not os.path.exists(args.index_json)):
        print(f"[build] Building JSON index via server embeddings → {args.index_json}")
        n_docs, n_chunks = build_json_index(args.data, args.index_json, args.embed_model, args.chunk_size, args.chunk_overlap, args.endpoint)
        print(f"[build] Done. docs={n_docs}, chunks={n_chunks}")
    else:
        print(f"[info] Using existing index: {args.index_json}")

    # Load index
    index = load_json_index(args.index_json)
    print(f"[info] Loaded index: dim={index['dim']}, rows={index['matrix'].shape[0]}, embed_model={index['embed_model']}")

    print(f"[info] Ready. model={args.model} endpoint={args.endpoint}")
    print("Type your question (Enter to quit). Commands: :reload (rebuild index)")

    while True:
        q = input("[user] ").strip()
        if not q:
            break
        if q.lower() == ":reload":
            print("[build] Rebuilding JSON index ...")
            n_docs, n_chunks = build_json_index(args.data, args.index_json, args.embed_model, args.chunk_size, args.chunk_overlap)
            index = load_json_index(args.index_json)
            print(f"[build] Done. docs={n_docs}, chunks={n_chunks}")
            continue

        # NumPy search
        try:
            top_idx, top_sims = search_numpy(q, index, args.embed_model, args.endpoint, top_k=args.k)
        except Exception as e:
            print(f"[error] Search failed: {e}")
            continue

        # Display retrieved items
        print("\n[retrieved]")
        for rank, (i, s) in enumerate(zip(top_idx.tolist(), top_sims.tolist()), start=1):
            src = os.path.basename(index["sources"][i])
            cid = index["chunk_ids"][i]
            snippet = index["texts"][i][:160].replace("\n", " ")
            print(f"  {rank}. {src}#chunk{cid}  sim={s:.4f}  {snippet}...")

        # Build chat messages with context
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

        print("\n[assistant]", end="", flush=True)
        try:
            for piece in call_nexa_chat(args.model, messages, args.endpoint, stream=True):
                print(piece, end="", flush=True)
            print()
        except requests.HTTPError as e:
            print(f"\n[warn] streaming failed, fallback to non-stream. Reason: {e}")
            try:
                full = call_nexa_chat(args.model, messages, args.endpoint, stream=False)
                print(full)
            except Exception as e2:
                print(f"[error] Non-stream request also failed: {e2}")

    print("[info] Bye.")


if __name__ == "__main__":
    main()
