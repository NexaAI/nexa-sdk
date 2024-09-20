import os
import time

from file_utils import (
    display_directory_tree,
    collect_file_paths,
    separate_files_by_type,
    read_text_file,
    read_pdf_file,
    read_docx_file  # Importing read_docx_file
)

from data_processing import (
    process_image_files,
    process_text_files,
    copy_and_rename_files
)

def main():
    # Paths configuration
    input_path = "Path/to/your/input/files/or/folder"
    output_path = "Path/to/your/output/files/or/folder"


    if not os.path.exists(input_path):
        print(f"Input path {input_path} does not exist. Please create it and add the necessary files.")
        return

    start_time = time.time()
    file_paths = collect_file_paths(input_path)
    end_time = time.time()

    print(f"Time taken to load file paths: {end_time - start_time:.2f} seconds")
    print("-" * 50)
    print("Directory tree before renaming:")
    display_directory_tree(input_path)

    # Inserted the custom message as requested
    print("*" * 50)
    print("The file upload successful. It will take some minutes")
    print("*" * 50)

    # Separate files by type
    image_files, text_files = separate_files_by_type(file_paths)

    # Process image files
    data_images = process_image_files(image_files)

    # Prepare text tuples for processing
    text_tuples = []
    for fp in text_files:
        ext = os.path.splitext(fp.lower())[1]
        if ext == '.txt':
            text_content = read_text_file(fp)
        elif ext == '.docx':
            text_content = read_docx_file(fp)
        elif ext == '.pdf':
            text_content = read_pdf_file(fp)
        else:
            print(f"Unsupported text file format: {fp}")
            continue  # Skip unsupported file formats
        text_tuples.append((fp, text_content))

    # Process text files
    data_texts = process_text_files(text_tuples)

    # Prepare for copying and renaming
    renamed_files = set()
    processed_files = set()
    os.makedirs(output_path, exist_ok=True)

    # Copy and rename image files
    copy_and_rename_files(data_images, output_path, renamed_files, processed_files)

    # Copy and rename text files
    copy_and_rename_files(data_texts, output_path, renamed_files, processed_files)

    print("Directory tree after copying and renaming:")
    display_directory_tree(output_path)

if __name__ == '__main__':
    main()
