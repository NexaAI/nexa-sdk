Environment Setup:
```shell
conda create -n granite-arm64 python==3.13
conda activate granite-arm64
pip install -r requirements.txt
```

Then, provide access token and start the server:
```shell
$env:NEXA_TOKEN=<your_token_here>
nexa serve
```

Run the following commands to start the agent:
```shell
python agent_nexa.py
python gradui_ui.py
```
