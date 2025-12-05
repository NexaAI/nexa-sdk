# Function Call Demo with Google Calendar

This demo showcases the function calling capabilities of NexaAI/OmniNeural-4B model, integrated with Google Calendar MCP application.

## Features

- Interactive command-line interface for chatting with the model
- Google Calendar integration via OAuth 2.0
- Function calling support for calendar operations:
  - Create calendar events
  - List calendar events
  - Update calendar events
  - Delete calendar events

## Setup

### 1. Install Dependencies

```bash
pip install -r requirements.txt
```

### 2. Configure Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Calendar API
4. Create OAuth 2.0 credentials (Desktop app)
5. Download the credentials JSON file and save it as `gcp-oauth.keys.json` in this directory

**Note**: This demo uses the existing MCP server `@cocal/google-calendar-mcp` via npx, which handles OAuth authentication automatically.

### 3. Run the Demo

```bash
python main.py
```

On first run, you'll be prompted to authorize the application to access your Google Calendar. The authorization token will be saved for future use.

## Usage

Once started, you can interact with the model through natural language:

```
User: Add a meeting tomorrow at 2pm called "Team Standup"
Assistant: [Function call: create_calendar_event]
          I've successfully created a calendar event "Team Standup" for tomorrow at 2:00 PM.
```

## Project Structure

```
solutions/function_call/
├── main.py                 # Main entry point with CLI interface
├── mcp/                    # MCP application modules
│   ├── __init__.py
│   ├── base.py            # MCP base classes and interfaces
│   └── google_calendar.py # Google Calendar MCP implementation
├── models/                 # Model configuration
│   └── config.py          # Model config and tool definitions
├── utils/                  # Utility functions
│   ├── __init__.py
│   └── function_parser.py  # Function call parser
├── requirements.txt        # Python dependencies
├── README.md              # This file
└── .env.example           # Environment variables template
```

## Notes

- The first authorization requires browser access for OAuth flow
- Tokens are stored locally in `.google_token.json`
- The model uses OmniNeural-4B by default, but can be configured via environment variables

