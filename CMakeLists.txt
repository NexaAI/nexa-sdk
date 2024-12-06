cmake_minimum_required(VERSION 3.16)
project(nexa_gguf)

include(ExternalProject)

# Platform-specific settings
if(WIN32)
    # Windows-specific settings
    add_definitions(-D_CRT_SECURE_NO_WARNINGS)
    # OpenMP is optional on Windows
    find_package(OpenMP QUIET)
    if(NOT OpenMP_FOUND)
        message(STATUS "OpenMP not found - OpenMP support will be disabled")
        set(OpenMP_C_FLAGS "")
        set(OpenMP_CXX_FLAGS "")
        set(OpenMP_EXE_LINKER_FLAGS "")
    endif()
elseif(APPLE)
    # macOS-specific settings
    find_package(OpenMP QUIET)
    if(NOT OpenMP_FOUND)
        message(STATUS "OpenMP not found - OpenMP support will be disabled")
        set(OpenMP_C_FLAGS "")
        set(OpenMP_CXX_FLAGS "")
        set(OpenMP_EXE_LINKER_FLAGS "")
    endif()
else()
    # Linux and other Unix systems
    find_package(OpenMP REQUIRED)
endif()

set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_CXX_STANDARD 17)

# Windows-specific configurations
if(WIN32)
    add_definitions(-D_CRT_SECURE_NO_WARNINGS)
    add_definitions(-DNOMINMAX)
    add_definitions(-D_WIN32_WINNT=0x0A00)  # Target Windows 10 or later
    set(CMAKE_WINDOWS_EXPORT_ALL_SYMBOLS ON)
endif()

# Function to collect all user-defined options
function(get_all_options output_var)
    get_cmake_property(variables VARIABLES)
    set(options)
    foreach(var ${variables})
        if(var MATCHES "^[A-Z]" AND NOT var MATCHES "^CMAKE_" AND NOT var MATCHES "^_")
            list(APPEND options "-D${var}=${${var}}")
        endif()
    endforeach()
    set(${output_var} ${options} PARENT_SCOPE)
endfunction()

# Create empty file if GGML_CUDA or GGML_METAL is ON
if (GGML_CUDA OR GGML_METAL OR GGML_HIPBLAS OR GGML_VULKAN)
    set(SOURCE_EMPTY_FILE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib/empty_file.txt")
    add_custom_command(
        OUTPUT ${SOURCE_EMPTY_FILE_PATH}
        COMMAND ${CMAKE_COMMAND} -E touch ${SOURCE_EMPTY_FILE_PATH}
        COMMENT "Creating an empty file to source folder because gpu option is ON"
    )
    set(WHEEL_EMPTY_FILE_PATH "${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib/empty_file.txt")
    add_custom_command(
        OUTPUT ${WHEEL_EMPTY_FILE_PATH}
        COMMAND ${CMAKE_COMMAND} -E touch ${WHEEL_EMPTY_FILE_PATH}
        COMMENT "Creating an empty file to lib folder because gpu option is ON"
    )    
    add_custom_target(create_empty_file ALL DEPENDS ${SOURCE_EMPTY_FILE_PATH} ${WHEEL_EMPTY_FILE_PATH})
endif()

if(WIN32 AND GGML_CUDA)
    set(BUILD_PARALLEL_LEVEL "2" CACHE STRING "Number of parallel jobs for MSBuild in CUDA compilation")
    set(MSBUILD_ARGS "/m:${BUILD_PARALLEL_LEVEL}")
else()
    set(MSBUILD_ARGS "")
endif()


# Function to set up common installation paths
function(setup_install_paths target install_dir)
    install(
        TARGETS ${target}
        LIBRARY DESTINATION ${install_dir}
        RUNTIME DESTINATION ${install_dir}
        ARCHIVE DESTINATION ${install_dir}
    )
endfunction()

# Collect all user-defined options
get_all_options(USER_DEFINED_OPTIONS)

if(APPLE)
    set(CMAKE_INSTALL_RPATH "@loader_path")
else()
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
    
    if(CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION)
        list(APPEND COMMON_CMAKE_OPTIONS
            -DCMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION=${CMAKE_VS_WINDOWS_TARGET_PLATFORM_VERSION}
        )
    endif()
endif()

# stable_diffusion_cpp project
option(STABLE_DIFFUSION_BUILD "Build stable-diffusion.cpp" ON)
if(STABLE_DIFFUSION_BUILD)
    ExternalProject_Add(stable_diffusion_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/stable-diffusion.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build
        CMAKE_ARGS
            ${USER_DEFINED_OPTIONS}
            ${COMMON_CMAKE_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_install
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DSD_BUILD_SHARED_LIBS=ON
            -DBUILD_SHARED_LIBS=ON
            -DSD_METAL=${GGML_METAL}
            -DSD_CUBLAS=${GGML_CUDA}
            -DSD_HIPBLAS=${GGML_HIPBLAS}
            -DSD_VULKAN=${GGML_VULKAN}
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# llama_cpp project
option(LLAMA_BUILD "Build llama.cpp" ON)
if(LLAMA_BUILD)
    set(LLAMA_CUDA ${GGML_CUDA})
    set(LLAMA_METAL ${GGML_METAL})

    if(WIN32)
        # Add Windows-specific definitions and flags for llama.cpp
        list(APPEND COMMON_CMAKE_OPTIONS
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
        CMAKE_ARGS
            ${USER_DEFINED_OPTIONS}
            ${COMMON_CMAKE_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/llama_install
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DBUILD_SHARED_LIBS=ON
            -DLLAMA_CUDA=${LLAMA_CUDA}
            -DLLAMA_METAL=${LLAMA_METAL}
            -DCMAKE_C_FLAGS=${OpenMP_C_FLAGS}
            -DCMAKE_CXX_FLAGS=${OpenMP_CXX_FLAGS}
            -DCMAKE_EXE_LINKER_FLAGS=${OpenMP_EXE_LINKER_FLAGS}
            -DGGML_AVX=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
            -DGGML_AVX2=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
            -DGGML_FMA=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
            -DGGML_F16C=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
            -DGGML_METAL_EMBED_LIBRARY=$<IF:$<PLATFORM_ID:Darwin>,ON,OFF>
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# bark_cpp project
# Temporarily disabled since version v0.0.9.3
option(BARK_BUILD "Build bark.cpp" OFF)
if(BARK_BUILD)
    # Filter out HIPBLAS and Vulkan options for bark.cpp since it doesn't support them
    set(BARK_CMAKE_OPTIONS ${USER_DEFINED_OPTIONS})
    list(FILTER BARK_CMAKE_OPTIONS EXCLUDE REGEX "GGML_HIPBLAS|GGML_VULKAN")
    
    ExternalProject_Add(bark_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/bark.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/bark_build
        CMAKE_ARGS
            ${USER_DEFINED_OPTIONS}
            ${COMMON_CMAKE_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/bark_install
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DGGML_CUDA=${GGML_CUDA}
            -DGGML_METAL=OFF
            -DBUILD_SHARED_LIBS=ON
            -DBARK_BUILD_EXAMPLES=OFF
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# Install the built libraries to the final destination
if(WIN32)
    install(
        DIRECTORY
            ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build/bin/Release/
            ${CMAKE_CURRENT_BINARY_DIR}/llama_build/bin/Release/
            ${CMAKE_CURRENT_BINARY_DIR}/bark_build/bin/Release/
            ${CMAKE_CURRENT_BINARY_DIR}/bark_build/Release/
        DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        USE_SOURCE_PERMISSIONS
        FILES_MATCHING
            PATTERN "*.dll"
    )

    install(
        DIRECTORY
            ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build/bin/Release/
            ${CMAKE_CURRENT_BINARY_DIR}/llama_build/bin/Release/
            ${CMAKE_CURRENT_BINARY_DIR}/bark_build/bin/Release/
            ${CMAKE_CURRENT_BINARY_DIR}/bark_build/Release/
        DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        USE_SOURCE_PERMISSIONS
        FILES_MATCHING
            PATTERN "*.dll"
    )
else()
    install(
        DIRECTORY
            ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build/bin/
            ${CMAKE_CURRENT_BINARY_DIR}/llama_install/lib/
            ${CMAKE_CURRENT_BINARY_DIR}/bark_install/lib/
        DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        USE_SOURCE_PERMISSIONS
        FILES_MATCHING
            PATTERN "*.so"
            PATTERN "*.dylib"
    )

    install(
        DIRECTORY
            ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build/bin/
            ${CMAKE_CURRENT_BINARY_DIR}/llama_install/lib/
            ${CMAKE_CURRENT_BINARY_DIR}/bark_install/lib/
        DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        USE_SOURCE_PERMISSIONS
        FILES_MATCHING
            PATTERN "*.so"
            PATTERN "*.dylib"
    )
endif()