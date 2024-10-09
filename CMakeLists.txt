cmake_minimum_required(VERSION 3.16)
project(nexa_gguf)

include(ExternalProject)

set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_CXX_STANDARD 17)

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

# stable_diffusion_cpp project
option(STABLE_DIFFUSION_BUILD "Build stable-diffusion.cpp" ON)
if(STABLE_DIFFUSION_BUILD)
    ExternalProject_Add(stable_diffusion_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/stable-diffusion.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build
        CMAKE_ARGS
            ${USER_DEFINED_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_install
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DSD_BUILD_SHARED_LIBS=ON
            -DBUILD_SHARED_LIBS=ON
            -DSD_METAL=${GGML_METAL}
            -DSD_CUBLAS=${GGML_CUDA}
        BUILD_ALWAYS 1
        INSTALL_COMMAND
            ${CMAKE_COMMAND} --build . --target install &&
            ${CMAKE_COMMAND} -E copy_if_different
                ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build/bin/libstable-diffusion${CMAKE_SHARED_LIBRARY_SUFFIX}
                ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_install/lib/
    )
endif()

# llama_cpp project
option(LLAMA_BUILD "Build llama.cpp" ON)
if(LLAMA_BUILD)
    set(LLAMA_CUDA ${GGML_CUDA})
    set(LLAMA_METAL ${GGML_METAL})
    
    ExternalProject_Add(llama_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/llama.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/llama_build
        CMAKE_ARGS
            ${USER_DEFINED_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/llama_install
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DLLAMA_CUDA=${LLAMA_CUDA}
            -DLLAMA_METAL=${LLAMA_METAL}
        BUILD_ALWAYS 1
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --target install
    )

    if (APPLE AND NOT CMAKE_SYSTEM_PROCESSOR MATCHES "arm64")
        ExternalProject_Add_Step(llama_project disable_optimizations
            COMMAND ${CMAKE_COMMAND} -E echo "Disabling optimizations for x86_64 Apple"
            COMMAND ${CMAKE_COMMAND} -DGGML_AVX=Off -DGGML_AVX2=Off -DGGML_FMA=Off -DGGML_F16C=Off ${CMAKE_CURRENT_BINARY_DIR}/llama_build
            DEPENDEES configure
            DEPENDERS build
        )
    endif()

    if (APPLE)
        ExternalProject_Add_Step(llama_project enable_metal_embed
            COMMAND ${CMAKE_COMMAND} -E echo "Enabling Metal library embedding for Apple"
            COMMAND ${CMAKE_COMMAND} -DGGML_METAL_EMBED_LIBRARY=On ${CMAKE_CURRENT_BINARY_DIR}/llama_build
            DEPENDEES configure
            DEPENDERS build
        )
    endif()
endif()

# bark_cpp project
option(BARK_BUILD "Build bark.cpp" ON)
if(BARK_BUILD)
    ExternalProject_Add(bark_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/bark.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/bark_build
        CMAKE_ARGS
            ${USER_DEFINED_OPTIONS}
            -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/bark_install
            -DCMAKE_POSITION_INDEPENDENT_CODE=ON
            -DCMAKE_CXX_STANDARD=17
            -DGGML_CUDA=${GGML_CUDA}
            -DGGML_METAL=${GGML_METAL}
            -DBUILD_SHARED_LIBS=ON
            -DBARK_BUILD_EXAMPLES=OFF
        BUILD_ALWAYS 1
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --target install
    )
endif()

# Install the built libraries to the final destination
install(
    DIRECTORY
        ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_install/lib/
        ${CMAKE_CURRENT_BINARY_DIR}/llama_install/lib/
        ${CMAKE_CURRENT_BINARY_DIR}/bark_install/lib/
    DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
    USE_SOURCE_PERMISSIONS
    FILES_MATCHING
        PATTERN "*.dll"
        PATTERN "*.so"
        PATTERN "*.dylib"
)

install(
    DIRECTORY
        ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_install/lib/
        ${CMAKE_CURRENT_BINARY_DIR}/llama_install/lib/
        ${CMAKE_CURRENT_BINARY_DIR}/bark_install/lib/
    DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
    USE_SOURCE_PERMISSIONS
    FILES_MATCHING
        PATTERN "*.dll"
        PATTERN "*.so"
        PATTERN "*.dylib"
)