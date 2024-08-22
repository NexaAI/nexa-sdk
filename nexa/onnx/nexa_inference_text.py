import argparse
import logging
import sys
import time
from pathlib import Path
from typing import Any, Tuple

from optimum.onnxruntime import ORTModelForCausalLM
from transformers import AutoTokenizer, TextStreamer
from nexa.general import pull_model
from nexa.constants import NEXA_RUN_MODEL_MAP_ONNX
from nexa.utils import nexa_prompt, SpinningCursorAnimation

logging.basicConfig(level=logging.INFO)


class NexaTextInference:
    """
    A class used for load text models and run text generation.

    Methods:
        run: Run the text generation loop.
        run_streamlit: Run the Streamlit UI.

    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str): Local path of the model.
    profiling (bool): Enable timing measurements for the generation process.
    streamlit (bool): Run the inference in Streamlit UI.
    temperature (float): Temperature for sampling.
    min_new_tokens (int): Minimum number of new tokens to generate.
    max_new_tokens (int): Maximum number of new tokens to generate.
    top_k (int): Top-k sampling parameter.
    top_p (float): Top-p sampling parameter
    """

    def __init__(self, model_path, local_path=None, **kwargs):
        self.model_path = NEXA_RUN_MODEL_MAP_ONNX.get(model_path, model_path)
        self.params = {
            "temperature": 0.5,
            "max_new_tokens": 256,
            "min_new_tokens": 8,
            "top_k": 50,
            "top_p": 1.0,
        }
        self.params.update(kwargs)
        self.model = None
        self.tokenizer = None
        self.streamer = None
        self.downloaded_onnx_folder = local_path
        self.timings = kwargs.get("timings", False)
        self.device = "cpu"

    @SpinningCursorAnimation()
    def _load_model_and_tokenizer(self) -> Tuple[Any, Any, Any, bool]:
        logging.debug(f"Loading model from {self.downloaded_onnx_folder}")
        start_time = time.time()
        self.tokenizer = AutoTokenizer.from_pretrained(self.downloaded_onnx_folder)

        if self.model is None:
            logging.debug(f"Loading model from {self.downloaded_onnx_folder}")
            try:
                self.model = ORTModelForCausalLM.from_pretrained(
                    self.downloaded_onnx_folder
                ).to(self.device)
                self.streamer = TextStreamer(
                    self.tokenizer, skip_prompt=True, skip_special_tokens=True
                )
            except Exception as e:
                logging.error(f"Error loading with ORTModel: {e}", exc_info=True)
                return

        load_time = time.time() - start_time
        logging.debug(f"Model loaded in {load_time:.2f} seconds")

    def start(self, chat_mode=True):
        conversation_history = [] if chat_mode else None
        chat_template = self.tokenizer.chat_template if chat_mode else None

        while True:
            try:
                if not (user_input := nexa_prompt()):
                    continue

                start_time = time.time()

                if chat_mode:
                    conversation_history.append({"role": "user", "content": user_input})
                    full_prompt = self.tokenizer.apply_chat_template(
                        conversation_history,
                        chat_template=chat_template,
                        tokenize=False,
                    )
                    inputs = self.tokenizer(full_prompt, return_tensors="pt")
                else:
                    inputs = self.tokenizer(user_input, return_tensors="pt")
                inputs = {k: v.to(self.device) for k, v in inputs.items()}

                prefill_end_time = time.time()
                prefill_time = prefill_end_time - start_time
                prefill_token_count = len(inputs["input_ids"][0])

                generation_start_time = time.time()
                output = self.model.generate(
                    **inputs,
                    min_new_tokens=self.params["min_new_tokens"],
                    max_new_tokens=self.params["max_new_tokens"],
                    do_sample=True,
                    temperature=self.params["temperature"],
                    streamer=self.streamer,
                    top_k=self.params["top_k"],
                    top_p=self.params["top_p"],
                    pad_token_id=self.tokenizer.eos_token_id,
                )
                end_time = time.time()

                total_inference_time = end_time - start_time
                generation_time = end_time - generation_start_time
                tokens_generated = len(output[0]) - prefill_token_count

                if self.timings:
                    logging.info("Timing statistics:")
                    logging.info(f"Total inference time: {total_inference_time:.2f}s")
                    logging.info(
                        f"Prefill speed: {prefill_token_count/prefill_time:.2f} tokens/s"
                    )
                    logging.info(
                        f"Decode speed: {tokens_generated/generation_time:.2f} tokens/s"
                    )

                response = self.tokenizer.decode(
                    output[0][prefill_token_count:], skip_special_tokens=True
                )
                if chat_mode:
                    conversation_history.append(
                        {"role": "assistant", "content": response}
                    )
            except KeyboardInterrupt:
                print("\n")
            except Exception as e:
                logging.error(f"Error during text generation: {e}", exc_info=True)

    def run(self):
        # Check if Streamlit mode should be run first
        if self.params.get("streamlit"):
            self.run_streamlit()
        else:
            if self.downloaded_onnx_folder is None:
                self.downloaded_onnx_folder, run_type = pull_model(self.model_path)

            if self.downloaded_onnx_folder is None:
                logging.error(
                    f"Model ({model_path}) is not applicable. Please refer to our docs for proper usage.",
                    exc_info=True,
                )
                exit(1)
            
            self._load_model_and_tokenizer()

            if self.model is None or self.tokenizer is None or self.streamer is None:
                logging.error(
                    "Failed to load model or tokenizer. Exiting.", exc_info=True
                )
                exit(1)

            chat_mode = (
                hasattr(self.tokenizer, "chat_template")
                and self.tokenizer.chat_template is not None
            )

            self.start(chat_mode=chat_mode)

    def run_streamlit(self, model_path: str):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")
        from streamlit.web import cli as stcli

        streamlit_script_path = (
            Path(__file__).resolve().parent / "streamlit" / "streamlit_text_chat.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path]
        sys.exit(stcli.main())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run text generation with a specified model"
    )
    parser.add_argument(
        "model_path", type=str, help="Path or identifier for the model in S3"
    )
    parser.add_argument(
        "-t", "--temperature", type=float, default=0.8, help="Temperature for sampling"
    )
    parser.add_argument(
        "-m",
        "--max_new_tokens",
        type=int,
        default=256,
        help="Maximum number of new tokens to generate",
    )
    parser.add_argument(
        "-n",
        "--min_new_tokens",
        type=int,
        default=8,
        help="Minimum number of new tokens to generate",
    )
    parser.add_argument(
        "-k", "--top_k", type=int, default=50, help="Top-k sampling parameter"
    )
    parser.add_argument(
        "-p", "--top_p", type=float, default=1.0, help="Top-p sampling parameter"
    )
    parser.add_argument(
        "-pf",
        "--profiling",
        action="store_true",
        help="Enable profiling logs for the inference process",
    )
    parser.add_argument(
        "st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
    )
    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    inference = NexaTextInference(model_path, **kwargs)
    inference.run()
