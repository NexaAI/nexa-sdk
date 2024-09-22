import os
import re
import shutil
from PIL import Image
import pytesseract
import fitz  # PyMuPDF
import docx

def sanitize_filename(name, max_length=50, max_words=5):
    """Sanitize the filename by removing unwanted words and characters."""
    # Remove file extension if present
    name = os.path.splitext(name)[0]
    # Remove unwanted words and data type words
    name = re.sub(
        r'\b(jpg|jpeg|png|gif|bmp|txt|pdf|docx|image|picture|photo|this|that|these|those|here|there|'
        r'please|note|additional|notes|folder|name|sure|heres|a|an|the|and|of|in|'
        r'to|for|on|with|your|answer|should|be|only|summary|summarize|text|category)\b',
        '',
        name,
        flags=re.IGNORECASE
    )
    # Remove non-word characters except underscores
    sanitized = re.sub(r'[^\w\s]', '', name).strip()
    # Replace multiple underscores or spaces with a single underscore
    sanitized = re.sub(r'[\s_]+', '_', sanitized)
    # Convert to lowercase
    sanitized = sanitized.lower()
    # Remove leading/trailing underscores
    sanitized = sanitized.strip('_')
    # Split into words and limit the number of words
    words = sanitized.split('_')
    limited_words = [word for word in words if word]  # Remove empty strings
    limited_words = limited_words[:max_words]
    limited_name = '_'.join(limited_words)
    # Limit length
    return limited_name[:max_length] if limited_name else 'untitled'

def read_docx_file(file_path):
    """Read text content from a .docx file."""
    try:
        doc = docx.Document(file_path)
        full_text = [para.text for para in doc.paragraphs]
        return '\n'.join(full_text)
    except Exception as e:
        print(f"Error reading DOCX file {file_path}: {e}")
        return ""

def read_pdf_file(file_path):
    """Read text content from a PDF file."""
    try:
        doc = fitz.open(file_path)
        # Read only the first few pages to speed up processing
        num_pages_to_read = 3  # Adjust as needed
        full_text = []
        for page_num in range(min(num_pages_to_read, len(doc))):
            page = doc.load_page(page_num)
            full_text.append(page.get_text())
        pdf_content = '\n'.join(full_text)
        return pdf_content
    except Exception as e:
        print(f"Error reading PDF file {file_path}: {e}")
        return ""

def read_image_file(file_path):
    """Extract text from an image file using OCR."""
    try:
        image = Image.open(file_path)
        text = pytesseract.image_to_string(image)
        return text
    except Exception as e:
        print(f"Error reading image file {file_path}: {e}")
        return ""

def read_text_file(file_path):
    """Read text content from a text file."""
    max_chars = 3000  # Limit processing time
    try:
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as file:
            text = file.read(max_chars)
        return text
    except Exception as e:
        print(f"Error reading text file {file_path}: {e}")
        return ""

def display_directory_tree(path):
    """Display the directory tree in a format similar to the 'tree' command, including the full path."""
    def tree(dir_path, prefix=''):
        contents = sorted([c for c in os.listdir(dir_path) if not c.startswith('.')])
        pointers = ['├── '] * (len(contents) - 1) + ['└── '] if contents else []
        for pointer, name in zip(pointers, contents):
            full_path = os.path.join(dir_path, name)
            print(prefix + pointer + name)
            if os.path.isdir(full_path):
                extension = '│   ' if pointer == '├── ' else '    '
                tree(full_path, prefix + extension)
    if os.path.isdir(path):
        print(os.path.abspath(path))
        tree(path)
    else:
        print(os.path.abspath(path))

def create_folder(base_path, foldername):
    """Create a directory for the given folder name."""
    sanitized_folder_name = sanitize_filename(foldername)
    dir_path = os.path.join(base_path, sanitized_folder_name)
    os.makedirs(dir_path, exist_ok=True)
    return dir_path

def collect_file_paths(base_path):
    """Collect all file paths from the base directory or single file."""
    if os.path.isfile(base_path):
        return [base_path]
    else:
        file_paths = [
            os.path.join(root, file)
            for root, _, files in os.walk(base_path)
            for file in files
        ]
        return file_paths

def separate_files_by_type(file_paths):
    """Separate files into images and text files based on their extensions."""
    image_extensions = ('.png', '.jpg', '.jpeg', '.gif', '.bmp')
    text_extensions = ('.txt', '.docx', '.pdf')

    image_files = [fp for fp in file_paths if os.path.splitext(fp.lower())[1] in image_extensions]
    text_files = [fp for fp in file_paths if os.path.splitext(fp.lower())[1] in text_extensions]

    return image_files, text_files  # Return only two values
