from nexa.gguf import NexaVLMInference
from tempfile import TemporaryDirectory
from .utils import download_model
import os

vlm = NexaVLMInference(
    model_path="llava-phi-3-mini:q4_0",
    local_path=None,
)

# Test create_chat_completion
def test_create_chat_completion():
    messages = [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "What is the capital of France?"}
    ]
    completion = vlm.create_chat_completion(
        messages=messages,
        max_tokens=50,
        temperature=0.7,
        top_p=0.95,
        top_k=40,
        stream=False
    )
    
    assert isinstance(completion, dict)
    assert "choices" in completion
    assert len(completion["choices"]) > 0
    assert "message" in completion["choices"][0]
    assert "content" in completion["choices"][0]["message"]
    print("create_chat_completion test passed")

# Test _chat method
def test_chat():
    with TemporaryDirectory() as temp_dir:
        # Download a sample image
        img_url = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
        img_path = download_model(img_url, temp_dir)

        # Test _chat with image
        chat_output = vlm._chat("Describe this image", image_path=img_path)
        
        # Check if the output is an iterator
        assert hasattr(chat_output, '__iter__')
        
        # Collect the output
        output_text = ""
        for chunk in chat_output:
            assert "choices" in chunk
            assert len(chunk["choices"]) > 0
            assert "delta" in chunk["choices"][0]
            delta = chunk["choices"][0]["delta"]
            if "content" in delta:
                output_text += delta["content"]
        
        assert len(output_text) > 0
        print("_chat test with image passed")

if __name__ == "__main__":
    print("=== Testing 1 ===")
    test_create_chat_completion()
    print("=== Testing 2 ===")
    test_chat()