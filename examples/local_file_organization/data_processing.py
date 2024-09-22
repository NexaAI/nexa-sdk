import re
from multiprocessing import Pool, cpu_count
from nexa.gguf import NexaVLMInference, NexaTextInference
from file_utils import sanitize_filename, create_folder
import os
import shutil
import sys
import contextlib

# Global variables to hold the models
image_inference = None
text_inference = None

@contextlib.contextmanager
def suppress_stdout_stderr():
    """A context manager that redirects stdout and stderr to devnull."""
    with open(os.devnull, 'w') as devnull:
        old_stdout = sys.stdout
        old_stderr = sys.stderr
        sys.stdout = devnull
        sys.stderr = devnull
        try:
            yield
        finally:
            sys.stdout = old_stdout
            sys.stderr = old_stderr

def initialize_models():
    """Initialize the models if they haven't been initialized yet."""
    global image_inference, text_inference
    if image_inference is None or text_inference is None:
        with suppress_stdout_stderr():
            # Initialize the models
            model_path = "llava-v1.6-vicuna-7b:q4_0"
            model_path_text = "gemma-2-2b-instruct:q4_0"

            # Initialize the image inference model
            image_inference = NexaVLMInference(
                model_path=model_path,
                local_path=None,
                stop_words=[],
                temperature=0.3,
                max_new_tokens=256,  # Reduced to speed up processing
                top_k=3,
                top_p=0.2,
                profiling=False
            )

            # Initialize the text inference model
            text_inference = NexaTextInference(
                model_path=model_path_text,
                local_path=None,
                stop_words=[],
                temperature=0.5,
                max_new_tokens=256,  # Reduced to speed up processing
                top_k=3,
                top_p=0.3,
                profiling=False
            )

def get_text_from_generator(generator):
    """Extract text from the generator response."""
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

def generate_image_metadata(image_path):
    """Generate description, folder name, and filename for an image file."""
    initialize_models()

    # Generate description
    description_prompt = "Please provide a detailed description of this image, focusing on the main subject and any important details."
    description_generator = image_inference._chat(description_prompt, image_path)
    description = get_text_from_generator(description_generator).strip()

    # Generate filename
    filename_prompt = f"""Based on the description below, generate a specific and descriptive filename (2-4 words) for the image.
Do not include any data type words like 'image', 'jpg', 'png', etc. Use only letters and connect words with underscores.

Description: {description}

Example:
Description: A photo of a sunset over the mountains.
Filename: sunset_over_mountains

Now generate the filename.

Filename:"""
    filename_response = text_inference.create_completion(filename_prompt)
    filename = filename_response['choices'][0]['text'].strip()
    filename = filename.replace('Filename:', '').strip()
    sanitized_filename = sanitize_filename(filename)

    if not sanitized_filename:
        sanitized_filename = 'untitled_image'

    # Generate folder name from description
    foldername_prompt = f"""Based on the description below, generate a general category or theme (1-2 words) for this image.
This will be used as the folder name. Do not include specific details or words from the filename.

Description: {description}

Example:
Description: A photo of a sunset over the mountains.
Category: landscapes

Now generate the category.

Category:"""
    foldername_response = text_inference.create_completion(foldername_prompt)
    foldername = foldername_response['choices'][0]['text'].strip()
    foldername = foldername.replace('Category:', '').strip()
    sanitized_foldername = sanitize_filename(foldername)

    if not sanitized_foldername:
        sanitized_foldername = 'images'

    return sanitized_foldername, sanitized_filename, description

def process_single_image(image_path):
    """Process a single image file to generate metadata."""
    foldername, filename, description = generate_image_metadata(image_path)
    print(f"File: {image_path}")
    print(f"Description: {description}")
    print(f"Folder name: {foldername}")
    print(f"Generated filename: {filename}")
    print("-" * 50)
    return {
        'file_path': image_path,
        'foldername': foldername,
        'filename': filename,
        'description': description
    }

def process_image_files(image_paths):
    """Process image files using multiprocessing."""
    with Pool(cpu_count()) as pool:
        data_list = pool.map(process_single_image, image_paths)
    return data_list

def summarize_text_content(text):
    """Summarize the given text content."""
    initialize_models()

    prompt = f"""Provide a concise and accurate summary of the following text, focusing on the main ideas and key details.
Limit your summary to a maximum of 150 words.

Text: {text}

Summary:"""

    response = text_inference.create_completion(prompt)
    summary = response['choices'][0]['text'].strip()
    return summary

def generate_text_metadata(input_text):
    """Generate description, folder name, and filename for a text document."""
    initialize_models()

    # Generate description
    description = summarize_text_content(input_text)

    # Generate filename
    filename_prompt = f"""Based on the summary below, generate a specific and descriptive filename (2-4 words) for the document.
Do not include any data type words like 'text', 'document', 'pdf', etc. Use only letters and connect words with underscores.

Summary: {description}

Example:
Summary: A research paper on the fundamentals of string theory.
Filename: string_theory_fundamentals

Now generate the filename.

Filename:"""
    filename_response = text_inference.create_completion(filename_prompt)
    filename = filename_response['choices'][0]['text'].strip()
    filename = filename.replace('Filename:', '').strip()
    sanitized_filename = sanitize_filename(filename)

    if not sanitized_filename:
        sanitized_filename = 'untitled_document'

    # Generate folder name from summary
    foldername_prompt = f"""Based on the summary below, generate a general category or theme (1-2 words) for this document.
This will be used as the folder name. Do not include specific details or words from the filename.

Summary: {description}

Example:
Summary: A research paper on the fundamentals of string theory.
Category: physics

Now generate the category.

Category:"""
    foldername_response = text_inference.create_completion(foldername_prompt)
    foldername = foldername_response['choices'][0]['text'].strip()
    foldername = foldername.replace('Category:', '').strip()
    sanitized_foldername = sanitize_filename(foldername)

    if not sanitized_foldername:
        sanitized_foldername = 'documents'

    return sanitized_foldername, sanitized_filename, description

def process_single_text_file(args):
    """Process a single text file to generate metadata."""
    file_path, text = args
    foldername, filename, description = generate_text_metadata(text)
    print(f"File: {file_path}")
    print(f"Description: {description}")
    print(f"Folder name: {foldername}")
    print(f"Generated filename: {filename}")
    print("-" * 50)
    return {
        'file_path': file_path,
        'foldername': foldername,
        'filename': filename,
        'description': description
    }

def process_text_files(text_tuples):
    """Process text files using multiprocessing."""
    with Pool(cpu_count()) as pool:
        results = pool.map(process_single_text_file, text_tuples)
    return results

def copy_and_rename_files(data_list, new_path, renamed_files, processed_files):
    """Copy and rename files based on generated metadata."""
    for data in data_list:
        file_path = data['file_path']
        if file_path in processed_files:
            continue
        processed_files.add(file_path)

        # Use folder name which generated from the description
        dir_path = create_folder(new_path, data['foldername'])

        # Use filename which generated from the  description
        new_file_name = data['filename'] + os.path.splitext(file_path)[1]
        new_file_path = os.path.join(dir_path, new_file_name)

        # Handle duplicates
        counter = 1
        while new_file_path in renamed_files or os.path.exists(new_file_path):
            new_file_name = f"{data['filename']}_{counter}" + os.path.splitext(file_path)[1]
            new_file_path = os.path.join(dir_path, new_file_name)
            counter += 1

        shutil.copy2(file_path, new_file_path)
        renamed_files.add(new_file_path)
        print(f"Copied and renamed to: {new_file_path}")
        print("-" * 50)