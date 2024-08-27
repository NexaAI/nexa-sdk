import warnings
warnings.filterwarnings("ignore", message=".*clean_up_tokenization_spaces.*")

import os
import pdfplumber
from sentence_transformers import SentenceTransformer
import faiss
import numpy as np
import re

input_dir = './assets/input/'
output_dir = './assets/output/processed_data/'

# extract pdf file one by one:
def extract_text_from_pdf(pdf_path):
    print(f"1️⃣ Extracting text from {pdf_path}")
    with pdfplumber.open(pdf_path) as pdf:
        text = ''
        for i, page in enumerate(pdf.pages):
            page_text = page.extract_text()
            text += page_text + '\n'
            print(f"Processed page {i+1}/{len(pdf.pages)}")
    return text

# chunk the file by tokens:
def chunk_text(text, model, max_tokens=256, overlap=20):
    print("2️⃣ Chunking extracted text ...")

    # split the text into definitions and other parts
    definitions = re.findall(r'"\w+(?:\s+\w+)*"\s+means.*?(?="\w+(?:\s+\w+)*"\s+means|\Z)', text, re.DOTALL)
    other_parts = re.split(r'"\w+(?:\s+\w+)*"\s+means.*?(?="\w+(?:\s+\w+)*"\s+means|\Z)', text)

    chunks = []

    for definition in definitions:
        if len(model.tokenizer.tokenize(definition)) <= max_tokens:
            chunks.append(definition.strip())
        else:
            # if a definition is too long, split it into smaller parts
            sentences = re.split(r'(?<=[.!?])\s+', definition)
            current_chunk = []
            current_tokens = 0
            for sentence in sentences:
                sentence_tokens = len(model.tokenizer.tokenize(sentence))
                if current_tokens + sentence_tokens > max_tokens:
                    chunks.append(' '.join(current_chunk).strip())
                    current_chunk = [sentence]
                    current_tokens = sentence_tokens
                else:
                    current_chunk.append(sentence)
                    current_tokens += sentence_tokens
            if current_chunk:
                chunks.append(' '.join(current_chunk).strip())

    # process other parts
    for part in other_parts:
        if part.strip():
            sentences = re.split(r'(?<=[.!?])\s+', part)
            current_chunk = []
            current_tokens = 0
            for sentence in sentences:
                sentence_tokens = len(model.tokenizer.tokenize(sentence))
                if current_tokens + sentence_tokens > max_tokens:
                    chunks.append(' '.join(current_chunk).strip())
                    current_chunk = [sentence]
                    current_tokens = sentence_tokens
                else:
                    current_chunk.append(sentence)
                    current_tokens += sentence_tokens
            if current_chunk:
                chunks.append(' '.join(current_chunk).strip())

    chunks = [chunk for chunk in chunks if chunk.strip()] # remove empty chunks, if any

    chunk_sizes = [len(model.tokenizer.tokenize(chunk)) for chunk in chunks]
    print(f"👉 Created {len(chunks)} chunks")
    print(f"   Chunk sizes: min={min(chunk_sizes)}, max={max(chunk_sizes)}, avg={sum(chunk_sizes)/len(chunk_sizes):.1f}")
    # print(f"👀{chunks}")

    return chunks

# create embeddings for all chunks at once:
def create_embeddings(chunks, model):
    print("3️⃣ Creating embeddings ...")
    embeddings = model.encode(chunks)
    print(f"👉 Created embeddings of shape: {embeddings.shape}")
    return embeddings

# add embeddings to FAISS index:
def build_faiss_index(embeddings):
    print("4️⃣ Building FAISS index ...")
    dimension = embeddings.shape[1]
    index = faiss.IndexFlatL2(dimension)
    index.add(embeddings.astype('float32'))
    print(f"👉 Added {len(embeddings)} vectors to FAISS index")
    return index

model = SentenceTransformer('all-MiniLM-L6-v2')

# process the pdf files:
all_chunks = []
for filename in os.listdir(input_dir):
    if filename.endswith('.pdf'):
        pdf_path = os.path.join(input_dir, filename)
        text = extract_text_from_pdf(pdf_path)
        file_chunks = chunk_text(text, model)  # using default overlap (20)
        all_chunks.extend(file_chunks)
        print(f"   File: {filename}, Chunks: {len(file_chunks)}")
print(f"✅ Total chunks from all PDFs: {len(all_chunks)}")

embeddings = create_embeddings(all_chunks, model)
faiss_index = build_faiss_index(embeddings)

# save the index and chunks:
print("5️⃣ Saving FAISS index and chunks ...")
os.makedirs(output_dir, exist_ok=True)
faiss.write_index(faiss_index, os.path.join(output_dir, 'pdf_index.faiss'))
np.save(os.path.join(output_dir, 'pdf_chunks.npy'), all_chunks)
