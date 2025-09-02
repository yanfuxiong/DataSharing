project(ddcci-utils LANGUAGES CXX)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)

add_library(${PROJECT_NAME} STATIC ${SRCS} ${HEADERS})
target_link_libraries(${PROJECT_NAME} PUBLIC psapi Setupapi dxva2 Qt5::Core)

