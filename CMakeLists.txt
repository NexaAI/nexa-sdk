cmake_minimum_required(VERSION 3.16)
project(nexa_gguf)

include(ExternalProject)

# Windows-specific settings
if(WIN32)
    # Windows-specific settings
    add_definitions(-D_CRT_SECURE_NO_WARNINGS)
endif()

# Common OpenMP handling for all platforms
find_package(OpenMP QUIET)
if(NOT OpenMP_FOUND)
    message(STATUS "OpenMP not found - OpenMP support will be disabled")
    set(OpenMP_C_FLAGS "")
    set(OpenMP_CXX_FLAGS "")
    set(OpenMP_EXE_LINKER_FLAGS "")
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

# Create empty file if GGML_CUDA, GGML_METAL or GGML_SYCL is ON
if (GGML_CUDA OR GGML_METAL OR GGML_HIPBLAS OR GGML_VULKAN OR GGML_SYCL)
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

# if(WIN32 AND GGML_CUDA)
#     set(BUILD_PARALLEL_LEVEL "2" CACHE STRING "Number of parallel jobs for MSBuild in CUDA compilation")
#     set(MSBUILD_ARGS "/m:${BUILD_PARALLEL_LEVEL}")
# else()
#     set(MSBUILD_ARGS "")
# endif()
set(BUILD_PARALLEL_LEVEL "4" CACHE STRING "Number of parallel jobs for MSBuild")
set(MSBUILD_ARGS "/m" "${BUILD_PARALLEL_LEVEL}")

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
        -DSD_HIPBLAS=${GGML_HIPBLAS}
        -DSD_VULKAN=${GGML_VULKAN}
        -DSD_SYCL=${GGML_SYCL}
    )

    if (GGML_SYCL)
        list(APPEND SD_CMAKE_ARGS
            -G "Ninja"
            -DGGML_SYCL=ON
            -DCMAKE_C_COMPILER=cl
            -DCMAKE_CXX_COMPILER=icx
        )
    endif()

    ExternalProject_Add(stable_diffusion_project
        SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/dependency/stable-diffusion.cpp
        BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/stable_diffusion_build
        CMAKE_ARGS ${SD_CMAKE_ARGS}  
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

# llama_cpp project
option(LLAMA_BUILD "Build llama.cpp" ON)
if(LLAMA_BUILD)

    set(LLAMA_CMAKE_ARGS
        ${USER_DEFINED_OPTIONS}
        ${COMMON_CMAKE_OPTIONS}
        -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/llama_install
        -DCMAKE_POSITION_INDEPENDENT_CODE=ON
        -DCMAKE_CXX_STANDARD=17
        -DCMAKE_BUILD_TYPE=Release
        -DBUILD_SHARED_LIBS=ON
        -DGGML_CUDA=${GGML_CUDA}
        -DGGML_METAL=${GGML_METAL}
        -DCMAKE_C_FLAGS=${OpenMP_C_FLAGS}
        -DCMAKE_CXX_FLAGS=${OpenMP_CXX_FLAGS}
        -DCMAKE_EXE_LINKER_FLAGS=${OpenMP_EXE_LINKER_FLAGS}
        -DGGML_AVX=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_AVX2=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_FMA=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_F16C=$<IF:$<AND:$<PLATFORM_ID:Darwin>,$<NOT:$<STREQUAL:${CMAKE_SYSTEM_PROCESSOR},arm64>>>,OFF,ON>
        -DGGML_METAL_EMBED_LIBRARY=$<IF:$<PLATFORM_ID:Darwin>,ON,OFF>
    )

    if (GGML_SYCL)
        list(APPEND LLAMA_CMAKE_ARGS
            -G "Ninja"
            -DGGML_SYCL=ON
            -DCMAKE_C_COMPILER=cl
            -DCMAKE_CXX_COMPILER=icx
        )
    endif()

    if(WIN32)
        # Add Windows-specific definitions and flags for llama.cpp
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

# bark_cpp project
option(BARK_BUILD "Build bark.cpp" ON)
if(BARK_BUILD)
    # Filter out HIPBLAS, Vulkan and SYCL options for bark.cpp since it doesn't support them
    set(BARK_CMAKE_OPTIONS ${USER_DEFINED_OPTIONS})
    list(FILTER BARK_CMAKE_OPTIONS EXCLUDE REGEX "GGML_HIPBLAS|GGML_VULKAN|GGML_SYCL")
    
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
            -DGGML_METAL=OFF
            -DBUILD_SHARED_LIBS=ON
            -DBARK_BUILD_EXAMPLES=OFF
        BUILD_ALWAYS 1
        BUILD_COMMAND ${CMAKE_COMMAND} --build . --config Release -- ${MSBUILD_ARGS}
        INSTALL_COMMAND ${CMAKE_COMMAND} --build . --config Release --target install
    )
endif()

macro(install_libraries TARGET DIRS)
    foreach(DIR ${DIRS})
        install(
            DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/${DIR}
            DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib/${TARGET}
            USE_SOURCE_PERMISSIONS
            FILES_MATCHING
            PATTERN "*.dll" 
            PATTERN "*.so"   
            PATTERN "*.dylib"
            PATTERN "cmake*" EXCLUDE
            PATTERN "pkgconfig*" EXCLUDE
            PATTERN "encodec*" EXCLUDE
        )
        install(
            DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}/${DIR}
            DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib/${TARGET}
            USE_SOURCE_PERMISSIONS
            FILES_MATCHING
            PATTERN "*.dll"
            PATTERN "*.so"
            PATTERN "*.dylib"
            PATTERN "cmake*" EXCLUDE
            PATTERN "pkgconfig*" EXCLUDE
            PATTERN "encodec*" EXCLUDE
        )
    endforeach()
endmacro()

if(CMAKE_CONFIGURATION_TYPES)  # Multi-config generators like MSBuild
    set(BUILD_SUBDIR "Release/")
else()  # Single-config generators like Ninja and Make
    set(BUILD_SUBDIR "")
endif()

install_libraries(stable_diffusion "stable_diffusion_build/bin/${BUILD_SUBDIR}")

if (WIN32)
    install_libraries(bark.cpp "bark_build/bin/${BUILD_SUBDIR}")
    install_libraries(bark.cpp "bark_build/${BUILD_SUBDIR}")
    install_libraries(llama "llama_build/bin/${BUILD_SUBDIR}")
else()
    install_libraries(bark.cpp "bark_install/lib/${BUILD_SUBDIR}")
    install_libraries(llama "llama_install/lib/${BUILD_SUBDIR}")
endif()