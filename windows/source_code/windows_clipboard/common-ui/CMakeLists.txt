project(common-ui LANGUAGES CXX)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)

#file(COPY
#    "${CMAKE_SOURCE_DIR}/config/cnf"
#    DESTINATION
#    ${CMAKE_CURRENT_BINARY_DIR}
#)

add_library(${PROJECT_NAME} STATIC ${SRCS} ${HEADERS})

target_link_libraries(${PROJECT_NAME} PUBLIC Qt5::Widgets common)
