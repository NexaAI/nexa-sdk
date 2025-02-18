import argparse
import json
from jsonschema import validate, ValidationError
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
from nexa.gguf.llama.llama_grammar import LlamaGrammar
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

    def __init__(self, model_path=None, local_path=None, stop_words=None, device="auto", function_calling: bool = False, **kwargs):
        if model_path is None and local_path is None:
            raise ValueError(
                "Either model_path or local_path must be provided.")

        self.params = DEFAULT_TEXT_GEN_PARAMS.copy()
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
        self.stop_words = (
            stop_words if stop_words else NEXA_STOP_WORDS_MAP.get(model_name, []))
        if function_calling:
            self.chat_format = 'functionary'
        else:
            self.chat_format = NEXA_RUN_CHAT_TEMPLATE_MAP.get(model_name, None)
        self.completion_template = NEXA_RUN_COMPLETION_TEMPLATE_MAP.get(
            model_name, None)

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
        logging.debug(
            f"Loading model from {self.downloaded_path}, use gpu : {is_gpu_available()}")
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
                    logits_all=self.params.get("logits_all", False),
                )
            except Exception as e:
                logging.error(
                    f"Failed to load model: {e}. Falling back to CPU.", exc_info=True)
                self.model = Llama(
                    model_path=self.downloaded_path,
                    verbose=self.profiling,
                    chat_format=self.chat_format,
                    n_ctx=self.params.get("nctx", 2048),
                    n_gpu_layers=0,  # hardcode to use CPU
                    lora_path=self.params.get("lora_path", ""),
                    logits_all=self.params.get("logits_all", False),
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
        from nexa.gguf.llama._utils_spinner import start_spinner, stop_spinner

        while True:
            generated_text = ""
            try:
                if not (user_input := nexa_prompt()):
                    continue

                generation_start_time = time.time()

                stop_event, spinner_thread = start_spinner(
                    style="default",
                    message=""
                )

                if self.chat_format:
                    output = self._chat(user_input)
                    first_token = True
                    stop_spinner(stop_event, spinner_thread)

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
                    stop_spinner(stop_event, spinner_thread)

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

                    self.conversation_history.append(
                        {"role": "user", "content": user_input})
                    self.conversation_history.append(
                        {"role": "assistant", "content": generated_text})
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
        current_messages = self.conversation_history + \
            [{"role": "user", "content": user_input}]
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

    def unload_lora(self):
        self.model.unload_lora()

    def reload_lora(self, lora_path: str, lora_scale: float = 1.0):
        self.model.reload_lora(lora_path, lora_scale)

    def structure_output(self, json_schema: str = None, json_schema_path: str = None, prompt: str = "", **kwargs):
        """
        Generate structured output from the model based on a given JSON schema.
        Args:
            json_schema (str): JSON schema as a string.
            json_schema_path (str): Path to a JSON schema file.
            prompt (str): The initial prompt or instructions for the model.
            **kwargs: Additional generation parameters.
        Returns:
            dict: The structured output conforming to the provided schema.
        """

        if not json_schema and not json_schema_path:
            raise ValueError(
                "Either json_schema or json_schema_path must be provided.")

        # Load schema from file if json_schema is not provided
        if json_schema_path and not json_schema:
            with open(json_schema_path, 'r') as f:
                json_schema = f.read()

        # Parse the schema
        try:
            schema_data = json.loads(json_schema)
        except json.JSONDecodeError as e:
            raise ValueError(f"Provided schema is not valid JSON: {e}")

        # print(f"schema_data: {schema_data}")
        grammar = LlamaGrammar.from_json_schema(
            json.dumps(schema_data), verbose=self.model.verbose
        )
        print(f"grammar: {grammar}")

        structured_prompt = f"Extract the following JSON from the text: {prompt}"

        params = {
            "temperature": self.params.get("temperature", 0.7),
            "max_tokens": self.params.get("max_new_tokens", 2048),
            "top_k": self.params.get("top_k", 50),
            "top_p": self.params.get("top_p", 1.0),
            "stop": self.stop_words,
            "logprobs": self.logprobs
        }
        params.update(kwargs)
        # We'll try to generate a completion that looks like JSON
        completion = self.model.create_completion(
            prompt=structured_prompt,
            grammar=grammar,
            **params
        )

        generated_text = completion["choices"][0]["text"]
        try:
            structured_data = json.loads(generated_text)
        except json.JSONDecodeError:
            logging.error(
                "Model output is not valid JSON. Consider retrying or adjusting your prompt.")
            raise

        # Validate against the schema
        try:
            validate(instance=structured_data, schema=schema_data)
        except ValidationError as e:
            logging.error("Generated JSON does not conform to the schema.")
            logging.debug(f"Generated JSON: {generated_text}")
            logging.debug(f"Validation error: {e}")
            raise

        return structured_data

    def function_calling(self, messages, tools) -> list[dict]:
        """
        Generate function calls based on input messages and available functions.

        This method generates a list of function calls in JSON format. 
        To use this method, the `function_calling` argument must be set to `True` 
        when initializing the `NexaTextInference` instance.

        Args:
            messages (list): A list of dictionaries representing the chat. Each dictionary should contain:
                - 'role' (str): The role of the message sender (e.g., 'user',
                'assistant').
                - 'content' (str): The textual content of the message.
            tools (list): A list of dictionaries, each defining an available
                function that can be called by the LLM. Each dictionary should include:
                - 'name' (str): The function's name.
                - 'description' (str): A brief description of the function's
                purpose.
                - 'parameters' (dict): A schema defining the function's parameters,
                including their types and descriptions.

        Returns:
            A list of dictionaries representing the assistant's responses.
            Each dictionary contains the function name and its corresponding arguments.
        """
        def process_output(output):
            processed_output = []
            for item in output:
                if "function" in item and isinstance(item["function"], dict):
                    try:
                        # llama-cpp-python's `create_chat_completion` produces incorrectly parsed output when only
                        # `messages` and `tools` are provided. Specifically, the function name is mistakenly treated
                        # as an argument, while the `function.name` field is an empty string.
                        # The following code corrects this issue.
                        function_data = json.loads(
                            item["function"]["arguments"])
                        function_name = function_data.get("function", "")
                        if function_name == "":
                            function_name = function_data.get("name", "")
                        function_args = {k: v for k, v in function_data.items() if k not in [
                            'type', 'function', 'name']}
                        if 'input' in function_args:
                            function_args = function_args['input']
                        else:
                            function_args = function_args['parameters']

                        processed_output.append({
                            "type": "function",
                            "function": {
                                "name": function_name,
                                "arguments": json.dumps(function_args)
                            }
                        })
                    except json.JSONDecodeError:
                        print("Error: Unable to parse JSON from function arguments")

            return processed_output

        response = self.model.create_chat_completion(
            messages=messages, tools=tools, function_call='none')
        response = response['choices'][0]['message']['tool_calls']
        try:
            # print(response)
            return process_output(response)
        except Exception as e:
            print(
                "Error: The model output does not match the expected function calling format. "
                "Consider trying a more capable model or adjusting your prompt."
            )
            return []

    def run_streamlit(self, model_path: str, is_local_path=False, hf=False):
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

    inference = NexaTextInference(
        model_path, stop_words=stop_words, device=device, **kwargs)
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()
