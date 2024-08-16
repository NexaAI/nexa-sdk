import base64
import os

from nexa.gguf.llama import llama
from nexa.gguf.llama.llama_chat_format import NanoLlavaChatHandler
from tests.utils import download_model

def image_to_base64_data_uri(file_path):
    """
    file_path = 'file_path.png'
    data_uri = image_to_base64_data_uri(file_path)
    """
    with open(file_path, "rb") as img_file:
        base64_data = base64.b64encode(img_file.read()).decode("utf-8")
        return f"data:image/png;base64,{base64_data}"

model_url = "https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nanoLLaVA/model-fp16.gguf"
mmproj_url = "https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nanoLLaVA/projector-fp16.gguf"

# Download paths
output_dir = os.getcwd()
model_path = download_model(model_url, output_dir)
mmproj_path = download_model(mmproj_url, output_dir)
print("Model downloaded to:", model_path)
print("MMProj downloaded to:", mmproj_path)

chat_handler = NanoLlavaChatHandler(clip_model_path=mmproj_path)

def test_image_generation():
    llm = llama.Llama(
        model_path=model_path,
        chat_handler=chat_handler,
        n_ctx=2048,  # n_ctx should be increased to accommodate the image embedding
        n_gpu_layers=-1,  # Uncomment to use GPU acceleration
        verbose=False,
    )
    output = llm.create_chat_completion(
        messages=[
            {
                "role": "system",
                "content": "You are an assistant who perfectly describes images.",
            },
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "What's in this image?"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
                        },
                    },
                ],
            },
        ],
        stream=True,
    )
    for chunk in output:
        delta = chunk["choices"][0]["delta"]
        if "role" in delta:
            print(delta["role"], end=": ")
        elif "content" in delta:
            print(delta["content"], end="")


# if __name__ == "__main__":
#     print("=== Testing 1 ===")
#     test1()
