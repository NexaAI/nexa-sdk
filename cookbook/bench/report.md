# Benchmark Report

Test prompt:

```
send_email to me@123.com say i have arrived
```

## Function call performance

| Model | Tokens | Throughput | Output Tokens | TTFT |
|-------|--------|------------|---------------|-------------|
| NexaAI/Qwen3-VL-4B-Instruct-NPU | 500 tokens | 13.7 tok/s | 47 | 0.8 s |
| NexaAI/Qwen3-VL-4B-Instruct-NPU | 1000 tokens | 13.8 tok/s | 47 | 1.6 s |
| NexaAI/Qwen3-VL-4B-Instruct-NPU | 2000 tokens | 13.7 tok/s | 47 | 2.4 s |
| NexaAI/Qwen3-VL-4B-Instruct-NPU | 4000 tokens | 13.8 tok/s | 152 | 5.2 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF | 500 tokens | 7.3 tok/s | 47 | 17.0 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF | 1000 tokens | 6.1 tok/s | 47 | 39.4 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF | 2000 tokens | 5.1 tok/s | 47 | 68.5 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF | 4000 tokens | 3.1 tok/s | 39 | 227.0 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF (ngl=0) | 500 tokens | 21.4 tok/s | 47 | 7.1 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF (ngl=0) | 1000 tokens | 16.0 tok/s | 47 | 17.5 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF (ngl=0) | 2000 tokens | 11.0 tok/s | 47 | 35.1 s |
| NexaAI/Qwen3-VL-4B-Instruct-GGUF (ngl=0) | 4000 tokens | 5.3 tok/s | 39 | 195.6 s |

