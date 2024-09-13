from multiprocessing import Pool, cpu_count
from llama_index.core import SimpleDirectoryReader, Document
from llama_index.core.node_parser import TokenTextSplitter

import time
import subprocess
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
    try:
        print(f"Attempting to read image file: {file_path}")
        image = Image.open(file_path)
        text = pytesseract.image_to_string(image)
        print(f"Successfully read image file: {file_path}")
        return text
    except Exception as e:
        print(f"Error reading image file {file_path}: {e}")
        return ""

def read_text_file(file_path):
    with open(file_path, 'r') as file:
        text = file.read()
    return text

def process_document(args):
    file_path, chunk_size = args
    _, file_ext = os.path.splitext(file_path.lower())
    if file_ext == '.docx':
        text = read_word_file(file_path)
    elif file_ext == '.pdf':
        text = read_pdf_file(file_path)
    elif file_ext in ('.png', '.jpg', '.jpeg'):
        text = read_image_file(file_path)
    elif file_ext == '.txt':
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
    description_generator = inference._chat(
        "Please provide a detailed description of this image in a sentence, emphasizing the meaning and context. Focus on capturing the key elements and underlying semantics.",
        image_path
    )
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

# Recursive summarization function
MAX_CHUNK_SIZE = 2048  # Adjust based on your model's context window size
MAX_RECURSION_DEPTH = 5  # Prevent infinite recursion

def summarize_text_recursively(text, max_chunk_size=MAX_CHUNK_SIZE, recursion_depth=0):
    if recursion_depth > MAX_RECURSION_DEPTH:
        # Stop recursion and return text as is
        return text
    splitter = TokenTextSplitter(chunk_size=max_chunk_size)
    chunks = splitter.split_text(text)
    if len(chunks) == 1:
        # Text is short enough, generate summary directly
        summary = generate_summary(chunks[0])
        return summary
    else:
        summaries = []
        for chunk in chunks:
            summary = summarize_text_recursively(chunk, max_chunk_size, recursion_depth+1)
            summaries.append(summary)
        # Combine summaries and summarize again
        combined_summaries = ' '.join(summaries)
        # Now check if combined_summaries is short enough
        if len(splitter.split_text(combined_summaries)) <= 1:
            final_summary = generate_summary(combined_summaries)
            return final_summary
        else:
            # Continue recursion
            return summarize_text_recursively(combined_summaries, max_chunk_size, recursion_depth+1)

def generate_summary(text):
    description_generator = inference._chat(
        "Please provide a detailed summary of the following text in a sentence, emphasizing the key points and context.",
        text
    )
    description = get_response_text_from_generator(description_generator)
    return description

def generate_text_description(input_text):
    summary = summarize_text_recursively(input_text)
    return summary

def get_descriptions_and_embeddings_for_texts(text_tuples):
    results = []
    for file_path, text in text_tuples:
        description = generate_text_description(text)
        embeddings = inference_text.create_embedding(description)["data"][0]['embedding']
        results.append({
            'file_path': file_path,
            'description': description,
            'embeddings': embeddings
        })
    return results

if __name__ == '__main__':
    path = "/Users/q/nexa_test/llama-fs/sample_data"
    
    start_time = time.time()
    documents, file_paths = load_documents_multiprocessing(path)
    end_time = time.time()
    
    print(f"Time taken to load documents: {end_time - start_time:.2f} seconds")
    print("-"*50)
    print_tree_with_subprocess(path)
    
    image_files = [doc.metadata['file_path'] for doc in documents if os.path.splitext(doc.metadata['file_path'].lower())[1] in ('.png', '.jpg', '.jpeg')]
    descriptions_and_embeddings_images = get_decriptions_and_embeddings_for_images(image_files)
    
    text_files = [doc.metadata['file_path'] for doc in documents if os.path.splitext(doc.metadata['file_path'].lower())[1] == '.txt']
    # Create a list of tuples (file_path, text_content)
    text_tuples = [(file_path, read_text_file(file_path)) for file_path in text_files]
    descriptions_and_embeddings_texts = get_descriptions_and_embeddings_for_texts(text_tuples)
    
    output_file_images = "data/images_with_embeddings.json"
    os.makedirs(os.path.dirname(output_file_images), exist_ok=True)  # Ensure the directory exists
    with open(output_file_images, 'w') as f:
        json.dump(descriptions_and_embeddings_images, f, indent=4)
    
    output_file_texts = "data/texts_with_embeddings.json"
    os.makedirs(os.path.dirname(output_file_texts), exist_ok=True)  # Ensure the directory exists
    with open(output_file_texts, 'w') as f:
        json.dump(descriptions_and_embeddings_texts, f, indent=4)
    
    for image_path, data in descriptions_and_embeddings_images.items():
        print(f"Image: {image_path}")
        print(f"Description: {data['description']}")
        # print(f"Embedding: {data['embedding']}")
        print("-"*50)
    
    for text_data in descriptions_and_embeddings_texts:
        print(f"File: {text_data['file_path']}")
        print(f"Description: {text_data['description']}")
        # print(f"Embedding: {text_data['embeddings']}")
        print("-"*50)