import os
from nexa.gguf.llama import llama
from tests.utils import download_model
from nexa.gguf.lib_utils import is_gpu_available
# Constants
TINY_LLAMA_URL = "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main/tinyllama-1.1b-chat-v1.0.Q4_0.gguf"
OUTPUT_DIR = os.getcwd()
MODEL_PATH = download_model(TINY_LLAMA_URL, OUTPUT_DIR)

# Initialize Llama model
def init_llama_model(verbose=False, n_gpu_layers=-1, chat_format=None, embedding=False):
    return llama.Llama(
        model_path=MODEL_PATH,
        verbose=verbose,
        n_gpu_layers=n_gpu_layers if is_gpu_available() else 0,
        chat_format=chat_format,
        embedding=embedding,
    )

# Test text generation from a prompt
def test_text_generation():
    model = init_llama_model()
    output = model(
        "Q: Name the planets in the solar system? A: ",
        max_tokens=512,
        stop=["Q:", "\n"],
        echo=True,
    )
    print(output)

# Test chat completion in streaming mode
def test_streaming():
    model = init_llama_model()
    output = model.create_completion(
        "Q: Name the planets in the solar system? A: ",
        max_tokens=512,
        echo=False,
        stream=True,
    )
    for chunk in output:
        if "choices" in chunk:
            print(chunk["choices"][0]["text"], end="", flush=True)

# Test conversation mode with chat format
def test_create_chat_completion():
    model = init_llama_model(chat_format="llama-2")
    output = model.create_chat_completion(
        messages=[
            {"role": "user", "content": "write a long 1000 word story about a detective"}
        ],
        stream=True,
    )
    for chunk in output:
        delta = chunk["choices"][0]["delta"]
        if "role" in delta:
            print(f'{delta["role"]}: ', end="", flush=True)
        elif "content" in delta:
            print(delta["content"], end="", flush=True)

def test_create_embedding():
    model = init_llama_model(embedding=True)
    embeddings = model.create_embedding("Hello, world!")
    print("Embeddings:\n", embeddings)

# Main execution
# if __name__ == "__main__":
#     print("=== Testing 1 ===")
#     test1()
#     print("=== Testing 2 ===")
#     test2()
#     print("=== Testing 3 ===")
#     test3()
#     print("=== Testing 4 ===")
#     test4()