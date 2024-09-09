from multiprocessing import Pool, cpu_count
from llama_index.core import SimpleDirectoryReader, Document
from llama_index.core.node_parser import TokenTextSplitter
import subprocess
import time

def process_document(args):
    # Unpack the arguments
    text, metadata, chunk_size = args
    
    # Create a TokenTextSplitter with the specified chunk size
    splitter = TokenTextSplitter(chunk_size=chunk_size)
    
    # Split the text into chunks
    contents = splitter.split_text(text)
    
    # Take the first chunk if available, otherwise use an empty string
    text = contents[0] if contents else ""
    
    # Return a Document object and the file path
    return Document(text=text, metadata=metadata), metadata['file_path']

def load_documents_multiprocessing(path: str):
    # Create a SimpleDirectoryReader to read documents from the specified path
    reader = SimpleDirectoryReader(
        input_dir=path,
        recursive=True,
        required_exts=[".pdf", ".txt", ".png", ".jpg", ".jpeg"]
    )
    
    chunk_size = 6144  # Define the chunk size for splitting text
    
    # Create a pool of worker processes, with the number of processes equal to the number of CPUs
    with Pool(cpu_count()) as pool:
        # Use pool.map to distribute the work of processing documents across the worker processes
        results = pool.map(process_document, [(d.text, d.metadata, chunk_size) for docs in reader.iter_data() for d in docs])
    
    # Extract documents and file paths from the results
    documents = [document for document, _ in results]
    file_paths = [file_path for _, file_path in results]
    
    return documents, file_paths

def print_tree_with_subprocess(path):
    # Run the 'tree' command and capture its output
    result = subprocess.run(['tree', path], capture_output=True, text=True)
    # Print the output of the 'tree' command
    print(result.stdout)

if __name__ == '__main__':
    path = "/Users/q/nexa_test/llama-fs/sample_data"  # Replace with your actual directory path
    
    # Measure the time taken to load documents
    start_time = time.time()
    documents, file_paths = load_documents_multiprocessing(path)
    end_time = time.time()
    
    # Print the time taken
    print(f"Time taken to load documents: {end_time - start_time:.2f} seconds")
    print("-"*50)
    # Print the directory tree
    print_tree_with_subprocess(path)