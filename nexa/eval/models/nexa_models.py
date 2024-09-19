from functools import cached_property
from nexa.eval.api.registry import register_model
from nexa.eval.models.api_models import TemplateAPI
from nexa.eval.models.openai_completions import LocalCompletionsAPI
from typing import Any, Dict, List, Optional, Tuple, Union
from nexa.eval.api.registry import register_model
from nexa.eval.utils import eval_logger

import logging
import time
import requests
from requests.exceptions import RequestException
from tqdm import tqdm
from nexa.eval.api.model import LM

@register_model("my-local-completions")
class nexa_models(TemplateAPI):
    def __init__(
        self,
        base_url=None,
        tokenizer_backend="huggingface",
        **kwargs,
    ):
        super().__init__(
            base_url=base_url, tokenizer_backend=tokenizer_backend, **kwargs
        )
        if self._batch_size > 1:
            eval_logger.warning(
                "Chat completions does not support batching. Defaulting to batch size 1."
            )
            self._batch_size = 1

    def _create_payload(
        self,
        messages: Union[List[List[int]], List[dict], List[str], str],
        generate=False,
        gen_kwargs: Optional[dict] = None,
        seed: int = 1234,
        **kwargs,
    ) -> dict:
        if generate:
            gen_kwargs.pop("do_sample", False)
            max_tokens = gen_kwargs.pop("max_gen_toks", self._max_gen_toks)
            temperature = gen_kwargs.pop("temperature", 0)
            stop = gen_kwargs.pop("until", ["<|endoftext|>"])
            payload = {
                "prompt": messages,
                "model": self.model,
                "max_tokens": max_tokens,
                "temperature": temperature,
                "stop": stop,
                "seed": seed,
                **gen_kwargs,
            }
            # print("Generated Payload:", payload)
            return payload
        else:
            payload = {
                "model": self.model,
                "prompt": messages,
                "temperature": 0,
                "max_tokens": 1,
                "logprobs": 1,
                "seed": seed,
                "echo": True,
            }
            # print("Logprobs Payload:", payload) 
            return payload
      
    @staticmethod
    def parse_logprobs(
        outputs: Union[Dict, List[Dict]],
        tokens: List[List[int]] = None,
        ctxlens: List[int] = None,
        **kwargs,
    ) -> List[Tuple[float, bool]]:
        res = []
        if not isinstance(outputs, list):
            outputs = [outputs]
        for out in outputs:
            for choice, ctxlen in zip(out["choices"], ctxlens):
                assert ctxlen > 0, "Context length must be greater than 0"
                logprobs = sum(choice["logprobs"]["token_logprobs"][ctxlen:-1])
                tokens = choice["logprobs"]["token_logprobs"][ctxlen:-1]
                top_logprobs = choice["logprobs"]["top_logprobs"][ctxlen:-1]
                is_greedy = True
                for tok, top in zip(tokens, top_logprobs):
                    if tok != max(top, key=top.get):
                        is_greedy = False
                        break
                res.append((logprobs, is_greedy))
        return res

    @staticmethod
    def parse_generations(outputs: Union[Dict, List[Dict]], **kwargs) -> List[str]:
        res = []
        if not isinstance(outputs, list):
            outputs = [outputs]
        for out in outputs:
            for choices in out["choices"]:
                res.append(choices["text"])
        return res
    
def get_result(logprobs, context_length):
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

logger = logging.getLogger(__name__)
@register_model("nexa-gguf", "nexa-ggml")
class GGUFLM(LM):
    def __init__(self, base_url=None, max_length=2048, **kwargs):
        super().__init__()
        self.base_url = base_url
        assert self.base_url, "must pass `base_url` to use GGUF LM!"
        self.logprobs = 10
        self.temperature = 0.0
        self.max_length = max_length

    def gguf_completion(
        self, context, continuation=None, stop=None, retries=3, delay=5, **kwargs
    ):
        for _ in range(retries):
            try:
                prompt = context
                request = {
                    "prompt": prompt,
                    "logprobs": True,
                    "top_logprobs": self.logprobs,
                    "temperature": self.temperature,
                }
                # print("request", request)
                if continuation:
                    prompt += continuation
                    request.update({"prompt": prompt, "max_tokens": 1, "echo": True})
                if stop is not None:
                    request["stop"] = stop
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
                    logprob, is_greedy = get_result(logprobs, len(context))
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
            response = self.gguf_completion(context=inp, stop=until)
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

    def loglikelihood_rolling(self, requests, disable_tqdm: bool = False):
        raise NotImplementedError(
            "loglikelihood_rolling not yet supported for GGUF models"
        )