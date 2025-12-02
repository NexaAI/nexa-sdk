"""
NexaAI CV OCR Example

This example demonstrates how to use the NexaAI SDK to perform OCR on an image.
"""

import argparse
import logging
import os

from nexaai.cv import CV


def setup_logging():
    """Setup logging with debug level."""
    logging.basicConfig(
        level=logging.DEBUG,
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    )


def main():
    setup_logging()
    parser = argparse.ArgumentParser(description="NexaAI CV OCR Example")
    parser.add_argument(
        "-m",
        "--model",
        default="NexaAI/paddleocr-npu",
        help="Model path",
    )
    parser.add_argument(
        "--image",
        default="path/to/image.png",
        help="Path to input image",
    )
    parser.add_argument("--plugin-id", default=None, help="Plugin ID to use")
    args = parser.parse_args()

    image_path = os.path.expanduser(args.image)

    if not os.path.exists(image_path):
        raise FileNotFoundError(f"Image file not found: {image_path}")

    cv: CV = CV.from_(
        model=os.path.expanduser(args.model),
        capabilities=0,  # 0=OCR
        plugin_id=args.plugin_id,
    )
    results = cv.infer(image_path)

    print(f"Number of results: {len(results.results)}")
    for result in results.results:
        print(f"[{result.confidence:.2f}] {result.text}")


if __name__ == "__main__":
    main()
