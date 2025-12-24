# Web Agent with Qwen3-VL

## Overview

This demo integrates Nexa SDK with Web-UI to enable local multimodal LLM-driven browser automation. The agent can interact with websites, perform searches, and execute complex web tasks autonomously using the `NexaAI/Qwen3-VL-4B-Instruct-GGUF` model (recommended).

Built upon [@browser-use/web-ui](https://github.com/browser-use/web-ui), this integration demonstrates the power of local vision-language models for web automation tasks.

## Prerequisites

- Python 3.11+
- Nexa SDK installed
- Hugging Face account with access token
- Sufficient storage space (4GB+ for model)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/mtilyxuegao/Nexa-Web-UI.git
cd Nexa-Web-UI
```

### 2. Install Nexa SDK

Download and install the Nexa SDK package from the official GitHub repository:

Visit [https://github.com/NexaAI/nexa-sdk/releases/tag/v0.2.49](https://github.com/NexaAI/nexa-sdk/releases/tag/v0.2.49) to download the appropriate installer for your platform and install it.

### 3. Set Up Python Environment

Choose one of the following methods to set up your Python environment:

#### Option A: Using uv (Recommended)

```bash
# Navigate to web-ui directory
cd web-ui

# Create virtual environment with uv
uv venv --python 3.11

# Activate virtual environment  
source .venv/bin/activate  # macOS/Linux
# or .\.venv\Scripts\Activate.ps1  # Windows PowerShell

# Install Python dependencies
uv pip install -r requirements.txt

# (Optional) Install memory features for enhanced agent learning capabilities
# This adds ~110MB of ML dependencies (torch, transformers, etc.)
uv pip install "browser-use[memory]"

# Install Playwright browsers (recommend Chromium only)
playwright install chromium --with-deps
```

#### Option B: Using conda

```bash
# Navigate to web-ui directory
cd web-ui

# Create conda environment
conda create -n nexa-webui python=3.11 -y
conda activate nexa-webui

# Install Python dependencies
pip install -r requirements.txt

# (Optional) Install memory features for enhanced agent learning capabilities
# This adds ~110MB of ML dependencies (torch, transformers, etc.)
pip install "browser-use[memory]"

# Install Playwright browsers (recommend Chromium only)
playwright install chromium --with-deps
```

### 4. Configure Environment Variables

The project includes a preconfigured `web-ui/.env` file with the following main settings:

```bash
# LLM Provider Settings
DEFAULT_LLM=nexa
NEXA_ENDPOINT=http://127.0.0.1:8080/v1

# Other API Keys (if using other LLMs)
# OPENAI_API_KEY=your_openai_key
# ANTHROPIC_API_KEY=your_anthropic_key
```

### 5. Download the Model

Download the multimodal VLM model:

```bash
# Download the model
nexa pull NexaAI/Qwen3-VL-4B-Instruct-GGUF
```

> **Optional**: If you want to set up Hugging Face token for accessing other private models:
> ```bash
> export HUGGINGFACE_HUB_TOKEN="your_huggingface_token"
> export NEXA_HFTOKEN="your_huggingface_token"
> ```

**Important Notes**:
- Model size is approximately 4GB
- Ensure sufficient storage and bandwidth

## Running the Demo

### Step 1: Clean Up Ports (Optional)

Before starting, ensure ports are clean:

```bash
# Kill all related processes
lsof -ti:8080,7788 | xargs kill -9 2>/dev/null
pkill -f "nexa serve"
pkill -f "webui.py"
```

### Step 2: Start Nexa Server

```bash
# Navigate to project root directory
cd Nexa-Web-UI
nexa serve --host 127.0.0.1:8080 --keepalive 600
```

Wait until you see the message: `Localhosting on http://127.0.0.1:8080/docs/ui`

### Step 3: Start Web-UI

In a **new terminal window**:

```bash
# Navigate to project root directory
cd Nexa-Web-UI

# Activate your Python environment
source web-ui/.venv/bin/activate  # or conda activate nexa-webui

# Start the web interface
python web-ui/webui.py --ip 127.0.0.1 --port 7788
```

Wait until you see the message: `Running on local URL: http://127.0.0.1:7788`

### Step 4: Access the Web Interface

Open your browser and visit: [http://127.0.0.1:7788](http://127.0.0.1:7788)

## Usage Example

Follow these steps to run the agent:

### Step 1: Verify Agent Settings
1. In the web interface, navigate to the **Agent Settings** tab
2. Verify that:
   - **LLM Provider** is set to `nexa`
   - **LLM Model Name** is set to `NexaAI/Qwen3-VL-4B-Instruct-GGUF`

### Step 2: Run the Agent
1. Click on the **Run Agent** tab (the third tab)
2. Enter your task in the input field

### Step 3: Try an Example Task

**Task**: `Go to google.com, search for 'nexa ai', and click the first element`

**Expected Behavior**:
1. Agent navigates to google.com
2. Enters 'nexa ai' in the search box and performs search
3. Clicks the first element in the search results
4. Reports completion status

The agent will automatically execute these steps using vision-language understanding to interact with the webpage.

## Features

- 🤖 **Autonomous Web Navigation**: AI agent can understand and interact with web pages
- 👁️ **Vision-Language Understanding**: Uses multimodal model to "see" and understand web content
- 🏠 **Local Execution**: Runs entirely on your machine with Nexa SDK
- 🎯 **Task-Based Interaction**: Simply describe what you want to accomplish
- 🔄 **Multi-Step Workflows**: Can execute complex sequences of actions

## Troubleshooting

### Port Already in Use
If you encounter port conflicts, use the cleanup command:
```bash
lsof -ti:8080,7788 | xargs kill -9 2>/dev/null
```

### Model Download Issues
- Ensure your Hugging Face token is valid and has access to the model
- Check your internet connection and available disk space
- Verify both `HUGGINGFACE_HUB_TOKEN` and `NEXA_HFTOKEN` are set

### Browser Issues
If Playwright browsers fail to install:
```bash
playwright install chromium --with-deps --force
```

## Acknowledgments

We would like to officially thank the [browser-use/web-ui](https://github.com/browser-use/web-ui) project and its contributors for providing the foundation that makes this integration possible.

## Additional Resources

- [Nexa SDK Repository](https://github.com/NexaAI/nexa-sdk)
- [Browser Use Web-UI](https://github.com/browser-use/web-ui)
- [Nexa SDK Documentation](https://docs.nexaai.com)

