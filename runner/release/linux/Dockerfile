FROM ubuntu:22.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y libgomp1 curl ffmpeg sox
COPY nexa_sdk.$TARGETARCH /opt/nexa_sdk
RUN ln -s /opt/nexa_sdk/nexa /usr/local/bin/nexa

CMD ["nexa", "serve"]
