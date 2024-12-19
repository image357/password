#include "TestStorage.h"
#include <password/cinterface.h>

#include <filesystem>

void TestStorage::SetUp() {
    Test::SetUp();
    auto ret_loglevel = CPWD__LogLevel(CPWD__LevelDebug);
    EXPECT_EQ(ret_loglevel, 0);
}

void TestStorage::TearDown() {
    // remove current manager
    CPWD__RegisterDefaultManager("old");

    Test::TearDown();
}

TEST_F(TestStorage, NormalizeId) {
    // success
    char buffer[256];
    auto ret = CPWD__NormalizeId("//TEST", buffer, 256);
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
    auto ret = CPWD__SetStorePath("test");
    ASSERT_EQ(ret, 0);

    // success
    char buffer[256];
    ret = CPWD__GetStorePath(buffer, 256);
    ASSERT_EQ(ret, 0);

    const char expected_string[] = "test";
    std::filesystem::path expected_path(expected_string);
    auto absolute_expected_path = std::filesystem::absolute(expected_path).make_preferred();
    ASSERT_EQ(std::filesystem::path(buffer), absolute_expected_path);

    // fail: nullptr
    ret = CPWD__GetStorePath(nullptr, 256);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__GetStorePath(buffer, strlen(expected_string));
    ASSERT_EQ(ret, -1);
}

TEST_F(TestStorage, FileEnding) {
    // set
    auto ret = CPWD__SetFileEnding("test");
    ASSERT_EQ(ret, 0);

    // success
    char buffer[256];
    ret = CPWD__GetFileEnding(buffer, 256);
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

    const char expected_string[] = "test/myid.end";
    std::filesystem::path expected_path(expected_string);
    auto absolute_expected_path = std::filesystem::absolute(expected_path).make_preferred();
    ASSERT_EQ(std::filesystem::path(buffer), absolute_expected_path);

    // fail: nullptr
    ret = CPWD__FilePath("myid", nullptr, 256);
    ASSERT_EQ(ret, -1);

    // fail: size
    ret = CPWD__FilePath("myid", buffer, strlen(expected_string));
    ASSERT_EQ(ret, -1);
}

TEST_F(TestStorage, SetTemporaryStorage) {
    // set temporary storage
    CPWD__SetTemporaryStorage();

    // create
    auto ret_overwrite = CPWD__Overwrite("foo", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // test
    char buffer[256];
    ASSERT_EQ(CPWD__Get("foo", "123", buffer, 256), 0);
}

TEST_F(TestStorage, WriteToDiskReadFromDisk) {
    CPWD__SetTemporaryStorage();

    // create
    auto ret_overwrite = CPWD__Overwrite("foo", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // confirm
    char buffer[256];
    ASSERT_EQ(CPWD__Get("foo", "123", buffer, 256), 0);

    // test
    char path[] = STORAGE_PATH "/cinterface_storage_ReadWriteDisk";
    auto ret_write = CPWD__WriteToDisk(path);
    EXPECT_EQ(ret_write, 0);

    auto ret_clean = CPWD__Clean();
    EXPECT_EQ(ret_clean, 0);

    auto ret_read = CPWD__ReadFromDisk(path);
    EXPECT_EQ(ret_read, 0);

    // confirm again
    EXPECT_EQ(CPWD__Get("foo", "123", buffer, 256), 0);

    // cleanup
    auto ret_remove = std::filesystem::remove_all(path);
    EXPECT_GE(ret_remove, 0);
}
