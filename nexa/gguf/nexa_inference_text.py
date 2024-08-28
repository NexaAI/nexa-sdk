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
    NEXA_RUN_MODEL_MAP,
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
    def __init__(self, model_path, local_path=None, stop_words=None, **kwargs):
        self.params = DEFAULT_TEXT_GEN_PARAMS
        self.params.update(kwargs)
        self.model = None
        
        self.model_path = model_path
        self.downloaded_path = local_path
        
        if self.downloaded_path is None:
            self.downloaded_path, run_type = pull_model(self.model_path)

        if self.downloaded_path is None:
            logging.error(
                f"Model ({model_path}) is not appicable. Please refer to our docs for proper usage.",
                exc_info=True,
            )
            exit(1)

        self.stop_words = (
            stop_words if stop_words else NEXA_STOP_WORDS_MAP.get(model_path, [])
        )
        self.profiling = kwargs.get("profiling", False)

        self.chat_format = NEXA_RUN_CHAT_TEMPLATE_MAP.get(model_path, None)
        self.completion_template = NEXA_RUN_COMPLETION_TEMPLATE_MAP.get(
            model_path, None
        )

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
    ):
        """Embed a string.

        Args:
            input: The utf-8 encoded string or a list of string to embed.

        Returns:
            A list of embeddings
        """
        return self.model.create_embedding(input)

    @SpinningCursorAnimation()
    def _load_model(self):
        logging.debug(f"Loading model from {self.downloaded_path}, use_cuda_or_metal : {is_gpu_available()}")
        start_time = time.time()
        with suppress_stdout_stderr():
            try:
                from nexa.gguf.llama.llama import Llama
                self.model = Llama(
                    embedding=self.params.get("embedding", False),
                    model_path=self.downloaded_path,
                    verbose=self.profiling,
                    chat_format=self.chat_format,
                    n_ctx=2048,
                    n_gpu_layers=-1 if is_gpu_available() else 0,
                )
            except Exception as e:
                logging.error(f"Failed to load model: {e}. Falling back to CPU.", exc_info=True)
                self.model = Llama(
                    model_path=self.downloaded_path,
                    verbose=self.profiling,
                    chat_format=self.chat_format,
                    n_ctx=2048,
                    n_gpu_layers=0,  # hardcode to use CPU
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
                        delta = chunk["choices"][0]["text"]
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

    def create_chat_completion(self, messages, temperature=0.7, max_tokens=2048, top_k=50, top_p=1.0, stream=False, stop=None):
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
        return self.model.create_chat_completion(messages=messages, temperature=temperature, max_tokens=max_tokens, top_k=top_k, top_p=top_p, stream=stream, stop=stop)

    def create_completion(self, prompt, temperature=0.7, max_tokens=2048, top_k=50, top_p=1.0, echo=False, stream=False, stop=None):
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
        return self.model.create_completion(prompt=prompt, temperature=temperature, max_tokens=max_tokens, top_k=top_k, top_p=top_p, echo=echo, stream=stream, stop=stop)


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
        )

    def run_streamlit(self, model_path: str):
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

        sys.argv = ["streamlit", "run", str(script_path), model_path]
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
    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    stop_words = kwargs.pop("stop_words", [])
    inference = NexaTextInference(model_path, stop_words=stop_words, **kwargs)
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()
