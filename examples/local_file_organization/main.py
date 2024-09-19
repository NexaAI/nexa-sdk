import os
import time

from file_utils import (
    display_directory_tree,
    collect_file_paths,
    separate_files_by_type,
    read_text_file,
    read_pdf_file
)

from data_processing import (
    process_image_files,
    process_text_files,
    copy_and_rename_files
)

def main():
    # Paths configuration
    base_path = "/Users/q/nexa/nexa_sdk_local_file_organization/nexa-sdk/examples/local_file_organization/sample_data"
    new_path = "/Users/q/nexa/nexa_sdk_local_file_organization/nexa-sdk/examples/local_file_organization/renamed_files"

    if not os.path.exists(base_path):
        print(f"Directory {base_path} does not exist. Please create it and add the necessary files.")
        return

    start_time = time.time()
    file_paths = collect_file_paths(base_path)
    end_time = time.time()

    print(f"Time taken to load file paths: {end_time - start_time:.2f} seconds")
    print("-" * 50)
    print("Directory tree before renaming:")
    display_directory_tree(base_path)

    # Separate files by type
    image_files, text_files, pdf_files = separate_files_by_type(file_paths)

    # Process image files
    data_images = process_image_files(image_files)

    # Process text files
    text_tuples = [(fp, read_text_file(fp)) for fp in text_files]

    # Process PDF files
    pdf_tuples = [(fp, read_pdf_file(fp)) for fp in pdf_files]

    # Combine text and PDF tuples
    text_pdf_tuples = text_tuples + pdf_tuples

    # Process text and PDF files
    data_texts = process_text_files(text_pdf_tuples)

    # Prepare for copying and renaming
    renamed_files = set()
    processed_files = set()
    os.makedirs(new_path, exist_ok=True)

    # Copy and rename image files
    copy_and_rename_files(data_images, new_path, renamed_files, processed_files)

    # Copy and rename text and PDF files
    copy_and_rename_files(data_texts, new_path, renamed_files, processed_files)

    print("Directory tree after copying and renaming:")
    display_directory_tree(new_path)

if __name__ == '__main__':
    main()
