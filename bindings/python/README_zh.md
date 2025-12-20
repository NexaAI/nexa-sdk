# NexaAI Python SDK

本目录包含 NexaAI Python SDK 及多种 AI 推理任务的全面示例。

## 快速开始

最简单的方式是通过我们的交互式 Jupyter notebook。你可以在 [`notebook/`](notebook/) 文件夹中找到示例 notebook。

每个 notebook 包含：
- **LLM（大型语言模型）**：文本生成与对话
- **VLM（视觉语言模型）**：多模态理解与生成
- **Embedder（嵌入模型）**：文本向量化与相似性计算
- **Reranker（重排序）**：文档重排序
- **ASR（自动语音识别）**：语音转文本
- **TTS（文本转语音）**：文本生成语音
- **Diarize（说话人分离）**：说话人分离
- **ImageGen（图像生成）**：文本生成图像或图像到图像转换
- **CV（计算机视觉）**：OCR/文本识别

## 前提条件

- Python 3
- 已安装 Nexa CLI

## 安装

```bash
pip install nexaai -v
```

## 命令行示例

如果你更喜欢命令行用法，这里有基础示例：

### LLM
```bash
python llm.py
```

### 多模态
```bash
python vlm.py
```

### 重排序
```bash
python rerank.py
```

### 嵌入模型
```bash
python embedder.py
```

### 计算机视觉
```bash
python cv_ocr.py
```

### TTS（文本转语音）
```bash
python tts.py --text "Hello, world!"
```

### 说话人分离
```bash
python diarize.py --audio path/to/audio.wav
```

### 图像生成
```bash
# 文本生成图像
python image_gen.py --prompt "A beautiful sunset over the ocean"

# 图像到图像
python image_gen.py --prompt "A beautiful sunset" --init-image path/to/image.png
```

## 通用参数

- `--model`：模型文件路径
- `--device`：推理设备（cpu，gpu等）
- `--max-tokens`：生成最大 tokens（适用 LLM/VLM）
- `--batch-size`：处理批量大小
- `--system`：聊天模型的 system 消息
- `--plugin-id`：指定插件 ID（默认 cpu_gpu）

## 插件 ID 选项

`--plugin-id` 参数支持不同后端：

- `cpu_gpu`：默认，同时支持 CPU 和 GPU
- `metal`：Apple Silicon 优化（支持相关模型）
- `npu`：高通 NPU 优化（支持相关模型）
- `nexaml`：NexaML 优化（支持相关模型）

### 各后端支持的模型

| 后端 | 支持的模型 |
|------|------------|
| `cpu_gpu` | GGUF 格式模型（默认后端） |
| `metal` | MLX 格式模型（如 Qwen3-VL-4B-MLX-4bit, gpt-oss-20b-MLX-4bit） |
| `npu` | **LLM：** Granite-4-Micro-NPU, phi4-mini-npu-turbo, Qwen3-4B-Instruct-2507-npu, Qwen3-4B-Thinking-2507-npu, Llama3.2-3B-NPU-Turbo, jan-v1-4B-npu, qwen3-4B-npu, phi3.5-mini-npu<br>**VLM：** Qwen3-VL-4B-Instruct-NPU, OmniNeural-4B, LFM2-1.2B-npu<br>**Embedder：** embeddinggemma-300m-npu<br>**ASR：** parakeet-tdt-0.6b-v3-npu<br>**CV：** convnext-tiny-npu, paddleocr-npu, yolov12-npu<br>**Reranker：** jina-v2-rerank-npu |
| `nexaml` | **VLM：** Qwen3-VL-4B-Instruct-GGUF:Q4_0, Qwen3-VL-4B-Thinking-GGUF:Q4_0 |

## 使用方法

1. **打开 notebook**（见 [`notebook/`](notebook/) 目录）
2. **按照 notebook 内设置说明进行环境准备**
3. **逐步运行示例**，体验各类 AI 能力
4. **根据自身需求自定义示例代码**

详细环境设置请参见具体 notebook 的说明。

