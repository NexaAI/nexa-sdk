import streamlit as st
from langchain_chroma import Chroma
from nexa_embedding import NexaEmbeddings
import os
from chart_data_generator import execute_chart_generation
from chart_generator import ChartGenerator
from PIL import Image
from nexa.gguf import NexaTextInference
from prompts import DECISION_MAKING_TEMPLATE
from build_db import create_chroma_db
import chromadb

avatar_path = "assets/avatar.jpeg"
persist_directory = "./chroma_db"

@st.cache_resource
def load_models():
    # Load the base model
    chat_model = NexaTextInference(model_path="llama3.2")
    print("Chat model loaded successfully!")

    # Load the decision model
    decision_model = NexaTextInference(model_path="DavidHandsome/Octopus-v2-PDF:gguf-q4_K_M")
    print("Decision model loaded successfully!")

    return chat_model, decision_model

def initialize_session_state():
    if "messages" not in st.session_state:
        st.session_state.messages = []
    if "last_response" not in st.session_state:
        st.session_state.last_response = ""
    if "file_uploaded" not in st.session_state:
        st.session_state.file_uploaded = False
    if "pdf_filename" not in st.session_state:
        st.session_state.pdf_filename = ""

def setup_retriever():
    embeddings = NexaEmbeddings(model_path="nomic")
    local_db = Chroma(
        persist_directory=persist_directory, embedding_function=embeddings
    )
    return local_db.as_retriever()


def retrieve_documents(retriever, query):
    docs = retriever.get_relevant_documents(query)
    return [doc.page_content for doc in docs]


def call_pdf_qa(query, context, chat_model):
    system_prompt = (
        "You are a QA assistant. Based on the following context, answer the question using bullet points and include necessary data.\n\n"
        f"Context:\n{context}"
    )

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": query},
    ]

    try:
        stream = chat_model.create_chat_completion(
            messages=messages,
            max_tokens=2048,
            stream=True
        )
        return stream
    except Exception as e:
        st.error(f"An error occurred while calling QA: {str(e)}")
        return None


def query_information_with_retrieval(prompt, retriever, chat_model):
    with st.chat_message("assistant", avatar=avatar_path):
        with st.spinner("Generating..."):
            # Retrieve documents and prepare context
            retrieved_docs = retrieve_documents(retriever, prompt)
            context = "\n\n".join(retrieved_docs)
            
            # Get the response stream
            stream = call_pdf_qa(prompt, context, chat_model)
            if stream is None:
                return
            
            # Create a placeholder for the streaming response
            response_placeholder = st.empty()
            full_response = ""
        
            # Stream the response and update the placeholder
            for chunk in stream:
                content = chunk["choices"][0]["delta"].get("content", "")
                full_response += content
                response_placeholder.markdown(full_response)
            
            # Update session state after streaming is complete
            st.session_state.messages.append({
                "role": "assistant",
                "content": full_response,
                "avatar": avatar_path
            })
            st.session_state.last_response = full_response

def irrelevant_function():
    with st.chat_message("assistant", avatar=avatar_path):
        with st.spinner("Generating..."):
            # Create a message indicating irrelevance
            irrelevance_message = "I apologize, but your question doesn't seem to be related to the PDF content."
            
            # Display the message
            st.markdown(irrelevance_message)

            # Update session state
            st.session_state.messages.append({
                "role": "assistant",
                "content": irrelevance_message,
                "avatar": avatar_path
            })
            st.session_state.last_response = irrelevance_message

def generate_chart(chart_type):
    """Helper function to generate a chart."""
    result = execute_chart_generation(st.session_state.last_response, chart_type)

    if result is None:
        st.warning("No valid json data was generated from the last response.")
        return None
    
    chart_generator = ChartGenerator()

    if chart_type and "chart_data" in result and result["chart_data"]:
        image_path = chart_generator.plot_chart(result["chart_data"])
        return image_path

    return None


def classify_user_intent(prompt, decision_model):
    if decision_model is None:
        st.error("Decision model is not loaded. Please refresh the page or contact support.")
        return None

    formatted_prompt = DECISION_MAKING_TEMPLATE.format(input=prompt)
    output = decision_model.create_completion(formatted_prompt, stop=["<nexa_end>"])

    return output["choices"][0]["text"].strip()



def add_to_slides(chart_type=None):
    if st.session_state.last_response:
        with st.spinner("Generating chart..."):
            image_path = generate_chart(chart_type)

            if image_path is None:
                return

        st.session_state.messages.append(
            {
                "role": "assistant",
                "avatar": avatar_path,
                "image_path": image_path,
            }
        )

        st.success("Chart generated successfully!")

        # Force a rerun to display the new message
        st.rerun()

def clear_chroma_collection(persist_directory: str):
    """
    Removes all collections from the Chroma database.

    Args:
    persist_directory (str): Path to the Chroma persistence directory
    """
    client = chromadb.PersistentClient(path=persist_directory)

    collections = client.list_collections()
    for collection in collections:
        print(f"Deleting collection: {collection.name}")
        client.delete_collection(collection.name)

    print("All collections have been deleted.")


# Main Streamlit App
def main():
    img = Image.open("assets/avatar.jpeg")

    st.set_page_config(
        page_title="Nexa AI PDF Chatbot",
        page_icon=img,
    )

    # Load the models once
    chat_model, decision_model = load_models()

    st.title("Nexa AI PDF Chatbot")
    initialize_session_state()
    retriever = setup_retriever()

    # Display the chat messages
    for message in st.session_state.messages:
        with st.chat_message(message["role"], avatar=message.get("avatar")):
            if message.get("content"):
                st.markdown(message["content"])
            if message.get("image_path"):
                st.image(message["image_path"], caption="Generated Chart")

    # Display uploaded PDF information
    if st.session_state.file_uploaded:
        st.info(f"PDF uploaded: {st.session_state.pdf_filename}")

    # File upload area
    if not st.session_state.file_uploaded:
        uploaded_file = st.file_uploader("Choose a PDF file", type="pdf")
        if uploaded_file is not None:
            with st.spinner("Processing the PDF file..."):
                # Clear existing collections
                clear_chroma_collection(persist_directory)
                # Save the uploaded file temporarily
                temp_file_path = os.path.join("temp", uploaded_file.name)
                os.makedirs("temp", exist_ok=True)
                with open(temp_file_path, "wb") as f:
                    f.write(uploaded_file.getbuffer())
                # Create the Chroma database
                db = create_chroma_db(pdf_path=temp_file_path)
                # Clean up the temporary file
                os.remove(temp_file_path)

            st.session_state.file_uploaded = True
            st.session_state.pdf_filename = uploaded_file.name
            st.success("File processed successfully!")
            st.rerun()

    if prompt := st.chat_input(placeholder="What would you like to know about the PDF? start your question with <pdf> to trigger RAG"):
        st.chat_message("user").markdown(prompt)
        intent = classify_user_intent(prompt, decision_model)
        st.session_state.messages.append({"role": "user", "content": prompt})
        
        print("intent", intent)
        if intent == "<nexa_0>":        # query_with_pdf 
            query_information_with_retrieval(prompt, retriever, chat_model)
        elif intent == "<nexa_2>":      # generate_slide_column_chart 
            add_to_slides("COLUMN_CLUSTERED")
        elif intent == "<nexa_4>":      # generate_slide_pie_chart
            add_to_slides("PIE")
        else:                           # irrelevant_function 
            irrelevant_function()


if __name__ == "__main__":
    main()
