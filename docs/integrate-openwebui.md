# OpenWebUI Integration Guide

This guide shows how to integrate Nexa-SDK with [OpenWebUI](https://openwebui.com/) for a web-based chat interface with locally running models.

## Overview

OpenWebUI is a modern, feature-rich web interface for interacting with language models. By integrating it with Nexa-SDK, you get:

- **Web-based chat interface** - Access models through any browser
- **Model management** - Switch between different models seamlessly
- **Local execution** - All models run locally on your machine
- **OpenAI API compatibility** - Works with any OpenAI-compatible API

## Prerequisites

Before starting, ensure you have:

- **Nexa-SDK installed** - Follow [Nexa-SDK installation guide](../README.md)
- **OpenWebUI installed** - See [OpenWebUI Installation Guide](https://docs.openwebui.com/getting-started/quick-start)

## Step 1: Start Nexa Serve

First, download the model you want to use:

```bash
nexa pull Qwen/Qwen3-VL-8B-Instruct-GGUF
```

Then start the Nexa Serve API server:

```bash
nexa serve
```

The server will start and listen on `http://127.0.0.1:18181`. You should see output like:

```
[INFO] Nexa Serve running on http://127.0.0.1:18181
[INFO] OpenAI-compatible API at http://127.0.0.1:18181/v1
```

## Step 2: Install OpenWebUI
Follow the [OpenWebUI Installation Guide](https://docs.openwebui.com/getting-started/quick-start) to set up OpenWebUI on your machine.

## Step 3: Configure OpenWebUI

1. Start OpenWebUI, open your browser and navigate to your OpenWebUI instance, typically at:
   ```
   http://localhost:8080
   ```

2. **Sign up and login**:
   - Click the sign-up button
   - Create an account with your email and password
   - Login with your credentials

3. **Navigate to Settings**:
   - Click the **Profile icon** in the top right corner
   - Select **Admin Panel** from the dropdown menu
   - Click **Settings** in the top navigation bar

4. **Configure Nexa-SDK Connection**:
   - In the left sidebar, click **Connections**
   - Scroll to the **OpenAI API** section
   - Click **+ Create** or **Add New Connection**

5. **Fill in Connection Details**:
   - **API URL**: `http://127.0.0.1:18181/v1`
   - **API Base**: `http://localhost:8080`
   - **API Key**: `any-key` (Nexa Serve doesn't require authentication)

6. **Add Model ID**:
   - In the **Model ID** field, enter the model you downloaded, for example:
     ```
     Qwen/Qwen3-VL-8B-Instruct-GGUF
     ```
   - Click **Save** to add the connection

## Step 4: Start Chatting

1. **Go back to the Chat interface** (click the OpenWebUI logo or close settings)
2. **Select the Nexa model** from the model dropdown at the top
3. **Start typing** your prompts and press Enter
4. **Enjoy local AI** without any data leaving your machine!

## Troubleshooting

### Connection Error: "Failed to connect to API"

**Problem**: OpenWebUI cannot reach Nexa Serve

**Solutions**:
- Ensure Nexa Serve is running: `nexa serve`
- Check the API URL is correct: `http://127.0.0.1:18181/v1`
- Check firewall settings - port 18181 should be accessible

### Model Not Found

**Problem**: The model ID is not recognized

**Solutions**:
- Download the model first: `nexa pull Qwen/Qwen3-VL-8B-Instruct-GGUF`
- Verify the model is downloaded: `nexa list`
- Use the exact model ID from the `nexa list` output
