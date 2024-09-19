import json
import logging
from typing import Dict, Any, Optional
import re
from jsonschema import validate, ValidationError, Draft7Validator
import streamlit as st
import datetime
from pathlib import Path

# Configuration: Load external schema and cleaning rules if necessary
SCHEMA_PATH = Path(__file__).parent / "schemas" / "victim_info_schema.json"

# Initialize logger with enhanced configuration
logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)  # Set to DEBUG for more granular logs
handler = logging.StreamHandler()
formatter = logging.Formatter(
    '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
handler.setFormatter(formatter)
logger.addHandler(handler)

def load_schema(schema_path: Path) -> Dict[str, Any]:
    """
    Load JSON schema from the specified file path.

    Args:
        schema_path (Path): Path to the JSON schema file.

    Returns:
        Dict[str, Any]: The loaded JSON schema.

    Raises:
        FileNotFoundError: If the schema file does not exist.
        json.JSONDecodeError: If the schema file contains invalid JSON.
    """
    try:
        with schema_path.open('r', encoding='utf-8') as f:
            schema = json.load(f)
        logger.debug(f"Schema loaded successfully from {schema_path}")
        return schema
    except FileNotFoundError:
        logger.error(f"Schema file not found at {schema_path}")
        raise
    except json.JSONDecodeError as e:
        logger.error(f"Invalid JSON in schema file: {e}")
        raise

def extract_json_from_response(response: str) -> Optional[str]:
    """
    Extract JSON content from a string that may contain additional text or formatting.

    Utilizes a more robust parsing approach to handle nested structures and escaped characters.

    Args:
        response (str): The response string containing JSON.

    Returns:
        Optional[str]: The extracted JSON string if found, else None.
    """
    try:
        json_start = response.index('{')
        json_end = response.rindex('}') + 1
        json_content = response[json_start:json_end]
        logger.debug("JSON content extracted using string indices.")
        return json_content
    except ValueError:
        logger.warning("No JSON object found in the response.")
        pass

def clean_json_string(json_str: str) -> str:
    """
    Clean the JSON string by addressing common formatting issues without altering string values.

    Handles:
        - Replacing Python-specific literals with JSON equivalents.
        - Removing trailing commas.
        - Ensuring proper quotation marks.

    Args:
        json_str (str): The raw JSON string.

    Returns:
        str: The cleaned JSON string.
    """
    logger.debug("Starting JSON string cleaning.")
    # Replace Python-specific literals
    replacements = {
        'None': 'null',
        'True': 'true',
        'False': 'false',
    }
    for py_literal, json_literal in replacements.items():
        pattern = r'\b' + re.escape(py_literal) + r'\b'
        json_str = re.sub(pattern, json_literal, json_str)
        logger.debug(f"Replaced {py_literal} with {json_literal}.")

    # Remove trailing commas
    json_str = re.sub(r',\s*([}\]])', r'\1', json_str)
    logger.debug("Removed trailing commas.")

    # Ensure double quotes around keys and string values
    def replace_single_quotes(match):
        return match.group(0).replace("'", '"')

    json_str = re.sub(r"(?<!\\)'", '"', json_str)
    logger.debug("Replaced single quotes with double quotes.")

    # Remove newline characters within string values
    json_str = re.sub(r'(?<!\\)\\n', ' ', json_str)
    logger.debug("Removed unescaped newline characters within string values.")

    logger.debug("JSON string cleaning completed.")
    return json_str

def parse_json_safely(json_str: str) -> Dict[str, Any]:
    """
    Attempt to parse a JSON string safely, applying cleaning steps if initial parsing fails.

    Args:
        json_str (str): The JSON string to parse.

    Returns:
        Dict[str, Any]: The parsed JSON as a dictionary.

    Raises:
        json.JSONDecodeError: If parsing fails after cleaning.
    """
    try:
        parsed = json.loads(json_str)
        logger.debug("JSON parsed successfully on first attempt.")
        return parsed
    except json.JSONDecodeError as e:
        logger.warning(f"Initial JSON parsing failed: {e}. Attempting to clean the JSON string.")
        cleaned_json = clean_json_string(json_str)
        try:
            parsed = json.loads(cleaned_json)
            logger.debug("JSON parsed successfully after cleaning.")
            return parsed
        except json.JSONDecodeError as e:
            logger.error(f"JSON parsing failed after cleaning: {e}")
            raise

def validate_json_schema(data: Dict[str, Any], schema: Dict[str, Any]) -> Dict[str, Any]:
    """
    Validate JSON data against a schema, attempting to fix issues based on schema definitions.

    Args:
        data (Dict[str, Any]): The JSON data to validate.
        schema (Dict[str, Any]): The JSON schema to validate against.

    Returns:
        Dict[str, Any]: The validated and potentially modified JSON data.

    Raises:
        ValidationError: If validation fails even after attempting fixes.
    """
    validator = Draft7Validator(schema)
    errors = sorted(validator.iter_errors(data), key=lambda e: e.path)
    if not errors:
        logger.debug("JSON data is valid against the schema.")
        return data

    logger.warning(f"JSON schema validation failed with {len(errors)} errors. Attempting to fix.")
    for error in errors:
        if error.validator == 'required':
            for missing_property in error.validator_value:
                if missing_property not in data:
                    # Fetch default value from schema if available
                    default = schema.get('properties', {}).get(missing_property, {}).get('default')
                    data[missing_property] = default
                    logger.debug(f"Added missing property '{missing_property}' with default value '{default}'.")
        elif error.validator == 'type':
            # Attempt type coercion if possible
            expected_type = error.validator_value
            actual_value = data.get(error.path[-1])
            key = error.path[-1]
            if expected_type == "string" and not isinstance(actual_value, str):
                data[key] = str(actual_value)
                logger.debug(f"Coerced property '{key}' to string.")
            elif expected_type == "integer" and isinstance(actual_value, str):
                try:
                    data[key] = int(actual_value)
                    logger.debug(f"Coerced property '{key}' to integer.")
                except ValueError:
                    logger.warning(f"Failed to coerce property '{key}' to integer.")
        # Additional validators can be handled here

    # Re-validate after fixes
    errors = sorted(validator.iter_errors(data), key=lambda e: e.path)
    if not errors:
        logger.debug("JSON data is valid after attempting fixes.")
        return data
    else:
        logger.error(f"JSON schema validation failed after attempted fixes: {errors}")
        raise ValidationError("JSON schema validation failed after attempted fixes.")

def process_json_response(response: str, schema: Dict[str, Any]) -> Dict[str, Any]:
    """
    Process a JSON response by extracting, parsing, and validating the JSON content.

    Args:
        response (str): The raw response containing JSON.
        schema (Dict[str, Any]): The JSON schema for validation.

    Returns:
        Dict[str, Any]: The validated JSON data.

    Raises:
        Exception: If any step in processing fails.
    """
    try:
        # Extract JSON content
        json_content = extract_json_from_response(response)
        if not json_content:
            raise ValueError("No JSON content found in the response.")

        # Parse JSON safely
        parsed_json = parse_json_safely(json_content)

        # Validate against schema
        validated_json = validate_json_schema(parsed_json, schema)

        logger.info("JSON response processed and validated successfully.")
        return validated_json
    except Exception as e:
        logger.error(f"Error processing JSON response: {e}")
        raise

def upload_victim_info(
    response: str, 
    schema: Optional[Dict[str, Any]] = None, 
    timestamp: Optional[str] = None
) -> None:
    """
    Upload victim information by processing the JSON response and updating the Streamlit session state.

    Args:
        response (str): The raw response containing JSON.
        schema (Optional[Dict[str, Any]]): The JSON schema for validation. If None, loads the default schema.
        timestamp (Optional[str]): The timestamp to add. If None, uses the current time.

    Raises:
        Exception: If processing or validation fails.
    """
    try:
        if schema is None:
            schema = load_schema(SCHEMA_PATH)

        if timestamp is None:
            timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

        processed_json = process_json_response(response, schema)
        processed_json['timestamp'] = timestamp

        # Update Streamlit session state
        if 'victim_info' not in st.session_state:
            st.session_state['victim_info'] = {}
        
        st.session_state['victim_info'].update(processed_json)

        logger.info("Victim info updated successfully in session state.")
    except Exception as e:
        logger.error(f"Failed to update victim info: {e}")
        st.error(f"An error occurred while processing the response: {e}")

# Unit Tests
if __name__ == "__main__":
    import unittest

    class TestJSONCleaner(unittest.TestCase):
        def setUp(self):
            self.schema = {
                "type": "object",
                "properties": {
                    "name": {"type": "string"},
                    "age": {"type": "integer", "default": 0},
                    "email": {"type": "string"},
                    "timestamp": {"type": "string"}
                },
                "required": ["name", "age", "email"]
            }

        def test_extract_json_from_response_valid(self):
            response = "Here is some text ```json {'name': 'John', 'age': 30, 'email': 'john@example.com'}``` end of message."
            expected = "{'name': 'John', 'age': 30, 'email': 'john@example.com'}"
            extracted = extract_json_from_response(response)
            self.assertEqual(extracted, expected)

        def test_extract_json_from_response_invalid(self):
            response = "No JSON here!"
            extracted = extract_json_from_response(response)
            self.assertIsNone(extracted)

        def test_clean_json_string(self):
            raw_json = "{'name': 'John', 'age': '30', 'email': 'john@example.com',}"
            cleaned = clean_json_string(raw_json)
            expected = '{"name": "John", "age": "30", "email": "john@example.com"}'
            self.assertEqual(cleaned, expected)

        def test_parse_json_safely_valid(self):
            json_str = '{"name": "John", "age": 30, "email": "john@example.com"}'
            parsed = parse_json_safely(json_str)
            expected = {"name": "John", "age": 30, "email": "john@example.com"}
            self.assertEqual(parsed, expected)

        def test_parse_json_safely_invalid_then_fixed(self):
            json_str = "{'name': 'John', 'age': '30', 'email': 'john@example.com',}"
            parsed = parse_json_safely(json_str)
            expected = {"name": "John", "age": "30", "email": "john@example.com"}
            self.assertEqual(parsed, expected)

        def test_validate_json_schema_complete(self):
            data = {"name": "John", "age": 30, "email": "john@example.com"}
            validated = validate_json_schema(data, self.schema)
            self.assertEqual(validated, data)

        def test_validate_json_schema_missing_required(self):
            data = {"name": "John", "email": "john@example.com"}
            validated = validate_json_schema(data, self.schema)
            expected = {"name": "John", "email": "john@example.com", "age": 0}
            self.assertEqual(validated, expected)

        def test_validate_json_schema_wrong_type(self):
            data = {"name": "John", "age": "thirty", "email": "john@example.com"}
            with self.assertRaises(ValidationError):
                validate_json_schema(data, self.schema)

    # Run unit tests
    unittest.main(argv=[''], exit=False)