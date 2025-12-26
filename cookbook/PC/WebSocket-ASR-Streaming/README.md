# WebSocket Streaming ASR API

This document describes the WebSocket-based streaming ASR (Automatic Speech Recognition) API for real-time audio transcription.

## Endpoint

```
ws://localhost:8080/v1/audio/stream
```

For HTTPS servers:
```
wss://your-server.com/v1/audio/stream
```

## Connection Flow

1. **Connect** to the WebSocket endpoint
2. **Send configuration** as the first message (JSON)
3. **Stream audio data** as binary messages
4. **Receive transcription results** as JSON messages
5. **Close connection** or send stop signal when done

## Configuration Message

The first message sent to the server must be a JSON configuration object:

```json
{
  "model": "NexaAI/parakeet-npu",
  "language": "en-US",
  "sample_rate": 16000,
  "enable_partial_results": true,
  "enable_word_timestamps": false,
  "vad_enabled": true,
  "chunk_duration": 0.5,
  "overlap_duration": 0.0,
  "beam_size": 5
}
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `model` | string | `"NexaAI/parakeet-npu"` | Model identifier to use for ASR |
| `language` | string | `"en-US"` | Language code (e.g., "en-US", "zh-CN") |
| `sample_rate` | integer | `16000` | Audio sample rate in Hz |
| `enable_partial_results` | boolean | `true` | Whether to send partial results during transcription |
| `enable_word_timestamps` | boolean | `false` | Whether to include word-level timestamps (if supported) |
| `vad_enabled` | boolean | `true` | Enable Voice Activity Detection |
| `chunk_duration` | float | `0.5` | Audio chunk duration in seconds |
| `overlap_duration` | float | `0.0` | Overlap between chunks in seconds |
| `beam_size` | integer | `5` | Beam size for decoding |

## Audio Data Format

After sending the configuration, stream audio data as **binary WebSocket messages**.

### Requirements:
- **Format**: 16-bit PCM (little-endian)
- **Sample Rate**: Must match the configured `sample_rate` (default: 16000 Hz)
- **Channels**: Mono (1 channel)
- **Encoding**: Raw PCM bytes

### Example (Python):
```python
import numpy as np

# Generate audio samples (float32, range [-1.0, 1.0])
audio_float = np.sin(2 * np.pi * 440 * t)  # 440 Hz sine wave

# Convert to 16-bit PCM
audio_int16 = (audio_float * 32767).astype(np.int16)
audio_bytes = audio_int16.tobytes()

# Send via WebSocket
await websocket.send(audio_bytes)
```

## Transcription Response

The server sends transcription results as JSON messages:

```json
{
  "type": "partial",
  "text": "hello world",
  "confidence": 0.95,
  "timestamp": 1703123456789,
  "is_final": false
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"partial"` or `"final"` |
| `text` | string | Transcribed text |
| `confidence` | float | Confidence score (0.0-1.0), optional |
| `timestamp` | integer | Unix timestamp in milliseconds |
| `is_final` | boolean | Whether this is a final result |

## Control Messages

Send control messages as JSON text messages:

### Stop Signal
```json
{
  "action": "stop"
}
```

## Error Handling

Errors are returned as JSON messages:

```json
{
  "type": "error",
  "error": "Error message description"
}
```

## Python Example

```python
import asyncio
import json
import websockets
import numpy as np

async def transcribe_audio():
    uri = "ws://localhost:8080/v1/audio/stream"
    
    async with websockets.connect(uri) as websocket:
        # Send configuration
        config = {
            "model": "NexaAI/parakeet-npu",
            "language": "en-US",
            "sample_rate": 16000,
            "enable_partial_results": True
        }
        await websocket.send(json.dumps(config))
        
        # Stream audio from microphone or file
        # (This is a simplified example)
        sample_rate = 16000
        duration = 1.0
        t = np.linspace(0, duration, int(sample_rate * duration))
        audio_float = np.sin(2 * np.pi * 440 * t)
        audio_int16 = (audio_float * 32767).astype(np.int16)
        
        # Send audio chunks
        chunk_size = 8000  # 0.5 seconds at 16kHz
        for i in range(0, len(audio_int16), chunk_size):
            chunk = audio_int16[i:i+chunk_size].tobytes()
            await websocket.send(chunk)
            
            # Receive transcription
            try:
                response = await asyncio.wait_for(websocket.recv(), timeout=0.1)
                result = json.loads(response)
                print(f"Transcription: {result['text']}")
            except asyncio.TimeoutError:
                pass
        
        # Send stop signal
        await websocket.send(json.dumps({"action": "stop"}))

asyncio.run(transcribe_audio())
```

## JavaScript/Browser Example

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/v1/audio/stream');

ws.onopen = () => {
    // Send configuration
    const config = {
        model: 'NexaAI/parakeet-npu',
        language: 'en-US',
        sample_rate: 16000,
        enable_partial_results: true
    };
    ws.send(JSON.stringify(config));
    
    // Start capturing audio from microphone
    navigator.mediaDevices.getUserMedia({ 
        audio: { 
            sampleRate: 16000, 
            channelCount: 1 
        } 
    }).then(stream => {
        const audioContext = new AudioContext({ sampleRate: 16000 });
        const source = audioContext.createMediaStreamSource(stream);
        const processor = audioContext.createScriptProcessor(4096, 1, 1);
        
        processor.onaudioprocess = (e) => {
            const audioData = e.inputBuffer.getChannelData(0);
            
            // Convert float32 to int16 PCM
            const int16Data = new Int16Array(audioData.length);
            for (let i = 0; i < audioData.length; i++) {
                const s = Math.max(-1, Math.min(1, audioData[i]));
                int16Data[i] = s < 0 ? s * 0x8000 : s * 0x7FFF;
            }
            
            ws.send(int16Data.buffer);
        };
        
        source.connect(processor);
        processor.connect(audioContext.destination);
    });
};

ws.onmessage = (event) => {
    const result = JSON.parse(event.data);
    if (result.text) {
        console.log('Transcription:', result.text);
    }
};
```

## C# Example

```csharp
using System;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

async Task TranscribeAudioAsync()
{
    using var ws = new ClientWebSocket();
    await ws.ConnectAsync(new Uri("ws://localhost:8080/v1/audio/stream"), CancellationToken.None);
    
    // Send configuration
    var config = new
    {
        model = "NexaAI/parakeet-npu",
        language = "en-US",
        sample_rate = 16000,
        enable_partial_results = true
    };
    var configJson = JsonSerializer.Serialize(config);
    var configBytes = Encoding.UTF8.GetBytes(configJson);
    await ws.SendAsync(configBytes, WebSocketMessageType.Text, true, CancellationToken.None);
    
    // Send audio chunks
    byte[] audioChunk = GetAudioChunk(); // Your audio data (16-bit PCM)
    await ws.SendAsync(audioChunk, WebSocketMessageType.Binary, true, CancellationToken.None);
    
    // Receive transcription
    var buffer = new byte[4096];
    var result = await ws.ReceiveAsync(buffer, CancellationToken.None);
    var json = Encoding.UTF8.GetString(buffer, 0, result.Count);
    var transcription = JsonSerializer.Deserialize<TranscriptionResponse>(json);
    Console.WriteLine($"Transcription: {transcription.text}");
    
    await ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "", CancellationToken.None);
}
```

## Performance Considerations

### Latency
- **Typical latency**: 300-800ms from audio input to transcription output
- Affected by: chunk duration, network latency, model complexity

### Resource Usage
- Memory usage is constant (no large buffer accumulation)
- Suitable for long-running transcription sessions

### Best Practices
1. Use appropriate `chunk_duration` (0.3-0.5s recommended)
2. Enable `vad_enabled` to filter out silence
3. Use `enable_partial_results: true` for real-time feedback
4. Handle WebSocket reconnection for production use
5. Implement proper error handling and timeouts

## Comparison with Batch API

| Metric | WebSocket Streaming | Batch API (`/v1/audio/transcriptions`) |
|--------|--------------------|-----------------------------------------|
| Latency | 300-800ms | 1-3s |
| Use Case | Real-time | Offline processing |
| Memory | Constant | Spikes with file size |
| Feedback | Incremental | Complete file only |

## Troubleshooting

### Connection Issues
- Verify WebSocket endpoint is accessible
- Check CORS settings if connecting from browser
- Ensure proper WebSocket protocol (ws:// or wss://)

### Audio Quality Issues
- Verify audio format (16-bit PCM, mono)
- Check sample rate matches configuration
- Ensure audio is not corrupted or silent

### No Transcription Output
- Verify model is loaded correctly
- Check server logs for errors
- Ensure audio data is being sent properly
- Confirm language is supported by the model

## Security Considerations

- Use WSS (WebSocket Secure) in production
- Implement authentication if needed
- Validate and sanitize all inputs
- Rate limit connections to prevent abuse
