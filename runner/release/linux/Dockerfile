# Multi-architecture Linux Docker image for Nexa SDK
# Supports both amd64 and arm64 platforms
# Usage: docker run --rm -it -v /path/to/data:/root/.cache/nexa.ai nexa4ai/nexasdk:latest

ARG BASE_IMAGE=ubuntu:22.04

FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.title="Nexa SDK"
LABEL org.opencontainers.image.description="Nexa SDK CLI for running multi models"
LABEL org.opencontainers.image.vendor="NexaAI"

RUN DEBIAN_FRONTEND=noninteractive \
    apt-get update \
    && apt-get install -y libgomp1 curl ffmpeg sox \
    && rm -rf /var/lib/apt/lists/*

ARG TARGETARCH
COPY nexa_sdk.$TARGETARCH/ /opt/nexa_sdk/
RUN ln -s /opt/nexa_sdk/nexa /usr/local/bin/nexa

RUN mkdir -p /data && mkdir -p /root/.cache && ln -s /data /root/.cache/nexa.ai

EXPOSE 18181
ENTRYPOINT ["nexa"]
