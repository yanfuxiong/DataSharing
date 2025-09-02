project(common LANGUAGES CXX)

file(GLOB SRCS *.cpp *.c)
file(GLOB HEADERS *.h *.hpp)

add_library(${PROJECT_NAME} STATIC ${SRCS} ${HEADERS})
target_link_libraries(${PROJECT_NAME} PUBLIC Qt5::Core Qt5::Sql Qt5::Network Qt5::Xml Qt5::AxContainer ddcci-utils
                    psapi ${Boost_LIBRARIES} Setupapi dxva2 PkgConfig::LIBARCHIVE)
