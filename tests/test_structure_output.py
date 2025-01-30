import json
from nexa.gguf import NexaTextInference
from pathlib import Path

def main():
    nexa = NexaTextInference(model_path="Meta-Llama-3.1-8B-Instruct:q4_0")
    prompt = "Emily Carter, a 32-year-old owner, drives a 2023 Audi Q5 available in White, Black, and Gray, equipped with Bang & Olufsen audio (19 speakers, Bluetooth), 10 airbags, 8 parking sensors, lane assist, and a turbocharged inline-4 engine delivering 261 horsepower and a top speed of 130 mph."
    schema_abs_path = (Path(__file__).parent / "structure_decoding_resources/schema.json").resolve()
    response = nexa.structure_output(prompt=prompt, json_schema_path=schema_abs_path)
    print(f"response: {json.dumps(response, indent=4)}")

if __name__ == "__main__":
    main()