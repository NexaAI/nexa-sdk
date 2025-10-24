"""
NexaAI CV OCR Example

This example demonstrates how to use the NexaAI SDK to perform OCR on an image.
"""

import argparse
import os
from nexaai.cv import CVCapabilities, CVModel, CVModelConfig, CVResults


def main():
    parser = argparse.ArgumentParser(description="NexaAI CV OCR Example")
    parser.add_argument(
        "--model", default="NexaAI/paddleocr-npu", help="model path")
    parser.add_argument("--image",
                        default="path/to/image.png",
                        help="Path to input image")
    parser.add_argument("--plugin-id", default="npu", help="Plugin ID to use")
    args = parser.parse_args()
    model_name= os.path.expanduser(args.model)
    image_path = os.path.expanduser(args.image)
    config = CVModelConfig(capabilities=CVCapabilities.OCR)
    cv = CVModel.from_(name_or_path=model_name,
                       config=config, plugin_id=args.plugin_id)
    results = cv.infer(image_path)

    print(f"Number of results: {results.result_count}")
    for result in results.results:
        print(f"[{result.confidence:.2f}] {result.text}")


if __name__ == "__main__":
    main()
