## NexaAI SDK Demo: On-Device Function Calling for Weather Retrieval

### Introduction
This demo highlights the **NexaAI SDK's on-device function calling** capability by fetching real-time weather information without relying on cloud-based AI models. It understands your instruction and retrieves weather data from `wttr.in`, a lightweight and publicly accessible weather API.  

By running the LLM model locally, this approach minimizes reliance on external LLM APIs that require authentication, making it a more efficient and privacy-friendly solution for offline applications.  

---

### How to Run the Demo

1. *(Optional)* Create and activate a virtual Python environment using `conda`:
    ```bash
    conda create -n function-calling python=3.12
    conda activate function-calling
    ```
2. Install Nexa SDK by following the instructions in the main repository's README.
3. Run the demo script:
    ```bash
    python main.py
    ```
