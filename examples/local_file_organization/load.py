from multiprocessing import Pool, cpu_count
from llama_index.core import SimpleDirectoryReader, Document
from llama_index.core.node_parser import TokenTextSplitter

import time
import subprocess
from collections import Counter
import fitz  # PyMuPDF
from PIL import Image
import pytesseract
import docx
from nexa.gguf import NexaVLMInference, NexaTextInference
import numpy as np
import json
import os

# Initialize the models
model_path = "llava-v1.6-vicuna-7b:q4_0"
model_path_text = "gemma-2b:q2_K"

inference = NexaVLMInference(
    model_path=model_path,
    local_path=None,
    stop_words=[],
    temperature=0.7,
    max_new_tokens=2048,
    top_k=50,
    top_p=1.0,
    profiling=True
)


inference_text = NexaTextInference(
    model_path=model_path_text,
    local_path=None,
    stop_words=[],
    temperature=0.7,
    max_new_tokens=512,
    top_k=50,
    top_p=0.9,
    profiling=True,
    embedding=True
)

def get_response_text_from_generator(generator):
    response_text = ""
    try:
        while True:
            response = next(generator)
            choices = response.get('choices', [])
            for choice in choices:
                delta = choice.get('delta', {})
                if 'content' in delta:
                    response_text += delta['content']
    except StopIteration:
        pass
    return response_text

def read_word_file(file_path):
    doc = docx.Document(file_path)
    full_text = []
    for para in doc.paragraphs:
        full_text.append(para.text)
    return '\n'.join(full_text)

def read_pdf_file(file_path):
    doc = fitz.open(file_path)
    full_text = []
    for page in doc:
        full_text.append(page.get_text())
    return '\n'.join(full_text)

def read_image_file(file_path):
    image = Image.open(file_path)
    text = pytesseract.image_to_string(image)
    return text

def read_text_file(file_path):
    with open(file_path, 'r') as file:
        text = file.read()
    return text

def process_document(args):
    file_path, chunk_size = args
    if file_path.endswith('.docx'):
        text = read_word_file(file_path)
    elif file_path.endswith('.pdf'):
        text = read_pdf_file(file_path)
    elif file_path.endswith(('.png', '.jpg', '.jpeg')):
        text = read_image_file(file_path)
    elif file_path.endswith('.txt'):
        text = read_text_file(file_path)
    else:
        raise ValueError(f"Unsupported file type: {file_path}")
    
    splitter = TokenTextSplitter(chunk_size=chunk_size)
    contents = splitter.split_text(text)
    combined_text = ' '.join(contents)
    return Document(text=combined_text, metadata={'file_path': file_path}), file_path

def load_documents_multiprocessing(path: str):
    reader = SimpleDirectoryReader(
        input_dir=path,
        recursive=True,
        required_exts=[".pdf", ".txt", ".png", ".jpg", ".jpeg", ".docx"]
    )
    
    chunk_size = 6144
    with Pool(cpu_count()) as pool:
        results = pool.map(process_document, [(d.metadata['file_path'], chunk_size) for docs in reader.iter_data() for d in docs])
    
    documents = [document for document, _ in results]
    file_paths = [file_path for _, file_path in results]
    
    return documents, file_paths

def print_tree_with_subprocess(path):
    result = subprocess.run(['tree', path], capture_output=True, text=True)
    print(result.stdout)

def generate_image_description(image_path):
    description_generator = inference._chat("Please provide a detailed description of this image in 10 sentences, emphasizing the meaning and context. Focus on capturing the key elements and underlying semantics.", image_path)
    description = get_response_text_from_generator(description_generator)
    return description

def get_decriptions_and_embeddings_for_images(image_paths):
    d = {}
    for image_path in image_paths:
        description = generate_image_description(image_path)
        embedding_result = inference_text.create_embedding(description)
        embedding = embedding_result["data"][0]['embedding']
        d[image_path] = {
            'description': description,
            'embedding': embedding
        }
    return d

if __name__ == '__main__':
    path = "/Users/q/nexa_test/llama-fs/sample_data"
    
    start_time = time.time()
    documents, file_paths = load_documents_multiprocessing(path)
    end_time = time.time()
    
    print(f"Time taken to load documents: {end_time - start_time:.2f} seconds")
    print("-"*50)
    print_tree_with_subprocess(path)
    
    image_files = [doc.metadata['file_path'] for doc in documents if doc.metadata['file_path'].endswith(('.png', '.jpg', '.jpeg'))]
    descriptions_and_embeddings = get_decriptions_and_embeddings_for_images(image_files)
    
    output_file = "data/images_with_embeddings.json"
    os.makedirs(os.path.dirname(output_file), exist_ok=True)  # Ensure the directory exists
    with open(output_file, 'w') as f:
        json.dump(descriptions_and_embeddings, f, indent=4)
    
    for image_path, data in descriptions_and_embeddings.items():
        print(f"Image: {image_path}")
        print(f"Description: {data['description']}")
        # print(f"Embedding: {data['embedding']}")
        print("-"*50)

