# Use a base image that includes Miniconda
FROM continuumio/miniconda3

# Set the working directory in the container
WORKDIR /app

RUN apt-get update && apt-get install -y cmake g++
# Copy the environment.yml file into the container at /app
COPY pyproject.toml CMakeLists.txt requirements.txt README.md ./
COPY dependency ./dependency

# Install the conda environment
RUN conda create -n nexa python=3.10 -y
RUN /bin/bash -c "source activate nexa && pip install -r requirements.txt && pip install -e ."

# Activate the environment
RUN echo "source activate nexa" > ~/.bashrc
ENV PATH /opt/conda/envs/nexa/bin:$PATH

# Copy your application code to the container
COPY . .

# Set the command to activate the environment and start your application
CMD ["bash", "-c", "source activate nexa && python -m nexa.cli.entry gen-text gemma"]