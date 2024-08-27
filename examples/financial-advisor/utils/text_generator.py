import os
import faiss
import numpy as np
import logging
from nexa.gguf import NexaTextInference
from langchain_community.embeddings import HuggingFaceEmbeddings
from langchain_core.documents import Document

# set up logging:
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# model initialization:
model_path = "gemma"
inference = NexaTextInference(
    model_path=model_path,
    stop_words=[],
    temperature=0.7,
    max_new_tokens=256,
    top_k=50,
    top_p=0.9,
    profiling=False
)

print(f"Model loaded: {inference.downloaded_path}")
print(f"Chat format: {inference.chat_format}")

# global variables:
embeddings = None
index = None
stored_docs = None

# load FAISS index:
def load_faiss_index():
    global embeddings, index, stored_docs
    try:
        embeddings = HuggingFaceEmbeddings(model_name="sentence-transformers/all-MiniLM-L6-v2")

        faiss_index_dir = "./assets/output/processed_data"

        if not os.path.exists(faiss_index_dir):
            logger.warning(f"FAISS index directory not found: {faiss_index_dir}")
            return None, None, None

        index_file = os.path.join(faiss_index_dir, "pdf_index.faiss")
        if not os.path.exists(index_file):
            logger.warning(f"FAISS index file not found: {index_file}")
            return None, None, None

        index = faiss.read_index(index_file)
        logger.info(f"FAISS index loaded successfully.")

        # load the chunks:
        doc_file = os.path.join(faiss_index_dir, "pdf_chunks.npy")
        stored_docs = np.load(doc_file, allow_pickle=True)
        logger.info(f"Loaded {len(stored_docs)} documents")

        # convert stored_docs to a list of Document objects:
        if not isinstance(stored_docs[0], Document):
            stored_docs = [Document(page_content=doc) for doc in stored_docs]

        return embeddings, index, stored_docs

    except Exception as e:
        logger.error(f"Error loading FAISS index: {str(e)}")
        return None, None, None

# load the index at module level:
embeddings, index, stored_docs = load_faiss_index()

def custom_search(query, k=3):
    global embeddings, index, stored_docs
    if embeddings is None or index is None or stored_docs is None:
        logger.error("FAISS index or embeddings not properly loaded")
        return []
    try:
        query_vector = embeddings.embed_query(query)
        scores, indices = index.search(np.array([query_vector]), k)
        docs = [stored_docs[i] for i in indices[0]]
        return list(zip(docs, scores[0]))
    except Exception as e:
        logger.error(f"Error in custom_search: {str(e)}")
        return []

# truncate text to a specific token limit:
def truncate_text(text, max_tokens=256):
    tokens = text.split()
    if len(tokens) <= max_tokens:
        return text
    return ' '.join(tokens[:max_tokens])

# query FAISS and generate LLM response:
def financial_analysis(query):
    global embeddings, index, stored_docs
    try:
        if embeddings is None or index is None or stored_docs is None:
            logger.error("FAISS index not loaded. Please process PDF files first.")
            return {"error": "FAISS index not loaded. Please process PDF files first."}

        relevant_docs = custom_search(query, k=1)
        if not relevant_docs:
            logger.warning("No relevant documents found for the query.")
            return {"error": "No relevant documents found for the query."}

        context = "\n".join([doc.page_content for doc, _ in relevant_docs])

        # truncate the context if it's too long:
        truncated_context = truncate_text(context)

        prompt = f"Financial context: {truncated_context}\n\nAnalyze: {query}"

        prompt_tokens = len(prompt.split())
        logger.info(f"Prompt length: {prompt_tokens} tokens")

        if prompt_tokens > 250:
            prompt = truncate_text(prompt, 250)
            logger.info(f"Truncated prompt length: {len(prompt.split())} tokens")

        llm_input = [
            {"role": "user", "content": prompt}
        ]

        # return the iterator
        return inference.create_chat_completion(llm_input, stream=True)

    except Exception as e:
        logger.error(f"Error in financial_analysis: {str(e)}")
        import traceback
        logger.error(traceback.format_exc())
        return {"error": str(e)}
