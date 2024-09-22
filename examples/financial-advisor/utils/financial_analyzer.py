import os
import faiss
import numpy as np
import logging
from nexa.gguf import NexaTextInference
from sentence_transformers import SentenceTransformer
from langchain_core.documents import Document

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class FinancialAnalyzer:
    def __init__(self, model_path="gemma"):
        self.model_path = model_path
        self.inference = NexaTextInference(
            model_path=self.model_path,
            stop_words=[],
            temperature=0.7,
            max_new_tokens=256,
            top_k=50,
            top_p=0.9,
            profiling=False
        )
        self.embeddings_model = SentenceTransformer('all-MiniLM-L6-v2')
        self.index = None
        self.stored_docs = None
        self.load_faiss_index()

    def get_params(self):
        return self.inference.params

    def set_params(self, **kwargs):
        self.inference.params.update(kwargs)

    def load_faiss_index(self):
        try:
            faiss_index_dir = "./assets/output/processed_data"
            if not os.path.exists(faiss_index_dir):
                logger.warning(f"FAISS index directory not found: {faiss_index_dir}")
                return

            index_file = os.path.join(faiss_index_dir, "pdf_index.faiss")
            if not os.path.exists(index_file):
                logger.warning(f"FAISS index file not found: {index_file}")
                return

            self.index = faiss.read_index(index_file)
            logger.info(f"FAISS index loaded successfully.")

            doc_file = os.path.join(faiss_index_dir, "pdf_chunks.npy")
            self.stored_docs = np.load(doc_file, allow_pickle=True)
            logger.info(f"Loaded {len(self.stored_docs)} documents")

            if not isinstance(self.stored_docs[0], Document):
                self.stored_docs = [Document(page_content=doc) for doc in self.stored_docs]

        except Exception as e:
            logger.error(f"Error loading FAISS index: {str(e)}")

    def custom_search(self, query, k=3):
        if self.embeddings_model is None or self.index is None or self.stored_docs is None:
            logger.error("FAISS index or embeddings model not properly loaded")
            return []
        try:
            query_vector = self.embeddings_model.encode([query])[0]
            scores, indices = self.index.search(np.array([query_vector]), k)
            docs = [self.stored_docs[i] for i in indices[0]]
            return list(zip(docs, scores[0]))
        except Exception as e:
            logger.error(f"Error in custom_search: {str(e)}")
            return []

    def truncate_text(self, text, max_tokens=256):
        tokens = text.split()
        if len(tokens) <= max_tokens:
            return text
        return ' '.join(tokens[:max_tokens])

    def financial_analysis(self, query):
        try:
            if self.embeddings_model is None or self.index is None or self.stored_docs is None:
                logger.error("FAISS index not loaded. Please process PDF files first.")
                return {"error": "FAISS index not loaded. Please process PDF files first."}

            relevant_docs = self.custom_search(query, k=1)
            if not relevant_docs:
                logger.warning("No relevant documents found for the query.")
                return {"error": "No relevant documents found for the query."}

            context = "\n".join([doc.page_content for doc, _ in relevant_docs])
            truncated_context = self.truncate_text(context)
            prompt = f"Financial context: {truncated_context}\n\nAnalyze: {query}"

            prompt_tokens = len(prompt.split())
            logger.info(f"Prompt length: {prompt_tokens} tokens")

            if prompt_tokens > 250:
                prompt = self.truncate_text(prompt, 250)
                logger.info(f"Truncated prompt length: {len(prompt.split())} tokens")

            llm_input = [
                {"role": "user", "content": prompt}
            ]

            return self.inference.create_chat_completion(llm_input, stream=True)

        except Exception as e:
            logger.error(f"Error in financial_analysis: {str(e)}")
            import traceback
            logger.error(traceback.format_exc())
            return {"error": str(e)}