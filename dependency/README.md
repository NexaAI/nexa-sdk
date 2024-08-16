# Add as submodule
```
# Step 1: Clone the main repository recursively to include existing submodules
git clone --recursive https://github.com/NexaAI/nexa-sdk-ggml

# Step 2: Navigate to the cloned repository
cd nexa-sdk-ggml

# Step 3: Add the first submodule
git submodule add https://github.com/ggerganov/llama.cpp dependency/llama.cpp

# Step 4: Add the second submodule
git submodule add https://github.com/leejet/stable-diffusion.cpp dependency/stable-diffusion.cpp

# Step 5: Initialize and update the submodules
git submodule update --init --recursive

# Step 6: Commit the changes to .gitmodules and the added submodules
git add .gitmodules dependency/
git commit -m "Added llama.cpp and stable-diffusion.cpp as submodules"
```