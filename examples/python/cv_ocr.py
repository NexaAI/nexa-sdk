"""
NexaAI CV OCR Example

This example demonstrates how to use the NexaAI SDK to perform OCR on an image.
"""

import argparse
import os
from nexaai.cv import CVCapabilities, CVModel, CVModelConfig, CVResults


def main():
    parser = argparse.ArgumentParser(description="NexaAI CV OCR Example")
    parser.add_argument("--det-model", 
                       default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/paddle-ocr-mlx/ch_ptocr_v4_det_infer.safetensors",
                       help="Path to detection model")
    parser.add_argument("--rec-model",
                       default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/paddle-ocr-mlx/ch_ptocr_v4_rec_infer.safetensors", 
                       help="Path to recognition model")
    parser.add_argument("--image",
                       default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/paddle-ocr-mlx/test_input.jpg",
                       help="Path to input image")
    parser.add_argument("--plugin-id", default="cpu_gpu", help="Plugin ID to use")
    args = parser.parse_args()

    det_model_path = os.path.expanduser(args.det_model)
    rec_model_path = os.path.expanduser(args.rec_model)
    image_path = os.path.expanduser(args.image)

    config = CVModelConfig(capabilities=CVCapabilities.OCR,
                           det_model_path=det_model_path, rec_model_path=rec_model_path)

    cv = CVModel.from_(name_or_path=det_model_path, config=config, plugin_id=args.plugin_id)
    results = cv.infer(image_path)

    print(f"Number of results: {results.result_count}")
    for result in results.results:
        print(f"[{result.confidence:.2f}] {result.text}")


if __name__ == "__main__":
    main()
