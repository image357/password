#include "TestRest.h"
#include <password/cinterface.h>

#include <filesystem>
#include <chrono>
#include <thread>

void TestRest::SetUp() {
    Test::SetUp();
    auto ret_loglevel = CPWD__LogLevel(CPWD__LevelDebug);
    EXPECT_EQ(ret_loglevel, 0);

    CPWD__SetStorePath(STORAGE_PATH "/cinterface_rest");
    CPWD__SetFileEnding("end");
}

void TestRest::TearDown() {
    auto ret_remove = std::filesystem::remove_all(STORAGE_PATH "/cinterface_rest");
    EXPECT_GE(ret_remove, 0);
    Test::TearDown();
}


bool test_callback(const char *token, const char *ip, const char *resource, const char *id) {
    return true;
}

TEST_F(TestRest, StartSimpleService) {
    // fail: nullptr
    auto ret = CPWD__StartSimpleService(":8080", "/prefix", "storage_key", nullptr);
    ASSERT_EQ(ret, -1);

    // start: success
    ret = CPWD__StartSimpleService(":8080", "/prefix", "storage_key", test_callback);
    ASSERT_EQ(ret, 0);

    // start: fail
    ret = CPWD__StartSimpleService(":8080", "/prefix", "storage_key", test_callback);
    ASSERT_EQ(ret, -1);

    // wait
    std::this_thread::sleep_for(std::chrono::milliseconds(1000));

    // stop: success
    ret = CPWD__StopService(1000, ":8080", "/prefix");
    ASSERT_EQ(ret, 0);

    // stop: fail
    ret = CPWD__StopService(1000, ":8080", "/prefix");
    ASSERT_EQ(ret, -1);
}

TEST_F(TestRest, StartMultiService) {
    // fail: nullptr
    auto ret = CPWD__StartMultiService(":8080", "/prefix", "storage_key", nullptr);
    ASSERT_EQ(ret, -1);

    // start: success
    ret = CPWD__StartMultiService(":8080", "/prefix", "storage_key", test_callback);
    ASSERT_EQ(ret, 0);

    // start: fail
    ret = CPWD__StartMultiService(":8080", "/prefix", "storage_key", test_callback);
    ASSERT_EQ(ret, -1);

    // wait
    std::this_thread::sleep_for(std::chrono::milliseconds(1000));

    // stop: success
    ret = CPWD__StopService(1000, ":8080", "/prefix");
    ASSERT_EQ(ret, 0);

    // stop: fail
    ret = CPWD__StopService(1000, ":8080", "/prefix");
    ASSERT_EQ(ret, -1);
}
