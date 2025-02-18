cmake_minimum_required(VERSION 3.16)
project(nexa_gguf)

include(ExternalProject)

# Windows-specific settings
if(WIN32)
    add_definitions(
        -D_CRT_SECURE_NO_WARNINGS  # Suppress security warnings
        -DNOMINMAX                 # Prevent Windows.h from defining min/max macros
        -D_WIN32_WINNT=0x0A00      # Target Windows 10 or later
    )
    # Auto-export all symbols in shared libraries on Windows
    set(CMAKE_WINDOWS_EXPORT_ALL_SYMBOLS ON)
endif()

# On Windows with CUDA backend, building the SDK with too many threads lead to out-of-memory crash.  
# To prevent this, limit the maximum number of parallel jobs to 2.
if(WIN32 AND GGML_CUDA)
    set(BUILD_PARALLEL_LEVEL "2" CACHE STRING "Number of parallel jobs for MSBuild in CUDA compilation")
    set(MSBUILD_ARGS "/m:${BUILD_PARALLEL_LEVEL}")
else()
    set(MSBUILD_ARGS "")
endif()

# Find OpenMP and configure settings accordingly
find_package(OpenMP QUIET)
if(NOT OpenMP_FOUND)
    message(STATUS "OpenMP not found - disabling OpenMP support")
    set(OpenMP_C_FLAGS "")
    set(OpenMP_CXX_FLAGS "")
    set(OpenMP_EXE_LINKER_FLAGS "")
endif()

# General project settings
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_CXX_STANDARD 17)

# Function to collect all user-defined CMake options
function(get_all_options output_var)
    # Retrieve all currently defined CMake variables
    get_cmake_property(variables VARIABLES)    
    # Initialize an empty list to store the user-defined options
    set(options)

    foreach(var ${variables})
        # Check if the variable:
        # 1. Starts with an uppercase letter (to exclude internal lowercase variables)
        # 2. Does NOT start with "CMAKE_" (to exclude built-in CMake variables)
        # 3. Does NOT start with "_" (to exclude internal/private variables)
        if(var MATCHES "^[A-Z]" AND NOT var MATCHES "^CMAKE_" AND NOT var MATCHES "^_")
            # Append the variable as a CMake command-line definition (-DVAR=VALUE)
            list(APPEND options "-D${var}=${${var}}")
        endif()
    endforeach()
    
    # Store the collected options in the output variable and make them accessible to the parent scope
    set(${output_var} ${options} PARENT_SCOPE)
endfunction()

# Generate an empty file if GGML_CUDA, GGML_METAL, or GGML_SYCL is enabled.
# This file serves as a runtime flag for the SDK to check if least one GPU backend is supported.
if (GGML_CUDA OR GGML_METAL OR GGML_HIPBLAS OR GGML_VULKAN OR GGML_SYCL)
    set(SOURCE_EMPTY_FILE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib/empty_file.txt")
    add_custom_command(
        OUTPUT ${SOURCE_EMPTY_FILE_PATH}
        COMMAND ${CMAKE_COMMAND} -E touch ${SOURCE_EMPTY_FILE_PATH}
        COMMENT "Generating an empty file in the source folder to indicate that GPU support is enabled."
    )
    set(WHEEL_EMPTY_FILE_PATH "${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib/empty_file.txt")
    add_custom_command(
        OUTPUT ${WHEEL_EMPTY_FILE_PATH}
        COMMAND ${CMAKE_COMMAND} -E touch ${WHEEL_EMPTY_FILE_PATH}
        COMMENT "Generating an empty file in the source folder to indicate that GPU support is enabled."
    )    
    add_custom_target(create_empty_file ALL DEPENDS ${SOURCE_EMPTY_FILE_PATH} ${WHEEL_EMPTY_FILE_PATH})
endif()

# Collect all user-defined options
get_all_options(USER_DEFINED_OPTIONS)

if(APPLE)
    # On macOS, use @loader_path to set the runtime path relative to the executable.
    # This means shared libraries will be searched in the same directory as the executable.
    set(CMAKE_INSTALL_RPATH "@loader_path")
else()
    # On Linux and other platforms, use $ORIGIN to set the runtime path relative to the executable.
    # This ensures that shared libraries are found relative to the executable's location,
    # improving portability and preventing conflicts with system-wide libraries.
    set(CMAKE_INSTALL_RPATH "$ORIGIN")
endif()

# Define common CMake options
set(COMMON_CMAKE_OPTIONS
    -DCMAKE_BUILD_WITH_INSTALL_RPATH=TRUE
    -DCMAKE_INSTALL_RPATH_USE_LINK_PATH=TRUE
    -DCMAKE_SKIP_BUILD_RPATH=FALSE
    -DCMAKE_SKIP_RPATH=FALSE
    -DCMAKE_INSTALL_RPATH=${CMAKE_INSTALL_RPATH}
)

if(WIN32)
    if(CMAKE_SYSTEM_VERSION)
        list(APPEND COMMON_CMAKE_OPTIONS
            -DCMAKE_SYSTEM_VERSION=${CMAKE_SYSTEM_VERSION}
        )
    endif()
    
    # Specifies the Windows SDK version being used in Visual Studio.
    # Ensures that the build system uses the correct SDK version for Windows applications.
    if(CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION)
        list(APPEND COMMON_CMAKE_OPTIONS
            -DCMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION=${CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION}
        )
    endif()
endif()

# stable_diffusion.cpp project
option(STABLE_DIFFUSION_BUILD "Build stable-diffusion.cpp" ON)
if(STABLE_DIFFUSION_BUILD)

    set(SD_CMAKE_ARGS
        ${USER_DEFINED_OPTIONS}
        ${COMMON_CMAKE_OPTIONS}
        -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_install
        -DCMAKE_BUILD_TYPE=Release
        -DCMAKE_POSITION_INDEPENDENT_CODE=ON
        -DCMAKE_CXX_STANDARD=17
        -DSD_BUILD_SHARED_LIBS=ON
        -DBUILD_SHARED_LIBS=ON
        -DSD_METAL=${GGML_METAL}
        -DSD_CUBLAS=${GGML_CUDA}
        -DSD_CUDA=${GGML_CUDA}
        -DSD_SYCL=${GGML_SYCL}
        -DSD_HIPBLAS=${GGML_HIPBLAS}
        -DSD_VULKAN=${GGML_VULKAN}
    )

    ExternalProject_Add(stable_diffusion_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/stable-diffusion.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build
        CMAKE_ARGS ${SD_CMAKE_ARGS}  
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# llama.cpp project
option(LLAMA_BUILD "Build llama.cpp" ON)
if(LLAMA_BUILD)

    set(LLAMA_CMAKE_ARGS
        ${USER_DEFINED_OPTIONS}
        ${COMMON_CMAKE_OPTIONS}
        -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/llama_install
        -DCMAKE_POSITION_INDEPENDENT_CODE=ON
        -DCMAKE_CXX_STANDARD=17
        -DCMAKE_BUILD_TYPE=Release
        -DLLAMA_BUILD_EXAMPLES=ON # Must be set to ON to build omni model libs
        -DLLAMA_BUILD_TESTS=OFF
        -DLLAMA_BUILD_SERVER=OFF
        -DBUILD_SHARED_LIBS=ON
        -DGGML_CUDA=${GGML_CUDA}
        -DGGML_METAL=${GGML_METAL}
        -DGGML_VULKAN=${GGML_VULKAN}
        -DGGML_SYCL=${GGML_SYCL}
        -DCMAKE_C_FLAGS=${OpenMP_C_FLAGS}
        -DCMAKE_CXX_FLAGS=${OpenMP_CXX_FLAGS}
        -DCMAKE_EXE_LINKER_FLAGS=${OpenMP_EXE_LINKER_FLAGS}
        -DGGML_AVX=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_AVX2=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_FMA=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_F16C=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_METAL_EMBED_LIBRARY=$<IF:$<PLATFORM_ID:Darwin>,ON,OFF>
    )

    # Add Windows-specific definitions and flags for llama.cpp
    if(WIN32)
        list(APPEND LLAMA_CMAKE_ARGS
            -DCMAKE_WINDOWS_EXPORT_ALL_SYMBOLS=ON
            -DLLAMA_NATIVE=OFF           # Disable native CPU optimizations on Windows
            -DLLAMA_DISABLE_CXXABI=ON    # Disable cxxabi.h dependency
        )
        # Add compile definition for all targets
        add_compile_definitions(LLAMA_DISABLE_CXXABI)
    endif()

    ExternalProject_Add(llama_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/llama.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/llama_build
        CMAKE_ARGS ${LLAMA_CMAKE_ARGS}
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# bark.cpp project
option(BARK_BUILD "Build bark.cpp" ON)
if(BARK_BUILD)
    # Filter out HIPBLAS, Metal, Vulkan and SYCL options for bark.cpp since it doesn't support them
    set(BARK_CMAKE_OPTIONS ${USER_DEFINED_OPTIONS})
    list(FILTER BARK_CMAKE_OPTIONS EXCLUDE REGEX "GGML_HIPBLAS|GGML_METAL|GGML_VULKAN|GGML_SYCL")

    ExternalProject_Add(bark_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/bark.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/bark_build
        CMAKE_ARGS
            ${BARK_CMAKE_OPTIONS}
            ${COMMON_CMAKE_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/bark_install
            -DCMAKE_BUILD_TYPE=Release
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DGGML_CUDA=${GGML_CUDA}
            -DBUILD_SHARED_LIBS=ON
            -DBARK_BUILD_EXAMPLES=OFF
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# Macro to install shared libraries for a given target.
# There exist two install locations:
# 1. ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib/ : for Python packaging
# 2. ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib/ : for local development (installed via `pip install -e`)
macro(install_libraries TARGET DIR)
    # Define the list of installation destinations
    set(DESTINATIONS
        "${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib/${TARGET}"
        "${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib/${TARGET}"
    )

    foreach(DEST ${DESTINATIONS})
        install(
            DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/${DIR}
            DESTINATION ${DEST}
            USE_SOURCE_PERMISSIONS  # Preserve original file permissions
            FILES_MATCHING  # Match only the following file patterns
            PATTERN "*.dll"  # Windows shared library
            PATTERN "*.so"  # Linux shared library
            PATTERN "*.dylib"  # macOS shared library
            PATTERN "cmake*" EXCLUDE  # Exclude CMake-related folders
            PATTERN "pkgconfig*" EXCLUDE  # Exclude pkgconfig folders
            PATTERN "encodec*" EXCLUDE  # Exclude encodec-related folders
        )
    endforeach()
endmacro()

if(CMAKE_CONFIGURATION_TYPES)  # Multi-config generators like MSBuild
    set(BUILD_SUBDIR "Release/")
else()  # Single-config generators like Ninja and Make
    set(BUILD_SUBDIR "")
endif()


# TODO: Currently, compiled libraries are placed in different paths on different platforms.
#       This inconsistency makes the following commands messy and harder to maintain.
#       Refactor in the future to unify library installation paths across platforms.
install_libraries(stable_diffusion "stable_diffusion_build/bin/${BUILD_SUBDIR}")
if (WIN32)
    install_libraries(bark.cpp "bark_build/bin/${BUILD_SUBDIR}")
    install_libraries(bark.cpp "bark_build/${BUILD_SUBDIR}")
    install_libraries(llama "llama_build/bin/${BUILD_SUBDIR}")
else()
    install_libraries(bark.cpp "bark_install/lib/${BUILD_SUBDIR}")
    install_libraries(llama "llama_install/lib/${BUILD_SUBDIR}")
endif()
