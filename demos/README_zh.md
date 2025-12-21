# Nexa SDK 演示

> **Android SDK 演示请参考** [`bindings/android`](../bindings/android)

## PC 演示（Python 绑定）

### 🤖 Agent-Granite

轻量级本地 AI 助手，支持函数调用（网页搜索），基于 Granite-4-Micro 模型。包含 Gradio 界面，可实现本地 Agent 流程。

- [Python绑定示例](./Agent-Granite/Python-Binding-Example)
- [服务示例](./Agent-Granite/Serve-Example)

### 🔧 函数调用

使用 NexaAI VLM 模型实现函数调用能力，并通过 MCP 协议集成 Google 日历。支持多模态输入（文本、图片、音频），包含 Web UI 和 CLI 界面。

- [演示](./function-calling)

### 📚 RAG-LLM

端到端的 检索增强生成（RAG）流程，集成嵌入、重排序与生成模型。可在本地查询 PDF、Word、文本等文档。

- [Python绑定示例](./RAG-LLM/Python-Binding-Example)
- [服务示例](./RAG-LLM/Serve-Example)

### 🖼️ 多模态-Qwen3VL

视觉语言模型（VLM），支持结构化 JSON 输出、多图片输入及自定义系统提示。

- [Python绑定示例](./Multimodal-Qwen3VL/Python-Binding-Example)

### 🔍 RAG-Qwen3VL

采用 Qwen3VL 多模态模型的轻量级 RAG 系统，支持 PDF、Word、文本、图片。带有 CLI 和 Gradio UI。

- [演示](./RAG-Qwen3VL)

### 🌐 Web Agent-Qwen3VL

本地多模态 LLM 驱动的网页自动化，基于 Qwen3-VL。支持自主网页导航、搜索及复杂网络任务。

- [演示](./Web%20Agent-Qwen3VL)

---

**平台支持**：Windows（x64/ARM64 Snapdragon）、macOS
**所有演示均在本地运行** — 数据不离开你的设备

