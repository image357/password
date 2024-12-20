#include "TestPassword.h"
#include <password/cinterface.h>

#include <filesystem>
#include <cstring>

void TestPassword::SetUp() {
    Test::SetUp();
    auto ret_loglevel = CPWD__LogLevel(CPWD__LevelDebug);
    EXPECT_EQ(ret_loglevel, 0);

    CPWD__SetStorePath(STORAGE_PATH "/cinterface_password");
    CPWD__SetFileEnding("end");

    // set HashPassword to false
    while (CPWD__ToggleHashPassword()) {
        // nothing
    }

    // enable recovery
    CPWD__EnableRecovery("recovery_key");

    // register manger
    CPWD__RegisterDefaultManager("current");
    auto ret_set_manager = CPWD__SetDefaultManager("current");
    ASSERT_EQ(ret_set_manager, 0);
}

void TestPassword::TearDown() {
    auto ret_clean = CPWD__Clean();
    EXPECT_EQ(ret_clean, 0);

    auto ret_remove = std::filesystem::remove_all(STORAGE_PATH "/cinterface_password");
    EXPECT_GE(ret_remove, 1);

    // disable recovery
    CPWD__DisableRecovery();

    // remove current manager
    CPWD__RegisterDefaultManager("old");

    Test::TearDown();
}

TEST_F(TestPassword, Overwrite) {
    // create
    auto ret_overwrite = CPWD__Overwrite("foo", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // test
    char buffer[256];
    ASSERT_EQ(CPWD__Get("foo", "123", buffer, 256), 0);
}

TEST_F(TestPassword, Get) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("get1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success
    char buffer[256];
    auto ret_get = CPWD__Get("get1", "123", buffer, 256);
    ASSERT_EQ(ret_get, 0);
    ASSERT_STREQ(buffer, "bar");

    // fail
    ret_get = CPWD__Get("get_invalid", "123", buffer, 256);
    ASSERT_EQ(ret_get, -1);
}

TEST_F(TestPassword, GetBufferSize) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("get2", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // fail
    char buffer[256];
    auto ret_get = CPWD__Get("get2", "123", buffer, 3);
    ASSERT_EQ(ret_get, -1);

    // success
    ret_get = CPWD__Get("get2", "123", buffer, 4);
    ASSERT_EQ(ret_get, 0);
    ASSERT_STREQ(buffer, "bar");
}

TEST_F(TestPassword, GetBufferNull) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("get3", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // fail
    auto ret_get = CPWD__Get("get3", "123", nullptr, 256);
    ASSERT_EQ(ret_get, -1);
}

TEST_F(TestPassword, Check) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("check1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success: true
    bool result = false;
    auto ret_check = CPWD__Check("check1", "bar", "123", &result);
    ASSERT_EQ(ret_check, 0);
    ASSERT_EQ(result, true);

    // success: false
    result = true;
    ret_check = CPWD__Check("check1", "foo", "123", &result);
    ASSERT_EQ(ret_check, 0);
    ASSERT_EQ(result, false);

    // fail
    result = true;
    ret_check = CPWD__Check("check_invalid", "bar", "123", &result);
    ASSERT_EQ(ret_check, -1);
    ASSERT_EQ(result, true);
}

TEST_F(TestPassword, CheckResultNull) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("check2", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // fail
    auto ret_check = CPWD__Check("check2", "bar", "123", nullptr);
    ASSERT_EQ(ret_check, -1);
}

TEST_F(TestPassword, Set) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("set1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success: change
    auto ret_set = CPWD__Set("set1", "bar", "foo", "123");
    ASSERT_EQ(ret_set, 0);
    char buffer[256];
    auto ret_get = CPWD__Get("set1", "123", buffer, 256);
    ASSERT_EQ(ret_get, 0);
    ASSERT_STREQ(buffer, "foo");

    // fail: change
    ret_set = CPWD__Set("set1", "bar", "foo", "123");
    ASSERT_EQ(ret_set, -1);
    ret_get = CPWD__Get("set1", "123", buffer, 256);
    ASSERT_EQ(ret_get, 0);
    ASSERT_STREQ(buffer, "foo");

    // success: create
    ret_set = CPWD__Set("set2", "irrelevant", "foobar", "123");
    ASSERT_EQ(ret_set, 0);
    ret_get = CPWD__Get("set2", "123", buffer, 256);
    ASSERT_EQ(ret_get, 0);
    ASSERT_STREQ(buffer, "foobar");
}

TEST_F(TestPassword, Unset) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("unset1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);
    ret_overwrite = CPWD__Overwrite("unset2", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success: delete
    auto ret_unset = CPWD__Unset("unset1", "bar", "123");
    ASSERT_EQ(ret_unset, 0);
    char buffer[256];
    ASSERT_EQ(CPWD__Get("unset1", "123", buffer, 256), -1);

    // fail: invalid
    ret_unset = CPWD__Unset("unset1", "bar", "123");
    ASSERT_EQ(ret_unset, -1);

    // fail: delete
    ret_unset = CPWD__Unset("unset2", "foo", "123");
    ASSERT_EQ(ret_unset, -1);
}

TEST_F(TestPassword, Exists) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("exists1", "foobar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success: true
    bool result = false;
    auto ret_exists = CPWD__Exists("exists1", &result);
    ASSERT_EQ(ret_exists, 0);
    ASSERT_EQ(result, true);

    // success: false
    result = true;
    ret_exists = CPWD__Exists("not_exists", &result);
    ASSERT_EQ(ret_exists, 0);
    ASSERT_EQ(result, false);
}

TEST_F(TestPassword, ExistsResultNull) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("exists2", "foobar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // fail
    auto ret_check = CPWD__Exists("exists2", nullptr);
    ASSERT_EQ(ret_check, -1);
}

TEST_F(TestPassword, List) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("list1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);
    ret_overwrite = CPWD__Overwrite("list2", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success
    char buffer[1024];
    auto ret_list = CPWD__List(buffer, 1024, ";;;");
    ASSERT_EQ(ret_list, 0);
    ASSERT_NE(std::strstr(buffer, ";;;"), nullptr);
    ASSERT_NE(std::strstr(buffer, "list1"), nullptr);
    ASSERT_NE(std::strstr(buffer, "list2"), nullptr);

    // fail: delim
    ret_list = CPWD__List(buffer, 1024, "list");
    ASSERT_EQ(ret_list, -1);
}

TEST_F(TestPassword, ListBufferSize) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("list3", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    const char expected_string[] = "list3;;;list3.recovery";
    // fail
    char buffer[256];
    auto ret_list = CPWD__List(buffer, strlen(expected_string), ";;;");
    ASSERT_EQ(ret_list, -1);

    // success
    ret_list = CPWD__List(buffer, strlen(expected_string) + 1, ";;;");
    ASSERT_EQ(ret_list, 0);
}

TEST_F(TestPassword, ListBufferNull) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("list4", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // fail
    auto ret_list = CPWD__List(nullptr, 1024, ";;;");
    ASSERT_EQ(ret_list, -1);
}

TEST_F(TestPassword, Delete) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("delete1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success
    auto ret_delete = CPWD__Delete("delete1");
    ASSERT_EQ(ret_delete, 0);
    char buffer[256];
    ASSERT_EQ(CPWD__Get("delete1", "123", buffer, 256), -1);

    // fail
    ret_delete = CPWD__Delete("delete1");
    ASSERT_EQ(ret_delete, -1);
}

TEST_F(TestPassword, Clean) {
    // prepare
    auto ret_overwrite = CPWD__Overwrite("clean1", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);
    ret_overwrite = CPWD__Overwrite("clean2", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // success
    auto ret_clean = CPWD__Clean();
    char buffer[256];
    ASSERT_EQ(CPWD__Get("clean1", "123", buffer, 256), -1);
    ASSERT_EQ(CPWD__Get("clean2", "123", buffer, 256), -1);
}

TEST_F(TestPassword, RewriteKey) {
    // create
    auto ret_overwrite = CPWD__Overwrite("foo", "bar", "123");
    ASSERT_EQ(ret_overwrite, 0);

    // confirm
    ASSERT_EQ(CPWD__Check("foo", "bar", "123"), 0);

    // rewrite key
    auto ret_rewrite = CPWD__RewriteKey("foo", "123", "456")
    ASSERT_EQ(ret_rewrite, 0);

    // test
    ASSERT_EQ(CPWD__Check("foo", "bar", "456"), 0);
}
