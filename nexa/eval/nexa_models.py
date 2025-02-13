from typing import Optional
from nexa.eval import utils
import logging
import time
from nexa.gguf.nexa_inference_text import NexaTextInference
from tqdm import tqdm
    

logger = logging.getLogger(__name__)

class GGUFLM:
    def __init__(self, model_path=None, local_path=None, **kwargs):
        if model_path is None and local_path is None:
            raise ValueError("model_path or local_path must be provided.")
        self.model = NexaTextInference(model_path, local_path, logits_all=True)
        self.logprobs = 10
        self.temperature = 0

    def gguf_completion(
        self, context, max_tokens = None, continuation = None, stop=None
    ):
        try:
            prompt = context
            params = {
                "prompt": prompt,
                "logprobs": self.logprobs,
                "temperature": self.temperature,
                "max_tokens": max_tokens,
                "eval_context_length": len(context)
            }
            if continuation:
                prompt += continuation
                params.update({"prompt": prompt, "max_tokens": 1, "echo": True})
            if stop is not None:
                params["stop"] = stop
            result = self.model.create_completion(**params)
            return result
        except Exception as e:
            logger.error(f"Unexpected error occured: {e}")

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
                    # FIXME: Using average logprob as fallback when logprobs data is invalid
                    # This is a temporary solution and might need better error handling
                    avg_logprob = sum(r[0] for r in res) / len(res) if res else -0.693
                    res.append((avg_logprob, True))
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
            max_tokens = request_args.get("max_gen_toks", None)
            response = self.gguf_completion(context=inp, stop=until, max_tokens=max_tokens)
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


    def get_result(self, logprobs, context_length=None):
        is_greedy = True
        tokens = logprobs["tokens"]
        tokens_logprobs = logprobs["token_logprobs"]

        # Now we can directly sum all logprobs except the last one
        continuation_logprobs = sum(tokens_logprobs[:-1])

        # Check if each token was the highest probability choice
        for i in range(len(tokens)):
            token = tokens[i]
            top_tokens = logprobs["top_logprobs"][i]
            top_token = max(top_tokens.keys(), key=lambda x: top_tokens[x])
            if top_token != token:
                is_greedy = False
                break

        return continuation_logprobs, is_greedy
    