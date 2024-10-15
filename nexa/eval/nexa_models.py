from typing import Optional
from nexa.eval import utils
import logging
import time
import requests
from requests.exceptions import RequestException
from tqdm import tqdm
    

logger = logging.getLogger(__name__)

class GGUFLM:
    def __init__(self, base_url=None, max_length=2048, **kwargs):
        super().__init__()
        self.base_url = base_url
        assert self.base_url, "must pass `base_url` to use GGUF LM!"
        self.logprobs = 10
        self.temperature = 0.0
        self.max_length = max_length
        self._rank = 0
        self._world_size = 1

    def gguf_completion(
        self, context, max_new_tokens = None, continuation=None, stop=None, retries=3, delay=5, **kwargs
    ):
        for _ in range(retries):
            try:
                prompt = context
                request = {
                    "prompt": prompt,
                    "logprobs": self.logprobs,
                    "temperature": self.temperature,
                }
                if continuation:
                    prompt += continuation
                    request.update({"prompt": prompt, "max_tokens": 1, "echo": True})
                if stop is not None:
                    request["stop_words"] = stop
                if max_new_tokens is not None:
                    request["max_new_tokens"] = max_new_tokens
                # print("request", request)
                response = requests.post(
                    f"{self.base_url}", json=request
                )
                # print("response", response.json())
                response.raise_for_status()
                return response.json()
            except RequestException as e:
                logger.error(f"RequestException: {e}")
                time.sleep(delay)  # wait before retrying
        else:
            raise Exception(f"Failed to get a valid response after {retries} retries.")

    def loglikelihood(self, requests, disable_tqdm: bool = False):
        if not requests:
            return []
        res = []
        for context, continuation in tqdm(
            [req.args for req in requests], disable=disable_tqdm
        ):
            response = self.gguf_completion(context=context, continuation=continuation)
            if response and "choices" in response and response["choices"]:
                choice = response["choices"][0]
                logprobs = choice.get("logprobs")
                if (
                    logprobs
                    and "token_logprobs" in logprobs
                    and logprobs["token_logprobs"]
                ):
                    logprob, is_greedy = self.get_result(logprobs, len(context))
                    res.append((logprob, is_greedy))
                else:
                    logger.warning(
                        "Invalid logprobs data. Expected 'logprobs' to contain 'token_logprobs' list."
                    )
            else:
                logger.error(
                    f"Invalid response for loglikelihood. Response: {response}"
                )
                assert False
        return res

    def generate_until(self, requests, disable_tqdm: bool = False):
        if not requests:
            return []

        res = []
        for request in tqdm([req.args for req in requests], disable=disable_tqdm):
            inp = request[0]
            request_args = request[1]
            until = request_args.get("until", ["</s>"])
            max_new_tokens = request_args.get("max_gen_toks", None)
            response = self.gguf_completion(context=inp, stop=until, max_new_tokens=max_new_tokens)
            if response and "choices" in response and response["choices"]:
                choice = response["choices"][0]
                if "text" in choice:
                    generated_text = choice["text"].strip()
                    res.append(generated_text)
                else:
                    logger.error(
                        f"Invalid response for greedy_until. Response: {response}"
                    )
                    res.append(None)  # Add default value in case of error
            else:
                logger.error(f"Invalid response for greedy_until. Response: {response}")
                res.append(None)  # Add default value in case of error
        return res


    def get_result(self, logprobs, context_length):
        is_greedy = True
        offsets = logprobs["text_offset"]
        tokens = logprobs["tokens"]
        tokens_logprobs = logprobs["token_logprobs"]

        idx = 0
        while idx < len(offsets) and offsets[idx] < context_length:
            idx += 1
        continuation_logprobs = sum(tokens_logprobs[idx:-1])
        for i in range(idx, len(tokens)):
            token = tokens[i]
            top_tokens = logprobs["top_logprobs"][i]
            top_token = max(top_tokens.keys(), key=lambda x: top_tokens[x])
            if top_token != token:
                is_greedy = False
                break

        return continuation_logprobs, is_greedy

    @property
    def rank(self):
        return self._rank

    @property
    def world_size(self):
        return self._world_size

    @classmethod
    def create_from_arg_string(
        cls, arg_string: str, additional_config: Optional[dict] = None
    ):
        additional_config = {} if additional_config is None else additional_config
        args = utils.simple_parse_args_string(arg_string)
        args2 = {k: v for k, v in additional_config.items() if v is not None}
        return cls(**args, **args2)
    
    