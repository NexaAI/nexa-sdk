## NexaAI SDK Demo: On-device Personal Finance advisor

### Introduction:

- Key features:

  - On-device processing for data privacy
  - Adjustable parameters (model, temperature, max tokens, top-k, top-p, etc.)
  - FAISS index for efficient similarity search
  - Interactive chat interface for financial queries

- File structure:

  - `app.py`: main Streamlit application
  - `utils/text_generator.py`: handles similarity search and text generation
  - `assets/fake_bank_statements`: fake bank statement for testing purpose

### Setup:

1. Install required packages:

```
pip install -r requirements.txt
```

2. Usage:

- Run the Streamlit app: `streamlit run app.py`
- Upload PDF financial docs (bank statements, SEC filings, etc.) and process them
- Use the chat interface to query your financial data

### Resources:

- [NexaAI | Model Hub](https://nexaai.com/models)
- [NexaAI | Inference with GGUF models](https://docs.nexaai.com/sdk/inference/gguf)
- [GitHub | FAISS](https://github.com/facebookresearch/faiss)
- [Local RAG with Unstructured, Ollama, FAISS and LangChain](https://medium.com/@dirakx/local-rag-with-unstructured-ollama-faiss-and-langchain-35e9dfeb56f1)
