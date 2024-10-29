# Nexa AI PDF Chatbot

## Introduction

This demo is a PDF chatbot that can answer PDF related questions and generate charts based on the response. It uses a combination of an embedding model, a vector database, a function calling model, two LLMs and two LoRA models for chart generation. It is built using Nexa SDK.

## Tech Stack

1. For local rag, we use NexaEmbeddings to embed the PDF content and store it in a ChromaDB.
2. For function calling, we use a Octopus-v2-PDF model that is finetuned from Octopus-V2-2B model for function calling.
3. For query with pdf, we use a Llama3.2-3B-Instruct model to generate the response according to the retrieved content.
4. For chart generation, we use a base model(gemma-2-2b-instruct) and two LoRA models(Column-Chart-LoRA and Pie-Chart-LoRA) for column/pie chart generation.


## Used Models

- [Llama3.2-3B-Instruct](https://nexa.ai/meta/Llama3.2-3B-Instruct/gguf-q4_0/readme)
- [Octopus-v2-PDF](https://nexa.ai/DavidHandsome/Octopus-v2-PDF/gguf-q4_K_M/readme)
- [gemma-2-2b-instruct](https://nexa.ai/google/gemma-2-2b-instruct/gguf-fp16/readme)
- [nomic-embed-text-v1.5](https://nexa.ai/nomic-ai/nomic-embed-text-v1.5/gguf-fp16/readme)
- [Column-Chart-LoRA](https://nexa.ai/DavidHandsome/Column-Chart-LoRA/gguf-fp16/readme)
- [Pie-Chart-LoRA](https://nexa.ai/DavidHandsome/Pie-Chart-LoRA/gguf-fp16/readme)

## Setup

Follow these steps to set up the project:

#### 1. Clone the Repository

```
git clone https://github.com/Davidqian123/Chat-with-PDF-locally.git
cd Chat-with-PDF-locally
```

#### 2. Create a New Conda Environment
Create and activate a new Conda environment to manage dependencies:

```
conda create --name pdf_chat python=3.10
conda activate pdf_chat
```

#### 3. Install Requirements
install the necessary dependencies:

```
pip install -r requirements.txt
```

#### 4. Install Nexa SDK

Follow docs [nexa-sdk](https://github.com/NexaAI/nexa-sdk) to install Nexa SDK.

#### 5. Run Streamlit
Run the application using Streamlit:

```
streamlit run app.py
```

## Resources
- [NexaAI Model Hub](https://nexa.ai/models)
- [Nexa-SDK](https://github.com/NexaAI/nexa-sdk)
- [LangChain](https://docs.langchain.com/docs/)
- [ChromaDB](https://docs.trychroma.com/getting-started)
