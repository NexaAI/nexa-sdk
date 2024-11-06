import json
from prompts import COLUMN_CHART_TEMPLATE, PIE_CHART_TEMPLATE
import logging
from nexa.gguf import NexaTextInference
from nexa.general import pull_model

# Set the logging level for the openai logger to WARNING
logging.getLogger().setLevel(logging.ERROR)

def get_template_and_model_path(chart_type: str) -> str:
    if chart_type == "COLUMN_CLUSTERED" or chart_type == None: # Will also use template for pure text calling
        return COLUMN_CHART_TEMPLATE, "DavidHandsome/Column-Chart-LoRA:gguf-fp16"
    elif chart_type == "PIE":
        return PIE_CHART_TEMPLATE, "DavidHandsome/Pie-Chart-LoRA:gguf-fp16"
    else:
        raise ValueError(f"Invalid chart type: {chart_type}")

def clean_response(raw_response: str) -> dict:
    start = raw_response.find('{')
    if start == -1:
        print("No JSON object found in the response.")
        return None  # No JSON object found

    brace_count = 0
    in_string = False
    escape = False
    end = start

    while end < len(raw_response):
        char = raw_response[end]
        if char == '"' and not escape:
            in_string = not in_string
        if not in_string:
            if char == '{':
                brace_count += 1
            elif char == '}':
                brace_count -= 1
                if brace_count == 0:
                    json_str = raw_response[start:end+1]
                    try:
                        return json.loads(json_str)
                    except json.JSONDecodeError as e:
                        print(f"JSON decoding failed: {e}")
                        return None
        if char == '\\' and not escape:
            escape = True
        else:
            escape = False
        end += 1

    print("No complete JSON object found in the response.")
    return None  # Didn't find a matching closing brace


def generation_chart_data(text, lora_model_path, chat_template):
    local_lora_path, run_type = pull_model(lora_model_path)
    chart_model = NexaTextInference(
        model_path="gemma-2-2b-instruct:fp16",
        lora_path=local_lora_path,
    )

    prompt = chat_template.format(input=text)
    response = chart_model.create_completion(prompt, stop=["<end>"])

    return response["choices"][0]["text"]

def execute_chart_generation(input_text, chart_type):
    template, lora_model_path = get_template_and_model_path(chart_type)
    raw_response = generation_chart_data(text = input_text, lora_model_path = lora_model_path, chat_template = template)
    print(raw_response)
    
    cleaned_response_dict = clean_response(raw_response)
    
    return cleaned_response_dict
    