# Copyright 2024-2025 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
NexaAI CV OCR Example

This example demonstrates how to use the NexaAI SDK to perform OCR on an image.
"""

import argparse
import logging
import os

from nexaai.cv import CV
from nexaai import setup_logging


def main():
    setup_logging(level=logging.DEBUG)
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
