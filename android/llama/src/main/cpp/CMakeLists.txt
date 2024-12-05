cmake_minimum_required(VERSION 3.22.1)
set(CMAKE_BUILD_TYPE Release)

project("llama-android")

include(FetchContent)

FetchContent_Declare(
        json
        GIT_REPOSITORY https://github.com/nlohmann/json
        GIT_TAG v3.11.3
)

# set(SOURCE_BASE_DIR /nexa-ai/llama.cpp)
##### from local #####
# FetchContent_Declare(
#         llama
#         SOURCE_DIR ${SOURCE_BASE_DIR}
# )
# FetchContent_Declare(
#         llava
#         SOURCE_DIR ${SOURCE_BASE_DIR}/examples/llava
# )
# FetchContent_Declare(
#         omni_vlm
#         SOURCE_DIR ${SOURCE_BASE_DIR}/examples/omni-vlm
# )
# FetchContent_Declare(
#         omni_audio
#         SOURCE_DIR ${SOURCE_BASE_DIR}/examples/nexa-omni-audio
# )

##### from remote #####

FetchContent_Declare(
       llama
       GIT_REPOSITORY https://github.com/NexaAI/llama.cpp.git
       GIT_TAG master
)

FetchContent_Declare(
       llava
       GIT_REPOSITORY https://github.com/NexaAI/llama.cpp.git
       GIT_TAG master
       SOURCE_SUBDIR examples/llava
)
FetchContent_Declare(
       omni_vlm
       GIT_REPOSITORY https://github.com/NexaAI/llama.cpp.git
       GIT_TAG master
       SOURCE_SUBDIR examples/omni-vlm
)
FetchContent_Declare(
       omni_audio
       GIT_REPOSITORY https://github.com/NexaAI/llama.cpp.git
       GIT_TAG T/dev
       SOURCE_SUBDIR examples/nexa-omni-audio
)

FetchContent_MakeAvailable(json llama llava omni_vlm omni_audio)

add_library(${CMAKE_PROJECT_NAME} SHARED
        llama-android.cpp
        common.cpp
        llava-android.cpp
)
target_link_libraries(${CMAKE_PROJECT_NAME}
        nlohmann_json
        llama
        common
        android
        log
        llava
)


##### vision  #####
add_library(omni-android SHARED
        common.cpp
        omni-android.cpp
)
target_link_libraries(omni-android
        nlohmann_json
        llama
        common
        android
        log
        omni_vlm
)


##### audio  #####
add_library(audio-android SHARED
        audio-android.cpp
        common.cpp
)
target_link_libraries(audio-android
        nlohmann_json
        llama
        common
        omni_audio
        android
        log
)

