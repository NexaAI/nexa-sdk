import base64
import os

from nexa.gguf import NexaVLMInference
from tests.utils import download_model
from nexa.gguf.lib_utils import is_gpu_available
import tempfile

def image_to_base64_data_uri(file_path):
    """
    file_path = 'file_path.png'
    data_uri = image_to_base64_data_uri(file_path)
    """
    with open(file_path, "rb") as img_file:
        base64_data = base64.b64encode(img_file.read()).decode("utf-8")
        return f"data:image/png;base64,{base64_data}"


def test_image_generation():
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_dir = os.path.dirname(os.path.abspath(__file__))
        model_url = "https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nanoLLaVA/model-fp16.gguf"
        mmproj_url = "https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nanoLLaVA/projector-fp16.gguf"

        model = NexaVLMInference(
            model_path="nanollava",
        )
        output = model.create_chat_completion(
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
