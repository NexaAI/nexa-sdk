# RAG with Qwen3VL and Nexa Serve

## 1. About
This project is a lightweight **Retrieval-Augmented Generation (RAG)** system built on top of **[nexa serve](https://github.com/NexaAI/nexa-sdk)** with the **Qwen3VL** multimodal model.  

The system lets you bring your own files — such as **PDFs, Word docs, text files, or images** — and automatically builds a small database from them. When you ask a question, and the model retrieves relevant chunks from your files and responds based on the resources you provided.  

You can run the system directly from the **CLI**, or launch a simple **Gradio UI** for an interactive experience.


## 2. Preparation
Before running this project, make sure you have the **Nexa SDK** installed. Please refer to the [Nexa SDK repository](https://github.com/NexaAI/nexa-sdk) for installation instructions.  

Once installed, you need to download the **Qwen3VL model** with the following command:

```bash
nexa pull NexaAI/Qwen3-VL-4B-Instruct-GGUF
nexa pull djuna/jina-embeddings-v2-small-en-Q5_K_M-GGUF
```

After the model is ready, start the Nexa server in a separate terminal:

```bash
nexa serve
```

Then back to this project, create a new conda environment (optional) and install dependencies:

```bash
# Create a new conda environment (optional)
conda create -n rag-nexa python=3.10 -y
conda activate rag-nexa

# install python dependencies
pip install gradio
pip install -r requirements.txt
```


## 3. Run from CLI
To run the RAG pipeline from the command line:

```bash
python rag_nexa.py --data ./docs
```

### Adding files
- Place your files into the `./docs` folder. Supported formats: **.pdf, .txt, .docx, .png, .jpg, .jpeg, .webp, .bmp**  
- After adding new files, you need to **rebuild** the index by restarting the script or triggering the rebuild function inside the UI.  
  Rebuilding is required because it re-indexes the new files so the model can use them.

Once running, simply type your question in the terminal and the system will answer using your documents.


## 4. Run with Gradio UI
You can also start an interactive **Gradio web UI**:

```bash
python gradio_ui.py
```

Open the browser at [http://127.0.0.1:7860](http://127.0.0.1:7860).  

### Using the UI
- On the **left panel**, you can:
  - Upload new files into the `./docs` folder (PDFs, docs, text, or images).
  - Click **Rebuild** after uploading to refresh the database.
- On the **right panel**, use the chat window to ask questions.
- The model will **stream answers** based on your documents.
