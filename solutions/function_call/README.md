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

## Configuration google calendar auth

1. Go to the [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project or select an existing one.
3. Enable the [Google Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com) for your project. Ensure that the right project is selected from the top bar before enabling the API.
4. Create OAuth 2.0 credentials:
   - Go to [Credentials](https://console.cloud.google.com/apis/credentials)
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Desktop app" as the application type (Important!)
   - Save the auth key, you'll need to add its path to the JSON in the next step, e.g. `gcp-oauth.keys.json`
- Add your email address as a test user under the [Audience screen](https://console.cloud.google.com/auth/audience)
    - Note: it might take a few minutes for the test user to be added. The OAuth consent will not allow you to proceed until the test user has propagated.
    - Note about test mode: While an app is in test mode the auth tokens will expire after 1 week and need to be refreshed (see Re-authentication section below).


tips: ensure https://console.cloud.google.com/apis/api/calendar-json.googleapis.com/credentials is enabled and you oauth client id is set on it

more details: https://github.com/nspady/google-calendar-mcp?tab=readme-ov-file#google-cloud-setup


## Usage

Run the demo with command line arguments:

```bash
# Text input only
python main.py --text "what is the time now?"

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
