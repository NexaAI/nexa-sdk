from nexa.gguf import NexaImageInference
from tempfile import TemporaryDirectory
import os
import subprocess

def download_model(url, output_dir):
    """
    Download a file from a given URL using curl, if it doesn't already exist.

    Args:
    - url: str, the URL of the file to download.
    - output_dir: str, the directory where the file should be saved.

    Returns:
    - str: The path to the downloaded file.
    """
    file_name = url.split("/")[-1]
    output_path = os.path.join(output_dir, file_name)

    if os.path.exists(output_path):
        print(f"File {file_name} already exists in {output_dir}. Skipping download.")
        return output_path

    try:
        subprocess.run(["curl", url, "--output", output_path], check=True)
        print(f"Downloaded {file_name} to {output_dir}")
    except subprocess.CalledProcessError as e:
        print(f"Failed to download {file_name}: {e}")
        raise

    return output_path

sd = NexaImageInference(
    model_path="sd1-4",
    local_path=None,
    wtype="q4_0",
)


# Test text-to-image generation
def test_txt_to_img():
    global sd
    output = sd.txt2img("a lovely cat", width=128, height=128, sample_steps=2)
    output[0].save("output_txt_to_img.png")

# Test image-to-image generation
def test_img_to_img():
    
    global sd
    img_url = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"        
    with TemporaryDirectory() as temp_dir:
        img_path = download_model(img_url, temp_dir)
        output = sd.img2img(
            image_path=img_path,  
            prompt="blue sky",  
            width=128,
            height=128,
            negative_prompt="black soil",
            sample_steps=2
        )

# Main execution
if __name__ == "__main__":
    test_txt_to_img()
    test_img_to_img()
