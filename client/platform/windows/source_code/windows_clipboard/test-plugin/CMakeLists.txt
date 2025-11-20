project(test-plugin LANGUAGES CXX)

qt5_add_resources(qrc_FILES res.qrc)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)
file(GLOB UI_FILES *.ui)

add_library(${PROJECT_NAME} SHARED ${SRCS} ${HEADERS} ${UI_FILES} ${qrc_FILES})
target_link_libraries(${PROJECT_NAME} PUBLIC Qt5::Widgets Qt5::Network common)
