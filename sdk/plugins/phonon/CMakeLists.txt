# Phonon is an explicitly supplied, reviewed dependency. Never fetch or execute
# source code from a model bundle. See README.md for the tested source revision.
set(GENIEX_PHONON_SOURCE "" CACHE PATH "Gradium xn-ptts source with Windows QNN support")
if(NOT EXISTS "${GENIEX_PHONON_SOURCE}/qnn/runner/phonon.hpp")
    message(FATAL_ERROR "GENIEX_PHONON_SOURCE must point to the Gradium xn-ptts source tree")
endif()
add_subdirectory("${GENIEX_PHONON_SOURCE}/qnn/runner" "${CMAKE_CURRENT_BINARY_DIR}/engine" EXCLUDE_FROM_ALL)
add_library(geniex_phonon SHARED plugin.cpp)
target_include_directories(geniex_phonon PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/../../include")
target_link_libraries(geniex_phonon PRIVATE geniex phonon)
target_compile_definitions(geniex_phonon PRIVATE GENIEX_SHARED)
set_target_properties(geniex_phonon PROPERTIES OUTPUT_NAME geniex_plugin)
install(TARGETS geniex_phonon LIBRARY DESTINATION lib/phonon_qnn RUNTIME DESTINATION lib/phonon_qnn)
install(FILES LICENSE-xn-ptts-MIT DESTINATION licenses/phonon_qnn)
include(FetchContent)
FetchContent_GetProperties(nlohmann_json)
install(FILES "${nlohmann_json_SOURCE_DIR}/LICENSE.MIT"
    DESTINATION licenses/phonon_qnn RENAME LICENSE-nlohmann-json-MIT)
