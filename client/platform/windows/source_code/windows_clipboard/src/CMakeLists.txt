project(windows_clipboard LANGUAGES CXX)

qt5_add_resources(qrc_FILES res.qrc)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)
file(GLOB UI_FILES *.ui)

file(COPY
    ${CMAKE_SOURCE_DIR}/licenses/README.txt
    ${CMAKE_SOURCE_DIR}/licenses/Qt_LGPLv3.txt
    ${CMAKE_SOURCE_DIR}/licenses/Boost_LICENSE_1_0.txt
    ${CMAKE_SOURCE_DIR}/licenses/nlohmann_json_MIT.txt
    DESTINATION
    ${CMAKE_CURRENT_BINARY_DIR}/licenses)

add_executable(${PROJECT_NAME} WIN32 ${SRCS} ${HEADERS} ${UI_FILES} ${qrc_FILES} resource/ico.rc)

target_link_libraries(${PROJECT_NAME} PRIVATE Qt5::Widgets Qt5::PrintSupport common common-ui)
install(TARGETS ${PROJECT_NAME} RUNTIME DESTINATION .)
