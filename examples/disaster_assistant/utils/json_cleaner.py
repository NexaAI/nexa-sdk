import json
import logging
from typing import Dict, Any
import re
from jsonschema import validate, ValidationError
import streamlit as st
import datetime 

logger = logging.getLogger(__name__)

def extract_json_from_response(response: str) -> str:
    """
    Extract JSON content from a string that may contain additional text.
    """
    json_match = re.search(r'```json\s*([\s\S]*?)\s*```', response)
    if json_match:
        return json_match.group(1)
    return response

def clean_json_string(json_str: str) -> str:
    """
    Clean the JSON string by removing problematic characters and formatting issues.
    """
    # Remove any leading/trailing whitespace
    json_str = json_str.strip()
    
    # Replace 'None' with 'null' for valid JSON
    json_str = json_str.replace('None', 'null')
    
    # Replace single quotes with double quotes for valid JSON
    json_str = json_str.replace("'", '"')
    
    # Remove any newline characters within string values
    json_str = re.sub(r'(?<!\\)\\n', ' ', json_str)
    
    return json_str

def parse_json_safely(json_str: str) -> Dict[str, Any]:
    """
    Attempt to parse JSON string, handling common issues.
    """
    try:
        return json.loads(json_str)
    except json.JSONDecodeError as e:
        logger.warning(f"Initial JSON parsing failed: {e}")
        
        # Try to fix common issues
        fixed_json = clean_json_string(json_str)
        try:
            return json.loads(fixed_json)
        except json.JSONDecodeError as e:
            logger.error(f"JSON parsing failed after cleaning: {e}")
            raise

def validate_json_schema(data: Dict[str, Any], schema: Dict[str, Any]) -> Dict[str, Any]:
    """
    Validate JSON data against a schema, attempting to fix issues.
    """
    try:
        validate(instance=data, schema=schema)
        return data
    except ValidationError as e:
        logger.warning(f"JSON schema validation failed: {e}")
        
        # Attempt to fix schema issues (this is a simplified example)
        for error in e.context:
            if error.validator == 'required':
                for missing_property in error.validator_value:
                    data[missing_property] = None
        
        # Validate again after fixes
        try:
            validate(instance=data, schema=schema)
            return data
        except ValidationError as e:
            logger.error(f"JSON schema validation failed after attempted fixes: {e}")
            raise

def process_json_response(response: str, schema: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process a JSON response, handling various potential issues.
    """
    try:
        # Extract JSON content
        json_content = extract_json_from_response(response)
        
        # Parse JSON safely
        parsed_json = parse_json_safely(json_content)
        
        # Validate against schema
        validated_json = validate_json_schema(parsed_json, schema)
        
        return validated_json
    except Exception as e:
        logger.error(f"Error processing JSON response: {e}")
        raise

# Usage example
def upload_victim_info(response: str, schema: Dict[str, Any], timestamp=datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")) -> None:
    try:
        processed_json = process_json_response(response, schema)
        st.session_state['victim_info'] = processed_json
        # add a timestamp
        st.session_state['victim_info']['timestamp'] = timestamp
        logger.info("Victim info updated successfully")
    except Exception as e:
        logger.error(f"Failed to update victim info: {e}")
        #st.error(f"An error occurred while processing the response: {e}")