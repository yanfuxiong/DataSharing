project(crosstool LANGUAGES CXX)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)

configure_file(
    ${CMAKE_SOURCE_DIR}/ico.rc.in
    ${CMAKE_CURRENT_SOURCE_DIR}/resource/ico.rc
    @ONLY
)

add_executable(${PROJECT_NAME} WIN32 ${SRCS} ${HEADERS} resource/ico.rc)

target_link_libraries(${PROJECT_NAME} PRIVATE common)
install(TARGETS ${PROJECT_NAME} RUNTIME DESTINATION .)
