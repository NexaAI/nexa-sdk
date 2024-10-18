import argparse
import logging
import os
import time
from pathlib import Path
from typing import Iterator, List, Union

from nexa.constants import (
    DEFAULT_TEXT_GEN_PARAMS,
    NEXA_RUN_CHAT_TEMPLATE_MAP,
    NEXA_RUN_COMPLETION_TEMPLATE_MAP,
    NEXA_STOP_WORDS_MAP,
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.general import pull_model
from nexa.utils import SpinningCursorAnimation, nexa_prompt
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr


logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s"
)


class NexaTextInference:
    """
    A class used for loading text models and running text generation.

    Methods:
        run: Run the text generation loop.
        run_streamlit: Run the Streamlit UI.
        create_embedding: Embed a string.
        create_chat_completion: Generate completion for a chat conversation.
        create_completion: Generate completion for a given prompt.
    Args:
    model_path (str): Path or identifier for the model in Nexa Model Hub.
    local_path (str, optional): Local path of the model.
    embedding (bool): Enable embedding generation.
    stop_words (list): List of stop words for early stopping.
    profiling (bool): Enable timing measurements for the generation process.
    streamlit (bool): Run the inference in Streamlit UI.
    temperature (float): Temperature for sampling.
    max_new_tokens (int): Maximum number of new tokens to generate.
    top_k (int): Top-k sampling parameter.
    top_p (float): Top-p sampling parameter
    """
    def __init__(self, model_path=None, local_path=None, stop_words=None, device="auto", **kwargs):
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")
        
        self.params = DEFAULT_TEXT_GEN_PARAMS
        self.params.update(kwargs)
        self.model = None
        self.device = device

        self.model_path = model_path
        self.downloaded_path = local_path

        self.logprobs = kwargs.get('logprobs', None)
        self.top_logprobs = kwargs.get('top_logprobs', None)

        if self.downloaded_path is None:
            self.downloaded_path, _ = pull_model(self.model_path, **kwargs)

        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not appicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)
        self.profiling = kwargs.get("profiling", False)

        model_name = model_path.split(":")[0].lower() if model_path else None
        self.stop_words = (stop_words if stop_words else NEXA_STOP_WORDS_MAP.get(model_name, []))
        self.chat_format = NEXA_RUN_CHAT_TEMPLATE_MAP.get(model_name, None)
        self.completion_template = NEXA_RUN_COMPLETION_TEMPLATE_MAP.get(model_name, None)

        if not kwargs.get("streamlit", False):
            self._load_model()
            if self.model is None:
                logging.error(
                    "Failed to load model or tokenizer. Exiting.", exc_info=True
                )
                exit(1)
    
    def create_embedding(
        self,
        input: Union[str, List[str]],
        normalize: bool = False,
        truncate: bool = True,
    ):
        """Embed a string.

        Args:
            input: The utf-8 encoded string or a list of string to embed.
            normalize: Normalize the embeddings.
            truncate: Truncate the embeddings.

        Returns:
            Embeddings or list of embeddings
        """
        return self.model.embed(input, normalize, truncate)

    @SpinningCursorAnimation()
    def _load_model(self):
        logging.debug(f"Loading model from {self.downloaded_path}, use_cuda_or_metal : {is_gpu_available()}")
        start_time = time.time()
        with suppress_stdout_stderr():
            from nexa.gguf.llama.llama import Llama
            try:
                if self.device == "auto" or self.device == "gpu":
                    n_gpu_layers = -1 if is_gpu_available() else 0
                elif self.device == "cpu":
                    n_gpu_layers = 0

                self.model = Llama(
                    embedding=self.params.get("embedding", False),
                    model_path=self.downloaded_path,
                    verbose=self.profiling,
                    chat_format=self.chat_format,
                    n_ctx=self.params.get("nctx", 2048),
                    n_gpu_layers=n_gpu_layers,
                    lora_path=self.params.get("lora_path", ""),
                )
            except Exception as e:
                logging.error(f"Failed to load model: {e}. Falling back to CPU.", exc_info=True)
                self.model = Llama(
                    model_path=self.downloaded_path,
                    verbose=self.profiling,
                    chat_format=self.chat_format,
                    n_ctx=self.params.get("nctx", 2048),
                    n_gpu_layers=0,  # hardcode to use CPU
                    lora_path=self.params.get("lora_path", ""),
                )

        load_time = time.time() - start_time
        if self.profiling:
            logging.debug(f"Model loaded in {load_time:.2f} seconds")
        if (
            self.completion_template is None
            and (
                chat_format := self.model.metadata.get("tokenizer.chat_template", None)
            )
            is not None
        ):
            self.chat_format = chat_format
            logging.debug("Chat format detected")

        self.conversation_history = [] if self.chat_format else None

    def run(self):
        """
        CLI interactive session. Not for SDK.
        """
        while True:
            generated_text = ""
            try:
                if not (user_input := nexa_prompt()):
                    continue

                generation_start_time = time.time()

                if self.chat_format:
                    output = self._chat(user_input)
                    first_token = True
                    for chunk in output:
                        if first_token:
                            decoding_start_time = time.time()
                            prefill_time = decoding_start_time - generation_start_time
                            first_token = False
                        delta = chunk["choices"][0]["delta"]
                        if "role" in delta:
                            print(delta["role"], end=": ", flush=True)
                            generated_text += delta["role"]
                        elif "content" in delta:
                            print(delta["content"], end="", flush=True)
                            generated_text += delta["content"]

                else:
                    output = self._complete(user_input)
                    first_token = True
                    for chunk in output:
                        if first_token:
                            decoding_start_time = time.time()
                            prefill_time = decoding_start_time - generation_start_time
                            first_token = False
                        choice = chunk["choices"][0]
                        if "text" in choice:
                            delta = choice["text"]
                        elif "delta" in choice:
                            delta = choice["delta"]["content"]
                        
                        print(delta, end="", flush=True)
                        generated_text += delta

                if self.chat_format:
                    if len(self.conversation_history) >= 2:
                        self.conversation_history = self.conversation_history[2:]

                    self.conversation_history.append({"role": "user", "content": user_input})
                    self.conversation_history.append({"role": "assistant", "content": generated_text})
            except KeyboardInterrupt:
                pass
            except Exception as e:
                logging.error(f"Error during generation: {e}", exc_info=True)
            print("\n")

    def create_chat_completion(self, messages, **kwargs):
        """
        Used for SDK. Generate completion for a chat conversation.

        Args:
            messages (list): List of messages in the conversation.
            temperature (float): Temperature for sampling.
            max_tokens (int): Maximum number of new tokens to generate.
            top_k (int): Top-k sampling parameter.
            top_p (float): Top-p sampling parameter.
            stream (bool): Stream the output.
            stop (list): List of stop words for early stopping.

        Returns:
            Iterator: Iterator for the completion.
        """
        params = {
            "temperature": self.params.get("temperature", 0.7),
            "max_tokens": self.params.get("max_new_tokens", 2048),
            "top_k": self.params.get("top_k", 50),
            "top_p": self.params.get("top_p", 1.0),
            "stop": self.stop_words,
            "logprobs": self.logprobs,
            "top_logprobs": self.top_logprobs
        }
        params.update(kwargs)
        if params['logprobs'] and params['top_logprobs'] is None:
            params['top_logprobs'] = 4

        return self.model.create_chat_completion(messages=messages, **params)

    def create_completion(self, prompt, **kwargs):
        """
        Used for SDK. Generate completion for a given prompt.

        Args:
            prompt (str): Prompt for the completion.
            temperature (float): Temperature for sampling.
            max_tokens (int): Maximum number of new tokens to generate.
            top_k (int): Top-k sampling parameter.
            top_p (float): Top-p sampling parameter.
            echo (bool): Echo the prompt back in the output.
            stream (bool): Stream the output.
            stop (list): List of stop words for early stopping.

        Returns:
            Iterator: Iterator for the completion.
        """
        params = {
            "temperature": self.params.get("temperature", 0.7),
            "max_tokens": self.params.get("max_new_tokens", 2048),
            "top_k": self.params.get("top_k", 50),
            "top_p": self.params.get("top_p", 1.0),
            "stop": self.stop_words,
            "logprobs": self.logprobs
        }
        params.update(kwargs)

        return self.model.create_completion(prompt=prompt, **params)


    def _chat(self, user_input: str) -> Iterator:
        current_messages = self.conversation_history + [{"role": "user", "content": user_input}]
        return self.model.create_chat_completion(
            messages=current_messages,
            temperature=self.params["temperature"],
            max_tokens=self.params["max_new_tokens"],
            top_k=self.params["top_k"],
            top_p=self.params["top_p"],
            stream=True,
            stop=self.stop_words,
            logprobs=self.logprobs,
            top_logprobs=self.top_logprobs,
        )

    def _complete(self, user_input: str) -> Iterator:
        prompt = (
            self.completion_template.format(input=user_input)
            if self.completion_template
            else user_input
        )
        return self.model.create_completion(
            prompt=prompt,
            temperature=self.params["temperature"],
            max_tokens=self.params["max_new_tokens"],
            top_k=self.params["top_k"],
            top_p=self.params["top_p"],
            echo=False,  # Echo the prompt back in the output
            stream=True,
            stop=self.stop_words,
            logprobs=self.logprobs,
        )

    def run_streamlit(self, model_path: str, is_local_path = False, hf = False):
        """
        Used for CLI. Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")

        script_path = (
            Path(os.path.abspath(__file__)).parent
            / "streamlit"
            / "streamlit_text_chat.py"
        )

        import sys
        from streamlit.web import cli as stcli

        # Convert all arguments to strings
        args = [
            "streamlit", "run", str(script_path),
            str(model_path),
            str(is_local_path),
            str(hf),
        ]

        sys.argv = args
        sys.exit(stcli.main())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run text generation with a specified model"
    )
    parser.add_argument(
        "model_path",
        type=str,
        help="Path or identifier for the model in Nexa Model Hub",
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
        "-k", "--top_k", type=int, default=50, help="Top-k sampling parameter"
    )
    parser.add_argument(
        "-p", "--top_p", type=float, default=1.0, help="Top-p sampling parameter"
    )
    parser.add_argument(
        "--nctx",
        type=int,
        default=2048,
        help="Maximum context length of the model you're using"
    )
    parser.add_argument(
        "-sw",
        "--stop_words",
        nargs="*",
        default=[],
        help="List of stop words for early stopping",
    )
    parser.add_argument(
        "-pf",
        "--profiling",
        action="store_true",
        help="Enable timing measurements for the generation process",
    )
    parser.add_argument(
        "-st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
    )
    parser.add_argument(
        "--lora_path",
        type=str,
        help="Path to a LoRA file to apply to the model.",
    )
    parser.add_argument(
        "-d",
        "--device",
        type=str,
        choices=["auto", "cpu", "gpu"],
        default="auto",
        help="Device to use for inference (auto, cpu, or gpu)",
    )
    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    stop_words = kwargs.pop("stop_words", [])
    device = kwargs.pop("device", "auto")

    inference = NexaTextInference(model_path, stop_words=stop_words, device=device, **kwargs)
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()
