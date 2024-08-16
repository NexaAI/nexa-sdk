cmake_minimum_required(VERSION 3.12)
project("stable-diffusion")

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

if (NOT XCODE AND NOT MSVC AND NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release CACHE STRING "Build type" FORCE)
    set_property(CACHE CMAKE_BUILD_TYPE PROPERTY STRINGS "Debug" "Release" "MinSizeRel" "RelWithDebInfo")
endif()

set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)

if(CMAKE_SOURCE_DIR STREQUAL CMAKE_CURRENT_SOURCE_DIR)
    set(SD_STANDALONE ON)
else()
    set(SD_STANDALONE OFF)
endif()

#
# Option list
#

# general
#option(SD_BUILD_TESTS                "sd: build tests"    ${SD_STANDALONE})
option(SD_BUILD_EXAMPLES             "sd: build examples" ${SD_STANDALONE})
option(SD_CUBLAS                     "sd: cuda backend" OFF)
option(SD_HIPBLAS                    "sd: rocm backend" OFF)
option(SD_METAL                      "sd: metal backend" OFF)
option(SD_FLASH_ATTN                 "sd: use flash attention for x4 less memory usage" OFF)
option(SD_FAST_SOFTMAX               "sd: x1.5 faster softmax, indeterministic (sometimes, same seed don't generate same image), cuda only" OFF)
option(SD_BUILD_SHARED_LIBS          "sd: build shared libs" OFF)
#option(SD_BUILD_SERVER               "sd: build server example"                           ON)

if(SD_CUBLAS)
	message("Use CUBLAS as backend stable-diffusion")
    set(GGML_CUDA ON)
    add_definitions(-DSD_USE_CUBLAS)
endif()

if(SD_METAL)
	message("Use Metal as backend stable-diffusion")
    set(GGML_METAL ON)
    add_definitions(-DSD_USE_METAL)
endif()

if (SD_HIPBLAS)
    message("Use HIPBLAS as backend stable-diffusion")
    set(GGML_HIPBLAS ON)
    add_definitions(-DSD_USE_CUBLAS)
    if(SD_FAST_SOFTMAX)
        set(GGML_CUDA_FAST_SOFTMAX ON)
    endif()
endif ()

if(SD_FLASH_ATTN)
    message("Use Flash Attention for memory optimization")
    add_definitions(-DSD_USE_FLASH_ATTENTION)
endif()

set(SD_LIB stable-diffusion)

file(GLOB SD_LIB_SOURCES 
    "*.h"
    "*.cpp"
    "*.hpp"
)

# we can get only one share lib
if(SD_BUILD_SHARED_LIBS)
    message("Build shared library")
    message(${SD_LIB_SOURCES})
    set(BUILD_SHARED_LIBS OFF)
    add_library(${SD_LIB} SHARED ${SD_LIB_SOURCES})
    add_definitions(-DSD_BUILD_SHARED_LIB)
    target_compile_definitions(${SD_LIB} PRIVATE -DSD_BUILD_DLL)
    set(CMAKE_POSITION_INDEPENDENT_CODE ON)
else()
    message("Build static library")
    set(BUILD_SHARED_LIBS OFF)
    add_library(${SD_LIB} STATIC ${SD_LIB_SOURCES})
endif()


set(CMAKE_POLICY_DEFAULT_CMP0077 NEW)

# see https://github.com/ggerganov/ggml/pull/682
add_definitions(-DGGML_MAX_NAME=128)

# deps
add_subdirectory(ggml)

add_subdirectory(thirdparty)

target_link_libraries(${SD_LIB} PUBLIC ggml zip)
target_include_directories(${SD_LIB} PUBLIC . thirdparty)
target_compile_features(${SD_LIB} PUBLIC cxx_std_11)


if (SD_BUILD_EXAMPLES)
    add_subdirectory(examples)
endif()

