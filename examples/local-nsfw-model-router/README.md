# Local NSFW Model Router
![image](https://github.com/MaokunZhang/AI-Soulmate-with-model-switcher/blob/main/preview_model_router.png)

## Introduction

The Local NSFW Model Router is a specialized tool designed for efficient comparison and utilization of AI language models optimized for NSFW (Not Safe For Work) content and role-playing scenarios. This project provides a local, privacy-focused solution for users to interact with and evaluate various AI models without relying on external servers or cloud-based services.

Key features include:
* Run models with [Nexa SDK](https://github.com/NexaAI/nexa-sdk) entirely on device for ultimate privacy
* Chat interface with [Streamlit](https://streamlit.io/) to chat with preset or customize character
* Built in model switcher
* Selection of uncensored language models at [Nexa Model Hub](https://nexa.ai/models?tasks=Uncensored) with RAM and disk size suggestions

## Installation

### Prerequisites
1. Set up [Miniconda](https://docs.anaconda.com/miniconda/miniconda-install/) and create a new conda virtual environment.
2. Ensure you have Git installed on your system.

### Step-by-Step Installation
1. Clone the repository:
   ```zsh
   git clone https://github.com/MaokunZhang/local-nsfw-model-router.git
   cd local-nsfw-model-router
   ```
2. Create and activate a new Conda environment:
   ```zsh
   conda create --name nsfw_model_router python=3.12
   conda activate nsfw_model_router
   ```
3. Install Nexa SDK:
   - For CPU:
     ```zsh
     pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cpu --extra-index-url https://pypi.org/simple --no-cache-dir
     ```
   - For GPU (Metal - macOS):
     ```zsh
     CMAKE_ARGS="-DGGML_METAL=ON -DSD_METAL=ON" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple --no-cache-dir
     ```
   - For CUDA and AMD GPU support, refer to the [Nexa SDK Installation Guide](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#installation).
4. Install other dependencies:
   ```zsh
   pip install -r requirements.txt
   ```

## Usage

Run the Streamlit app:
```zsh
streamlit run app.py
```
Follow the on-screen instructions to start chatting with the default model or switch between different models.

## File Structure
- `app.py`: Main Streamlit application
- `utils/`:
  - `initialize.py`: Initializes chat and loads models
  - `gen_response.py`: Handles output generation
  - `customize.py`: Allows users to customize character roles

## Resources
- [NexaAI Model Hub](https://nexaai.com/models)
- [Nexa-SDK GitHub Repository](https://github.com/NexaAI/nexa-sdk)

## Disclaimer
This project is intended for adult users and contains NSFW content. Use responsibly and in accordance with local laws and regulations.
