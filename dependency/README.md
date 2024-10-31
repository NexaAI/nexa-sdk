# Add as submodule
Step 1: Clone the main repository recursively to include existing submodules
```shell
git clone --recursive https://github.com/NexaAI/nexa-sdk-ggml
```

Step 2: Add submodule in a specific branch, for example `nexa-audio-lm`
```shell
git submodule add -b nexa-audio-lm https://github.com/NexaAI/llama.cpp dependency/nexa_llama.cpp
```

Step 3: Initialize and update the submodules
```shell
git submodule update --init --recursive
```

Step 4: Commit the changes to .gitmodules and the added submodules
```shell
git add .gitmodules dependency/
git commit -m "Added llama.cpp and stable-diffusion.cpp as submodules"
```

# Update submodules
pull the latest change
```shell
git submodule sync
git submodule update --init --recursive --remote
```