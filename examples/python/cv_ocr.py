"""
NexaAI CV OCR Example

This example demonstrates how to use the NexaAI SDK to perform OCR on an image.
"""

import os
from nexaai.cv import CVCapabilities, CVModel, CVModelConfig, CVResults


def main():
    det_model_path = os.path.expanduser(
        "~/.cache/nexa.ai/nexa_sdk/models/nexaml/paddle-ocr-mlx/ch_ptocr_v4_det_infer.safetensors")
    rec_model_path = os.path.expanduser(
        "~/.cache/nexa.ai/nexa_sdk/models/nexaml/paddle-ocr-mlx/ch_ptocr_v4_rec_infer.safetensors")

    config = CVModelConfig(capabilities=CVCapabilities.OCR,
                           det_model_path=det_model_path, rec_model_path=rec_model_path)

    cv: CVModel = CVModel.from_(
        name_or_path=det_model_path, config=config, plugin_id="mlx")

    results: CVResults = cv.infer(os.path.expanduser(
        "~/.cache/nexa.ai/nexa_sdk/models/nexaml/paddle-ocr-mlx/20250406-170821.jpeg"))

    print(f"Number of results: {results.result_count}")
    for result in results.results:
        print(f"[{result.confidence:.2f}] {result.text}")


if __name__ == "__main__":
    main()
