# setup qairt

set(BUILD_EXAMPLES OFF)
set(GENIEX_BUILD_VLM ON)

# Keep Rust artifact paths short on Windows to avoid MAX_PATH linker failures
# in third-party/tokenizers-cpp when building QAIRT dependencies.
if(WIN32)
    set(TOKENIZERS_CPP_CARGO_TARGET_DIR "${CMAKE_BINARY_DIR}/tok")
endif()

add_subdirectory(
    ../../../third-party/geniex-qairt
    ${CMAKE_BINARY_DIR}/third-party/geniex-qairt
    EXCLUDE_FROM_ALL)

# setup plugin

add_subdirectory(./src)

# setup install
#
# One build, one directory: the plugin negotiates the QNN C API at load time, so there
# are no per-ABI variants to install side by side.
set(QAIRT_PLUGIN_DIR "lib/qairt")

set(QAIRT_LIBS geniex_core geniex-proc geniex-proc-vision geniex-proc-audio geniex_vlm)
foreach(dep IN LISTS QAIRT_LIBS)
    if(TARGET ${dep})
        install(TARGETS ${dep}
            LIBRARY DESTINATION ${QAIRT_PLUGIN_DIR}
            RUNTIME DESTINATION ${QAIRT_PLUGIN_DIR})
    endif()
endforeach()

# Install the bundled HTP runtime from the plugin's build tree. Must land beside
# geniex_core as htp-files/ -- that is where the plugin's resolveHtpPaths() looks.
install(DIRECTORY "$<TARGET_FILE_DIR:geniex_core>/htp-files/"
    DESTINATION ${QAIRT_PLUGIN_DIR}/htp-files
    OPTIONAL)

# Linux only: the QNN HTP stubs link against the unversioned `libcdsprpc.so`, while Debian/Ubuntu
# FastRPC packages ship only `libcdsprpc.so.1`. An empty alias with SONAME libcdsprpc.so that
# depends on libcdsprpc.so.1 bridges the two; the plugin loads it only when the bare name is missing
# (see ensure_cdsprpc_alias() in src/plugin.cpp). Named differently from the SONAME so that it never
# shadows a real libcdsprpc.so found through LD_LIBRARY_PATH.
if(UNIX AND NOT APPLE AND NOT ANDROID)
    # Build-time stand-in for libcdsprpc.so.1 so the alias records the right DT_NEEDED.
    add_library(geniex_cdsprpc_v1 SHARED src/cdsprpc_alias.cpp)
    set_target_properties(geniex_cdsprpc_v1 PROPERTIES
        OUTPUT_NAME cdsprpc VERSION 1.0.0 SOVERSION 1
        LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/cdsprpc_stub")

    add_library(geniex_cdsprpc_alias SHARED src/cdsprpc_alias.cpp)
    set_target_properties(geniex_cdsprpc_alias PROPERTIES
        OUTPUT_NAME geniex_cdsprpc_alias
        NO_SONAME TRUE
        LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/plugins/qairt")
    target_link_options(geniex_cdsprpc_alias PRIVATE
        -Wl,-soname,libcdsprpc.so -Wl,--no-as-needed)
    target_link_libraries(geniex_cdsprpc_alias PRIVATE geniex_cdsprpc_v1)

    install(TARGETS geniex_cdsprpc_alias LIBRARY DESTINATION ${QAIRT_PLUGIN_DIR})
endif()
