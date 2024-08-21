from nexa.gguf import NexaVLMInference
import tempfile
from .utils import download_model

def test_image_generation():
    with tempfile.TemporaryDirectory() as temp_dir:
        model = NexaVLMInference(
            model_path="nanollava",
        )
        image_path = download_model(
            "https://www.wikipedia.org/portal/wikipedia.org/assets/img/Wikipedia-logo-v2.png",
            temp_dir,
        )
        output = model._chat("what's in this image?", image_path)
        for chunk in output:
            delta = chunk["choices"][0]["delta"]
            if "role" in delta:
                print(delta["role"], end=": ")
            elif "content" in delta:
                print(delta["content"], end="")


# if __name__ == "__main__":
#     print("=== Testing 1 ===")
#     test1()
