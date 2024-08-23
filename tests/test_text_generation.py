from nexa.gguf import NexaTextInference
from nexa.gguf.lib_utils import is_gpu_available

model = NexaTextInference(
    model_path="gemma",
    local_path=None,
    verbose=False,
    n_gpu_layers=-1 if is_gpu_available() else 0,
    chat_format="llama-2",
)

# Test text generation from a prompt
def test_text_generation():
    global model
    output = model.create_completion(
        "Q: Name the planets in the solar system? A: ",
        max_tokens=512,
        stop=["Q:", "\n"],
        echo=True,
    )
    # print(output)
    # TODO: add assertions here

# Test chat completion in streaming mode
def test_streaming():
    global model
    output = model.create_completion(
        "Q: Name the planets in the solar system? A: ",
        max_tokens=512,
        echo=False,
        stream=True,
    )
    for chunk in output:
        if "choices" in chunk:
            print(chunk["choices"][0]["text"], end="", flush=True)
    # TODO: add assertions here

# Test conversation mode with chat format
def test_create_chat_completion():
    global model

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
    model = NexaTextInference(
        model_path="gemma",
        verbose=False,
        n_gpu_layers=-1 if is_gpu_available() else 0,
        chat_format="llama-2",
        embedding=True,
    )    
    embeddings = model.create_embedding("Hello, world!")
    print("Embeddings:\n", embeddings)

# Main execution
if __name__ == "__main__":
    print("=== Testing 1 ===")
    test_text_generation()
    print("=== Testing 2 ===")
    test_streaming()
    print("=== Testing 3 ===")
    test_create_chat_completion()
    print("=== Testing 4 ===")
    test_create_embedding()