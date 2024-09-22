cmake_minimum_required(VERSION 3.16)

if (GGML_CUDA OR GGML_METAL)
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

# Project: stable_diffusion_cpp
project(stable_diffusion_cpp)

option(STABLE_DIFFUSION_BUILD "Build stable-diffusion.cpp shared library and install alongside python package" ON)

if (STABLE_DIFFUSION_BUILD)
    set(BUILD_SHARED_LIBS "ON")
    option(SD_BUILD_SHARED_LIBS "" "ON")

    # Building llama
    if (APPLE AND NOT CMAKE_SYSTEM_PROCESSOR MATCHES "arm64")
        # Need to disable these llama.cpp flags on Apple x86_64,
        # otherwise users may encounter invalid instruction errors
        set(GGML_AVX "Off" CACHE BOOL "ggml: enable AVX" FORCE)
        set(GGML_AVX2 "Off" CACHE BOOL "ggml: enable AVX2" FORCE)
        set(GGML_FMA "Off" CACHE BOOL "ggml: enable FMA" FORCE)
        set(GGML_F16C "Off" CACHE BOOL "ggml: enable F16C" FORCE)
    endif()

    add_subdirectory(dependency/stable-diffusion.cpp)
    install(
        TARGETS stable-diffusion
        LIBRARY DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        RUNTIME DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        ARCHIVE DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        FRAMEWORK DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        RESOURCE DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
    )

    message(STATUS "SKBUILD_PLATLIB_DIR: ${SKBUILD_PLATLIB_DIR}")
    # Temporary fix for https://github.com/scikit-build/scikit-build-core/issues/374
    install(
        TARGETS stable-diffusion
        LIBRARY DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        RUNTIME DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        ARCHIVE DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        FRAMEWORK DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        RESOURCE DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
    )
    # Workaround for Windows + CUDA
    if (WIN32)
        install(
            FILES $<TARGET_RUNTIME_DLLS:stable-diffusion>
            DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        )
        install(
            FILES $<TARGET_RUNTIME_DLLS:stable-diffusion>
            DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        )
    endif()
endif()

# Project: llama_cpp
project(llama_cpp)

option(LLAMA_BUILD "Build llama.cpp shared library and install alongside python package" ON)
option(LLAVA_BUILD "Build llava shared library and install alongside python package" ON)

function(llama_cpp_python_install_target target)
    install(
        TARGETS ${target}
        LIBRARY DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        RUNTIME DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        ARCHIVE DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        FRAMEWORK DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        RESOURCE DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
    )
    install(
        TARGETS ${target}
        LIBRARY DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        RUNTIME DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        ARCHIVE DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        FRAMEWORK DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        RESOURCE DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
    )
    set_target_properties(${target} PROPERTIES
        INSTALL_RPATH "$ORIGIN"
        BUILD_WITH_INSTALL_RPATH TRUE
    )
    if(UNIX)
        if(APPLE)
            set_target_properties(${target} PROPERTIES
                INSTALL_RPATH "@loader_path"
                BUILD_WITH_INSTALL_RPATH TRUE
            )
        else()
            set_target_properties(${target} PROPERTIES
                INSTALL_RPATH "$ORIGIN"
                BUILD_WITH_INSTALL_RPATH TRUE
            )
        endif()
    endif()
endfunction()

if (LLAMA_BUILD)
    set(BUILD_SHARED_LIBS "On")

    set(CMAKE_SKIP_BUILD_RPATH FALSE)

    # When building, don't use the install RPATH already
    # (but later on when installing)
    set(CMAKE_BUILD_WITH_INSTALL_RPATH FALSE)

    # Add the automatically determined parts of the RPATH
    # which point to directories outside the build tree to the install RPATH
    set(CMAKE_INSTALL_RPATH_USE_LINK_PATH TRUE)
    set(CMAKE_SKIP_RPATH FALSE)

    # Building llama
    if (APPLE AND NOT CMAKE_SYSTEM_PROCESSOR MATCHES "arm64")
        # Need to disable these llama.cpp flags on Apple x86_64,
        # otherwise users may encounter invalid instruction errors
        set(GGML_AVX "Off" CACHE BOOL "ggml: enable AVX" FORCE)
        set(GGML_AVX2 "Off" CACHE BOOL "ggml: enable AVX2" FORCE)
        set(GGML_FMA "Off" CACHE BOOL "ggml: enable FMA" FORCE)
        set(GGML_F16C "Off" CACHE BOOL "ggml: enable F16C" FORCE)
    endif()

    if (APPLE)
        set(GGML_METAL_EMBED_LIBRARY "On" CACHE BOOL "llama: embed metal library" FORCE)
    endif()

    add_subdirectory(dependency/llama.cpp)
    llama_cpp_python_install_target(llama)
    llama_cpp_python_install_target(ggml_llama)

    # Workaround for Windows + CUDA https://github.com/abetlen/llama-cpp-python/issues/563
    if (WIN32)
        install(
            FILES $<TARGET_RUNTIME_DLLS:llama>
            DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        )
        install(
            FILES $<TARGET_RUNTIME_DLLS:llama>
            DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        )
        install(
            FILES $<TARGET_RUNTIME_DLLS:ggml_llama>
            DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
        )
        install(
            FILES $<TARGET_RUNTIME_DLLS:ggml_llama>
            DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
        )
    endif()

    if (LLAVA_BUILD)
        if (LLAMA_CUBLAS OR LLAMA_CUDA)
            add_compile_definitions(GGML_USE_CUBLAS)
            add_compile_definitions(GGML_USE_CUDA)
        endif()

        if (LLAMA_METAL)
            add_compile_definitions(GGML_USE_METAL)
        endif()

        add_subdirectory(dependency/llama.cpp/examples/llava)
        set_target_properties(llava_shared PROPERTIES OUTPUT_NAME "llava")
        if (WIN32)
            set_target_properties(llava_shared PROPERTIES CUDA_ARCHITECTURES OFF)
        endif()
        llama_cpp_python_install_target(llava_shared)
        if (WIN32)
            install(
                FILES $<TARGET_RUNTIME_DLLS:llava_shared>
                DESTINATION ${CMAKE_CURRENT_SOURCE_DIR}/nexa/gguf/lib
            )
            install(
                FILES $<TARGET_RUNTIME_DLLS:llava_shared>
                DESTINATION ${SKBUILD_PLATLIB_DIR}/nexa/gguf/lib
            )
        endif()
    endif()
endif()

