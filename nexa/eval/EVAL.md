# GGUF Model Benchmark
Benchmark GGUF models with a single line of code. The fastest benchmarking tool for quantized GGUF models, featuring multiprocessing support and 8 evaluation tasks.

## 🚀 Quick Start

```bash
# Evaluate a pre-configured model
nexa eval phi3 --tasks ifeval

# Evaluate your custom GGUF model
nexa eval <model_name> --tasks ifeval --num_workers 8
```

## 🔧 Installation

Supports Windows, Linux, and macOS.

1. Install Nexa SDK

2. Install Nexa Eval Package
```bash
pip install 'nexaai[eval]'
```
3. Run your first evaluation

## 📊 Evaluation Tasks

- **General Tasks**
  - `ifeval`: General language understanding
  - `mmlu_pro`: Massive multitask language understanding

- **Math Tasks**
  - `math`: Mathematical reasoning
  - `mgsm_direct`: Grade school math problems

- **Reasoning Tasks**
  - `gpqa`: General purpose question answering

- **Coding Tasks**
  - `openai_humaneval`: Code generation and completion

- **Safety Tasks**
  - `do-not-answer`: Adversarial question handling
  - `truthfulqa`: Model truthfulness evaluation

## 🔄 Advanced Usage

### Multiprocessing
Optimize performance by specifying worker count:
```bash
nexa eval Llama3.2-1B-Instruct:q3_K_M --tasks ifeval --num_workers 8
```

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

## 📚 Resources

- [Nexa SDK Documentation](https://github.com/NexaAI/nexa-sdk)
- [Model Hub](https://example.com/model-hub) - Upload and share your GGUF models

## 📝 License

[License information here]
