# Nexa SDK Cookbook

This repository contains comprehensive demos and examples for Nexa SDK across different platforms: **PC (Python)**, **Android**, and **Linux**.

---

## 📁 Directory Structure

- **`PC/`** - Python bindings demos for Windows, macOS, and Linux
- **`android/`** - Android platform-specific demos (coming soon, see [bindings/android](../bindings/android) for Android SDK)
- **`linux/`** - Linux-specific demos (coming soon)

---

## 🖥️ PC Demos (Python Bindings)

All PC demos are located in the [`PC/`](./PC) directory and run on **Windows (x64/ARM64 Snapdragon)**, **macOS**, and **Linux**.

### 🤖 Agent-Granite

Lightweight on-device AI assistant with function calling (web search) using Granite-4-Micro model. Includes Gradio interface for local agentic workflows.

- [Python-Binding-Example](./PC/Agent-Granite/Python-Binding-Example)
- [Serve-Example](./PC/Agent-Granite/Serve-Example)

### 🔧 Function-Calling

Function calling capabilities with NexaAI VLM model, integrated with Google Calendar via MCP protocol. Supports multi-modal input (text, image, audio) with Web UI and CLI interfaces.

- [Demo](./PC/function-calling)

### 📚 RAG-LLM

End-to-end Retrieval-Augmented Generation pipeline with embeddings, reranking, and generation models. Query your own documents (PDFs, Word, text) locally on device.

- [Python-Binding-Example](./PC/RAG-LLM/Python-Binding-Example)
- [Serve-Example](./PC/RAG-LLM/Serve-Example)

### 🖼️ Multimodal-Qwen3VL

Vision-Language Model (VLM) with structured JSON outputs, multi-image input, and custom system prompt control.

- [Python-Binding-Example](./PC/Multimodal-Qwen3VL/Python-Binding-Example)

### 🔍 RAG-VLM

Lightweight RAG system with Qwen3VL multimodal model powered by Nexa Serve. Supports PDFs, Word docs, text files, and images. Includes CLI and Gradio UI.

- [Demo](./PC/RAG-VLM)

### 🌐 Web-Agent-Qwen3VL

Local multimodal LLM-driven browser automation using Qwen3-VL. Enables autonomous web navigation, searches, and complex web tasks.

- [Demo](./PC/Web-Agent-Qwen3VL)

---

## 📱 Android Demos

For Android SDK demos and examples, please refer to [`bindings/android`](../bindings/android).

---

## 🐧 Linux Demos

Linux-specific demos coming soon in the [`linux/`](./linux) directory.

---

## 🔒 Privacy First

**All demos run locally** — no data leaves your device.
