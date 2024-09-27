# Local File Organizer: AI File Management Run Entirely on Your Device, Privacy Assured

Tired of digital clutter? Overwhelmed by disorganized files scattered across your computer? Let AI do the heavy lifting! The Local File Organizer is your personal organizing assistant, using cutting-edge AI to bring order to your file chaos - all while respecting your privacy.

## A Glimpse of How It Works

```
--------------------------------------------------
Enter the path of the directory you want to organize: /home/user/documents/input_files
--------------------------------------------------
Enter the path to store organized files and folders (press Enter to use 'organized_folder' in the input directory)
Output path successfully upload: /home/user/documents/organzied_folder
--------------------------------------------------
Time taken to load file paths: 0.00 seconds
--------------------------------------------------
Directory tree before renaming:
Path/to/your/input/files/or/folder
├── image.jpg
├── document.pdf
├── notes.txt
└── sub_directory
    └── picture.png

1 directory, 4 files
*****************
The files have been uploaded successfully. Processing will take a few minutes.
*****************
File: Path/to/your/input/files/or/folder/image1.jpg
Description: [Generated description]
Folder name: [Generated folder name]
Generated filename: [Generated filename]
--------------------------------------------------
File: Path/to/your/input/files/or/folder/document.pdf
Description: [Generated description]
Folder name: [Generated folder name]
Generated filename: [Generated filename]
--------------------------------------------------
... [Additional files processed]
Directory tree after copying and renaming:
Path/to/your/output/files/or/folder
├── category1
│   └── generated_filename.jpg
├── category2
│   └── generated_filename.pdf
└── category3
    └── generated_filename.png

3 directories, 3 files
```

## What It Does

This intelligent file organizer harnesses the power of advanced AI models, including language models (LMs) and vision-language models (VLMs), to automate the process of organizing files by:


* Scanning a specified input directory for files.
* Content Understanding: 
  - **Textual Analysis**: Uses the [Gemma-2-2B](https://nexaai.com/google/gemma-2-2b-instruct/gguf-q4_0/file) language model (LM) to analyze and summarize text-based content, generating relevant descriptions and filenames.
  - **Visual Content Analysis**: Uses the [LLaVA-v1.6](https://nexaai.com/liuhaotian/llava-v1.6-vicuna-7b/gguf-q4_0/file) vision-language model (VLM), based on Vicuna-7B, to interpret visual files such as images, providing context-aware categorization and descriptions.

* Understanding the content of your files (text, images, and more) to generate relevant descriptions, folder names, and filenames.
* Organizing the files into a new directory structure based on the generated metadata.

The best part? All AI processing happens 100% on your local device using the [Nexa SDK](https://github.com/NexaAI/nexa-sdk). No internet connection required, no data leaves your computer, and no AI API is needed - keeping your files completely private and secure.

We hope this tool can help bring some order to your digital life, making file management a little easier and more efficient.

## Features

- **Automated File Organization:** Automatically sorts files into folders based on AI-generated categories.
- **Intelligent Metadata Generation:** Creates descriptions and filenames using advanced AI models.
- **Support for Multiple File Types:** Handles images, text files, and PDFs.
- **Parallel Processing:** Utilizes multiprocessing to speed up file processing.
- **Customizable Prompts:** Prompts used for AI model interactions can be customized.

## Supported file types

- **Images:** `.png`, `.jpg`, `.jpeg`, `.gif`, `.bmp`
- **Text Files:** `.txt`, `.docx`
- **PDFs:** `.pdf`

## Prerequisites

- **Operating System:** Compatible with Windows, macOS, and Linux.
- **Python Version:** Python 3.12
- **Conda:** Anaconda or Miniconda installed.
- **Git:** For cloning the repository (or you can download the code as a ZIP file).

## Installation

### 1. Clone the Repository

Clone this repository to your local machine using Git:

```zsh
git clone https://github.com/QiuYannnn/Local-File-Organizer.git
```

Or download the repository as a ZIP file and extract it to your desired location.

### 2. Set Up the Python Environment

Create a new Conda environment named `local_file_organizer` with Python 3.12:

```zsh
conda create --name local_file_organizer python=3.12
```

Activate the environment:

```zsh
conda activate local_file_organizer
```

### 3. Install Nexa SDK 🛠️

#### CPU Installation
To install the CPU version of Nexa SDK, run:
```bash
pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cpu --extra-index-url https://pypi.org/simple --no-cache-dir
```

#### GPU Installation (Metal - macOS)
For the GPU version supporting Metal (macOS), run:
```bash
CMAKE_ARGS="-DGGML_METAL=ON -DSD_METAL=ON" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple --no-cache-dir
```
For detailed installation instructions of Nexa SDK for **CUDA** and **AMD GPU** support, please refer to the [Installation section](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#installation) in the main README.


### 4. Install Dependencies

Ensure you are in the project directory and install the required dependencies using `requirements.txt`:

```zsh
pip install -r requirements.txt
```

**Note:** If you encounter issues with any packages, install them individually:

```zsh
pip install nexa Pillow pytesseract PyMuPDF python-docx
```

With the environment activated and dependencies installed, run the script using:
## Running the Script
```zsh
python main.py
```

The script will:

1. Display the directory tree of your input directory.
2. Inform you that the files have been uploaded and processing will begin.
3. Process each file, generating metadata.
4. Copy and rename the files into the output directory based on the generated metadata.
5. Display the directory tree of your output directory after processing.

**Note:** The actual descriptions, folder names, and filenames will be generated by the AI models based on your files' content.

#### Enter the Input Path
You will be prompted to enter the path of the directory where the files you want to organize are stored. Enter the full path to that directory and press Enter.

```zsh
Enter the path of the directory you want to organize: /path/to/your/input_folder
```

#### Enter the Output Path
Next, you will be prompted to enter the path where you want the organized files to be stored. You can either specify a directory or press Enter to use the default directory (organzied_folder) inside the input directory.

```zsh
Enter the path to store organized files and folders (press Enter to use 'organzied_folder' in the input directory): /path/to/your/output_folder
```
If you press Enter without specifying a path, the script will create a folder named organzied_folder in the input directory to store the organized files.

## Notes

- **SDK Models:**
  - The script uses `NexaVLMInference` and `NexaTextInference` models.
  - Ensure you have access to these models and they are correctly set up.
  - You may need to download model files or configure paths.

- **Dependencies:**
  - **pytesseract:** Requires Tesseract OCR installed on your system.
    - **macOS:** `brew install tesseract`
    - **Ubuntu/Linux:** `sudo apt-get install tesseract-ocr`
    - **Windows:** Download from [Tesseract OCR Windows Installer](https://github.com/UB-Mannheim/tesseract/wiki)
  - **PyMuPDF (fitz):** Used for reading PDFs.

- **Processing Time:**
  - Processing may take time depending on the number and size of files.
  - The script uses multiprocessing to improve performance.

- **Customizing Prompts:**
  - You can adjust prompts in `data_processing.py` to change how metadata is generated.