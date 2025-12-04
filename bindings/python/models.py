"""
Example script demonstrating how to use the model management API.

This script shows how to:
1. Download models from HuggingFace Hub
2. List downloaded models
3. Remove models from local store
4. Use progress callbacks for download tracking
"""

import logging

from nexaai import download_model, get_plugin_list, list_models, remove_model, setup_logging, version

setup_logging(logging.INFO)

repo_id = 'NexaAI/yolov12-npu'


def progress_callback(info: dict) -> None:
    filename = info['filename']
    percentage = info['percentage']
    downloaded_mb = info['downloaded_bytes'] / (1024 * 1024)
    total_mb = info['total_bytes'] / (1024 * 1024) if info['total_bytes'] > 0 else 0

    print(f'{filename}: {percentage:.1f}% ({downloaded_mb:.1f} MB / {total_mb:.1f} MB)')


def main():
    print('NexaAI Model Management Examples')
    print('=' * 50)
    print(f'\nSDK Version: {version()}')

    plugins = get_plugin_list()
    print(f'Available plugins: {", ".join(plugins)}')

    # Download model
    print(f'Downloading model: {repo_id}')
    download_model(
        repo_id=repo_id,
        quant_spec=None,  # Use default quant, or specify like 'Q4_K_M'
        token=None,  # Optional: HuggingFace token for private repos
        progress_callback=progress_callback,
    )
    print(f'✓ Successfully downloaded {repo_id}')

    # List models
    models = list_models()
    print(f'Found {len(models)} model(s) in local store:\n')
    for model in models:
        size_gb = model.size / (1024**3) if model.size > 0 else 0
        print(f'  Repository ID: {model.repo_id}')
        print(f'  Model Name: {model.model_name}')
        print(f'  Model Type: {model.model_type}')
        print(f'  Size: {size_gb:.2f} GB')
        print(f'  Plugin ID: {model.plugin_id or "default"}')
        print(f'  Device ID: {model.device_id or "default"}')
        print()

    # Remove model
    print(f'Removing model: {repo_id}')
    success = remove_model(repo_id)
    if success:
        print(f'✓ Successfully removed {repo_id}')
    else:
        print(f'✗ Failed to remove {repo_id} (model may not exist)')

    print('\n' + '=' * 50)
    print('Examples completed!')


if __name__ == '__main__':
    main()
