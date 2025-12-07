# NexaAI VLM Function Call Demo with Google Calendar MCP

Demonstrates the function calling capabilities of NexaAI/OmniNeural-4B model, integrated with Google Calendar MCP server.

## Features

- Uses NexaAI VLM model for function calling
- Connects to Google Calendar via MCP protocol
- Automatically parses and executes function calls
- Supports text, image, and audio inputs via command line

## Prerequisites

- Python 3 in architecture arm64
- Node.js and npm (for running MCP server)
```
winget install OpenJS.NodeJS.LTS
```
Restart terminal or your IDE after installation.


## Installation

```bash
pip install -r requirements.txt
```

## Configuration

1. Create OAuth 2.0 credentials in Google Cloud Console: https://github.com/nspady/google-calendar-mcp?tab=readme-ov-file#google-cloud-setup
2. Download the credentials file and save as `gcp-oauth.keys.json` (or specify via `--credentials` argument)

## Usage

Run the demo with command line arguments:

```bash
# Text input only
python main.py --text "help me add a meeting tomorrow at 2pm"

# Image input with text
python main.py --image image.png --text "help me add this event to my calendar"

# Audio input with text
python main.py --audio audio.mp3 --text "transcribe and add to calendar"
```

On first run, you'll be prompted to authorize the application to access your Google Calendar. The authorization token will be saved for future use.

### Command Line Arguments

- `--credentials`: Google OAuth credentials file path (default: `gcp-oauth.keys.json`)
- `--text`: Text input
- `--image`: Image file path
- `--audio`: Audio file path

### Environment Variables

- `MODEL_PATH`: Path to the VLM model (default: `NexaAI/OmniNeural-4B`)
- `MAX_TOKENS`: Maximum tokens to generate (default: `2048`)

### Example

```bash
python main.py --image event.png --text "add this event to my calendar"
```

Output:
```
[Function call: create-event]
[Arguments: {'summary': '...', 'start': {...}, 'end': {...}}]
[Function result: {...}]
Assistant: I've successfully added the event to your calendar.
```

## Project Structure

```
solutions/function_call/
├── main.py           # Main entry point
├── mcp_utils.py     # MCP utility functions
├── function_parser.py # Function call parser
├── requirements.txt  # Python dependencies
└── README.md        # This file
```
