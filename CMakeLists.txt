cmake_minimum_required(VERSION 3.5)
project(vtt)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(GOOGLE_CLOUD_CPP_ROOT "extern/google-cloud-cpp")
list(APPEND CMAKE_PREFIX_PATH "${GOOGLE_CLOUD_CPP_ROOT}/lib/cmake")

find_package(google_cloud_cpp_speech REQUIRED)

add_executable(stream stream.cpp)
target_link_libraries(stream google-cloud-cpp::speech)
target_include_directories(stream PRIVATE ${GOOGLE_CLOUD_CPP_ROOT}/include/)
