import json
import jsonschema
import datetime
from typing import Dict, Any, Optional, Union
import streamlit as st
import base64
import os
import requests
import logging
import io
from utils.schemas import schema

logger = logging.getLogger(__name__)


def text_to_speech_elevenlabs(text):
    text = ''.join(text)
    url = f"https://api.elevenlabs.io/v1/text-to-speech/21m00Tcm4TlvDq8ikWAM"
    elevenlab_api = os.getenv('elevenlabs_api')
    headers = {
    "accept": "audio/mpeg",
    "xi-api-key": elevenlab_api,
    "Content-Type": "application/json",
    }

    data = {"text": text}
    response = requests.post(url, headers = headers, json=data)

    if response.status_code == 200:
        print(response.status_code)
        return response.content
    else:
        print(f"Error: {response.status_code}, {response.content}")
        return None
    

def play_audio(text):
    try:
        if '```' in text:
            text = text.split('```')[-1]
        print(text)
        audio_content = text_to_speech_elevenlabs(text)
        # audio_content = text_to_speech_openai(text)

        if audio_content is not None:
            audio_content = io.BytesIO(audio_content)
            audio_content.seek(0) 
            audio_base_64 = base64.b64encode(audio_content.read()).decode("utf-8")
            st.audio(audio_content, format='audio/mpeg', autoplay=True)
    except Exception as e:
        print(f'Error :{e}')



import os
import requests
from dotenv import load_dotenv

load_dotenv()

groq_api_key = os.getenv('GROQ_API_KEY')
helicone_api_key = os.getenv('HELICONE_API_KEY')

headers = {
    'Content-Type': 'application/json',
    'Authorization': f'Bearer {groq_api_key}',
    'Helicone-Auth': f"Bearer {helicone_api_key}",
    'Helicone-Target-URL': 'https://api.groq.com',
}

url = "https://groq.helicone.ai/openai/v1/chat/completions"

def send_message(message, system_instruction=None):
    messages = []
    if system_instruction:
        messages.append({"role": "system", "content": system_instruction})
    messages.append({"role": "user", "content": message})

    data = {
        #"model": "llama-3.1-70b-versatile",  # or whichever model you're using
        "model": "llama-3.1-8b-instant",
        "messages": messages,
        "temperature": 0.8,
        "max_tokens": 4096,
        "stream": False,
    }

    response = requests.post(url, headers=headers, json=data)
    response.raise_for_status()
    logger.info(f"Response: {response.json()}")
    return response.json()


    
def update_victim_json(new_infos: Optional[Dict[str, Any]]) -> Optional[Dict[str, Any]]:
    """
    Updates the victim JSON structure with new information by interacting with a language model API.

    Args:
        new_infos (Optional[Dict[str, Any]]): New information to update in the JSON.

    Returns:
        Optional[Dict[str, Any]]: Extracted JSON content if successful, None otherwise.
    """
     
    json_template = st.session_state.get('json_template', {})
    history_infos = st.session_state.get('victim_info', {})
    
    prompt = (
        f"Update the JSON structure: {json.dumps(schema)}\n\n"
        f"with accurate information based on history: {json.dumps(history_infos)}\n\n"
        f"and new information: {json.dumps(new_infos)}\n\n"
        "Output should be a valid JSON file. Fit new information in the main structure of the template "
        f"{json.dumps(list(json_template.keys()))}. Leave blank (e.g., \"\") when there is no information. "
        "Do not overwrite existing information provided, unless it's to update it into something more informative. "
        "NEVER replace existing information with blank values! Ask follow-up questions to keep filling the JSON file, "
        "but in a natural way and prioritizing the most important ones for rescue. Always update emergency_status "
        "[unknown, stable, urgent, very_urgent, critical], but keep it low by default. Output the JSON without any additional text:"
    )

    try:
        response = send_message(prompt)
        content = response.get('choices', [{}])[0].get('message', {}).get('content', '')
        
        # Try to parse the entire content as JSON first
        try:
            json_content = json.loads(content)
            logger.info("Successfully parsed JSON content from response.")
            return json_content
        except json.JSONDecodeError:
            # If parsing the entire content fails, try to extract JSON from code block
            if '```json' in content and '```' in content.split('```json')[-1]:
                json_str = content.split('```json')[1].split('```')[0].strip()
                json_content = json.loads(json_str)
                logger.info("Successfully extracted JSON content from code block in response.")
                return json_content
            else:
                logger.error("No valid JSON found in the response.")
                logger.debug(f"Response content: {content}")
                return None
        
    except requests.exceptions.RequestException as e:
        logger.error(f"An error occurred while sending the message: {e}")
        return None
    except json.JSONDecodeError as e:
        logger.error(f"Failed to decode JSON from response: {e}")
        logger.debug(f"Response content: {content}")
        return None


# validate and upload new informations
def process_and_upload_victim_info(response: Union[str, Dict[str, Any]], schema: Dict[str, Any]):
    """
    Validates the JSON response against a schema and uploads the victim information to the session state.

    Args:
        response (Union[str, Dict[str, Any]]): The response containing the JSON data, either as a string or a dictionary.
        schema (Dict[str, Any]): The JSON schema to validate against.

    Returns:
        Optional[str]: The JSON string if validation and upload are successful, None otherwise.
    """
    try:
        if isinstance(response, str):
            json_str = extract_json_from_response(response)
            if not json_str:
                logger.info("No JSON block found in response string.")
                return None
            json_data = json.loads(json_str)
        elif isinstance(response, dict):
            json_data = response
            json_str = json.dumps(json_data)
        else:
            logger.error(f"Unexpected response type: {type(response)}")
            return None

        # Validate JSON against schema
        jsonschema.validate(instance=json_data, schema=schema)
        logger.info("JSON data validated successfully against the schema.")

        if 'message' in json_data and len(json_data) == 1:
            logger.debug("Returning specific message from JSON data.")
            return json_data['message']
        else:
            # Update session state with the new victim info
            victim_info = json_data
            victim_info['timestamp'] = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            st.session_state['victim_info'] = victim_info
            logger.info("Victim info updated successfully in session state.")
            return json_str

    except jsonschema.ValidationError as ve:
        logger.error(f"JSON validation error: {ve.message}")
        st.warning("Received data does not match the required schema.")
    except json.JSONDecodeError as je:
        logger.error(f"Error decoding JSON: {je.msg}")
        st.warning("Invalid JSON format in response.")
    except Exception as e:
        logger.error(f"An unexpected error occurred: {e}")
        st.warning("An unexpected error occurred while processing the response.")

    return None


def extract_json_from_response(response: str) -> Optional[str]:
    """
    Extracts JSON content from a response string.

    Args:
        response (str): The raw response string containing the JSON block.

    Returns:
        Optional[str]: The extracted JSON string if found, None otherwise.
    """
    try:
        start = response.index('```json') + len('```json')
        end = response.index('```', start)
        json_str = response[start:end].strip()
        logger.debug("Successfully extracted JSON string from response.")
        return json_str
    except ValueError:
        logger.debug("JSON block delimiters not found in the response.")
        return None
    



def validate_json(schema: Dict[str, Any], new_infos: Optional[Dict[str, Any]] = None) -> Optional[Dict[str, Any]]:
    """
    Validates and updates the victim JSON structure based on new information.

    Args:
        schema (Dict[str, Any]): The JSON schema to validate against.
        new_infos (Optional[Dict[str, Any]]): New information to update in the JSON.

    Returns:
        Optional[Dict[str, Any]]: Updated JSON if successful, None otherwise.
    """

    json_template = st.session_state.get('victim_info', {})
    history_infos = st.session_state.get('victim_info', {})
    
    #[{list(json_template.keys())}]. 

    prompt = f'''Update the JSON structure: {schema}
    with accurate information based on history: {history_infos}
    and new information: {new_infos}. 
    Output should be a JSON file. Fit new information in the main structure of the template. Leave blank (e.g., "") 
    when there is no information. Do not overwrite existing information provided, unless its to update it into something more informative. NEVER replace existing information with blank values! 
    Ask follow-up questions to keep filling the JSON file, but in a natural way and prioritizing the most important ones for rescue. 
    Always update emergency_status [unknown, stable, urgent, very_urgent, critical], but keep it low by default. Output:'''
        
    try:
        response = send_message(prompt)
        content = response['choices'][0]['message']['content']
        
        try:
            # Extract JSON from response
            json_content = json.loads(content)
            return json_content
        except json.JSONDecodeError:
            st.error("Failed to parse response as JSON. Raw response:")
            st.text(content)
            return None
    except requests.exceptions.RequestException as e:
        st.error(f"An error occurred: {e}")
        return None

