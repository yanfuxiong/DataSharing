project(clipboard LANGUAGES CXX)

file(GLOB SRCS *.cpp *.c MSPaste/*.cpp MSPipeServer/*.cpp pipeserver-asio/*.cpp ddcci/*.cpp)
file(GLOB HEADERS *.h *.hpp MSPaste/*.h MSPipeServer/*.h pipeserver-asio/*.h ddcci/*.h)

add_library(${PROJECT_NAME} STATIC ${SRCS} ${HEADERS})
target_link_libraries(${PROJECT_NAME} PUBLIC gdi32 ws2_32 dxva2)
