project(cross_share_serv LANGUAGES CXX)

qt5_add_resources(qrc_FILES res.qrc)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)
file(GLOB UI_FILES *.ui)

add_executable(${PROJECT_NAME} WIN32 ${SRCS} ${HEADERS} ${UI_FILES} ${qrc_FILES} resource/ico.rc)

target_link_libraries(${PROJECT_NAME} PRIVATE Qt5::Widgets common common-ui)
install(TARGETS ${PROJECT_NAME} RUNTIME DESTINATION .)
