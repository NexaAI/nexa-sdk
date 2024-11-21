from nexa.gguf.llama.llama_cache import LlamaDiskCache

def run_inference_with_disk_cache(
    model, cache_prompt, total_prompt, use_cache=True, cache_dir="llama.cache", **kwargs
):
    temperature = kwargs.get('temperature', 0.7)
    max_tokens = kwargs.get('max_tokens', 2048)
    top_p = kwargs.get('top_p', 1.0)
    top_k = kwargs.get('top_k', 50)
    repeat_penalty = kwargs.get('repeat_penalty', 1.0)

    if use_cache:
        # Initialize disk cache with specified directory
        cache_context = LlamaDiskCache(cache_dir=cache_dir)
        model.set_cache(cache_context)
        # Convert prompt to tokens for cache key
        prompt_tokens = model.tokenize(cache_prompt.encode("utf-8"))

        try:
            # Try to load existing cache
            cached_state = cache_context[prompt_tokens]
            model.load_state(cached_state)

            output = model(
                total_prompt,
                max_tokens=max_tokens,
                temperature=temperature,
                stream=True,
            )
        except KeyError:
            # If cache doesn't exist, create it
            model.reset()
            # Run initial inference to populate cache
            _ = model(
                cache_prompt,
                max_tokens=1,  # Minimal tokens for cache creation
                temperature=temperature,
                echo=False,
            )
            # Save the state to cache
            cache_context[prompt_tokens] = model.save_state()

            # Generate output after creating cache
            output = model(
                total_prompt,
                max_tokens=max_tokens,
                temperature=temperature,
                top_p=top_p,
                top_k=top_k,
                repeat_penalty=repeat_penalty,
                stream=True,
            )
    else:
        model.reset()
        model.set_cache(None)

        output = model(
            total_prompt,
            max_tokens=max_tokens,
            temperature=temperature,
            top_p=top_p,
            top_k=top_k,
            repeat_penalty=repeat_penalty,
            stream=True,
        )
    return output