#!/usr/bin/env python3
"""
Test client for WebSocket ASR streaming
"""
import asyncio
import json
import struct
import websockets
import numpy as np

async def test_websocket_stream():
    uri = "ws://localhost:8080/v1/audio/stream"
    
    try:
        async with websockets.connect(uri) as websocket:
            print("Connected to WebSocket server")
            
            # Send configuration
            config = {
                "model": "NexaAI/parakeet-npu",
                "language": "en-US",
                "sample_rate": 16000,
                "enable_partial_results": True,
                "enable_word_timestamps": False,
                "vad_enabled": True,
                "chunk_duration": 0.5,
                "beam_size": 5
            }
            
            await websocket.send(json.dumps(config))
            print(f"Sent config: {config}")
            
            # Generate test audio data (sine wave)
            # This simulates 1 second of audio at 16kHz
            sample_rate = 16000
            duration = 1.0  # seconds
            frequency = 440  # Hz (A4 note)
            
            t = np.linspace(0, duration, int(sample_rate * duration))
            audio_float = np.sin(2 * np.pi * frequency * t)
            
            # Convert to 16-bit PCM
            audio_int16 = (audio_float * 32767).astype(np.int16)
            audio_bytes = audio_int16.tobytes()
            
            # Send audio in chunks
            chunk_size = 8000  # Send 0.5 seconds at a time
            for i in range(0, len(audio_bytes), chunk_size):
                chunk = audio_bytes[i:i+chunk_size]
                await websocket.send(chunk)
                print(f"Sent audio chunk {i//chunk_size + 1} ({len(chunk)} bytes)")
                
                # Try to receive any responses
                try:
                    response = await asyncio.wait_for(websocket.recv(), timeout=0.1)
                    result = json.loads(response)
                    print(f"Received: {result}")
                except asyncio.TimeoutError:
                    pass
                
                await asyncio.sleep(0.1)
            
            # Send stop signal
            stop_msg = json.dumps({"action": "stop"})
            await websocket.send(stop_msg)
            print("Sent stop signal")
            
            # Wait for final responses
            try:
                while True:
                    response = await asyncio.wait_for(websocket.recv(), timeout=1.0)
                    result = json.loads(response)
                    print(f"Final response: {result}")
            except asyncio.TimeoutError:
                print("No more responses")
            
            print("Test completed successfully!")
            
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    print("WebSocket ASR Streaming Test Client")
    print("=" * 50)
    asyncio.run(test_websocket_stream())
