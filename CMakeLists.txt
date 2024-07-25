cmake_minimum_required(VERSION 3.5)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

project(vtt)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(GOOGLE_CLOUD_CPP_ROOT "extern/google-cloud-cpp")
list(APPEND CMAKE_PREFIX_PATH "${GOOGLE_CLOUD_CPP_ROOT}/lib/cmake")

find_package(google_cloud_cpp_speech REQUIRED)

add_executable(stream stream.cpp)
target_link_libraries(stream google-cloud-cpp::speech)
target_include_directories(stream PRIVATE ${GOOGLE_CLOUD_CPP_ROOT}/include/)

add_custom_command(TARGET stream POST_BUILD
                   COMMAND ${CMAKE_COMMAND} -E create_symlink
                   ${CMAKE_BINARY_DIR}/compile_commands.json
                   ${CMAKE_SOURCE_DIR}/compile_commands.json)

include(ExternalProject)

ExternalProject_Add(google-cloud-cpp
  SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/extern/google-cloud-cpp
  BINARY_DIR ${CMAKE_CURRENT_BINARY_DIR}/google-cloud-cpp_build
  CONFIGURE_COMMAND cmake -S ${CMAKE_CURRENT_SOURCE_DIR}/extern/google-cloud-cpp -B ${CMAKE_CURRENT_BINARY_DIR}/google-cloud-cpp_build -DBUILD_TESTING=OFF -DGOOGLE_CLOUD_CPP_ENABLE_EXAMPLES=OFF -DGOOGLE_CLOUD_CPP_ENABLE=speech -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_BINARY_DIR}/google-cloud-cpp_install
  BUILD_COMMAND cmake --build ${CMAKE_CURRENT_BINARY_DIR}/google-cloud-cpp_build
  INSTALL_COMMAND cmake --build ${CMAKE_CURRENT_BINARY_DIR}/google-cloud-cpp_build --target install
)

# Make sure our main build depends on Google Cloud C++ build
add_dependencies(stream google-cloud-cpp)
