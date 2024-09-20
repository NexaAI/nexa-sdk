
# Local File Organizer

This project is a local file organizer that processes files in a specified input directory, generates metadata using AI models, and organizes them into a structured output directory based on the generated metadata.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Set Up the Python Environment](#2-set-up-the-python-environment)
  - [3. Install Dependencies](#3-install-dependencies)
- [Configuration](#configuration)
- [Usage](#usage)
- [Example Output](#example-output)
- [Notes and Troubleshooting](#notes-and-troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Introduction

The Local File Organizer automates the process of organizing files by:

- Scanning a specified input directory for files.
- Generating descriptions, folder names, and filenames using AI models.
- Organizing the files into a new directory structure based on the generated metadata.

Supported file types include:

- **Images:** `.png`, `.jpg`, `.jpeg`, `.gif`, `.bmp`
- **Text Files:** `.txt`, `.docx`
- **PDFs:** `.pdf`

## Features

- **Automated File Organization:** Automatically sorts files into folders based on AI-generated categories.
- **Metadata Generation:** Generates descriptions and filenames using AI models.
- **Support for Multiple File Types:** Handles images, text files, and PDFs.
- **Parallel Processing:** Utilizes multiprocessing to speed up file processing.
- **Customizable Prompts:** Prompts used for AI model interactions can be customized.

## Prerequisites

- **Operating System:** Compatible with Windows, macOS, and Linux.
- **Python Version:** Python 3.12
- **Conda:** Anaconda or Miniconda installed.
- **Git:** For cloning the repository (or you can download the code as a ZIP file).

## Installation

### 1. Clone the Repository

Clone this repository to your local machine using Git:

```bash
git clone https://github.com/NexaAI/nexa-sdk.git
```

Or download the repository as a ZIP file and extract it to your desired location.

### 2. Set Up the Python Environment

Create a new Conda environment named `local_file_manager` with Python 3.12:

```bash
conda create --name local_file_manager python=3.12
```

Activate the environment:

```bash
conda activate local_file_manager
```

### 3. Install Dependencies

Ensure you are in the project directory and install the required dependencies using `requirements.txt`:

```bash
pip install -r requirements.txt
```

If you don't have a `requirements.txt` file, create one with the following content:

```txt
nexa
Pillow
pytesseract
PyMuPDF
python-docx
```

Then run the install command again.

**Note:** If you encounter issues with any packages, install them individually:

```bash
pip install nexa Pillow pytesseract PyMuPDF python-docx
```

## Configuration

Before running the script, you need to set the input (`base_path`) and output (`new_path`) directories in `main.py`.

Open `main.py` and locate and modify the following lines to point to your desired input and output directories:

```python
# Paths configuration
base_path = "Path/to/your/input/files/or/folder"
new_path = "Path/to/your/output/files/or/folder"
```

**Example:**

```python
# Paths configuration
base_path = "/home/user/documents/input_files"
new_path = "/home/user/documents/organized_files"
```

## Usage

With the environment activated and dependencies installed, run the script using:

```bash
python main.py
```

The script will:

1. Display the directory tree of your input directory.
2. Inform you that the files have been uploaded and processing will begin.
3. Process each file, generating metadata.
4. Copy and rename the files into the output directory based on the generated metadata.
5. Display the directory tree of your output directory after processing.

## Example Output

```
Directory tree before renaming:
Path/to/your/input/files/or/folder
├── image1.jpg
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
│   └── generated_filename1.jpg
├── category2
│   └── generated_filename2.pdf
└── category3
    └── generated_filename3.png

3 directories, 3 files
```

**Note:** The actual descriptions, folder names, and filenames will be generated by the AI models based on your files' content.

<!-- ## Notes and Troubleshooting

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

- **File Permissions:**
  - Ensure the script has read access to the input directory and write access to the output directory.

- **Processing Time:**
  - Processing may take time depending on the number and size of files.
  - The script uses multiprocessing to improve performance.

- **Customizing Prompts:**
  - You can adjust prompts in `data_processing.py` to change how metadata is generated.

- **Error Handling:**
  - The script includes error handling for file reading operations.
  - Check console output for error messages if processing fails.

- **Suppressing Output:**
  - The script suppresses output from the AI models to keep the console output clean.
  - Modify the `suppress_stdout_stderr` function in `data_processing.py` to change this behavior.
 -->
