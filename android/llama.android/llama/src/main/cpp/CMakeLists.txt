# Sets the minimum CMake version required for this project.
cmake_minimum_required(VERSION 3.22.1)
set(CMAKE_BUILD_TYPE  Release)
# Declares the project name
project("llama-android")

# Enable FetchContent module
include(FetchContent)

FetchContent_Declare(
        json
        GIT_REPOSITORY https://github.com/nlohmann/json
        GIT_TAG        v3.11.3
)
FetchContent_MakeAvailable(json)

# Declare llama.cpp repository
FetchContent_Declare(
        llama
#        GIT_REPOSITORY https://github.com/ggerganov/llama.cpp
        GIT_REPOSITORY https://github.com/NexaAI/llama.cpp
        GIT_TAG master
)

# Declare llava repository (if needed)
FetchContent_Declare(
        llava
#        GIT_REPOSITORY https://github.com/ggerganov/llama.cpp
        GIT_REPOSITORY https://github.com/NexaAI/llama.cpp
        GIT_TAG master
        SOURCE_SUBDIR examples/llava
)

# Make the content available
FetchContent_MakeAvailable(llama llava)

# Create the main library
add_library(${CMAKE_PROJECT_NAME} SHARED
        llama-android.cpp
        llava-android.cpp
        common.cpp
)


# Link the required libraries
target_link_libraries(${CMAKE_PROJECT_NAME}
        nlohmann_json
        llama
        common
        android
        log
        llava
)