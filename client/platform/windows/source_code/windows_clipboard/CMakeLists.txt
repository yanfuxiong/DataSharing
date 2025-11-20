cmake_minimum_required(VERSION 3.5)

project(windows_clipboard_project LANGUAGES CXX)

set(CMAKE_INCLUDE_CURRENT_DIR ON)

set(CMAKE_AUTOUIC ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

set(VERSION_MAJOR 2)
set(VERSION_MINOR 1)
set(VERSION_PATCH 50)

if (CMAKE_INSTALL_PREFIX_INITIALIZED_TO_DEFAULT)
    set(CMAKE_INSTALL_PREFIX "${CMAKE_BINARY_DIR}/install" CACHE PATH "Install path" FORCE)
endif()
option(INSTALL_TEST_PLUGIN "Whether to install the test plugin" OFF)
message (STATUS "------------------CMAKE_BINARY_DIR: ${CMAKE_BINARY_DIR}")

if (MSVC)
    add_compile_options("$<$<C_COMPILER_ID:MSVC>:/utf-8>")
    add_compile_options("$<$<CXX_COMPILER_ID:MSVC>:/utf-8>")
    add_definitions(-DWIN32_LEAN_AND_MEAN -D_CRT_SECURE_NO_WARNINGS)
    add_compile_options(/wd4309)
    add_compile_options(/wd4307)
endif()

if (CMAKE_BUILD_TYPE STREQUAL "Release")
    add_definitions(-DNDEBUG)
endif()

add_definitions(-DQT_MESSAGELOGCONTEXT -DQT_NO_FLOAT16_OPERATORS)
find_package(Qt5 COMPONENTS LinguistTools Widgets Network Sql Xml PrintSupport Test AxContainer REQUIRED)
find_package(PkgConfig REQUIRED)
pkg_check_modules(LIBARCHIVE REQUIRED IMPORTED_TARGET libarchive)

set(Boost_USE_STATIC_LIBS ON)
if (CMAKE_CXX_COMPILER_ID STREQUAL "GNU")
    find_package(Boost 1.87.0 CONFIG COMPONENTS
    system
    filesystem
    program_options
    iostreams
    regex
    serialization
    timer
    thread
    random
    log
    log_setup
    headers
    date_time
    chrono
    stacktrace
    REQUIRED)
elseif (MSVC)
    find_package(Boost 1.86.0 CONFIG COMPONENTS
    system
    filesystem
    program_options
    iostreams
    regex
    serialization
    timer
    thread
    random
    log
    headers
    date_time
    chrono
    REQUIRED)
endif()

add_definitions(-DBOOST_DATE_TIME_HAS_REENTRANT_STD_FUNCTIONS)
include_directories(${CMAKE_SOURCE_DIR})
include_directories(${CMAKE_SOURCE_DIR}/3rdparty)
include_directories(${CMAKE_SOURCE_DIR}/common)
include_directories(${CMAKE_SOURCE_DIR}/common-ui)

enable_testing()
add_subdirectory(3rdparty)
add_subdirectory(src)
#add_subdirectory(image-paste-progress)
add_subdirectory(status-tips)
add_subdirectory(helper-server)
add_subdirectory(crosstool)
add_subdirectory(common)
add_subdirectory(common-ui)
add_subdirectory(monitor_info_tool)
add_subdirectory(test)
add_subdirectory(test-plugin)

install(DIRECTORY licenses/ DESTINATION licenses PATTERN "*.txt")
