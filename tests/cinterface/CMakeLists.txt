set(CMAKE_CXX_STANDARD 17)
find_package(GTest 1.15.2 REQUIRED)

set(TEST_NAME "${LIB_NAME}_test")
add_executable(
        "${TEST_NAME}"
        TestPassword.cpp
        TestPassword.h
        TestStorage.cpp
        TestStorage.h
        TestRest.cpp
        TestRest.h
)

target_link_libraries(
        "${TEST_NAME}" PRIVATE
        "${LIB_NAME}"
        GTest::gmock_main
        GTest::gtest
        GTest::gmock
)

target_compile_definitions(
        "${TEST_NAME}" PRIVATE
        STORAGE_PATH="${CMAKE_SOURCE_DIR}/tests/workdir"
)
