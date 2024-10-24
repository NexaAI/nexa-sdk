# GGUF Benchmark
Benchmark GGUF models with a ONE line of code. The fastest benchmarking tool for quantized GGUF models, featuring multiprocessing support and 8 evaluation tasks.

> Currently supports text GGUF models.

## 🔧 Installation

Supports Windows, Linux, and macOS.

1. Install [Nexa SDK Python Pacakage](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#installation---python-package)

2. Install Nexa Eval Package
    ```bash
    pip install 'nexaai[eval]'
    ```


## 🚀 Quick Start

Choose a GGUF model from [Nexa Model Hub](https://nexa.ai/models) to benchmark. You can also upload your own GGUF models.

```bash
# Evaluate Llama3.2-1B Q4_K_M quantization with "ifeval" task
nexa eval Llama3.2-1B-Instruct:q4_K_M --tasks ifeval


# Use Multiprocessing. You can specify number of workerse to optimize performance.
nexa eval Llama3.2-1B-Instruct:q4_K_M --tasks ifeval --num_workers 4
```

## CLI Reference for EVAL

```bash
usage: nexa eval model_path [-h] [--tasks TASKS] [--limit LIMIT]

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub. Text after 'nexa run'.

options:
  -h, --help            show this help message and exit
  --tasks TASKS         Tasks to evaluate, comma-separated
  --limit LIMIT         Limit the number of examples per task. If <1, limit is a percentage of the total number of examples.
```

## 📊 Evaluation Tasks

- **General Tasks**
  - [`ifeval`](https://arxiv.org/abs/2311.07911): General language understanding
  - [`mmlu_pro`](https://arxiv.org/abs/2406.01574): Massive multitask language understanding

- **Math Tasks**
  - [`math`](https://arxiv.org/pdf/2103.03874): Mathematical reasoning
  - [`mgsm_direct`](https://arxiv.org/abs/2210.03057): Grade school math problems

- **Reasoning Tasks**
  - [`gpqa`](https://arxiv.org/abs/2311.12022): General purpose question answering

- **Coding Tasks**
  - [`openai_humaneval`](https://arxiv.org/abs/2107.03374): Code generation and completion

- **Safety Tasks**
  - [`do-not-answer`](https://aclanthology.org/2024.findings-eacl.61): Adversarial question handling
  - [`truthfulqa`](https://arxiv.org/abs/2109.07958): Model truthfulness evaluation


## 💡 Why GGUF Models?

GGUF (GGML Universal Format) models are optimized for on-device AI deployment:
- Reduced memory footprint through quantization
- Cross-platform compatibility via llama.cpp
- No external dependencies
- Supported by popular projects: llama.cpp, whisper.cpp, stable-diffusion.cpp, and more

## 📈 Why Benchmark?

Quantization affects three key factors:
- File size
- Model quality
- Performance

Benchmarking helps you:
1. Verify accuracy retention after quantization
2. Select the optimal model for your specific use case
3. Make informed decisions about quantization levels

## Acknowledgements
Adapted From [Language Model Evaluation Harness](https://github.com/EleutherAI/lm-evaluation-harness).
