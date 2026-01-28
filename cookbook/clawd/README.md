# Clawdbot + Nexa SDK Examples

Practical examples of integrating Nexa SDK with Clawdbot.

## Prerequisites
- Node.js ≥ 22
- [Nexa SDK installed](../../README.md#installation)
- Clawdbot installed: `npm install -g clawdbot@latest`

## Quick Start

Configure Clawdbot to use Nexa SDK's OpenAI-compatible API.

1. Run nexa serve in separate terminal.
```bash
nexa pull NexaAI/Qwen3-4B-GGUF
nexa serve
```

2. Configure Clawdbot by copying the example configuration:
```bash
cp clawdbot.example.json ~/.clawdbot/clawdbot.json
```

3. Start Clawdbot gateway:
```bash
clawdbot gateway
```

4. Test the integration:
   - Open browser control UI: http://127.0.0.1:18789/
   - Or send a test message via CLI (requires channel setup):
   ```bash
   clawdbot message send --target <your-phone-number> --message "Hello from Nexa SDK"
   ```

## Testing

Once the gateway is running, you can test the integration:

1. **Browser Control UI** (Recommended for testing):
   - Open http://127.0.0.1:18789/ in your browser
   - This web interface allows you to chat directly without configuring channels
   - Simply type a message and the bot will respond using Nexa SDK

2. **Check Model Status**:
   ```bash
   clawdbot models status
   ```

3. **Configure Channels** (Optional, for WhatsApp/Telegram/etc.):
   ```bash
   # Configure a channel (e.g., WhatsApp)
   clawdbot channels login
   
   # Then you can send messages via CLI
   clawdbot message send --target <phone-number> --message "Hello"
   ```

**Note**: 
- The Browser Control UI is the easiest way to test without channel setup
- Make sure Nexa server is running (`nexa serve`) before testing, as Clawdbot will call the Nexa API when processing messages
- For production use, configure channels (WhatsApp, Telegram, etc.) using `clawdbot channels login`

## Configuration

The example configuration (`clawdbot.example.json`) shows how to configure a custom provider using `models.providers` to connect to Nexa SDK's OpenAI-compatible API endpoint at `http://localhost:18181/v1`.

**Important Notes:**
- The model `id` must match the exact model name returned by Nexa server (including quantization info, e.g., `NexaAI/Qwen3-4B-GGUF:Q4_0`)
- The model reference format is `provider/model-id` (e.g., `nexa/NexaAI/Qwen3-4B-GGUF:Q4_0`)
- Check available models with `nexa list` or `curl http://localhost:18181/v1/models`
- If you use a different model or quantization, update both the `id` in `models.providers.nexa.models[].id` and the `primary` in `agents.defaults.model.primary`
