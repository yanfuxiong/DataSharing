cmake_minimum_required(VERSION 3.10)

project(test)

add_definitions(-DQT_WIDGETS_LIB)
link_libraries(Qt5::Widgets Qt5::Core Qt5::Gui Qt5::Network Qt5::Test common)

file(COPY
    "${CMAKE_SOURCE_DIR}/src/resource/cross_share_client.sql.txt"
    DESTINATION
    ${CMAKE_CURRENT_BINARY_DIR}
)

set(TestTarget
test_qstring
test_message_parse
test_sqlite_db
)

foreach(target ${TestTarget})
    message(STATUS "------------------------>>> ADD UNIT TEST : ${target}.cpp")
    add_executable(${target} ${target}.cpp)
    add_test(NAME ${target} COMMAND ${target} WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR})
endforeach()



