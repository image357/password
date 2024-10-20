#include "TestStorage.h"
#include <password/cinterface.h>

#include <filesystem>

void TestStorage::SetUp() {
    Test::SetUp();
    auto ret_loglevel = CPWD__LogLevel(CPWD__LevelDebug);
    EXPECT_EQ(ret_loglevel, 0);
}

void TestStorage::TearDown() {
    Test::TearDown();
}

TEST_F(TestStorage, NormalizeId) {
    // success
    char buffer[256];
    int ret = CPWD__NormalizeId("//TEST", buffer, 256);
    ASSERT_EQ(ret, 0);
    ASSERT_STREQ(buffer, "test");

    // fail: nullptr
    ret = CPWD__NormalizeId("//TEST", nullptr, 256);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__NormalizeId("//TEST", buffer, 4);
    ASSERT_EQ(ret, -1);

    // success: size
    ret = CPWD__NormalizeId("//TEST", buffer, 5);
    ASSERT_EQ(ret, 0);
}

TEST_F(TestStorage, StorePath) {
    // set
    CPWD__SetStorePath("test");

    // success
    char buffer[256];
    int ret = CPWD__GetStorePath(buffer, 256);
    ASSERT_EQ(ret, 0);

    const char exptected_string[] = "test";
    std::filesystem::path expected_path(exptected_string);
    auto absolute_expected_path = std::filesystem::absolute(expected_path).make_preferred();
    ASSERT_EQ(std::filesystem::path(buffer), absolute_expected_path);

    // fail: nullptr
    ret = CPWD__GetStorePath(nullptr, 256);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__GetStorePath(buffer, strlen(exptected_string));
    ASSERT_EQ(ret, -1);
}

TEST_F(TestStorage, FileEnding) {
    // set
    CPWD__SetFileEnding("test");

    // success
    char buffer[256];
    int ret = CPWD__GetFileEnding(buffer, 256);
    ASSERT_EQ(ret, 0);
    ASSERT_STREQ(buffer, "test");

    // fail: nullptr
    ret = CPWD__GetFileEnding(nullptr, 256);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__GetFileEnding(buffer, 4);
    ASSERT_EQ(ret, -1);

    // success: size
    ret = CPWD__GetFileEnding(buffer, 5);
    ASSERT_EQ(ret, 0);
}

TEST_F(TestStorage, FilePath) {
    CPWD__SetStorePath("test");
    CPWD__SetFileEnding("end");

    // success
    char buffer[256];
    auto ret = CPWD__FilePath("myid", buffer, 256);
    ASSERT_EQ(ret, 0);

    const char exptected_string[] = "test/myid.end";
    std::filesystem::path expected_path(exptected_string);
    auto absolute_expected_path = std::filesystem::absolute(expected_path).make_preferred();
    ASSERT_EQ(std::filesystem::path(buffer), absolute_expected_path);

    // fail: nullptr
    ret = CPWD__FilePath("myid", nullptr, 256);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__FilePath("myid", buffer, strlen(exptected_string));
    ASSERT_EQ(ret, -1);
}
