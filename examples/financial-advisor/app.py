import sys
import os
import streamlit as st
import shutil
import pdfplumber
from sentence_transformers import SentenceTransformer
import faiss
import numpy as np
import re
import traceback
import logging
from utils.financial_analyzer import FinancialAnalyzer

# set up logging:
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# set a default model path & allow override from command line:
default_model = "gemma"
if len(sys.argv) > 1:
    default_model = sys.argv[1]

@st.cache_resource
def load_model(model_path):
    return FinancialAnalyzer(model_path)

def generate_response(query: str) -> str:
    result = st.session_state.nexa_model.financial_analysis(query)
    if isinstance(result, dict) and "error" in result:
        return f"An error occurred: {result['error']}"
    return result

def extract_text_from_pdf(pdf_path):
    try:
        with pdfplumber.open(pdf_path) as pdf:
            text = ''
            for page in pdf.pages:
                text += page.extract_text() + '\n'
        return text
    except Exception as e:
        logger.error(f"Error extracting text from PDF {pdf_path}: {str(e)}")
        return None

def chunk_text(text, model, max_tokens=256, overlap=20):
    try:
        if not text:
            logger.warning("Empty text provided to chunk_text function")
            return []

        sentences = re.split(r'(?<=[.!?])\s+', text)
        chunks = []
        current_chunk = []
        current_tokens = 0

        for sentence in sentences:
            sentence_tokens = len(model.tokenize(sentence))
            if current_tokens + sentence_tokens > max_tokens:
                if current_chunk:
                    chunks.append(' '.join(current_chunk))
                current_chunk = [sentence]
                current_tokens = sentence_tokens
            else:
                current_chunk.append(sentence)
                current_tokens += sentence_tokens

        if current_chunk:
            chunks.append(' '.join(current_chunk))

        logger.info(f"Created {len(chunks)} chunks from text")
        return chunks
    except Exception as e:
        logger.error(f"Error chunking text: {str(e)}")
        logger.error(traceback.format_exc())
        return []

def create_embeddings(chunks, model):
    try:
        if not chunks:
            logger.warning("No chunks provided for embedding creation")
            return None
        embeddings = model.encode(chunks)
        logger.info(f"Created embeddings of shape: {embeddings.shape}")
        return embeddings
    except Exception as e:
        logger.error(f"Error creating embeddings: {str(e)}")
        logger.error(traceback.format_exc())
        return None

def build_faiss_index(embeddings):
    try:
        if embeddings is None or embeddings.shape[0] == 0:
            logger.warning("No valid embeddings provided for FAISS index creation")
            return None
        dimension = embeddings.shape[1]
        index = faiss.IndexFlatL2(dimension)
        index.add(embeddings.astype('float32'))
        logger.info(f"Built FAISS index with {index.ntotal} vectors")
        return index
    except Exception as e:
        logger.error(f"Error building FAISS index: {str(e)}")
        logger.error(traceback.format_exc())
        return None

def process_pdfs(uploaded_files):
    if not uploaded_files:
        st.warning("Please upload PDF files first.")
        return False

    input_dir = "./assets/input"
    output_dir = "./assets/output/processed_data"

    # clear existing files in the input directory:
    if os.path.exists(input_dir):
        shutil.rmtree(input_dir)
    os.makedirs(input_dir, exist_ok=True)

    # save uploaded files to the input directory:
    for uploaded_file in uploaded_files:
        with open(os.path.join(input_dir, uploaded_file.name), "wb") as f:
            f.write(uploaded_file.getbuffer())

    # process PDFs:
    try:
        model = SentenceTransformer('all-MiniLM-L6-v2')
        all_chunks = []

        for filename in os.listdir(input_dir):
            if filename.endswith('.pdf'):
                pdf_path = os.path.join(input_dir, filename)
                text = extract_text_from_pdf(pdf_path)
                if text is None:
                    logger.warning(f"Skipping {filename} due to extraction error")
                    continue
                file_chunks = chunk_text(text, model)
                if file_chunks:
                    all_chunks.extend(file_chunks)
                    st.write(f"Processed {filename}: {len(file_chunks)} chunks")
                else:
                    logger.warning(f"No chunks created for {filename}")

        if not all_chunks:
            st.warning("No valid content found in the uploaded PDFs.")
            return False

        embeddings = create_embeddings(all_chunks, model)
        if embeddings is None:
            st.error("Failed to create embeddings.")
            return False

        index = build_faiss_index(embeddings)
        if index is None:
            st.error("Failed to build FAISS index.")
            return False

        # save the index and chunks:
        os.makedirs(output_dir, exist_ok=True)
        faiss.write_index(index, os.path.join(output_dir, 'pdf_index.faiss'))
        np.save(os.path.join(output_dir, 'pdf_chunks.npy'), all_chunks)

        # verify files were saved & reload the FAISS index:
        if os.path.exists(os.path.join(output_dir, 'pdf_index.faiss')) and \
            os.path.exists(os.path.join(output_dir, 'pdf_chunks.npy')):
                # Reload the FAISS index
                st.session_state.nexa_model.load_faiss_index()
                st.success("PDFs processed and FAISS index reloaded successfully!")
                return True
        else:
            st.error("Error: Processed files not found after saving.")
            return False

    except Exception as e:
        st.error(f"Error processing PDFs: {str(e)}")
        logger.error(f"Error processing PDFs: {str(e)}")
        logger.error(traceback.format_exc())
        return False

def check_faiss_index():
    if "nexa_model" not in st.session_state:
        return False
    return (st.session_state.nexa_model.embeddings_model is not None and 
            st.session_state.nexa_model.index is not None and 
            st.session_state.nexa_model.stored_docs is not None)

# Streamlit app:
def main():
    st.markdown("<h1 style='font-size: 43px;'>On-Device Personal Finance Advisor</h1>", unsafe_allow_html=True)
    st.caption("Powered by Nexa AI SDK🐙")

    # add an empty line:
    st.markdown("<br>", unsafe_allow_html=True)

    if "nexa_model" not in st.session_state:
        st.session_state.nexa_model = load_model(default_model)

    # check if FAISS index exists:
    if not check_faiss_index():
        st.info("No processed financial documents found. Please upload and process PDFs.")

    # step 1 - file upload:
    uploaded_files = st.file_uploader("Choose PDF files", accept_multiple_files=True, type="pdf")

    # step 2 - process PDFs:
    if st.button("Process PDFs"):
        with st.spinner("Processing PDFs..."):
            if process_pdfs(uploaded_files):
                st.success("PDFs processed successfully! You can now use the chat feature.")
                st.rerun()
            else:
                st.error("Failed to process PDFs. Please check the logs for more information.")

    # add a horizontal line:
    st.markdown("---")

    # original sidebar configuration:
    st.sidebar.header("Model Configuration")
    model_path = st.sidebar.text_input("Model path", default_model)

    if not model_path:
        st.warning("Please enter a valid path or identifier for the model in Nexa Model Hub to proceed.")
        st.stop()

    if "nexa_model" not in st.session_state or "current_model_path" not in st.session_state or st.session_state.current_model_path != model_path:
        st.session_state.current_model_path = model_path
        st.session_state.nexa_model = load_model(model_path)
        if st.session_state.nexa_model is None:
            st.stop()

    st.sidebar.header("Generation Parameters")
    params = st.session_state.nexa_model.get_params()
    temperature = st.sidebar.slider("Temperature", 0.0, 1.0, params["temperature"])
    max_new_tokens = st.sidebar.slider("Max New Tokens", 1, 500, params["max_new_tokens"])
    top_k = st.sidebar.slider("Top K", 1, 100, params["top_k"])
    top_p = st.sidebar.slider("Top P", 0.0, 1.0, params["top_p"])

    st.session_state.nexa_model.set_params(
        temperature=temperature,
        max_new_tokens=max_new_tokens,
        top_k=top_k,
        top_p=top_p
    )

    # step 3 - interactive financial analysis chat:
    st.header("Let's discuss your finances🧑‍💼")

    if check_faiss_index():
        if "messages" not in st.session_state:
            st.session_state.messages = []

        for message in st.session_state.messages:
            with st.chat_message(message["role"]):
                st.markdown(message["content"])

        if prompt := st.chat_input("Ask about your financial documents..."):
            st.session_state.messages.append({"role": "user", "content": prompt})
            with st.chat_message("user"):
                st.markdown(prompt)

            with st.chat_message("assistant"):
                response_placeholder = st.empty()
                full_response = ""
                for chunk in generate_response(prompt):
                    choice = chunk["choices"][0]
                    if "delta" in choice:
                        delta = choice["delta"]
                        content = delta.get("content", "")
                    elif "text" in choice:
                        delta = choice["text"]
                        content = delta

                    full_response += content
                    response_placeholder.markdown(full_response, unsafe_allow_html=True)

            st.session_state.messages.append({"role": "assistant", "content": full_response})
    else:
        st.info("Please upload and process PDF files before using the chat feature.")

if __name__ == "__main__":
    main()