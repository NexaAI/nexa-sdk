# Image Search with Nexa SDK

This is an image search application using Nexa SDK's OpenAI-compatible embedding API.

## Prerequisites

1. **Start Nexa Serve**: Before running this application, you need to start the nexa serve service:

```bash
nexa pull NexaAI/EmbedNeural
nexa serve
```

2. **Install Dependencies**: Install the required Python packages:

```bash
pip install -r requirements.txt
```

## Usage

1. Start the Gradio UI:

```bash
gradio gradio_ui.py
```

2. **Upload Images**: Use the file uploader to upload image files (png, jpg, jpeg, etc.)

3. **Index Images**: Click the "Index" button to calculate embeddings for all uploaded images. This step is required before searching.

4. **Search**: Enter a text query in the search box and click the search button or press Enter to find similar images.

## Features

- **Text-to-Image Search**: Search images using natural language queries
- **L2 Distance**: Uses L2 (Euclidean) distance for similarity calculation
- **Embedding Caching**: Image embeddings are cached after indexing to avoid redundant calculations
- **Top-K Results**: Configurable number of search results (default: 5)

## Configuration

- **API Endpoint**: Default is `http://localhost:18181`
- **Model**: Default is `NexaAI/EmbedNeural`
- **Top-K**: Adjustable in the UI (default: 5)
- **Metric**: Currently supports L2 distance only

## Architecture

- `nexa_client.py`: Client for Nexa API using OpenAI library
- `search.py`: Search implementation with embedding caching
- `viewmodel.py`: Business logic and state management
- `gradio_ui.py`: Gradio user interface
- `style.py`: CSS styles for the UI

