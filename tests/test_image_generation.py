from nexa.gguf import NexaImageInference
from tempfile import TemporaryDirectory
from .utils import download_model

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
