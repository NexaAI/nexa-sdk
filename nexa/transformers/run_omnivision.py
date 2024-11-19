from nexa.transformers.omnivision.processing import NanoVLMProcessor
from nexa.transformers.omnivision.modeling import OminiVLMForConditionalGeneration
import argparse
import torch


model_name = "NexaAIDev/omnivlm-dpo"
image_url = "https://public-storage.nexa4ai.com/public-images/cat.png"


def get_device():
    if torch.cuda.is_available():
        return "cuda"
    elif torch.backends.mps.is_available():
        return "mps"
    return "cpu"


def load_model_and_processor(model_path):
    device = get_device()
    proc_path = "nexa-collaboration/nano-vlm-processor"
    processor = NanoVLMProcessor.from_pretrained(proc_path)
    processor.tokenizer.pad_token = processor.tokenizer.eos_token
    processor.tokenizer.padding_side = "right"

    model_kwargs = {}
    # Adjust dtype based on device
    dtype = torch.bfloat16 if device == "cuda" else torch.float32
    local_model = OminiVLMForConditionalGeneration.from_pretrained(
        model_path,
        torch_dtype=dtype,
        **model_kwargs
    )
    local_model = local_model.to(device)
    return local_model, processor


def process_single_image(processor, image_path, input_prompt=None):
    text = f"<|im_start|>system\nYou are Nano-Omni-VLM, created by Nexa AI. You are a helpful assistant.<|im_end|>\n<|im_start|>user\n{input_prompt}\n<|vision_start|><|image_pad|><|vision_end|><|im_end|>"
    # Changed from Image.open() to handle URLs
    if image_path.startswith('http'):
        from PIL import Image
        import requests
        from io import BytesIO
        response = requests.get(image_path)
        image = Image.open(BytesIO(response.content)).convert('RGB')
    else:
        image = Image.open(image_path).convert('RGB')
    inputs = processor(
        text=[text],
        images=[image],
        padding=True,
        return_tensors="pt",
    )
    return inputs.to(get_device())


def generate_output(model, processor, inputs, max_tokens):
    cur_ids = inputs['input_ids']
    cur_attention_mask = inputs['attention_mask']
    input_token_length = cur_ids.shape[-1]
    for _ in range(max_tokens):
        out = model(
            cur_ids,
            attention_mask=cur_attention_mask,
            pixel_values=inputs['pixel_values'],
            use_cache=False
        )
        next_token = out.logits[:, -1].argmax()
        next_word = processor.decode(next_token)
        cur_ids = torch.cat([cur_ids, next_token.unsqueeze(0).unsqueeze(0)], dim=-1)
        cur_attention_mask = torch.cat([cur_attention_mask, torch.ones_like(next_token).unsqueeze(0).unsqueeze(0)], dim=-1)
        if next_word in ("<|im_end|>"):
            break
    return processor.batch_decode(cur_ids[:, input_token_length:])[0]

def main(args):
    model, processor = load_model_and_processor(args.model_path)
    inputs = process_single_image(processor, args.image_path, args.input_prompt)
    output = generate_output(model, processor, inputs, args.max_tokens)
    print("=== Inference Result ===\n", output)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Inference script for Nano-Omni-VLM")
    parser.add_argument("--model_path", default=model_name, help="Path to the model checkpoint")
    # Add image_path argument
    parser.add_argument("--image_path", default=image_url, help="Path to input image or image URL")
    parser.add_argument("--input_prompt", type=str, default="Describe this image for me", help="Input prompt for instruct task")
    parser.add_argument("--max_tokens", type=int, default=512, help="Maximum number of tokens to generate")

    args = parser.parse_args()
    main(args)