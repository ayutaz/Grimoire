cmake_minimum_required(VERSION 3.20)
project(grimoire VERSION 0.1.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 最適化フラグ
if(CMAKE_BUILD_TYPE STREQUAL "Release")
    if(MSVC)
        set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /O2 /GL /LTCG")
    else()
        set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O3 -march=native -flto")
    endif()
endif()

# vcpkgを使用
if(DEFINED ENV{VCPKG_ROOT})
    set(CMAKE_TOOLCHAIN_FILE "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake")
endif()

# OpenCV
find_package(OpenCV REQUIRED)

# CLI11 (コマンドライン解析)
find_package(CLI11 CONFIG REQUIRED)

# ソースファイル
set(SOURCES
    src/main.cpp
    src/detector.cpp
    src/parser.cpp
    src/compiler.cpp
)

add_executable(grimoire ${SOURCES})

target_link_libraries(grimoire PRIVATE
    ${OpenCV_LIBS}
    CLI11::CLI11
)

# 静的リンク設定
if(WIN32)
    set_property(TARGET grimoire PROPERTY MSVC_RUNTIME_LIBRARY "MultiThreaded")
elseif(UNIX AND NOT APPLE)
    target_link_options(grimoire PRIVATE -static-libgcc -static-libstdc++)
endif()

# インストール設定
install(TARGETS grimoire DESTINATION bin)