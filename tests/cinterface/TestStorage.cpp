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
    char buffer[100];
    int ret = CPWD__NormalizeId("//TEST", buffer, 100);
    ASSERT_EQ(ret, 0);
    ASSERT_STREQ(buffer, "test");

    // fail: nullptr
    ret = CPWD__NormalizeId("//TEST", nullptr, 100);
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
    char buffer[100];
    int ret = CPWD__GetStorePath(buffer, 100);
    ASSERT_EQ(ret, 0);
    ASSERT_STREQ(buffer, "test");

    // fail: nullptr
    ret = CPWD__GetStorePath(nullptr, 100);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__GetStorePath(buffer, 4);
    ASSERT_EQ(ret, -1);

    // success: size
    ret = CPWD__GetStorePath(buffer, 5);
    ASSERT_EQ(ret, 0);
}

TEST_F(TestStorage, FileEnding) {
    // set
    CPWD__SetFileEnding("test");

    // success
    char buffer[100];
    int ret = CPWD__GetFileEnding(buffer, 100);
    ASSERT_EQ(ret, 0);
    ASSERT_STREQ(buffer, "test");

    // fail: nullptr
    ret = CPWD__GetFileEnding(nullptr, 100);
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
    char buffer[100];
    auto ret = CPWD__FilePath("myid", buffer, 100);
    ASSERT_EQ(ret, 0);

    std::filesystem::path expected_path("test/myid.end");
    auto expected_string = expected_path.make_preferred().string().c_str();
    ASSERT_STREQ(buffer, expected_string);

    // fail: nullptr
    ret = CPWD__FilePath("myid", nullptr, 100);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__FilePath("myid", buffer, 13);
    ASSERT_EQ(ret, -1);

    // success: size
    ret = CPWD__FilePath("myid", buffer, 14);
    ASSERT_EQ(ret, 0);
}
