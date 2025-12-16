# NexaAI VLM Function Call Demo with Google Calendar MCP

Demonstrates function calling capabilities of NexaAI/OmniNeural-4B model, integrated with Google Calendar via MCP protocol.

## Features

- Function calling with NexaAI VLM model
- Google Calendar integration via MCP
- Automatic function call parsing and execution
- Multi-modal input support (text, image, audio)
- Web UI and command-line interfaces

## Prerequisites

- **Python 3** (arm64 architecture recommended)
  - See [bindings/python/notebook/windows(arm64).ipynb](../../bindings/python/notebook/windows(arm64).ipynb) for setup
- **Node.js and npm** (for MCP server)
  ```powershell
  winget install OpenJS.NodeJS.LTS
  ```
  Restart terminal/IDE after installation.

## Installation

```bash
pip install -r requirements.txt
```

## Google Calendar Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create/select a project and enable [Google Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com)
3. Go to [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent) follow the instructions to configure your consent screen.
3. Create OAuth 2.0 credentials:
   - Go to [Credentials](https://console.cloud.google.com/apis/credentials)
   - Create "Create Credentials" > "OAuth client ID" > Select "Desktop app" > click "Create"
   - Click "Download JSON" and save as `gcp-oauth.keys.json` on the same directory as this README.md
4. Add your email as a test user in [Audience](https://console.cloud.google.com/auth/audience)
   - Click "Add User" > enter your email address > click "Save"
   - Note: Test mode tokens expire after 1 week

5. Authenication (only need to do once)
```powershell
$env:GOOGLE_OAUTH_CREDENTIALS="gcp-oauth.keys.json"
npx @cocal/google-calendar-mcp auth
```
follow the instructions to authorize the application to access your Google Calendar.

**Tip**: Ensure the OAuth client ID is enabled for Calendar API at [Credentials](https://console.cloud.google.com/apis/api/calendar-json.googleapis.com/credentials)
For detailed setup, see: https://github.com/nspady/google-calendar-mcp?tab=readme-ov-file#google-cloud-setup

## Usage

### Command Line

```bash
# Text only
python main.py --text "what is the time now?"

# Image with text
python main.py --image image.png --text "help me add this event to my calendar"

# Audio with text
python main.py --audio audio.mp3 --text "transcribe and add to calendar"
```
### Web UI

```powershell
python .\web\flask_ui.py
```

### Http server
```powershell
python main.py --serve --port 8088
```

Example curl request:
```bash
curl -X POST http://192.168.0.102:8088/api/function-call -H "Content-Type: application/json" -d '{"text": "what is the time now?"}'
```