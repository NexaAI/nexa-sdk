# Function-Calling Benchmarks

Comprehensive performance benchmarks for NexaAI VLM models on ARM64 Windows platform, measuring function-calling capabilities with varying system prompt sizes.

## 📊 Contents

- **report.md** - Detailed benchmark results and analysis
- **run_bench.ps1** - Automated benchmark execution script
- **count_words.py** - Token counter utility for system prompts
- **prompt*.txt** - System prompt files at different token sizes (500, 1000, 2000, 4000)

## 🚀 Quick Start

### Prerequisites

1. Build the nexa CLI tool:
   ```powershell
   cd runner
   make build
   ```

2. Ensure prompt files exist:
   ```
   prompt500.txt, prompt1000.txt, prompt2000.txt, prompt4000.txt
   ```

### Run Benchmarks

```powershell
cd runner
.\run_bench.ps1
```

This will:
1. Test NPU model (Qwen3-VL-4B-Instruct-NPU) 
2. Test GGUF model with GPU acceleration
3. Test GGUF model with CPU-only mode (ngl=0)

Results are saved to `nexa_infer_*.log` files.

## 📈 Key Metrics

- **Throughput** - Tokens generated per second
- **Output Tokens** - Total tokens in model response
- **TTFT** - Time To First Token (model latency)

## 🔍 Analyzing Results

Check `report.md` for:
- Performance comparison table
- Analysis of each model variant
- Recommendations for different use cases

## 🛠️ Token Counting

Verify prompt token counts:

```bash
# Using OpenAI tokenizer (default)
python count_words.py prompt500.txt

# Using Hugging Face tokenizer
python count_words.py prompt500.txt --backend transformers --model meta-llama/Llama-2-7b-hf
```

## 📝 System Prompt Sizes

| File | Size | Use Case |
|------|------|----------|
| prompt500.txt | ~500 tokens | Light system prompts |
| prompt1000.txt | ~1000 tokens | Standard function definitions |
| prompt2000.txt | ~2000 tokens | Complex prompts with context |
| prompt4000.txt | ~4000 tokens | Heavy system prompts |

## ⚡ Performance Summary

| Model | Best For | Throughput | Latency |
|-------|----------|-----------|---------|
| **NPU** | Production, low latency | 13.7-13.8 tok/s | 0.8-5.2s TTFT |
| **GGUF (GPU)** | Batch processing | 7.3-3.1 tok/s | 17-227s TTFT |
| **GGUF (CPU)** | Balanced, CPU-constrained | 21.4-5.3 tok/s | 7.1-195.6s TTFT |

## 📋 Log Files

Each benchmark run generates detailed logs with:
- Model loading information
- Tokenization details
- Generated function call results
- Performance metrics (tokens/sec, TTFT)

Example:
```
nexa_infer_NPU_500.log     # 500-token prompt, NPU model
nexa_infer_GGUF_1000.log   # 1000-token prompt, GGUF GPU
nexa_infer_GGUF_ngl0_4000.log  # 4000-token prompt, GGUF CPU
```

## 🔧 Customization

To modify the test prompt, edit `run_bench.ps1`:

```powershell
$TestPrompt = 'your custom test prompt here'
```

To test different models, update the model names in the script:

```powershell
& ".\build\nexa" infer "YOUR-MODEL-NAME" -s $promptFile -p $TestPrompt
```

## 📖 References

- [report.md](report.md) - Full benchmark report with analysis
- [count_words.py](count_words.py) - Token counting utility documentation

---

**Last Updated:** January 2026  
**Platform:** Windows ARM64 (Snapdragon X Elite)  
**Models Tested:** Qwen3-VL-4B-Instruct (NPU and GGUF variants)
