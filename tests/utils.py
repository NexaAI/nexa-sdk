import subprocess
import os

def download_model(url, output_dir):
    """
    Download a file from a given URL using curl, if it doesn't already exist.

    Args:
    - url: str, the URL of the file to download.
    - output_dir: str, the directory where the file should be saved.

    Returns:
    - str: The path to the downloaded file.
    """
    file_name = url.split("/")[-1]
    output_path = os.path.join(output_dir, file_name)

    if os.path.exists(output_path):
        print(f"File {file_name} already exists in {output_dir}. Skipping download.")
        return output_path

    try:
        subprocess.run(["curl", url, "--output", output_path], check=True)
        print(f"Downloaded {file_name} to {output_dir}")
    except subprocess.CalledProcessError as e:
        print(f"Failed to download {file_name}: {e}")
        raise

    return output_path