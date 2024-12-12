project(test-server LANGUAGES CXX)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)
file(GLOB UI_FILES *.ui)

add_executable(${PROJECT_NAME} WIN32 ${SRCS} ${HEADERS} ${UI_FILES})
target_link_libraries(${PROJECT_NAME} PUBLIC Qt5::Widgets Qt5::Network common-ui)
