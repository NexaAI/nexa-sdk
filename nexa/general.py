import json
import logging
from pathlib import Path
from typing import Tuple
import shutil
import requests

from nexa.constants import (
    NEXA_API_URL,
    NEXA_LOGO,
    NEXA_MODEL_LIST_PATH,
    NEXA_MODELS_HUB_DIR,
    NEXA_MODELS_HUB_OFFICIAL_DIR,
    NEXA_MODELS_HUB_HF_DIR,
    NEXA_OFFICIAL_BUCKET,
    NEXA_RUN_MODEL_MAP,
    NEXA_TOKEN_PATH,
    NEXA_OFFICIAL_MODELS_TYPE,
)
from nexa.constants import ModelType

def login():
    """
    Login the machine to access the Nexa Model Hub.
    """
    print(NEXA_LOGO)

    # Check if a token already exists
    if NEXA_TOKEN_PATH.exists():
        print(
            "A token is already saved on your machine. Run `nexa-cli whoami` to get more information or `nexa-cli logout` if you want to log out."
        )
        overwrite = (
            input("Do you want to overwrite the existing token? (y/n): ")
            .lower()
            .strip()
        )
        if overwrite != "y":
            print("Login cancelled.")
            return

    token = input("Please enter your Nexa token: ").strip()

    # Validate the token
    if not token:
        print("Error: Token cannot be empty.")
        return
    if not verify_token(token):
        print("Error: Invalid token. Please try again.")
        return

    NEXA_TOKEN_PATH.parent.mkdir(parents=True, exist_ok=True)
    NEXA_TOKEN_PATH.write_text(token)
    print("Success: Login successful.")


def logout():
    """
    Logout and delete the stored token.
    """
    if NEXA_TOKEN_PATH.exists():
        NEXA_TOKEN_PATH.unlink()
        print("Success: Logged out successfully.")
    else:
        print("Info: No active account found.")


def whoami():
    if not NEXA_TOKEN_PATH.exists():
        print("Error: No active account. Please login first.")
        return

    token = NEXA_TOKEN_PATH.read_text().strip()
    user_info = get_user_info(token)
    if user_info:
        print(f"Username: {user_info.get('username', 'N/A')}")
        print(f"Email: {user_info.get('email', 'N/A')}")
    else:
        print("Error: Failed to retrieve user information.")


def verify_token(token):
    """
    Verify the given toke.
    """
    user_info = get_user_info(token)
    return user_info is not None and "username" in user_info


def get_user_info(token):
    endpoint = f"{NEXA_API_URL}/user/verify-token"
    headers = {"Content-Type": "application/json"}
    data = {"token": token}

    try:
        response = requests.post(endpoint, json=data, headers=headers, timeout=10)
        response.raise_for_status()
        return response.json()
    except requests.RequestException as e:
        print(f"Error: An error occurred while verifying the token: {e}")
        return None


def pull_model(model_path, hf = False):
    model_path = NEXA_RUN_MODEL_MAP.get(model_path, model_path)

    try:
        if hf == True:
            result = pull_model_from_hf(model_path)
        else: 
            if is_model_exists(model_path):
                location, run_type = get_model_info(model_path)
                print(f"Model {model_path} already exists at {location}")
                return location, run_type

            if "/" in model_path:
                result = pull_model_from_hub(model_path)
            else:
                result = pull_model_from_official(model_path)

        if result["success"]:
            add_model_to_list(model_path, result["local_path"], result["model_type"], result["run_type"])
            print(f"Successfully pulled model {model_path} to {result['local_path']}, run_type: {result['run_type']}")
            return result["local_path"], result["run_type"]
        else:
            print(f"Failed to pull model {model_path}")
            return None, "NLP"
    except Exception as e:
        logging.error(f"An error occurred while pulling the model: {e}")
        return None, "NLP"


def pull_model_from_hub(model_path):
    NEXA_MODELS_HUB_DIR.mkdir(parents=True, exist_ok=True)

    token = ""
    if Path(NEXA_TOKEN_PATH).exists():
        with open(NEXA_TOKEN_PATH, "r") as file:
            token = file.read().strip()

    try:
        result = get_model_presigned_link(model_path, token)
        run_type = result['type']
        presigned_links = result['presigned_urls']
    except Exception as e:
        print(f"Failed to get download models: {e}")
        return {
            "success": False,
            "local_path": None,
            "model_type": None,
            "run_type": None
        }

    success = True
    local_path = None
    model_type = "undefined"

    # Determine model_type
    for file_path in presigned_links.keys():
        if file_path.endswith(".onnx"):
            model_type = "onnx"
            break
        elif file_path.endswith(".gguf"):
            model_type = "gguf"
            break
        elif file_path.endswith(".bin"):
            model_type = "bin"
            break

    for file_path, presigned_link in presigned_links.items():
        try:
            download_path = NEXA_MODELS_HUB_DIR / file_path
            download_file_with_progress(presigned_link, download_path)

            if local_path is None:
                if model_type == "onnx" or model_type == "bin":
                    local_path = str(download_path.parent)
                elif model_type == "gguf":
                    local_path = str(download_path)
                else:  # undefined
                    local_path = str(download_path.parent)

        except Exception as e:
            print(f"Failed to download {file_path}: {e}")
            success = False

    return {
        "success": success,
        "local_path": local_path,
        "model_type": model_type,
        "run_type": run_type
    }


def pull_model_from_official(model_path):
    NEXA_MODELS_HUB_OFFICIAL_DIR.mkdir(parents=True, exist_ok=True)

    if "onnx" in model_path:
        model_type = "onnx"
    elif "bin" in model_path:
        model_type = "bin"
    else:
        model_type = "gguf"

    run_type = get_run_type_from_model_path(model_path)
    run_type_str = run_type.value if isinstance(run_type, ModelType) else str(run_type)
    success, location = download_model_from_official(model_path, model_type)
    
    return {
        "success": success,
        "local_path": location,
        "model_type": model_type,
        "run_type": run_type_str
    }

def pull_model_from_hf(repo_id):
    repo_id, filename = select_gguf_in_hf_repo(repo_id)
    success, model_path = download_gguf_from_hf(repo_id, filename)

    # For beta version, we only support NLP gguf models
    return {
        "success": success,
        "local_path": model_path,
        "model_type": "gguf",
        "run_type": "NLP"
    }


def get_run_type_from_model_path(model_path):
    model_name, _ = model_path.split(":")
    return NEXA_OFFICIAL_MODELS_TYPE.get(model_name, ModelType.NLP).value


def get_model_presigned_link(full_path, token):
    """
    Get the presigned links for downloading the contents of a model folder.

    Args:
    full_path (str): The full path of the folder to download (e.g., "openai/gpt2:gguf-q2_K").
    token (str, optional): The authentication token. Defaults to None.

    Returns:
    dict: A dictionary containing the model type and presigned URLs.
    """

    url = f"{NEXA_API_URL}/model/download-tag-folder"
    headers = {"Content-Type": "application/json"}

    if token:
        headers["Authorization"] = f"Bearer {token}"

    body = {"full_path": full_path, "need_type": True}

    try:
        response = requests.post(url, headers=headers, json=body)
        response.raise_for_status()
        result = response.json()

        run_type = result.get("type", [])[0] if result.get("type") else None
        presigned_urls = result.get("presigned_urls", {})
        
        return {
            "run_type": run_type,
            "presigned_urls": presigned_urls
        }

    except requests.exceptions.RequestException as e:
        print(f"API request failed: {e}")
        raise


def download_file_with_progress(url, file_path: Path):
    file_path.parent.mkdir(parents=True, exist_ok=True)

    response = requests.get(url, stream=True)
    response.raise_for_status()

    total_size = int(response.headers.get("content-length", 0))
    block_size = 1024

    from tqdm import tqdm

    with open(file_path, "wb") as file, tqdm(
        desc=file_path.name,
        total=total_size,
        unit="iB",
        unit_scale=True,
        unit_divisor=1024,
    ) as progress_bar:
        for data in response.iter_content(block_size):
            size = file.write(data)
            progress_bar.update(size)


def download_model_from_official(model_path, model_type):
    try:
        model_name, model_version = model_path.split(":")
        file_extension = ".zip" if model_type == "onnx" or model_type == "bin" else ".gguf"
        filename = f"{model_version}{file_extension}"

        filepath = f"{model_name}/{filename}"
        print(f"Downloading {filepath}...")
        full_path = NEXA_MODELS_HUB_OFFICIAL_DIR / filepath
        download_url = f"{NEXA_OFFICIAL_BUCKET}{filepath}"

        full_path.parent.mkdir(parents=True, exist_ok=True)
        download_file_with_progress(download_url, full_path)

        if model_type == "onnx" or model_type == "bin":
            unzipped_folder = full_path.parent / model_version
            unzipped_folder.mkdir(parents=True, exist_ok=True)
            import zipfile

            with zipfile.ZipFile(full_path, "r") as zip_ref:
                zip_ref.extractall(unzipped_folder)
            full_path.unlink()
            final_path = unzipped_folder
            print(f"Successfully downloaded and unzipped {filepath} to {final_path}")
        else:
            final_path = full_path
            print(f"Successfully downloaded {filepath} to {final_path}")

        return True, str(final_path)

    except Exception as e:
        print(f"An error occurred while downloading or processing the model: {e}")
        return False, None

def download_gguf_from_hf(repo_id, filename):
    try:
        from huggingface_hub import hf_hub_download
        from pathlib import Path
    except ImportError:
        print("The huggingface-hub package is required. Please install it with `pip install huggingface-hub`.")
        return None

    # Define the local directory to save the model
    local_dir = NEXA_MODELS_HUB_HF_DIR / Path(repo_id)
    local_dir.mkdir(parents=True, exist_ok=True)

    # Download the model
    try:
        model_path = hf_hub_download(
            repo_id=repo_id,
            filename=filename,
            local_dir=local_dir,
            local_files_only=False,
        )
        return True, model_path
    except Exception as e:
        print(f"Failed to download the model: {e}")
        return False, None

def is_model_exists(model_name):
    if not NEXA_MODEL_LIST_PATH.exists():
        return False

    with open(NEXA_MODEL_LIST_PATH, "r") as f:
        model_list = json.load(f)

    return model_name in model_list


def add_model_to_list(model_name, model_location, model_type, run_type):
    NEXA_MODEL_LIST_PATH.parent.mkdir(parents=True, exist_ok=True)

    if NEXA_MODEL_LIST_PATH.exists():
        with open(NEXA_MODEL_LIST_PATH, "r") as f:
            model_list = json.load(f)
    else:
        model_list = {}

    model_list[model_name] = {
        "type": model_type,
        "location": model_location,
        "run_type": run_type
    }

    with open(NEXA_MODEL_LIST_PATH, "w") as f:
        json.dump(model_list, f, indent=2)


def get_model_info(model_name):
    if not NEXA_MODEL_LIST_PATH.exists():
        return None, None

    with open(NEXA_MODEL_LIST_PATH, "r") as f:
        model_list = json.load(f)

    model_data = model_list.get(model_name, {})
    location = model_data.get("location")
    run_type = model_data.get("run_type")

    return location, run_type


def list_models():
    if not NEXA_MODEL_LIST_PATH.exists():
        print("No models found.")
        return
    try:
        with open(NEXA_MODEL_LIST_PATH, "r") as f:
            model_list = json.load(f)

        table = [
            (model_name, model_info["type"], model_info["run_type"], model_info["location"])
            for model_name, model_info in model_list.items()
        ]
        headers = ["Model Name", "Type", "Run Type", "Location"]
        from tabulate import tabulate

        print(
            tabulate(
                table, headers, tablefmt="pretty", colalign=("left", "left", "left", "left")
            )
        )
    except Exception as e:
        print(f"An error occurred while listing the models: {e}")


def remove_model(model_path):
    model_path = NEXA_RUN_MODEL_MAP.get(model_path, model_path)

    if not NEXA_MODEL_LIST_PATH.exists():
        print("No models found.")
        return

    try:
        with open(NEXA_MODEL_LIST_PATH, "r") as f:
            model_list = json.load(f)

        if model_path not in model_list:
            print(f"Model {model_path} not found.")
            return

        model_info = model_list.pop(model_path)
        model_location = model_info['location']
        model_path = Path(model_location)

        # Delete the model files
        if model_path.is_file():
            model_path.unlink()
            print(f"Deleted model file: {model_path}")
        elif model_path.is_dir():
            shutil.rmtree(model_path)
            print(f"Deleted model directory: {model_path}")
        else:
            print(f"Warning: Model location not found: {model_path}")

        # Update the model list file
        with open(NEXA_MODEL_LIST_PATH, "w") as f:
            json.dump(model_list, f, indent=2)

        print(f"Model {model_path} removed from the list.")
        return model_location
    except Exception as e:
        print(f"An error occurred while removing the model: {e}")
        return None
    
def clean():
    if not NEXA_MODELS_HUB_DIR.exists():
        print(f"Nothing to clean.")
        return
    
    # Ask for user confirmation
    confirmation = input(f"This will remove all downloaded models and the model list. Are you sure? (y/N): ").lower().strip()
    
    if confirmation != 'y':
        print("Operation cancelled.")
        return

    try:
        # Remove all contents of the directory
        for item in NEXA_MODELS_HUB_DIR.iterdir():
            if item.is_file():
                item.unlink()
            elif item.is_dir():
                shutil.rmtree(item)
        
        print(f"Successfully removed all contents from {NEXA_MODELS_HUB_DIR}")
    
    except Exception as e:
        print(f"An error occurred while cleaning the directory: {e}")

def select_gguf_in_hf_repo(repo_id: str) -> Tuple[str, str]:
    """
    Lists all files ending with .gguf in the given Hugging Face repository,
    prompts the user to select one, and returns the repo_id and the selected filename.

    Args:
        repo_id (str): The Hugging Face repository ID.

    Returns:
        Tuple[str, str]: A tuple containing the repo_id and the selected filename.
    """
    try:
        from huggingface_hub import HfFileSystem
        from huggingface_hub.utils import validate_repo_id
        from pathlib import Path
    except ImportError:
        print("The huggingface-hub package is required. Please install it with `pip install huggingface-hub`.")
        exit(1)

    validate_repo_id(repo_id)
    hffs = HfFileSystem()

    try:
        files = [
            file["name"] if isinstance(file, dict) else file
            for file in hffs.ls(repo_id, recursive=True)
        ]
    except Exception as e:
        print(f"Error accessing repository '{repo_id}'. Please make sure you have access to the Hugging Face repository first.")
        exit(1)

    # Remove the repo prefix from files
    file_list = []
    for file in files:
        rel_path = Path(file).relative_to(repo_id)
        file_list.append(str(rel_path))

    # Filter for files ending with .gguf
    gguf_files = [file for file in file_list if file.endswith('.gguf')]

    if not gguf_files:
        print(f"No gguf models found in repository '{repo_id}'.")
        exit(1)

    print("Available gguf models in the repository:")
    for i, file in enumerate(gguf_files, 1):
        print(f"{i}. {file}")

    # Prompt the user to select a file
    while True:
        try:
            selected_index = int(input("Please enter the number of the model you want to download and use: "))
            if 1 <= selected_index <= len(gguf_files):
                filename = gguf_files[selected_index - 1]
                print(f"You have selected: {filename}")
                break
            else:
                print(f"Please enter a number between 1 and {len(gguf_files)}")
        except ValueError:
            print("Invalid input. Please enter a number.")

    return repo_id, filename
