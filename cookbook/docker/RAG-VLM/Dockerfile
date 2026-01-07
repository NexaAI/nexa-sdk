# Dockerfile for AutoNeural Video Inference Demo
# Based on nexa4ai/nexasdk:latest

FROM nexa4ai/nexasdk:latest

LABEL org.opencontainers.image.title="AutoNeural Video Inference Demo"
LABEL org.opencontainers.image.description="Gradio UI for AutoNeural video analysis using Nexa SDK"

# Install Python 3 and pip if not already present
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get update \
    && apt-get install -y python3 python3-pip python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy application files
COPY requirements.txt .
COPY gradio_ui.py .
COPY start.sh .

# Install Python dependencies
RUN pip3 install --no-cache-dir -r requirements.txt

# Make start script executable
RUN chmod +x start.sh

# Expose ports
EXPOSE 18181 7860

# Use start script as entrypoint
ENTRYPOINT ["/app/start.sh"]

